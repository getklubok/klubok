import logger from './logger';

const smartCaptchaScriptURL = 'https://smartcaptcha.cloud.yandex.ru/captcha.js';

interface SmartCaptchaConfig {
    sitekey: string;
    hl: string;
}

interface SmartCaptchaWidget {
    render: (containerId: string, config: SmartCaptchaConfig) => void;
}

declare global {
    interface Window {
        smartCaptcha?: SmartCaptchaWidget;
    }
}

let smartCaptchaLoadPromise: Promise<void> | null = null;

/**
 * Загрузить скрипт капчи. Вызывается только когда капча включена.
 */
const loadSmartCaptcha = (): Promise<void> => {
    if (window.smartCaptcha) {
        return Promise.resolve();
    }

    if (smartCaptchaLoadPromise) {
        return smartCaptchaLoadPromise;
    }

    smartCaptchaLoadPromise = new Promise<void>((resolve, reject) => {
        const script = document.createElement('script');
        script.src = smartCaptchaScriptURL;
        script.async = true;
        script.onload = () => resolve();
        script.onerror = () => {
            script.remove();
            smartCaptchaLoadPromise = null;
            reject(new Error('Не удалось загрузить SmartCaptcha'));
        };
        document.head.appendChild(script);
    });

    return smartCaptchaLoadPromise;
};

/**
 * Отрисовать виджет капчи, загрузив скрипт при необходимости.
 */
export const renderSmartCaptcha = (containerId: string, config: SmartCaptchaConfig): void => {
    void loadSmartCaptcha()
        .then(() => {
            if (!document.getElementById(containerId)) {
                return;
            }

            window.smartCaptcha?.render(containerId, config);
        })
        .catch((err) => logger.error('Ошибка загрузки SmartCaptcha:', err));
};

/**
 * Получить токен капчи
 */
export const getCaptchaToken = (captchaContainerId: string): string => {
    const container = document.getElementById(captchaContainerId);
    if (!container) {
        return '';
    }

    const input = container.querySelector('input[name="smart-token"]');
    if (!input) {
        return '';
    }

    //@ts-expect-error TODO: fix
    return input.value;
};
