<script lang="ts">
    import { onMount, onDestroy } from 'svelte';

    import AuthModal from './lib/components/AuthModal.svelte';
    import ActivationScreen from './lib/components/ActivationScreen.svelte';
    import Canvas from './lib/components/Canvas.svelte';
    import CookieConsent from './lib/components/CookieConsent.svelte';
    import ActivationError from './lib/components/ActivationError.svelte';
    import DiagramListModal from './lib/components/DiagramListModal.svelte';
    import Alert from './lib/components/Alert.svelte';
    import ConfirmDialog from './lib/components/ConfirmDialog.svelte';
    import PromptDialog from './lib/components/PromptDialog.svelte';
    import EmbedPage from './lib/components/EmbedPage.svelte';

    import { webSocketManager } from './lib/services/WebSocketManager';
    import { onlineUserStoreManager } from './lib/services/OnlineUserStoreManager';

    import logger from './lib/services/logger';

    import { api } from './lib/Api';
    import type { DiagramInfo } from './lib/ApiSpec';
    import type { RespDataMsg, RespStatus } from './lib/types';
    import { diagramTitle, diagramPermission, currentUser, loadFeaturesOnce } from './lib/stores';
    import { resolvedThemeStore, type ResolvedTheme } from './lib/stores/theme';

    let unsubscribeTheme: (() => void) | undefined;

    let authStatus: 'loading' | 'authenticated' | 'unauthenticated' | 'inactive' | 'activationerror' =
        $state('loading');

    let showAuthModal = $state(false);
    let activationError = $state<string | null>(null);
    let currentDiagramId = $state<string | null>(null);
    let diagramInfo = $state<DiagramInfo | null>(null);
    let embedToken: string | null = $state(null);
    // Тема embed-страницы (локальная, не из localStorage). Читается из ?theme=... в hash
    // и отдаётся в EmbedPage; null = не embed. Не управляется тем же applyTheme, что основной
    // режим, — у embed собственная локальная тема.
    let embedTheme: 'system' | 'light' | 'dark' | null = $state(null);

    // Разбирает hash на embed-маршрут /token/TOKEN?theme=... (query-параметры идут внутри hash,
    // window.location.search при этом пуст). Возвращает { token, theme } или null, если это не embed.
    function parseEmbedRoute(hash: string): { token: string; theme: 'light' | 'dark' | 'system' } | null {
        // Явная проверка намерения: путь начинается с /token/. Количество сегментов ниже — вторичная
        // защита от чужого hash, у которого случайно ровно 3 части (например, /foo/bar/baz).
        if (!hash.startsWith('/token/')) return null;

        const parts = hash.split('/');
        if (parts.length !== 3) return null;

        // Query-параметры (например, ?theme=dark) идут после токена в hash — отрезаем их.
        // Тему парсим здесь, чтобы App не применял тему из localStorage поверх локальной темы embed.
        const embedPart = parts[2];
        const queryStart = embedPart.indexOf('?');
        const token = queryStart === -1 ? embedPart : embedPart.substring(0, queryStart);
        const themeParam = queryStart === -1 ? null : new URLSearchParams(embedPart.substring(queryStart)).get('theme');
        const theme: 'light' | 'dark' | 'system' =
            themeParam === 'light' || themeParam === 'dark' ? themeParam : 'system';
        return { token, theme };
    }

    onMount(async () => {
        void loadFeaturesOnce();
        const hash = window.location.hash.substring(1);

        // Check for embed link first (no auth required)
        const embedRoute = parseEmbedRoute(hash);
        if (embedRoute) {
            embedToken = embedRoute.token;
            embedTheme = embedRoute.theme;
            authStatus = 'authenticated'; // Skip auth for embed
            return;
        }

        // Не-embed режим: тема управляется основным приложением. Подписываемся только здесь —
        // в embed-режиме тему локально задаёт EmbedPage, и applyTheme не должен применяться.
        // (Также это убирает TDZ: applyTheme читает embedTheme, который инициализируется выше.)
        unsubscribeTheme = resolvedThemeStore.subscribe(applyTheme);

        // Check for activation link
        if (hash.startsWith('/activate/')) {
            const parts = hash.split('/');

            if (parts.length === 4 && parts[1] === 'activate') {
                const userId = parts[2];
                const code = parts[3];

                try {
                    await api.userService.userActivate({ userId, code });
                    // Success, reload to clear hash
                    window.location.href = window.location.origin;
                    return;
                } catch (err) {
                    authStatus = 'activationerror';
                    activationError = (err as RespDataMsg)?.response?.data?.message || 'Ошибка активации';
                    return;
                }
            }
        }

        try {
            const user = await api.userService.userMe();
            currentUser.set(user);
            authStatus = 'authenticated';
            await updateCurrentDiagramId();
        } catch (err) {
            currentUser.set(null);
            const axiosErr = err as RespStatus;
            if (axiosErr.response?.status === 401) {
                // D7: при 401 пробуем автоматически создать обезличенный аккаунт (POST /users/fake).
                // Успех → перезагружаемся (дальше обычный поток userMe), ошибка → как раньше форма входа.
                try {
                    await api.userService.userFake();
                    // Успех → перезагрузка: дальше обычный поток userMe, сессия уже установлена (D7).
                    window.location.reload();
                    return;
                } catch (fakeErr) {
                    // Флаг выключен (404) или иная ошибка — показываем форму входа/регистрации.
                    logger.error('Ошибка создания автоматического аккаунта:', fakeErr);
                    authStatus = 'unauthenticated';
                    showAuthModal = true;
                }
            } else if (axiosErr.response?.status === 403) {
                authStatus = 'inactive';
            } else {
                logger.error('Ошибка проверки статуса:', err);
                authStatus = 'unauthenticated';
                showAuthModal = true;
            }
        }

        // Обработчик изменения hash
        window.addEventListener('hashchange', updateCurrentDiagramId);
    });

    onDestroy(() => unsubscribeTheme?.());

    // Применяем разрешённую тему к <html> и CSS-переменной при каждом изменении.
    // В embed-режиме тему локально управляет EmbedPage (embedTheme), поэтому не перезаписываем её.
    function applyTheme(resolved: ResolvedTheme) {
        if (embedTheme !== null) return;
        const html = document.documentElement;
        const isDark = resolved === 'dark';
        html.setAttribute('data-bs-theme', resolved);
        html.classList.toggle('theme-dark', isDark);
        html.style.setProperty('--color-scheme', resolved);
    }

    function handleAuthSuccess(isRegister: boolean = false) {
        if (isRegister) {
            authStatus = 'inactive';
        } else {
            window.location.reload();
        }
    }

    async function fetchDiagramInfo(diagramId: string) {
        try {
            logger.log('Получаем диаграмму');
            const respD = await api.diagramService.diagramInfo(diagramId);
            currentDiagramId = diagramId;
            diagramInfo = respD.diagramInfo;
            diagramTitle.set(respD.diagramInfo.title);
            diagramPermission.set(respD.diagramInfo.permission);

            logger.log('Подключаемся к websocket');
            await webSocketManager.connect(diagramId);

            logger.log('Получаем пользователей онлайн');
            try {
                const respU = await api.diagramService.diagramOnlineUsers(diagramId);
                onlineUserStoreManager.load(respU.onlineUsers);
            } catch (err) {
                logger.error('Ошибка получения пользователей онлайн:', err);
                onlineUserStoreManager.clear();
            }
        } catch (err) {
            logger.error('Ошибка получения диаграммы:', err);
            currentDiagramId = null;
            diagramInfo = null;
            diagramTitle.set(null);
            diagramPermission.set(null);
        }
    }

    async function updateCurrentDiagramId() {
        logger.log('Обновляем currentDiagramId');
        const hash = window.location.hash.substring(1);
        if (hash.startsWith('/diagrams/') && hash.length > 1) {
            const diagramId = hash.substring('/diagrams/'.length);
            logger.log('Активная диаграмма', diagramId);
            fetchDiagramInfo(diagramId);
        } else {
            logger.log('Нет активной диаграммы');
            currentDiagramId = null;
            diagramInfo = null;
            diagramTitle.set(null);
            diagramPermission.set(null);
            await webSocketManager.disconnect();
        }
    }
</script>

<!-- Не показываем согласие до завершения авторизации, чтобы модалка не мигала, и для fake-аккаунтов. -->
<CookieConsent skipConsent={authStatus === 'loading' || $currentUser?.isFake === true}>
    {#if authStatus === 'loading'}
        <div class="d-flex justify-content-center align-items-center vh-100">
            <div class="spinner-border" role="status">
                <span class="visually-hidden">Загрузка...</span>
            </div>
        </div>
    {:else if embedToken}
        <EmbedPage token={embedToken} embedTheme={embedTheme ?? 'system'} />
    {:else if authStatus === 'authenticated'}
        {#if currentDiagramId && diagramInfo}
            <Canvas {currentDiagramId} {diagramInfo} />
        {:else}
            <DiagramListModal />
        {/if}
    {:else if authStatus === 'inactive'}
        <ActivationScreen />
    {:else if authStatus === 'activationerror'}
        <ActivationError error={activationError} />
    {:else}
        <AuthModal bind:show={showAuthModal} onSuccess={handleAuthSuccess} />
    {/if}

    <Alert />

    <ConfirmDialog />
    <PromptDialog />
</CookieConsent>
