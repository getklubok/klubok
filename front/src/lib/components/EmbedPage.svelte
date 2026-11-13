<script lang="ts">
    import Fa from 'svelte-fa';
    import {
        faCrosshairs,
        faExpand,
        faQuestionCircle,
        faDownload,
        faFileExport,
        faMoon,
        faSun,
    } from '@fortawesome/free-solid-svg-icons';

    import { Canvas as FabricCanvas, FabricObject, ActiveSelection as FabricActiveSelection } from 'fabric/es';
    import { onMount } from 'svelte';

    import { canvasManager } from '../services/CanvasManager';
    import { coordinatesManager } from '../services/CoordinatesManager';
    import { viewportManager } from '../services/ViewportManager';
    import { commentBadgeManager } from '../services/CommentBadgeManager';
    import { exporterManager } from '../services/ExporterManager';

    import { api } from '../Api';
    import type { DiagramInfo, EmbedDataResponse } from '../ApiSpec';
    import type { RespDataMsg } from '../types';
    import { coordinates, diagramTitle } from '../stores';

    interface Props {
        token: string;
        // Начальная тема embed. App.svelte разбирает ?theme=... из hash (system — если параметр отсутствует).
        // Переключатель ниже меняет её локально; в localStorage не пишется и на тему основного приложения не влияет.
        embedTheme?: 'system' | 'light' | 'dark';
    }

    const { token, embedTheme: initialEmbedTheme = 'system' }: Props = $props();

    let canvasElement: HTMLCanvasElement;
    let canvas: FabricCanvas;

    let loading = $state(true);
    let error: string | null = $state(null);
    let diagramInfo: DiagramInfo | null = $state(null);
    let currentZoom = $state(100);
    let showHelp = $state(false);
    // Локальная тема embed, инициализируется пропом. Изменяется только кнопкой-переключателем.
    // Проп читается один раз при создании компонента — тема берётся из URL и не меняется реактивно.
    // svelte-ignore state_referenced_locally
    let embedTheme = $state<'system' | 'light' | 'dark'>(initialEmbedTheme);
    // Флаг готовности canvas — реактивный, чтобы $effect ниже перезапускался
    // после асинхронного создания canvas в loadEmbedData().
    let canvasReady = $state(false);
    // Разрешённая тема: учитывает prefers-color-scheme при 'system'. Используется
    // и эффектом, и кнопкой, чтобы показывать фактическое состояние, а не намерение.
    const isDark = $derived(
        embedTheme === 'dark' || (embedTheme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches),
    );

    coordinates.subscribe((v) => {
        currentZoom = Math.round(v.zoom * 100);
    });

    // Применяем локальную тему embed: CSS-переменные, data-bs-theme и перекраска canvas.
    // Канвас создаётся асинхронно в onMount, поэтому эффект обходит $state canvas
    // (не реактивный) и сам перечитывает DOM при каждом срабатывании.
    $effect(() => {
        const html = document.documentElement;
        html.setAttribute('data-bs-theme', isDark ? 'dark' : '');
        html.classList.toggle('theme-dark', isDark);
        html.style.setProperty('--color-scheme', isDark ? 'dark' : 'light');

        // Canvas создаётся асинхронно; canvasReady перезапускает эффект после инициализации.
        // Фон берём из CSS-переменной --diagram-bg (обновлена выше), как в Canvas.svelte.
        if (canvasReady && canvas) {
            canvas.backgroundColor = getDiagramBackground();
            canvasManager.recolorAll(isDark);
            canvas.requestRenderAll();
        }
    });

    function toggleEmbedTheme() {
        // Переключаем только локально, не сохраняя в localStorage.
        // Опираемся на разрешённую тему: system с тёмной prefers-color-scheme уже тёмная,
        // поэтому переключение идёт в светлую, а не в «тёмную» (без изменения).
        embedTheme = isDark ? 'light' : 'dark';
    }

    // Фон канваса берём из CSS-переменной темы (--diagram-bg), как в Canvas.svelte
    function getDiagramBackground(): string {
        const bg = getComputedStyle(document.documentElement).getPropertyValue('--diagram-bg').trim();
        return bg || '#ffffff';
    }

    onMount(() => {
        async function loadEmbedData() {
            try {
                const response: EmbedDataResponse = await api.diagramService.embedData(token, {
                    withCredentials: false,
                });
                diagramInfo = response.diagramInfo;
                diagramTitle.set(response.diagramInfo.title);
            } catch (err) {
                error = (err as RespDataMsg)?.response?.data?.message || 'Ошибка загрузки диаграммы';
            } finally {
                loading = false;

                FabricObject.prototype.lockRotation = true;
                FabricObject.prototype.originX = 'center';
                FabricObject.prototype.originY = 'center';

                FabricActiveSelection.prototype.set({
                    hasControls: false,
                    hasBorders: false,
                });

                canvas = new FabricCanvas(canvasElement, {
                    width: window.innerWidth,
                    height: window.innerHeight,
                    selection: false,
                    stateful: true,
                });

                canvas.on('mouse:wheel', canvasManager.onMouseWheel);
                canvas.on('mouse:down', canvasManager.onMouseDown);
                canvas.on('mouse:move', canvasManager.onMouseMove);
                canvas.on('mouse:up', canvasManager.onMouseUp);

                canvasManager.init(canvas);

                coordinatesManager.setCanvas(canvas);
                viewportManager.setCanvas(canvas);
                commentBadgeManager.setCanvas(canvas);
                exporterManager.setCanvas(canvas);

                coordinatesManager.update();

                canvasManager.setDiagramInfo(diagramInfo as DiagramInfo);

                // Канвас готов — $effect темы перекрасит его под embedTheme
                canvasReady = true;

                viewportManager.fitView({
                    padding: {
                        left: 50,
                        right: 50,
                        top: 50,
                        bottom: 50,
                    },
                    maxZoom: 5,
                    minZoom: 0.1,
                });
            }
        }

        void loadEmbedData();

        const handleResize = () => {
            canvas.setDimensions({ width: window.innerWidth, height: window.innerHeight });
            canvas.requestRenderAll();
        };
        window.addEventListener('resize', handleResize);

        return () => {
            window.removeEventListener('resize', handleResize);
            canvas.dispose();
            canvasManager.reset();
            // Сброс флага: при повторном монтировании (переход обратно на embed) $effect
            // должен ждать новый canvas, а не красить ссылку на старый.
            canvasReady = false;
        };
    });
