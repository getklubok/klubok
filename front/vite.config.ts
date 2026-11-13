import { defineConfig } from 'vitest/config';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import { playwright } from '@vitest/browser-playwright';

export default defineConfig(({ mode }) => {
    if (mode === 'development') {
        const csp =
            "default-src 'self'; script-src 'self'; script-src-elem 'self' https://smartcaptcha.cloud.yandex.ru 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self' http://localhost:* ws://localhost:* data:; font-src 'self'; object-src 'none'; media-src 'none'; frame-src 'self' https://smartcaptcha.cloud.yandex.ru; child-src 'none'; form-action 'none'; worker-src 'self'; manifest-src 'self'; frame-ancestors http: https: file:";

        return {
            plugins: [svelte()],
            server: {
                headers: {
                    'Content-Security-Policy': csp,
                },
            },
            test: {
                expect: { requireAssertions: true },
                projects: [
                    {
                        extends: './vite.config.ts',
                        test: {
                            name: 'client',
                            browser: {
                                enabled: true,
                                provider: playwright(),
                                instances: [{ browser: 'chromium' }],
                            },
                            include: ['src/**/*.svelte.{test,spec}.{js,ts}'],
                            exclude: ['src/lib/server/**'],
                            setupFiles: ['./vitest-setup-client.ts'],
                        },
                    },
                    {
                        extends: './vite.config.ts',
                        test: {
                            name: 'server',
                            environment: 'node',
                            include: ['src/**/*.{test,spec}.{js,ts}'],
                            exclude: ['src/**/*.svelte.{test,spec}.{js,ts}'],
                        },
                    },
                ],
            },
        };
    }

    return {
        plugins: [svelte()],
        server: {
            fs: {
                allow: ['.'],
            },
        },
        test: {
            expect: { requireAssertions: true },
            projects: [
                {
                    extends: './vite.config.ts',
                    test: {
                        name: 'client',
                        browser: {
                            enabled: true,
                            provider: playwright(),
                            instances: [{ browser: 'chromium' }],
                        },
                        include: ['src/**/*.svelte.{test,spec}.{js,ts}'],
                        exclude: ['src/lib/server/**'],
                        setupFiles: ['./vitest-setup-client.ts'],
                    },
                },
                {
                    extends: './vite.config.ts',
                    test: {
                        name: 'server',
                        environment: 'node',
                        include: ['src/**/*.{test,spec}.{js,ts}'],
                        exclude: ['src/**/*.svelte.{test,spec}.{js,ts}'],
                    },
                },
            ],
        },
    };
});
