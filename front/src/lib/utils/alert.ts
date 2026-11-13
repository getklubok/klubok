interface AlertOptions {
    message: string;
    title?: string;
    type?: 'success' | 'danger' | 'warning' | 'info' | 'error';
    duration?: number;
}

export function alert(options: AlertOptions | string): void {
    window.dispatchEvent(
        new CustomEvent('show-alert', {
            detail: typeof options === 'string' ? { message: options } : options,
        }),
    );
}
