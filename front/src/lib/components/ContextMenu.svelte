<script lang="ts">
    import Fa from 'svelte-fa';
    import {
        type IconDefinition,
        faArrowsDownToLine,
        faArrowsUpToLine,
        faArrowDown,
        faArrowUp,
        faTrash,
        faEdit,
        faComment,
        faClone,
        faExpand,
        faDownload,
    } from '@fortawesome/free-solid-svg-icons';

    import { Group as FabricGroup } from 'fabric/es';

    import { canvasManager } from '../services/CanvasManager';
    import { groupActionManager } from '../services/GroupActionManager';

    import hasPermission from '../services/hasPermission';

    import type { DiagramCollaboratorPermissionEnum } from '../ApiSpec';
    import { getImageUrl } from '../Api';
    import { imageViewerUrl, shapes } from '../stores';

    interface MenuItem {
        label?: string;
        icon?: IconDefinition;
        action?: () => void;
        variant?: 'default' | 'danger';
        separator?: boolean;
    }

    interface Props {
        x: number;
        y: number;
        target: FabricGroup;
        permission: DiagramCollaboratorPermissionEnum | null | undefined;
        diagramId: string;
        onClose: () => void;
    }

    // Пункты слоёв: один обработчик на направление перемещения.
    function layerItems(group: FabricGroup, close: () => void): MenuItem[] {
        const moves: { label: string; icon: IconDefinition; direction: 'back' | 'down' | 'up' | 'front' }[] = [
            { label: 'На задний план', icon: faArrowsDownToLine, direction: 'back' },
            { label: 'Переместить назад', icon: faArrowDown, direction: 'down' },
            { label: 'Переместить вперёд', icon: faArrowUp, direction: 'up' },
            { label: 'На передний план', icon: faArrowsUpToLine, direction: 'front' },
        ];

        return moves.map(({ label, icon, direction }) => ({
            label,
            icon,
            action: () => {
                if (!group.name) return;
                canvasManager.moveObject(group.name, direction);
                close();
            },
        }));
    }
    const { x, y, target, permission, diagramId, onClose }: Props = $props();

    const selectedShape = $derived($shapes.find((s) => s.id === target.name));
    const imageId = $derived(selectedShape && 'img' in selectedShape ? selectedShape.img : '');
    const menuItems = $derived.by<MenuItem[]>(() => {
        const items: MenuItem[] = [
            {
                label: 'Открыть комментарии',
                icon: faComment,
                action: () => {
                    groupActionManager.showComments(target);
                    onClose();
                },
            },
        ];

        if (imageId) {
            const fullUrl = getImageUrl(diagramId, imageId, 'full');
            const downloadUrl = getImageUrl(diagramId, imageId, 'download');

            items.push(
                {
                    separator: true,
                },
                {
                    label: 'На весь экран',
                    icon: faExpand,
                    action: () => {
                        imageViewerUrl.set(fullUrl);
                        onClose();
                    },
                },
                {
                    label: 'Скачать оригинал',
                    icon: faDownload,
                    action: () => {
                        const link = document.createElement('a');
                        link.href = downloadUrl;
                        link.click();
                        onClose();
                    },
                },
            );
        }

        if (hasPermission(permission)) {
            items.push({
                separator: true,
            });

            // У Image-шейпа нет текста для правки и он не клонируется.
            if (!imageId) {
                items.push(
                    {
                        label: 'Редактировать текст',
                        icon: faEdit,
                        action: () => {
                            groupActionManager.startEditing(target);
                            onClose();
                        },
                    },
                    {
                        label: 'Клонировать',
                        icon: faClone,
                        action: () => {
                            groupActionManager.cloneObject(target);
                            onClose();
                        },
                    },
                );
            }

            items.push({
                label: 'Удалить',
                icon: faTrash,
                variant: 'danger',
                action: () => {
                    groupActionManager.deleteObject(target);
                    onClose();
                },
            });
            items.push(
                {
                    separator: true,
                },
                ...layerItems(target, onClose),
            );
        }

        return items;
    });

    function handleClickOutside(event: PointerEvent) {
        if (!(event.target as HTMLElement).closest('.context-menu')) {
            onClose();
        }
    }
</script>

<svelte:window onpointerdown={handleClickOutside} />

<div
    class="context-menu"
    style="left: {x}px; top: {y}px; transform: translate(
        min(0px, calc(100vw - {x}px - 100% - 8px)),
        min(0px, calc(100vh - {y}px - 100% - 8px))
    );"
>
    <!-- eslint-disable-next-line svelte/require-each-key -->
    {#each menuItems as item}
        {#if item.separator}
            <div class="separator"></div>
        {:else}
            <button class="menu-item" class:danger={item.variant === 'danger'} onclick={item.action} type="button">
                {#if item.icon}
                    <Fa fw icon={item.icon} />
                {/if}
                <span>{item.label}</span>
            </button>
        {/if}
    {/each}
</div>

<style>
    .context-menu {
        position: fixed;
        background: var(--surface-bg);
        border: 1px solid var(--border-color);
        border-radius: 8px;
        box-shadow: var(--shadow);
        padding: 4px 0;
        min-width: 160px;
        z-index: 2000;
    }

    .menu-item {
        display: flex;
        align-items: center;
        gap: 8px;
        width: 100%;
        padding: 8px 12px;
        border: none;
        background: none;
        cursor: pointer;
        font-size: 14px;
        text-align: left;
        color: var(--bs-emphasis-color);
        transition: background-color 0.15s;
    }

    .menu-item:hover {
        background-color: var(--bs-tertiary-bg);
    }

    .menu-item.danger {
        color: var(--bs-danger-text-emphasis);
    }

    .menu-item.danger:hover {
        background-color: var(--bs-danger-bg-subtle);
    }

    .separator {
        height: 1px;
        background-color: var(--border-color);
        margin: 4px 8px;
    }
</style>
