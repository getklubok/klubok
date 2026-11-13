<script lang="ts">
    import { onMount } from 'svelte';

    interface AlertOptions {
        message: string;
        title?: string;
        type?: 'success' | 'danger' | 'warning' | 'info' | 'error';
        duration?: number;
    }

    let alerts: Array<AlertOptions & { id: number; visible: boolean; progress: number }> = $state([]);
    let nextId = 1;

    onMount(() => {
        // Слушаем кастомные события для показа алертов
        window.addEventListener('show-alert', ((event: Event) => {
            const detail = (event as CustomEvent).detail;
            showAlert(detail);
        }) as EventListener);
    });

    export function showAlert(options: AlertOptions | string) {
        const opts = typeof options === 'string' ? { message: options } : options;
        const id = nextId++;
        const duration = opts.duration ?? 0;
        alerts.push({
            id,
            title: opts.title || '',
            message: opts.message,
            type: opts.type || 'info',
            duration: duration,
            visible: true,
            progress: 100,
        });

        if (duration !== 0) {
            const startTime = Date.now();
            const interval = setInterval(() => {
                const elapsed = Date.now() - startTime;
                const remaining = 100 - (elapsed / duration) * 100;
                const alert = alerts.find((a) => a.id === id);
                if (alert) {
                    alert.progress = Math.max(0, remaining);
                }
                if (remaining <= 0) {
                    clearInterval(interval);
                }
            }, 16);

            setTimeout(() => {
                clearInterval(interval);
                closeAlert(id);
            }, duration);
        }
    }

    function closeAlert(id: number) {
        const alert = alerts.find((a) => a.id === id);
        if (alert) {
            alert.visible = false;
            setTimeout(() => {
                alerts = alerts.filter((a) => a.id !== id);
            }, 300);
        }
    }
</script>

<div class="alerts-container position-fixed" style="top: 20px; right: 20px; z-index: 9999;">
    {#each alerts as alert (alert.id)}
        {#if alert.visible}
            <div
                class="alert alert-{alert.type === 'error' ? 'danger' : alert.type} alert-dismissible fade show mb-2"
                role="alert"
                class:with-progress={alert.duration ? alert.duration > 0 : false}
                style="min-width: 300px; max-width: 500px; box-shadow: 0 0.5rem 1rem rgba(0, 0, 0, 0.15); animation: slideIn 0.3s ease-out;"
            >
                {#if alert.title}
                    <strong>{alert.title}</strong>
                    <br />
                {/if}
                {alert.message}
                <button type="button" class="btn-close" aria-label="Close" onclick={() => closeAlert(alert.id)}
                ></button>
                {#if alert.duration && alert.duration > 0}
                    <div class="progress-bar" style="width: {alert.progress}%;"></div>
                {/if}
            </div>
        {/if}
    {/each}
</div>

<style>
    @keyframes slideIn {
        from {
            transform: translateX(100%);
            opacity: 0;
        }
        to {
            transform: translateX(0);
            opacity: 1;
        }
    }

    .alert {
        padding: 0.5rem 2.5rem 0.5rem 0.75rem !important;
        position: relative;
        overflow: hidden;
    }

    .btn-close {
        position: absolute !important;
        padding: 0.75rem 0.75rem !important;
    }

    .progress-bar {
        position: absolute;
        bottom: 0;
        left: 0;
        height: 3px;
        background: rgba(0, 0, 0, 0.25);
        transition: width 0.016s linear;
    }
</style>
