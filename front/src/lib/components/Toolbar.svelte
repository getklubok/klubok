<script lang="ts">
    import Fa from 'svelte-fa';
    import {
        faHandPointer,
        faSquare,
        faCircle,
        faDiamond,
        faArrowTrendUp,
        faAnglesLeft,
        faAnglesRight,
        faAnglesUp,
        faAnglesDown,
        faDownload,
        faFileExport,
        faBoxOpen,
        faT,
    } from '@fortawesome/free-solid-svg-icons';

    import { drawingMode, type DrawingMode } from '../stores';

    import { canvasManager } from '../services/CanvasManager';
    import { exporterManager } from '../services/ExporterManager';
    import hasPermission from '../services/hasPermission';

    import { type DiagramCollaboratorPermissionEnum } from '../ApiSpec';

    interface Props {
        permission: DiagramCollaboratorPermissionEnum | null | undefined;
    }
    const { permission }: Props = $props();

    // Реактивная переменная для текущего режима
    let currentMode: string = $state('select');
    drawingMode.subscribe((value) => {
        currentMode = value;
    });

    // Функция для установки режима рисования
    function setDrawingMode(mode: DrawingMode) {
        drawingMode.set(mode);
    }

    const isSelectActive = $derived(currentMode === 'select');
    const isSquareActive = $derived(currentMode === 'square');
    const isEllipseActive = $derived(currentMode === 'ellipse');
    const isDiamondActive = $derived(currentMode === 'diamond');
    const isBoxActive = $derived(currentMode === 'box');
    const isTextActive = $derived(currentMode === 'text');
    const isArrowActive = $derived(currentMode === 'arrow');
</script>

<div class="toolbar">
    <div class="btn-group-vertical">
        <button
            class="btn text-start"
            class:active={isSelectActive}
            class:btn-outline-primary={!isSelectActive}
            class:btn-primary={isSelectActive}
            type="button"
            onclick={() => setDrawingMode('select')}
            title="Выбор"
        >
            <Fa fw icon={faHandPointer} />
        </button>

        {#if hasPermission(permission)}
            <button
                class="btn text-start"
                class:active={isSquareActive}
                class:btn-outline-primary={!isSquareActive}
                class:btn-primary={isSquareActive}
                type="button"
                onclick={() => setDrawingMode('square')}
                title="Квадрат"
            >
                <Fa fw icon={faSquare} />
            </button>

            <button
                class="btn text-start"
                class:active={isEllipseActive}
                class:btn-outline-primary={!isEllipseActive}
                class:btn-primary={isEllipseActive}
                type="button"
                onclick={() => setDrawingMode('ellipse')}
                title="Круг"
            >
                <Fa fw icon={faCircle} />
            </button>

            <button
                class="btn text-start"
                class:active={isDiamondActive}
                class:btn-outline-primary={!isDiamondActive}
                class:btn-primary={isDiamondActive}
                type="button"
                onclick={() => setDrawingMode('diamond')}
                title="Ромб"
            >
                <Fa fw icon={faDiamond} />
            </button>

            <button
                class="btn text-start"
                class:active={isBoxActive}
                class:btn-outline-primary={!isBoxActive}
                class:btn-primary={isBoxActive}
                type="button"
                onclick={() => setDrawingMode('box')}
                title="Контейнер"
            >
                <Fa fw icon={faBoxOpen} />
            </button>

            <button
                class="btn text-start"
                class:active={isTextActive}
                class:btn-outline-primary={!isTextActive}
                class:btn-primary={isTextActive}
                type="button"
                onclick={() => setDrawingMode('text')}
                title="Текст"
            >
                <Fa fw icon={faT} />
            </button>

            <button
                class="btn text-start"
                class:active={isArrowActive}
                class:btn-outline-primary={!isArrowActive}
                class:btn-primary={isArrowActive}
                type="button"
                onclick={() => setDrawingMode('arrow')}
                title="Стрелка"
            >
                <Fa fw icon={faArrowTrendUp} />
            </button>
        {/if}
    </div>

    <div class="btn-group-vertical">
        <button
            class="btn btn-sm btn-outline-secondary"
            type="button"
            onclick={() => canvasManager.expandPage('left')}
            title="Расширить влево"
        >
            <Fa fw icon={faAnglesLeft} />
        </button>
        <button
            class="btn btn-sm btn-outline-secondary"
            type="button"
            onclick={() => canvasManager.expandPage('top')}
            title="Расширить вверх"
        >
            <Fa fw icon={faAnglesUp} />
        </button>
        <button
            class="btn btn-sm btn-outline-secondary"
            type="button"
            onclick={() => canvasManager.expandPage('bottom')}
            title="Расширить вниз"
        >
            <Fa fw icon={faAnglesDown} />
        </button>
        <button
            class="btn btn-sm btn-outline-secondary"
            type="button"
            onclick={() => canvasManager.expandPage('right')}
            title="Расширить вправо"
        >
            <Fa fw icon={faAnglesRight} />
        </button>
    </div>

    <div class="btn-group-vertical">
        <button
            class="btn btn-sm btn-outline-primary"
            type="button"
            onclick={() => exporterManager.toPNG()}
            title="Экспорт в PNG"
        >
            <Fa fw icon={faDownload} />
        </button>
        <button
            class="btn btn-sm btn-outline-primary"
            type="button"
            onclick={() => exporterManager.toFile()}
            title="Экспорт в файл"
        >
            <Fa fw icon={faFileExport} />
        </button>
    </div>
</div>

<style>
    .toolbar {
        display: flex;
        flex-direction: column;
        gap: 8px;
    }
</style>
