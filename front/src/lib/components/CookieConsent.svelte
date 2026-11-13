<script lang="ts">
    import { onMount, type Snippet } from 'svelte';
    import Privacy from './Privacy.svelte';

    // Для fake-аккаунтов согласие не требуется: персональные данные не обрабатываются.
    let { children, skipConsent = false }: { children: Snippet; skipConsent?: boolean } = $props();

    const COOKIE_KEY = 'cookieConsent';

    let consentStatus = $state<string | null>(null);
    let showPrivacy = $state(false);

    onMount(() => {
        consentStatus = localStorage.getItem(COOKIE_KEY);
    });

    function accept() {
        localStorage.setItem(COOKIE_KEY, 'accepted');
        window.location.reload();
    }

    function reject() {
        localStorage.setItem(COOKIE_KEY, 'rejected');
        consentStatus = 'rejected';
    }

    function openPrivacy(e: Event) {
        e.preventDefault();
        showPrivacy = true;
    }
</script>

{#if skipConsent || consentStatus === 'accepted'}
    {@render children()}
{:else if consentStatus === 'rejected'}
    <div class="cookie-overlay">
        <div class="cookie-modal bg-dark text-white p-4 rounded">
            <h4 class="mb-3">Cookie-файлы отключены</h4>
            <p class="mb-3">
                Вы отклонили использование cookie-файлов. Доступ к приложению невозможен без согласия на использование
                cookie.
            </p>
            <p class="mb-4">Если вы хотите дать согласие, нажмите кнопку «Согласиться».</p>
            <button onclick={accept} class="btn btn-primary">Согласиться</button>
        </div>
    </div>
{:else}
    <div class="cookie-overlay">
        <div class="cookie-modal bg-dark text-white p-4 rounded">
            <h4 class="mb-3">Использование cookie-файлов</h4>
            <p class="mb-4">
                Мы используем cookie-файлы для улучшения работы сайта. Продолжая использовать сайт, вы соглашаетесь с
                <a href="/privacy" class="text-light" onclick={openPrivacy}>политикой конфиденциальности</a>.
            </p>
            <button onclick={accept} class="btn btn-primary me-2">Согласиться</button>
            <button onclick={reject} class="btn btn-secondary">Отказаться</button>
        </div>
    </div>
{/if}

<Privacy bind:show={showPrivacy} />

<style>
    .cookie-overlay {
        position: fixed;
        inset: 0;
        display: flex;
        align-items: center;
        justify-content: center;
        background: rgba(0, 0, 0, 0.6);
    }

    .cookie-modal {
        max-width: 480px;
        width: 90%;
        box-shadow: 0 4px 24px rgba(0, 0, 0, 0.3);
    }

    .cookie-modal a {
        text-decoration: underline;
    }
</style>
