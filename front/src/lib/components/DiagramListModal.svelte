<script lang="ts">
    import Fa from 'svelte-fa';
    import {
        faUserGroup,
        faEdit,
        faKey,
        faTrash,
        faClone,
        faFileImport,
        faPlus,
        faCode,
    } from '@fortawesome/free-solid-svg-icons';

    import { onMount } from 'svelte';

    import { api } from '../Api';
    import type { Diagram } from '../ApiSpec';
    import type { RespDataMsg } from '../types';
    import { currentUser } from '../stores';

    import DiagramAccessModal from './DiagramAccessModal.svelte';
    import EmbedModal from './EmbedModal.svelte';
    import PermissionBadge from './PermissionBadge.svelte';

    import hasPermission from '../services/hasPermission';

    import { alert } from '../utils/alert';
    import { confirm } from '../utils/confirm';
    import { prompt } from '../utils/prompt';
    import { importerManager } from '../services/ImporterManager';

    let { show = $bindable() } = $props();

    let diagrams: Diagram[] = $state([]);
    let loading = $state(true);
    let error: string | null = $state(null);
    let accessModalShow = $state(false);
    let embedModalShow = $state(false);
    let selectedDiagram: Diagram | null = $state(null);

    // стражи однократности автосоздания и параллельных запросов создания (D3/D4 design.md)
    let creating = false;
    let autoCreateAttempted = $state(false);
    let autoCreatedNavigated = false; // автосоздание этого показа успешно и уводило на диаграмму (D6); сбрасывается при повторном открытии

    const AUTO_CREATE_TITLE = 'Моя диаграмма';

    onMount(async () => {
        try {
            const response = await api.diagramService.diagramList();
            diagrams = response.diagrams;
            if (error === null && diagrams.length === 0 && !autoCreateAttempted) {
                autoCreateAttempted = true; // одна попытка за показ списка (D3)
                const newDiagram = await createDiagram(AUTO_CREATE_TITLE);
                if (newDiagram) {
                    autoCreatedNavigated = true;
                    openDiagram(newDiagram.id);
                }
            }
        } catch (err) {
            // eslint-disable-next-line require-atomic-updates
            error = (err as RespDataMsg)?.response?.data?.message || 'Ошибка загрузки диаграмм';
        } finally {
            loading = false;
        }
    });

    // Повторное открытие списка (Tools → «Мои диаграммы») — новый ручной контекст:
    // D6-страж актуален только для prompt'а, подтверждённого сразу после автоперехода
    $effect(() => {
        if (show) autoCreatedNavigated = false;
    });

    function openDiagram(id: string) {
        window.location.hash = `/diagrams/${id}`;
        closeModal();
    }

    // Общий хелпер создания (D2): возвращает созданную диаграмму или null при ошибке/отказе.
    async function createDiagram(title: string): Promise<Diagram | null> {
        if (creating) return null; // повторный вызов во время запроса игнорируется (D4)
        creating = true;
        try {
            const response = await api.diagramService.diagramCreate({ title });
            return response.diagram;
        } catch (err) {
            alert({
                title: 'Ошибка создания диаграммы',
                message: (err as RespDataMsg)?.response?.data?.message || (err as Error).message,
                type: 'error',
            });
            return null;
        } finally {
            // eslint-disable-next-line require-atomic-updates
            creating = false;
        }
    }

    async function handleCreateButton() {
        const title = await prompt({
            title: 'Новая диаграмма',
            message: 'Введите название диаграммы',
        });
        if (!title) return;
        // Если автосоздание этого показа уже уводило на диаграмму, подтверждённый prompt закрывается без создания (D6)
        if (autoCreatedNavigated) return;
        const newDiagram = await createDiagram(title);
        if (newDiagram) openDiagram(newDiagram.id);
    }

    async function updateTitle(diagram: Diagram) {
        const newTitle = await prompt({
            title: 'Обновление названия диаграммы',
            message: 'Введите новое название',
            defaultValue: diagram.title,
        });
        if (!newTitle || newTitle === diagram.title) return;

        try {
            await api.diagramService.diagramSetTitle(diagram.id, { title: newTitle });
            // eslint-disable-next-line require-atomic-updates
            diagram.title = newTitle;
            diagrams = [...diagrams]; // trigger reactivity
        } catch (err) {
            alert({
                title: 'Ошибка обновления названия',
                message: (err as RespDataMsg)?.response?.data?.message || (err as Error).message,
                type: 'error',
            });
        }
    }

    async function deleteDiagram(diagram: Diagram) {
        const confirmed = await confirm({
            title: 'Удаление диаграммы',
            message: `Вы уверены, что хотите удалить диаграмму "${diagram.title}"?`,
        });
        if (!confirmed) return;

        try {
            await api.diagramService.diagramDelete(diagram.id);
            diagrams = diagrams.filter((d) => d.id !== diagram.id);
            if (window.location.hash === `#/diagrams/${diagram.id}`) {
                window.location.hash = '#/diagrams';
            }
        } catch (err) {
            alert({
                title: 'Ошибка удаления диаграммы',
                message: (err as RespDataMsg)?.response?.data?.message || (err as Error).message,
                type: 'error',
            });
        }
    }

    async function cloneDiagram(diagram: Diagram) {
        const confirmed = await confirm({
            title: 'Клонирование диаграммы',
            message: `Клонировать диаграмму "${diagram.title}"?`,
        });
        if (!confirmed) return;

        try {
            const response = await api.diagramService.diagramClone(diagram.id);
            const clonedDiagram = response.diagram;
            diagrams = [...diagrams, clonedDiagram];
            window.location.hash = `/diagrams/${clonedDiagram.id}`;
            show = false;
        } catch (err) {
            alert({
                title: 'Ошибка клонирования диаграммы',
                message: (err as RespDataMsg)?.response?.data?.message || (err as Error).message,
                type: 'error',
            });
        }
    }

    function closeModal() {
        show = false;
    }

    function openAccessModal(diagram: Diagram) {
        selectedDiagram = diagram;
        accessModalShow = true;
    }

    function openEmbedModal(diagram: Diagram) {
        selectedDiagram = diagram;
        embedModalShow = true;
    }

    function importDiagram() {
        const input = document.createElement('input');
        input.type = 'file';
        input.accept = '.klubok';
        input.onchange = async (e) => {
            const file = (e.target as HTMLInputElement).files?.[0];
            if (file) {
                await importerManager.fromFile(file);
            }
        };
        input.click();
    }
