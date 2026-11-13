// @vitest-environment jsdom
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { getCaptchaToken } from './SmartCaptcha';

const smartCaptchaScriptURL = 'https://smartcaptcha.cloud.yandex.ru/captcha.js';

const flushPromises = async () => {
    await new Promise<void>((resolve) => {
        setTimeout(resolve, 0);
    });
};

describe('SmartCaptcha Service', () => {
    let testContainer: HTMLElement;

    beforeEach(() => {
        // Создаем тестовый контейнер перед каждым тестом
        testContainer = document.createElement('div');
        testContainer.id = 'test-captcha-container';
        document.body.appendChild(testContainer);
    });

    afterEach(() => {
        // Очищаем DOM после каждого теста
        if (testContainer && testContainer.parentNode) {
            testContainer.parentNode.removeChild(testContainer);
        }
        // Очищаем весь body от других элементов
        document.body.innerHTML = '';
    });

    describe('getCaptchaToken', () => {
        it('должен вернуть пустую строку когда контейнер не найден', () => {
            const result = getCaptchaToken('non-existent-container');
            expect(result).toBe('');
        });

        it('должен вернуть пустую строку когда контейнер найден, но input не найден', () => {
            const result = getCaptchaToken('test-captcha-container');
            expect(result).toBe('');
        });

        it('должен вернуть значение input когда контейнер и input найдены', () => {
            // Создаем input с нужным name и значением
            const input = document.createElement('input');
            input.name = 'smart-token';
            input.value = 'test-captcha-token-12345';
            testContainer.appendChild(input);

            const result = getCaptchaToken('test-captcha-container');
            expect(result).toBe('test-captcha-token-12345');
        });

        it('должен вернуть пустую строку когда input существует но пустой', () => {
            // Создаем input с нужным name но без значения
            const input = document.createElement('input');
            input.name = 'smart-token';
            input.value = '';
            testContainer.appendChild(input);

            const result = getCaptchaToken('test-captcha-container');
            expect(result).toBe('');
        });

        it('должен вернуть значение input игнорируя другие input элементы', () => {
            // Создаем несколько input элементов, но только один с нужным name
            const otherInput1 = document.createElement('input');
            otherInput1.name = 'other-input';
            otherInput1.value = 'other-value';
            testContainer.appendChild(otherInput1);

            const captchaInput = document.createElement('input');
            captchaInput.name = 'smart-token';
            captchaInput.value = 'captcha-token-67890';
            testContainer.appendChild(captchaInput);

            const otherInput2 = document.createElement('input');
            otherInput2.name = 'another-input';
            otherInput2.value = 'another-value';
            testContainer.appendChild(otherInput2);

            const result = getCaptchaToken('test-captcha-container');
            expect(result).toBe('captcha-token-67890');
        });
    });

    describe('renderSmartCaptcha', () => {
        let renderSmartCaptcha: typeof import('./SmartCaptcha').renderSmartCaptcha;

        beforeEach(async () => {
            vi.resetModules();
            ({ renderSmartCaptcha } = await import('./SmartCaptcha'));

            document.head
                .querySelectorAll(`script[src="${smartCaptchaScriptURL}"]`)
                .forEach((script) => script.remove());
            delete window.smartCaptcha;
        });

        it('должен отрисовать капчу сразу, если скрипт уже загружен', async () => {
            const render = vi.fn();
            window.smartCaptcha = { render };

            renderSmartCaptcha('test-captcha-container', { sitekey: 'site-key', hl: 'ru' });
            await flushPromises();

            expect(render).toHaveBeenCalledWith('test-captcha-container', { sitekey: 'site-key', hl: 'ru' });
            expect(document.head.querySelector(`script[src="${smartCaptchaScriptURL}"]`)).toBeNull();
        });

        it('должен загрузить скрипт один раз и отрисовать капчу после загрузки', async () => {
            renderSmartCaptcha('test-captcha-container', { sitekey: 'site-key', hl: 'ru' });
            renderSmartCaptcha('test-captcha-container', { sitekey: 'site-key', hl: 'ru' });

            const scripts = document.head.querySelectorAll<HTMLScriptElement>(`script[src="${smartCaptchaScriptURL}"]`);
            expect(scripts).toHaveLength(1);

            const render = vi.fn();
            window.smartCaptcha = { render };
            scripts[0].dispatchEvent(new Event('load'));
            await flushPromises();

            expect(render).toHaveBeenCalledWith('test-captcha-container', { sitekey: 'site-key', hl: 'ru' });
        });

        it('должен пропустить отрисовку, если контейнер исчез до загрузки скрипта', async () => {
            renderSmartCaptcha('missing-container', { sitekey: 'site-key', hl: 'ru' });

            const script = document.head.querySelector<HTMLScriptElement>(`script[src="${smartCaptchaScriptURL}"]`);
            expect(script).not.toBeNull();

            const render = vi.fn();
            window.smartCaptcha = { render };
            script?.dispatchEvent(new Event('load'));
            await flushPromises();

            expect(render).not.toHaveBeenCalled();
        });
    });
});
