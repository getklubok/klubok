<script lang="ts">
    import { onMount } from 'svelte';
    import { get } from 'svelte/store';
    import {
        Canvas as FabricCanvas,
        FabricObject,
        ActiveSelection as FabricActiveSelection,
        Group as FabricGroup,
    } from 'fabric/es';

    import Title from './Title.svelte';
    import Toolbar from './Toolbar.svelte';
    import StatusBar from './StatusBar.svelte';
    import Tools from './Tools.svelte';
    import Comments from './Comments.svelte';
    import DebugPanel from './DebugPanel.svelte';
    import GroupPanel from './GroupPanel.svelte';
    import ContextMenu from './ContextMenu.svelte';

    import { textEditorManager } from '../services/TextEditorManager';
    import { commentBadgeManager } from '../services/CommentBadgeManager';
    import { coordinatesManager } from '../services/CoordinatesManager';
    import { viewportManager } from '../services/ViewportManager';
    import { canvasManager } from '../services/CanvasManager';
    import { exporterManager } from '../services/ExporterManager';
    import { autoSaveService } from '../services/AutoSaveService';

    import hasPermission from '../services/hasPermission';
    import logger from '../services/logger';

    import type { DiagramInfo } from '../ApiSpec';

    import {
        drawingMode,
        selectedObject,
        shapes,
        connections,
        commentsShow,
        commentsSelectedGroup,
        cursors,
        contextMenu,
    } from '../stores';

    import { resolvedThemeStore, type ResolvedTheme } from '../stores/theme';

    interface Props {
        currentDiagramId: string;
        diagramInfo: DiagramInfo | null;
    }

    const { currentDiagramId, diagramInfo }: Props = $props();

    let canvasElement: HTMLCanvasElement;
    let canvas: FabricCanvas;

    // Фон канваса берём из CSS-переменной темы (--diagram-bg), а не дублируем палитру в JS.
    function getDiagramBackground(): string {
        const bg = getComputedStyle(document.documentElement).getPropertyValue('--diagram-bg').trim();
        return bg || '#ffffff';
    }

    onMount(() => {
        logger.info('Canvas mounted');

        FabricObject.prototype.lockRotation = true;
        FabricObject.prototype.originX = 'center';
        FabricObject.prototype.originY = 'center';

        FabricActiveSelection.prototype.set({
            hasControls: false,
            hasBorders: true,
        });

        // Инициализируем Canvas
        canvas = new FabricCanvas(canvasElement, {
            width: window.innerWidth,
            height: window.innerHeight,
            selection: true,
            stateful: true,
        });

        canvas.on('mouse:down', canvasManager.onMouseDown);
        canvas.on('mouse:move', canvasManager.onMouseMove);
        canvas.on('mouse:up', canvasManager.onMouseUp);
        canvas.on('mouse:wheel', canvasManager.onMouseWheel);

        if (hasPermission(diagramInfo?.permission)) {
            canvas.on('mouse:dblclick', (opt) => textEditorManager.startEditing(opt.target as FabricGroup));

            canvas.on('text:editing:exited', (_opt) => textEditorManager.finishEditing());

            canvas.on('object:moving', canvasManager.onObjectMoving);
            canvas.on('object:scaling', canvasManager.onObjectScaling);
            canvas.on('object:modified', canvasManager.onObjectModified);

            canvas.on('contextmenu', canvasManager.onContextMenu);
        }

        canvas.on('selection:created', canvasManager.onSelectionCreated);
        canvas.on('selection:updated', canvasManager.onSelectionUpdated);
        canvas.on('selection:cleared', canvasManager.onSelectionCleared);

        canvasManager.init(canvas);

        coordinatesManager.setCanvas(canvas);
        viewportManager.setCanvas(canvas);
        textEditorManager.setCanvas(canvas);
        commentBadgeManager.setCanvas(canvas);
        exporterManager.setCanvas(canvas);

        // Применяем фоновый цвет сразу после создания канваса
        applyCanvasBackground();

        // Инициализируем координаты
        coordinatesManager.update();

        const unsubscribeTheme = resolvedThemeStore.subscribe(applyCanvasBackground);

        return () => {
            // Отменяем подписку до очистки canvas, чтобы не обращаться к устаревшему экземпляру
            unsubscribeTheme();
            // Очищаем canvas при размонтировании компонента
            canvas.dispose();
            canvasManager.reset();
        };
    });

    $effect(() => {
        if (diagramInfo && canvas) {
            canvasManager.setDiagramInfo(diagramInfo as DiagramInfo);
        }
    });

    // Фон канваса и цвета объектов подстраиваются под тему. Подписка на
    // resolvedThemeStore срабатывает и при начальной инициализации (в onMount),
    // и при каждой смене темы (включая системную через prefers-color-scheme).
    // CanvasManager — класс без Svelte-реактивности, поэтому подписка здесь, в Canvas.svelte.
    function applyCanvasBackground(resolved: ResolvedTheme = get(resolvedThemeStore)) {
        if (canvas) {
            canvas.backgroundColor = getDiagramBackground();
            canvasManager.recolorAll(resolved === 'dark');
            canvas.requestRenderAll();
        }
    }

    $effect(() => {
        autoSaveService.setDiagramId(currentDiagramId);
    });

    // Подписываемся на изменения режима рисования
    drawingMode.subscribe((mode) => {
        canvasManager.setCurrentDrawingMode(mode);
    });

    // Подписки на изменения для автосохранения
    shapes.subscribe(() => {
        autoSaveService.triggerAutoSave();
    });
    connections.subscribe(() => {
        autoSaveService.triggerAutoSave();
    });

    // Подписываемся на изменения курсоров
    cursors.subscribe((items) => {
        canvasManager.upsertCursors(items);
    });