</script>

<div class="embed-root">
    <div id="embed-canvas-container">
        <canvas bind:this={canvasElement}></canvas>

        <div class="embed-topbar">
            <span class="diagram-title">{diagramInfo?.title}</span>
            <a href="/" target="_top" class="brand-link">Klubok</a>
        </div>

        <div class="zoom-panel">
            <span class="zoom-label"><strong>{currentZoom}%</strong></span>
            <div class="btn-group">
                <button
                    class="btn btn-sm btn-outline-primary"
                    type="button"
                    onclick={() => viewportManager.centerView()}
                    title="Центрировать"
                >
                    <Fa icon={faCrosshairs} style="vertical-align: middle;" />
                </button>
                <button
                    class="btn btn-sm btn-outline-primary"
                    type="button"
                    onclick={() =>
                        viewportManager.fitView({
                            padding: {
                                left: 50,
                                right: 50,
                                top: 50,
                                bottom: 50,
                            },
                            maxZoom: 5,
                            minZoom: 0.1,
                        })}
                    title="Показать всё"
                >
                    <Fa icon={faExpand} style="vertical-align: middle;" />
                </button>
            </div>

            <div class="btn-group">
                <button
                    class="btn btn-sm btn-outline-primary"
                    type="button"
                    onclick={() => exporterManager.toPNG()}
                    title="Экспорт в PNG"
                >
                    <Fa icon={faDownload} style="vertical-align: middle;" />
                </button>
                <button
                    class="btn btn-sm btn-outline-primary"
                    type="button"
                    onclick={() => exporterManager.toFile()}
                    title="Экспорт в файл"
                >
                    <Fa icon={faFileExport} style="vertical-align: middle;" />
                </button>
            </div>

            <div class="btn-group">
                <button
                    class="btn btn-sm btn-outline-secondary"
                    type="button"
                    onclick={toggleEmbedTheme}
                    title={isDark ? 'Сменить на светлую тему' : 'Сменить на тёмную тему'}
                >
                    <Fa icon={isDark ? faMoon : faSun} style="vertical-align: middle;" />
                </button>
            </div>

            <button
                class="btn btn-sm btn-outline-secondary btn-help"
                type="button"
                onclick={() => (showHelp = true)}
                title="Управление"
            >
                <Fa icon={faQuestionCircle} style="vertical-align: middle;" />
            </button>
        </div>
    </div>

    {#if showHelp}
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <div class="help-overlay" onclick={() => (showHelp = false)}>
            <div class="help-modal" onclick={(e) => e.stopPropagation()}>
                <h3>Управление диаграммой</h3>
                <ul>
                    <li><kbd>Колесико мыши</kbd> — масштабирование</li>
                    <li><kbd>CTRL</kbd> + <kbd>ЛКМ</kbd> — перемещение диаграммы</li>
                </ul>
                <button class="btn btn-sm btn-primary" onclick={() => (showHelp = false)}>Закрыть</button>
            </div>
        </div>
    {/if}

    {#if loading}
        <div class="loading-screen">
            <div class="spinner-border text-primary" role="status">
                <span class="visually-hidden">Загрузка...</span>
            </div>
        </div>
    {/if}

    {#if error}
        <div class="error-screen">
            <h3>Ошибка</h3>
            <p>{error}</p>
        </div>
    {/if}
</div>

<style>
    :global(body) {
        margin: 0;
        padding: 0;
        overflow: hidden;
    }

    .embed-root {
        position: relative;
        width: 100vw;
        height: 100vh;
        overflow: hidden;
        background-color: var(--diagram-bg);
    }

    .loading-screen {
        position: absolute;
        inset: 0;
        display: flex;
        justify-content: center;
        align-items: center;
        /* Фолбэк для браузеров без поддержки color-mix() */
        background-color: var(--diagram-bg-overlay);
        background-color: color-mix(in srgb, var(--diagram-bg) 95%, transparent);
        z-index: 10;
    }

    .error-screen {
        position: absolute;
        inset: 0;
        display: flex;
        flex-direction: column;
        justify-content: center;
        align-items: center;
        /* Фолбэк для браузеров без поддержки color-mix() */
        background-color: var(--diagram-bg-overlay);
        background-color: color-mix(in srgb, var(--diagram-bg) 95%, transparent);
        color: var(--text-color);
        z-index: 10;
    }

    .error-screen h3 {
        color: #dc3545;
    }

    #embed-canvas-container {
        width: 100%;
        height: 100%;
    }

    canvas {
        display: block;
        width: 100%;
        height: 100%;
    }

    .embed-topbar {
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        padding: 0.35rem 0.75rem 0.75rem 0.75rem;
        display: flex;
        align-items: center;
        justify-content: space-between;
        background: var(--surface-bg);
        border-bottom: 1px solid var(--border-color);
        box-shadow: var(--shadow);
        z-index: 1000;
        font-size: 0.8rem;
    }

    .zoom-panel {
        position: absolute;
        top: calc(2.5rem + 8px);
        left: 0.75rem;
        display: flex;
        align-items: center;
        gap: 0.5rem;
        padding: 0.35rem 0.6rem;
        background: var(--surface-bg);
        border: 1px solid var(--border-color);
        border-radius: 6px;
        box-shadow: var(--shadow);
        z-index: 1000;
        font-size: 0.8rem;
    }

    .zoom-label {
        color: var(--text-color);
        font-size: 0.85rem;
    }

    .btn-help {
        border: none !important;
        color: #999 !important;
        padding: 0 0.25rem !important;
    }

    .btn-help:hover {
        color: #333 !important;
    }

    .help-overlay {
        position: fixed;
        inset: 0;
        display: flex;
        justify-content: center;
        align-items: center;
        background: rgba(0, 0, 0, 0.4);
        z-index: 2000;
    }

    .help-modal {
        background: var(--surface-bg);
        border-radius: 10px;
        padding: 1.5rem 2rem;
        max-width: 380px;
        width: 90%;
        box-shadow: 0 4px 24px rgba(0, 0, 0, 0.2);
        text-align: center;
    }

    .help-modal h3 {
        margin: 0 0 1rem;
        font-size: 1rem;
        color: var(--text-color);
    }

    .help-modal ul {
        list-style: none;
        padding: 0;
        margin: 0 0 1.25rem;
        text-align: left;
    }

    .help-modal li {
        padding: 0.4rem 0;
        font-size: 0.9rem;
        color: var(--text-color);
    }

    .help-modal kbd {
        display: inline-block;
        padding: 0.15rem 0.4rem;
        font-size: 0.8rem;
        font-family: inherit;
        color: var(--text-color);
        background: var(--diagram-bg);
        border: 1px solid var(--border-color);
        border-radius: 4px;
        box-shadow: var(--shadow);
    }

    .diagram-title {
        font-weight: 500;
        max-width: 200px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    .brand-link {
        color: #0d6efd;
        text-decoration: none;
        font-weight: 500;
        font-size: 0.75rem;
    }

    .brand-link:hover {
        text-decoration: underline;
    }
</style>
