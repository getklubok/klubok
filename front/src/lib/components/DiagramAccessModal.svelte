<script lang="ts">
    import { api } from '../Api';
    import type { DiagramCollaborator, DiagramCollaboratorPermissionEnum } from '../ApiSpec';
    import type { RespDataMsg } from '../types';
    import PermissionBadge from './PermissionBadge.svelte';
    import { alert } from '../utils/alert';

    let { show = $bindable(), diagram } = $props();

    let collaborators: DiagramCollaborator[] = $state([]);
    let loading = $state(false);
    let error: string | null = $state(null);
    let newUser = $state('');
    let newPermission: DiagramCollaboratorPermissionEnum = $state('VIEW');

    $effect(() => {
        if (show && diagram) {
            loadCollaborators();
        }
    });

    async function loadCollaborators() {
        if (!diagram) return;
        loading = true;
        error = null;
        try {
            const response = await api.diagramService.collaboratorList(diagram.id);
            collaborators = response.collaborators;
        } catch (err) {
            error = (err as RespDataMsg)?.response?.data?.message || 'Ошибка загрузки участников';
        } finally {
            loading = false;
        }
    }

    async function addCollaborator() {
        if (!diagram || !newUser.trim()) return;
        try {
            await api.diagramService.collaboratorUpsert(diagram.id, {
                userNameOrEmail: newUser.trim(),
                permission: newPermission,
            });
            // eslint-disable-next-line require-atomic-updates
            newUser = '';
            // eslint-disable-next-line require-atomic-updates
            newPermission = 'VIEW';
            await loadCollaborators();
        } catch (err) {
            alert({
                title: 'Ошибка добавления участника',
                message: (err as RespDataMsg)?.response?.data?.message || (err as Error).message,
                type: 'error',
            });
        }
    }

    async function removeCollaborator(collaboratorId: string) {
        if (!diagram) return;
        try {
            await api.diagramService.collaboratorDelete(diagram.id, collaboratorId);
            await loadCollaborators();
        } catch (err) {
            alert({
                title: 'Ошибка удаления участника',
                message: (err as RespDataMsg)?.response?.data?.message || (err as Error).message,
                type: 'error',
            });
        }
    }

    function closeModal() {
        show = false;
        collaborators = [];
        error = null;
        newUser = '';
        newPermission = 'VIEW';
    }
</script>

<!-- Modal -->
{#if show}
    <div class="modal fade show d-block" tabindex="-1" style="background-color: rgba(0,0,0,0.5); z-index: 1060;">
        <div class="modal-dialog modal-lg modal-dialog-centered">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title d-flex">
                        Общий доступ к диаграмме: &nbsp;
                        <span class="diagram-title">{diagram?.title}</span>
                    </h5>
                    <!-- svelte-ignore a11y_consider_explicit_label -->
                    <button type="button" class="btn-close" onclick={closeModal}></button>
                </div>
                <div class="modal-body" style="max-height: 70vh; overflow-y: auto;">
                    {#if loading}
                        <p>Загрузка...</p>
                    {:else if error}
                        <p class="text-danger">{error}</p>
                    {:else}
                        <div class="mb-3">
                            <h6>Участники</h6>
                            {#if collaborators.length === 0}
                                <p>Нет участников</p>
                            {:else}
                                <ul class="list-group">
                                    {#each collaborators as collaborator (collaborator.id)}
                                        <li class="list-group-item d-flex justify-content-between align-items-center">
                                            <span>
                                                <strong>{collaborator.userName}</strong>
                                                <PermissionBadge permission={collaborator.permission} />
                                            </span>
                                            <button
                                                class="btn btn-sm btn-outline-danger"
                                                onclick={() => removeCollaborator(collaborator.id)}
                                            >
                                                Удалить
                                            </button>
                                        </li>
                                    {/each}
                                </ul>
                            {/if}
                        </div>
                        <div class="mb-3">
                            <h6>Пригласить пользователя</h6>
                            <div class="input-group mb-2">
                                <input
                                    type="text"
                                    class="form-control"
                                    placeholder="Введите никнейм"
                                    bind:value={newUser}
                                />
                                <select class="form-select select-permission" bind:value={newPermission}>
                                    <option value="VIEW">Просмотр</option>
                                    <option value="EDIT">Редактирование</option>
                                </select>
                                <button class="btn btn-primary" onclick={addCollaborator}> Пригласить </button>
                            </div>
                        </div>
                    {/if}
                </div>
            </div>
        </div>
    </div>
{/if}

<style>
    .modal {
        z-index: 1060;
    }

    .diagram-title {
        display: inline-block;
        max-width: 300px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    .select-permission {
        max-width: 190px;
    }
</style>
