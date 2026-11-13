<script lang="ts">
    import { onMount } from 'svelte';

    let visible = $state(false);
    let title = $state('');
    let message = $state('');
    let defaultValue = $state('');
    let inputValue = $state('');
    let pendingResolve: ((value: string | null) => void) | null = $state(null);

    function show(options: { title?: string; message: string; defaultValue?: string }): Promise<string | null> {
        title = options.title || 'Ввод';
        message = options.message;
        defaultValue = options.defaultValue || '';
        inputValue = defaultValue;
        visible = true;

        return new Promise((res) => {
            pendingResolve = res;
        });
    }

    onMount(() => {
        window.addEventListener('show-prompt', (e: Event) => {
            const detail = (e as CustomEvent).detail;
            show(detail).then((result) => {
                window.dispatchEvent(new CustomEvent('prompt-result', { detail: result }));
            });
        });
    });

    function onSubmit() {
        const value = inputValue.trim() || null;
        pendingResolve?.(value);
        close();
    }

    function onCancel() {
        pendingResolve?.(null);
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
                    <p class="mb-2">{message}</p>
                    <input
                        type="text"
                        class="form-control"
                        bind:value={inputValue}
                        onkeydown={(e) => e.key === 'Enter' && onSubmit()}
                    />
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" onclick={onCancel}>Отмена</button>
                    <button type="button" class="btn btn-primary" onclick={onSubmit}>ОК</button>
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
