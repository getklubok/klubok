<script lang="ts">
    import Fa from 'svelte-fa';
    import {
        faSpinner,
        faCheck,
        faExclamationTriangle,
        faCloud,
        faCrosshairs,
        faExpand,
    } from '@fortawesome/free-solid-svg-icons';

    import { coordinates, diagramTitle, diagramPermission, syncStatus, onlineUsers } from '../stores';

    import Avatar from './Avatar.svelte';
    import PermissionBadge from './PermissionBadge.svelte';

    import { viewportManager } from '../services/ViewportManager';

    let currentZoom = 100;
    let currentSyncStatus = 'idle';

    coordinates.subscribe((v) => {
        currentZoom = Math.round(v.zoom * 100);
    });

    syncStatus.subscribe((status) => {
        currentSyncStatus = status;
    });
</script>

<div class="statusbar">
    <div class="statusbar-item">
        <strong>Zoom:</strong>
        {currentZoom}%
        <div class="btn-group ms-2">
            <button
                class="btn btn-sm btn-outline-primary"
                type="button"
                onclick={() => viewportManager.centerView()}
                title="Центрировать"
            >
                <Fa icon={faCrosshairs} />
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
                <Fa icon={faExpand} />
            </button>
        </div>
    </div>
    {#if $diagramTitle}
        <div class="statusbar-item">|</div>
        <div class="statusbar-item">
            <div class="d-flex align-items-center gap-2">
                <span class="diagram-title" title={$diagramTitle}>
                    {$diagramTitle}
                </span>
                <PermissionBadge permission={$diagramPermission} />
            </div>
        </div>
    {/if}
    <div class="statusbar-item">|</div>
    <div class="statusbar-item">
        <span
            class="sync-status"
            class:saving={currentSyncStatus === 'saving'}
            class:saved={currentSyncStatus === 'saved'}
            class:error={currentSyncStatus === 'error'}
        >
            {#if currentSyncStatus === 'saving'}
                <Fa pulse icon={faSpinner} /> Сохранение...
            {:else if currentSyncStatus === 'saved'}
                <Fa icon={faCheck} /> Сохранено
            {:else if currentSyncStatus === 'error'}
                <Fa icon={faExclamationTriangle} /> Ошибка
            {:else}
                <Fa icon={faCloud} /> Синхронизировано
            {/if}
        </span>
    </div>
    {#if $onlineUsers.length > 0}
        <div class="statusbar-item">|</div>
        <div class="statusbar-item">
            <div class="online-users">
                {#each $onlineUsers as user, index (index)}
                    <Avatar {user} {index} />
                {/each}
            </div>
        </div>
    {/if}
</div>

<style>
    .statusbar {
        font-family: monospace;
        font-size: 13px;
        display: flex;
        gap: 15px;
    }

    .statusbar-item {
        display: flex;
        align-items: center;
    }

    .statusbar-item strong {
        margin-right: 5px;
        color: #4caf50;
    }

    .diagram-title {
        color: #904caf;
        max-width: 200px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    .sync-status {
        color: #4caf50;
        font-size: 12px;
    }

    .sync-status.saving {
        color: #ff9800;
    }

    .sync-status.saved {
        color: #4caf50;
    }

    .sync-status.error {
        color: #f44336;
    }

    .online-users {
        display: flex;
        align-items: center;
    }

    /* ./Avatar.svelte */
    .online-users:hover :global {
        .user-avatar {
            margin-left: 0 !important;
        }
    }
</style>