</script>

<div id="canvas-container">
    <canvas bind:this={canvasElement}></canvas>

    <nav class="navbar">
        <div class="container-fluid">
            <div class="me-4">
                <Title />
            </div>

            <Tools />
        </div>
    </nav>

    <div class="statusbar">
        <StatusBar />
    </div>

    <div class="toolbar">
        <Toolbar permission={diagramInfo?.permission} />
    </div>

    {#if $selectedObject !== null && $selectedObject?.name !== ''}
        <div class="group-panel">
            <GroupPanel permission={diagramInfo?.permission} />
        </div>
    {/if}

    <DebugPanel />

    {#if $commentsShow && $commentsSelectedGroup}
        <Comments diagramId={currentDiagramId} />
    {/if}

    {#if $contextMenu.visible && $contextMenu.target}
        <ContextMenu
            x={$contextMenu.x}
            y={$contextMenu.y}
            target={$contextMenu.target}
            permission={diagramInfo?.permission}
            onClose={() => contextMenu.set({ visible: false, x: 0, y: 0, target: null })}
        />
    {/if}
</div>

<style>
    #canvas-container {
        background-color: var(--diagram-bg);
        position: relative;
        width: 100vw;
        height: 100vh;
        overflow: hidden;
    }

    canvas {
        display: block;
        width: 100%;
        height: 100%;
    }

    .navbar {
        position: absolute;
        top: 0;
        left: 0;
        width: 100%;
        background: var(--surface-bg);
        box-shadow: var(--shadow);
        z-index: 1000;
    }

    .statusbar {
        position: absolute;
        padding: 0.5rem 1rem;
        top: 58px;
        left: 10px;
        background: var(--surface-bg);
        border: 1px solid var(--border-color);
        border-radius: 8px;
        box-shadow: var(--shadow);
        z-index: 999;
    }

    .toolbar {
        position: absolute;
        padding: 0.5rem;
        top: 114px;
        left: 10px;
        background: var(--surface-bg);
        border: 1px solid var(--border-color);
        border-radius: 8px;
        box-shadow: var(--shadow);
        z-index: 998;
    }

    .group-panel {
        position: absolute;
        padding: 0.5rem;
        top: 114px;
        left: 82px;
        background: var(--surface-bg);
        border: 1px solid var(--border-color);
        border-radius: 8px;
        box-shadow: var(--shadow);
        z-index: 998;
    }
</style>
