import { vi } from 'vitest';

// Подмена localStorage для jsdom
if (typeof window !== 'undefined') {
    // Если localStorage уже существует, перезаписываем его
    Object.defineProperty(window, 'localStorage', {
        value: {
            getItem: vi.fn(),
            setItem: vi.fn(),
            removeItem: vi.fn(),
            clear: vi.fn(),
            key: vi.fn(),
            length: 0,
        },
        writable: true,
        configurable: true,
    });
}
