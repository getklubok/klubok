<script lang="ts">
    import Fa from 'svelte-fa';
    import {
        faArrowsDownToLine,
        faArrowsUpToLine,
        faArrowDown,
        faArrowUp,
        faTrash,
        faEdit,
        faComment,
        faClone,
        faAlignLeft,
        faAlignCenter,
        faAlignRight,
    } from '@fortawesome/free-solid-svg-icons';

    import { canvasManager } from '../services/CanvasManager';

    import hasPermission from '../services/hasPermission';

    import { selectedObject, shapes } from '../stores';
    import type { DiagramCollaboratorPermissionEnum } from '../ApiSpec';
    import { groupActionManager } from '../services/GroupActionManager';

    interface Props {
        permission: DiagramCollaboratorPermissionEnum | null | undefined;
    }
    const { permission }: Props = $props();

    const textAlign = $derived($shapes.find((s) => s.id === $selectedObject?.name)?.textAlign ?? 'center');
</script>

<div class="group-panel">
    {#if $selectedObject}
        <div class="section-title">Действия</div>
        <div class="btn-group-sm">
            <button
                class="btn btn-outline-primary"
                type="button"
                onclick={() => groupActionManager.showComments($selectedObject.object)}
                title="Открыть комментарии"
            >
                <Fa fw icon={faComment} />
            </button>
            {#if hasPermission(permission)}
                <button
                    class="btn btn-outline-primary"
                    type="button"
                    onclick={() => groupActionManager.startEditing($selectedObject.object)}
                    title="Редактировать текст"
                >
                    <Fa fw icon={faEdit} />
                </button>
                <button
                    class="btn btn-outline-primary"
                    type="button"
                    onclick={() => groupActionManager.cloneObject($selectedObject.object)}
                    title="Клонировать"
                >
                    <Fa fw icon={faClone} />
                </button>
                <button
                    class="btn btn-outline-danger"
                    type="button"
                    onclick={() => groupActionManager.deleteObject($selectedObject.object)}
                    title="Удалить"
                >
                    <Fa fw icon={faTrash} />
                </button>
            {/if}
        </div>
        {#if hasPermission(permission) && ($selectedObject.type === 'box' || $selectedObject.type === 'text')}
            <div class="section-title">Текст</div>
            <div class="btn-group-sm">
                <button
                    class="btn btn-outline-primary"
                    class:active={textAlign === 'left'}
                    type="button"
                    onclick={() => canvasManager.setTextAlign($selectedObject.name, 'left')}
                    title="По левому краю"
                    aria-pressed={textAlign === 'left'}
                >
                    <Fa fw icon={faAlignLeft} />
                </button>
                <button
                    class="btn btn-outline-primary"
                    class:active={textAlign === 'center'}
                    type="button"
                    onclick={() => canvasManager.setTextAlign($selectedObject.name, 'center')}
                    title="По центру"
                    aria-pressed={textAlign === 'center'}
                >
                    <Fa fw icon={faAlignCenter} />
                </button>
                <button
                    class="btn btn-outline-primary"
                    class:active={textAlign === 'right'}
                    type="button"
                    onclick={() => canvasManager.setTextAlign($selectedObject.name, 'right')}
                    title="По правому краю"
                    aria-pressed={textAlign === 'right'}
                >
                    <Fa fw icon={faAlignRight} />
                </button>
            </div>
        {/if}
        {#if hasPermission(permission)}
            <div class="section-title">Слои</div>
            <div class="btn-group-sm">
                <button
                    class="btn btn-outline-primary"
                    type="button"
                    onclick={() => canvasManager.moveObject($selectedObject.name, 'back')}
                    title="На задний план"
                >
                    <Fa fw icon={faArrowsDownToLine} />
                </button>
                <button
                    class="btn btn-outline-primary"
                    type="button"
                    onclick={() => canvasManager.moveObject($selectedObject.name, 'down')}
                    title="Переместить назад"
                >
                    <Fa fw icon={faArrowDown} />
                </button>
                <button
                    class="btn btn-outline-primary"
                    type="button"
                    onclick={() => canvasManager.moveObject($selectedObject.name, 'up')}
                    title="Переместить вперёд"
                >
                    <Fa fw icon={faArrowUp} />
                </button>
                <button
                    class="btn btn-outline-primary"
                    type="button"
                    onclick={() => canvasManager.moveObject($selectedObject.name, 'front')}
                    title="На передний план"
                >
                    <Fa fw icon={faArrowsUpToLine} />
                </button>
            </div>
        {/if}
    {/if}
</div>

<style>
    .group-panel {
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .section-title {
        font-family: monospace;
        font-size: 14px;
    }
</style>