</script>

<!-- Modal -->
<div class="modal fade show d-block" tabindex="-1" style="background-color: rgba(0,0,0,0.5);">
    <div class="modal-dialog modal-lg modal-dialog-centered">
        <div class="modal-content">
            <div class="modal-header">
                <h5 class="modal-title">Список диаграмм</h5>
                <!-- svelte-ignore a11y_consider_explicit_label -->
                <button type="button" class="btn-close" onclick={closeModal}></button>
            </div>
            <div class="modal-body" style="max-height: 90vh; overflow-y: auto;">
                {#if $currentUser?.isFake}
                    <div class="alert alert-warning fake-account-banner" role="alert">
                        Включён режим автоматических аккаунтов. Данные доступны до тех пор, пока в браузере хранится
                        сессионная кука. Рекомендуется делать бэкапы диаграмм.
                    </div>
                {/if}
                {#if loading}
                    <p>Загрузка...</p>
                {:else if error}
                    <p class="text-danger">{error}</p>
                {:else if diagrams.length === 0}
                    <p>Нет диаграмм</p>
                {:else}
                    <div class="list-group">
                        {#each diagrams as diagram (diagram.id)}
                            <div class="list-group-item diagram-row">
                                <!-- Колонка названия -->
                                <div class="diagram-col-title">
                                    <a
                                        href="#/diagrams/{diagram.id}"
                                        class="diagram-title text-decoration-none"
                                        onclick={() => openDiagram(diagram.id)}
                                        title={diagram.title}
                                    >
                                        {diagram.title}
                                    </a>
                                </div>

                                <!-- Колонка индикаторов -->
                                <div class="diagram-col-indicators">
                                    {#if diagram.hasCollaborators}
                                        <span class="badge bg-info" title="Есть участники">
                                            <Fa icon={faUserGroup} />
                                        </span>
                                    {/if}
                                    <PermissionBadge permission={diagram.permission} />
                                </div>

                                <!-- Колонка действий -->
                                <div class="diagram-col-actions">
                                    <button
                                        class="btn btn-sm btn-outline-info"
                                        onclick={() => cloneDiagram(diagram)}
                                        title="Клонировать диаграмму"
                                    >
                                        <Fa fw icon={faClone} />
                                    </button>
                                    {#if hasPermission(diagram.permission)}
                                        <button
                                            class="btn btn-sm btn-outline-secondary"
                                            onclick={() => updateTitle(diagram)}
                                            title="Изменить название"
                                        >
                                            <Fa fw icon={faEdit} />
                                        </button>
                                    {:else}
                                        <!-- Пустой плейсхолдер для выравнивания -->
                                        <span class="btn-placeholder"></span>
                                    {/if}
                                    {#if diagram.permission === 'OWNER'}
                                        <button
                                            class="btn btn-sm btn-outline-success"
                                            onclick={() => openEmbedModal(diagram)}
                                            title="Embed диаграммы"
                                        >
                                            <Fa fw icon={faCode} />
                                        </button>
                                        <button
                                            class="btn btn-sm btn-outline-primary"
                                            onclick={() => openAccessModal(diagram)}
                                            title="Настроить доступ"
                                        >
                                            <Fa fw icon={faKey} />
                                        </button>
                                        <button
                                            class="btn btn-sm btn-outline-danger"
                                            onclick={() => deleteDiagram(diagram)}
                                            title="Удалить диаграмму"
                                        >
                                            <Fa fw icon={faTrash} />
                                        </button>
                                    {:else}
                                        <!-- Пустые плейсхолдеры для выравнивания -->
                                        <span class="btn-placeholder"></span>
                                        <span class="btn-placeholder"></span>
                                        <span class="btn-placeholder"></span>
                                    {/if}
                                </div>
                            </div>
                        {/each}
                    </div>
                {/if}
            </div>
            <div class="modal-footer">
                <button class="btn btn-primary" onclick={handleCreateButton}>
                    <Fa icon={faPlus} />
                    Создать
                </button>
                <button class="btn btn-outline-primary" onclick={importDiagram}>
                    <Fa icon={faFileImport} />
                    Импортировать
                </button>
            </div>
        </div>
    </div>
</div>

<DiagramAccessModal bind:show={accessModalShow} diagram={selectedDiagram} />
<EmbedModal bind:show={embedModalShow} diagram={selectedDiagram} />

<style>
    .modal {
        z-index: 1055;
    }

    .fake-account-banner {
        margin-bottom: 1rem;
    }

    .diagram-title {
        max-width: 250px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    .diagram-row {
        display: grid;
        grid-template-columns: 1fr auto auto;
        gap: 1rem;
        align-items: center;
    }

    .diagram-col-title {
        min-width: 0;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    .diagram-col-indicators {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        min-width: 67.09px; /* фиксированная ширина для индикаторов */
    }

    .diagram-col-actions {
        display: flex;
        gap: 0.25rem;
        min-width: 180px; /* фиксированная ширина для кнопок */
        justify-content: flex-end;
    }

    .btn-placeholder {
        display: inline-block;
        width: 35.5px; /* ширина как у кнопки */
        height: 31px; /* высота как у кнопки */
    }
</style>
