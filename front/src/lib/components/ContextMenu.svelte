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
    } from '@fortawesome/free-solid-svg-icons';

    import { Group as FabricGroup } from 'fabric/es';

    import { canvasManager } from '../services/CanvasManager';
    import { groupActionManager } from '../services/GroupActionManager';

    import hasPermission from '../services/hasPermission';

    import type { DiagramCollaboratorPermissionEnum } from '../ApiSpec';

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
        onClose: () => void;
    }
    const { x, y, target, permission, onClose }: Props = $props();

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

        if (hasPermission(permission)) {
            items.push(
                {
                    separator: true,
                },
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
                {
                    label: 'Удалить',
                    icon: faTrash,
                    variant: 'danger',
                    action: () => {
                        groupActionManager.deleteObject(target);
                        onClose();
                    },
                },
                {
                    separator: true,
                },
                {
                    label: 'На задний план',
                    icon: faArrowsDownToLine,
                    action: () => {
                        if (!target.name) return;
                        canvasManager.moveObject(target.name, 'back');
                        onClose();
                    },
                },
                {
                    label: 'Переместить назад',
                    icon: faArrowDown,
                    action: () => {
                        if (!target.name) return;
                        canvasManager.moveObject(target.name, 'down');
                        onClose();
                    },
                },
                {
                    label: 'Переместить вперёд',
                    icon: faArrowUp,
                    action: () => {
                        if (!target.name) return;
                        canvasManager.moveObject(target.name, 'up');
                        onClose();
                    },
                },
                {
                    label: 'На передний план',
                    icon: faArrowsUpToLine,
                    action: () => {
                        if (!target.name) return;
                        canvasManager.moveObject(target.name, 'front');
                        onClose();
                    },
                },
            );
        }

        return items;
    });

    function handleClickOutside(event: MouseEvent) {
        if (!(event.target as HTMLElement).closest('.context-menu')) {
            onClose();
        }
    }
</script>

<svelte:window onclick={handleClickOutside} />

<div class="context-menu" style="left: {x}px; top: {y}px;">
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
        background: white;
        border: 1px solid #ddd;
        border-radius: 8px;
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
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
        color: #333;
        transition: background-color 0.15s;
    }

    .menu-item:hover {
        background-color: #f0f0f0;
    }

    .menu-item.danger {
        color: #dc3545;
    }

    .menu-item.danger:hover {
        background-color: #fee;
    }

    .separator {
        height: 1px;
        background-color: #e0e0e0;
        margin: 4px 8px;
    }
</style>
