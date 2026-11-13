<script lang="ts">
    import Fa from 'svelte-fa';
    import { faCopy } from '@fortawesome/free-solid-svg-icons';
    import { selectedObject } from '../stores';
    import { trimStart } from '../utils/trimStart';

    const isDev = import.meta.env.DEV;

    let copySuccess = $state(false);

    async function copyToClipboard() {
        if (!$selectedObject) return;
        const text = JSON.stringify($selectedObject, null, 2);
        await navigator.clipboard.writeText(text);
        copySuccess = true;
        setTimeout(() => (copySuccess = false), 2000);
    }
</script>

{#if $selectedObject && isDev}
    <div class="debug-panel">
        <div class="debug-header">
            <div class="debug-title">DEBUG</div>
            <button class="copy-btn" onclick={copyToClipboard}>
                {#if copySuccess}
                    <span class="copy-success">Скопировано!</span>
                {/if}
                <Fa icon={faCopy} />
            </button>
        </div>
        <div class="debug-row">
            <span class="debug-label">name:</span>
            <span class="debug-value">{trimStart($selectedObject.name, 15)}</span>
        </div>
        <div class="debug-row">
            <span class="debug-label">type:</span>
            <span class="debug-value">{$selectedObject.type}</span>
        </div>
        <div class="debug-row">
            <span class="debug-label">order:</span>
            <span class="debug-value">{$selectedObject.order}</span>
        </div>
        <div class="debug-divider"></div>
        <div class="debug-row">
            <span class="debug-title">group:</span>
        </div>
        <div class="debug-row">
            <span class="debug-label">left:</span>
            <span class="debug-value">{Math.round($selectedObject.group.left)}</span>
        </div>
        <div class="debug-row">
            <span class="debug-label">top:</span>
            <span class="debug-value">{Math.round($selectedObject.group.top)}</span>
        </div>
        <div class="debug-row">
            <span class="debug-label">width:</span>
            <span class="debug-value">{Math.round($selectedObject.group.width)}</span>
        </div>
        <div class="debug-row">
            <span class="debug-label">height:</span>
            <span class="debug-value">{Math.round($selectedObject.group.height)}</span>
        </div>
        <div class="debug-row">
            <span class="debug-label">scaleX:</span>
            <span class="debug-value">{$selectedObject.group.scaleX.toFixed(2)}</span>
        </div>
        <div class="debug-row">
            <span class="debug-label">scaleY:</span>
            <span class="debug-value">{$selectedObject.group.scaleY.toFixed(2)}</span>
        </div>
        <div class="debug-divider"></div>
        <div class="debug-row">
            <span class="debug-title">shape:</span>
        </div>
        <div class="debug-row">
            <span class="debug-label">left:</span>
            <span class="debug-value">{Math.round($selectedObject.shape.left)}</span>
        </div>
        <div class="debug-row">
            <span class="debug-label">top:</span>
            <span class="debug-value">{Math.round($selectedObject.shape.top)}</span>
        </div>
        <div class="debug-row">
            <span class="debug-label">width:</span>
            <span class="debug-value">{Math.round($selectedObject.shape.width)}</span>
        </div>
        <div class="debug-row">
            <span class="debug-label">height:</span>
            <span class="debug-value">{Math.round($selectedObject.shape.height)}</span>
        </div>
        <div class="debug-row">
            <span class="debug-label">scaleX:</span>
            <span class="debug-value">{$selectedObject.shape.scaleX.toFixed(2)}</span>
        </div>
        <div class="debug-row">
            <span class="debug-label">scaleY:</span>
            <span class="debug-value">{$selectedObject.shape.scaleY.toFixed(2)}</span>
        </div>
        {#if $selectedObject.shape.rx && $selectedObject.shape.ry}
            <div class="debug-row">
                <span class="debug-label">rx:</span>
                <span class="debug-value">{$selectedObject.shape.rx.toFixed(2)}</span>
            </div>
            <div class="debug-row">
                <span class="debug-label">ry:</span>
                <span class="debug-value">{$selectedObject.shape.ry.toFixed(2)}</span>
            </div>
        {/if}
    </div>
{/if}

<style>
    .debug-panel {
        position: absolute;
        top: 58px;
        right: 10px;
        background: rgba(30, 30, 30, 0.95);
        border-radius: 8px;
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
        padding: 12px;
        font-family: 'Fira Code', monospace;
        font-size: 12px;
        color: #e0e0e0;
        z-index: 998;
        min-width: 180px;
        user-select: none;
    }

    .debug-title {
        font-weight: 600;
        color: #ff9800;
        font-size: 13px;
        line-height: 1;
    }

    .debug-row {
        display: flex;
        justify-content: space-between;
        margin: 3px 0;
    }

    .debug-label {
        color: #888;
        padding-right: 8px;
    }

    .debug-value {
        color: #4fc3f7;
        text-align: right;
    }

    .debug-divider {
        height: 1px;
        background: #444;
        margin: 8px 0;
    }

    .debug-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 8px;
    }

    .copy-btn {
        background: transparent;
        border: none;
        color: #888;
        cursor: pointer;
        padding: 2px;
        display: flex;
        align-items: center;
        border-radius: 4px;
        transition:
            color 0.2s,
            background 0.2s;
        position: relative;
    }

    .copy-btn:hover {
        color: #4fc3f7;
        background: rgba(255, 255, 255, 0.1);
    }

    .copy-success {
        color: #4ade80;
        font-size: 11px;
        position: absolute;
        right: 100%;
        white-space: nowrap;
        margin-right: 6px;
        background: rgba(30, 30, 30, 0.95);
        padding: 2px 6px;
        border-radius: 4px;
    }
</style>
