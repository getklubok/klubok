<script lang="ts">
    let { show = $bindable(false), panAlways = false } = $props();
</script>

{#if show}
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div class="help-overlay" onclick={() => (show = false)}>
        <div class="help-modal" onclick={(e) => e.stopPropagation()}>
            <h3>Управление диаграммой</h3>
            <ul>
                <li><kbd>Колесико мыши</kbd> — масштабирование</li>
                {#if panAlways}
                    <li><kbd>ЛКМ</kbd> — панорамирование</li>
                {:else}
                    <li><kbd>Инструмент «Рука»</kbd> — панорамирование</li>
                    <li><kbd>Пробел</kbd> + <kbd>ЛКМ</kbd> — панорамирование в любом режиме</li>
                {/if}
            </ul>
            <button class="btn btn-sm btn-primary" onclick={() => (show = false)}>Закрыть</button>
        </div>
    </div>
{/if}

<style>
    .help-overlay {
        position: fixed;
        inset: 0;
        display: flex;
        justify-content: center;
        align-items: center;
        background: rgba(0, 0, 0, 0.4);
        z-index: 2000;
    }

    .help-modal {
        background: var(--surface-bg);
        border-radius: 10px;
        padding: 1.5rem 2rem;
        max-width: 380px;
        width: 90%;
        box-shadow: 0 4px 24px rgba(0, 0, 0, 0.2);
        text-align: center;
    }

    .help-modal h3 {
        margin: 0 0 1rem;
        font-size: 1rem;
        color: var(--text-color);
    }

    .help-modal ul {
        list-style: none;
        padding: 0;
        margin: 0 0 1.25rem;
        text-align: left;
    }

    .help-modal li {
        padding: 0.4rem 0;
        font-size: 0.9rem;
        color: var(--text-color);
    }

    .help-modal kbd {
        display: inline-block;
        padding: 0.15rem 0.4rem;
        font-size: 0.8rem;
        font-family: inherit;
        color: var(--text-color);
        background: var(--diagram-bg);
        border: 1px solid var(--border-color);
        border-radius: 4px;
        box-shadow: var(--shadow);
    }
</style>
