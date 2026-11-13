interface PromptOptions {
    message: string;
    title?: string;
    defaultValue?: string;
}

export function prompt(options: PromptOptions | string, defaultValue?: string): Promise<string | null> {
    let detail: PromptOptions;

    if (typeof options === 'string') {
        detail = { message: options, defaultValue };
    } else {
        detail = options;
    }

    return new Promise((resolve) => {
        function handler(e: Event) {
            window.removeEventListener('prompt-result', handler);
            resolve((e as CustomEvent).detail);
        }

        window.addEventListener('prompt-result', handler);
        window.dispatchEvent(new CustomEvent('show-prompt', { detail }));
    });
}
