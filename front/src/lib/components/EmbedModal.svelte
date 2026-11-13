<script lang="ts">
    import Fa from 'svelte-fa';
    import { faTrash, faCopy, faCode } from '@fortawesome/free-solid-svg-icons';

    import { api } from '../Api';
    import type { EmbedInfo } from '../ApiSpec';
    import type { RespDataMsg } from '../types';

    import { alert } from '../utils/alert';
    import { confirm } from '../utils/confirm';

    let { show = $bindable(), diagram } = $props();

    let embeds: EmbedInfo[] = $state([]);
    let loading = $state(false);
    let error: string | null = $state(null);

    $effect(() => {
        if (show && diagram) {
            loadEmbeds();
        }
    });

    async function loadEmbeds() {
        if (!diagram) return;
        loading = true;
        error = null;
        try {
            const response = await api.diagramService.embedList(diagram.id);
            embeds = response.embeds;
        } catch (err) {
            error = (err as RespDataMsg)?.response?.data?.message || 'Ошибка загрузки embed-кодов';
        } finally {
            loading = false;
        }
    }

    async function createEmbed() {
        if (!diagram) return;
        try {
            const response = await api.diagramService.embedCreate(diagram.id);
            embeds = [...embeds, response.embed];
        } catch (err) {
            alert({
                title: 'Ошибка создания embed-кода',
                message: (err as RespDataMsg)?.response?.data?.message || (err as Error).message,
                type: 'error',
            });
        }
    }

    async function revokeEmbed(embedId: string) {
        if (!diagram) return;
        const confirmed = await confirm({
            title: 'Отозвать embed-код',
            message: 'Вы уверены, что хотите отозвать этот embed-код?',
        });
        if (!confirmed) return;

        try {
            await api.diagramService.embedDelete(diagram.id, embedId);
            embeds = embeds.map((e) => (e.id === embedId ? { ...e, isActive: false } : e));
        } catch (err) {
            alert({
                title: 'Ошибка отзыва embed-кода',
                message: (err as RespDataMsg)?.response?.data?.message || (err as Error).message,
                type: 'error',
            });
        }
    }

    function copyToClipboard(text: string) {
        navigator.clipboard.writeText(text).then(() => {
            alert({ message: 'Скопировано в буфер обмена', type: 'success', duration: 2000 });
        });
    }

    function closeModal() {
        show = false;
        embeds = [];
        error = null;
    }

    function formatDate(dateStr: string | null | undefined): string {
        if (!dateStr) return '—';
        return new Date(dateStr).toLocaleString('ru-RU');
    }

    function getEmbedUrl(token: string): string {
        const baseUrl = window.location.origin;
        return `${baseUrl}/embed/#/token/${token}`;
    }

    function generateIframeCode(embed: EmbedInfo): string {
        const url = getEmbedUrl(embed.embedCode);
        return `<iframe src="${url}" width="800" height="600" frameborder="0" allowfullscreen></iframe>`;
    }
</script>

{#if show}
    <div class="modal fade show d-block" tabindex="-1" style="background-color: rgba(0,0,0,0.5); z-index: 1060;">
        <div class="modal-dialog modal-lg modal-dialog-centered">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title d-flex align-items-center gap-2">
                        <Fa icon={faCode} /> Embed диаграммы:
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
                        <div class="form-section">
                            <div class="form-section-header">
                                <h6 class="mb-0">Создание embed-кода</h6>
                            </div>

                            <button class="btn btn-primary" onclick={createEmbed}>Создать</button>
                        </div>

                        <!-- Подсказка по параметрам URL -->
                        <div class="form-section">
                            <div class="form-section-header">
                                <h6 class="mb-0">Параметры URL</h6>
                            </div>
                            <p class="mb-1 text-muted small">
                                К embed-адресу (после <code>/embed/#/token/&lt;token&gt;</code>) можно добавить
                                опциональный параметр <code>?theme=</code>. Если его не указывать, используется
                                системная тема:
                            </p>
                            <ul class="mb-0 text-muted small">
                                <li><code>?theme=system</code> — системная (по умолчанию, можно не указывать)</li>
                                <li><code>?theme=light</code> — светлая тема</li>
                                <li><code>?theme=dark</code> — тёмная тема</li>
                            </ul>
                        </div>

                        <!-- Список существующих embed-кодов -->
                        <div>
                            <div class="form-section-header">
                                <h6>Существующие embed-коды</h6>
                            </div>
                            {#if embeds.length === 0}
                                <p>Нет созданных embed-кодов</p>
                            {:else}
                                <div class="embed-list">
                                    {#each embeds as embed (embed.id)}
                                        <div class="embed-item card mb-2" class:inactive={!embed.isActive}>
                                            <div class="card-body">
                                                <div class="d-flex justify-content-between align-items-start">
                                                    <div>
                                                        <span
                                                            class="badge"
                                                            class:bg-success={embed.isActive}
                                                            class:bg-secondary={!embed.isActive}
                                                        >
                                                            {embed.isActive ? 'Активен' : 'Отозван'}
                                                        </span>
                                                        <div class="mt-1">
                                                            <small>
                                                                Просмотров: {embed.viewsCount}
                                                                {#if embed.lastViewedAt}
                                                                    <br />
                                                                    Последний просмотр: {formatDate(embed.lastViewedAt)}
                                                                {/if}
                                                            </small>
                                                        </div>
                                                    </div>
                                                    <div class="d-flex gap-1">
                                                        {#if embed.isActive}
                                                            <button
                                                                class="btn btn-sm btn-outline-primary"
                                                                onclick={() =>
                                                                    copyToClipboard(generateIframeCode(embed))}
                                                                title="Скопировать код"
                                                            >
                                                                <Fa icon={faCopy} />
                                                            </button>
                                                            <button
                                                                class="btn btn-sm btn-outline-danger"
                                                                onclick={() => revokeEmbed(embed.id)}
                                                                title="Отозвать"
                                                            >
                                                                <Fa icon={faTrash} />
                                                            </button>
                                                        {:else}
                                                            <span class="text-muted">
                                                                <small>Отозван {formatDate(embed.updatedAt)}</small>
                                                            </span>
                                                        {/if}
                                                    </div>
                                                </div>
                                            </div>
                                        </div>
                                    {/each}
                                </div>
                            {/if}
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

    .embed-item.inactive {
        opacity: 0.6;
    }

    .form-section {
        background: var(--surface-bg);
        border: 1px solid var(--border-color);
        border-radius: 0.5rem;
        padding: 1.25rem;
        margin-bottom: 1.5rem;
    }

    .form-section-header {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        margin-bottom: 1rem;
        padding-bottom: 0.75rem;
        border-bottom: 1px solid var(--border-color);
    }

    .form-section-header h6 {
        font-weight: 600;
        color: var(--text-color);
    }
</style>
