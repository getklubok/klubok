<script lang="ts">
    import { onMount } from 'svelte';
    let visible = $state(false);
    let title = $state('');
    let message = $state('');
    let pendingResolve: ((value: boolean) => void) | null = $state(null);

    function show(options: { title?: string; message: string }): Promise<boolean> {
        title = options.title || 'Подтверждение';
        message = options.message;
        visible = true;

        return new Promise((res) => {
            pendingResolve = res;
        });
    }

    onMount(() => {
        window.addEventListener('show-confirm', (e: Event) => {
            const detail = (e as CustomEvent).detail;
            show(detail).then((result) => {
                window.dispatchEvent(new CustomEvent('confirm-result', { detail: result }));
            });
        });
    });

    function onConfirm() {
        pendingResolve?.(true);
        close();
    }

    function onCancel() {
        pendingResolve?.(false);
        close();
    }

    function close() {
        visible = false;
        pendingResolve = null;
    }
</script>

{#if visible}
    <div class="modal fade show d-block" tabindex="-1" style="background-color: rgba(0,0,0,0.5);">
        <div class="modal-dialog modal-dialog-centered">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title">{title}</h5>
                    <button type="button" class="btn-close" onclick={onCancel} aria-label="Close"></button>
                </div>
                <div class="modal-body">
                    <p>{message}</p>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" onclick={onCancel}>Нет</button>
                    <button type="button" class="btn btn-primary" onclick={onConfirm}>Да</button>
                </div>
            </div>
        </div>
    </div>
{/if}

<style>
    .modal {
        z-index: 1060;
    }
</style>
