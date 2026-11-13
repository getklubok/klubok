<script lang="ts">
    import DOMPurify from 'dompurify';
    import { marked } from 'marked';
    import { untrack } from 'svelte';

    let { show = $bindable() } = $props();

    let content: string = $state('');
    let loadState: 'loading' | 'loaded' | 'error' = $state('loading');

    function closeModal() {
        show = false;
    }

    $effect(() => {
        if (show) {
            untrack(() => loadPolicy());
        }
    });

    async function loadPolicy() {
        loadState = 'loading';
        try {
            const res = await fetch('/privacy.md');
            if (!res.ok) {
                loadState = 'error';
                return;
            }
            const md = await res.text();
            content = DOMPurify.sanitize(marked.parse(md) as string);
            loadState = 'loaded';
        } catch {
            loadState = 'error';
        }
    }
</script>

{#if show}
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div class="modal fade show d-block" tabindex="-1" style="background-color: rgba(0,0,0,0.5);" onclick={closeModal}>
        <div class="modal-dialog modal-dialog-centered modal-lg" onclick={(e) => e.stopPropagation()}>
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title">Политика в отношении обработки персональных данных</h5>
                    <button type="button" class="btn-close" aria-label="Close" onclick={closeModal}></button>
                </div>
                <div class="modal-body" style="max-height: 90vh; overflow-y: auto;">
                    {#if loadState === 'loading'}
                        <div class="text-center py-4">
                            <div class="spinner-border text-secondary" role="status">
                                <span class="visually-hidden">Загрузка...</span>
                            </div>
                        </div>
                    {:else if loadState === 'loaded'}
                        <div class="policy-content">
                            <!-- eslint-disable-next-line svelte/no-at-html-tags -->
                            {@html content}
                        </div>
                    {:else}
                        <div class="text-center py-4">
                            <p class="text-muted">Файл политики не найден.</p>
                        </div>
                    {/if}
                </div>
            </div>
        </div>
    </div>
{/if}
