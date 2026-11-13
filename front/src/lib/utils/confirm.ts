interface ConfirmOptions {
    message: string;
    title?: string;
}

export function confirm(options: ConfirmOptions | string): Promise<boolean> {
    const detail = typeof options === 'string' ? { message: options } : options;

    return new Promise((resolve) => {
        function handler(e: Event) {
            window.removeEventListener('confirm-result', handler);
            resolve((e as CustomEvent).detail);
        }

        window.addEventListener('confirm-result', handler);
        window.dispatchEvent(new CustomEvent('show-confirm', { detail }));
    });
}
