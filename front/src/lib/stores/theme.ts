import { writable, derived } from 'svelte/store';

export type Theme = 'system' | 'light' | 'dark';

export type ResolvedTheme = 'light' | 'dark';

const STORAGE_KEY = 'klubok-theme';

/** Считывает тему из localStorage, fallback — 'system'. */
function readStoredTheme(): Theme {
    if (typeof window === 'undefined') return 'system';
    const stored = window.localStorage.getItem(STORAGE_KEY);
    if (stored === 'light' || stored === 'dark' || stored === 'system') return stored;
    return 'system';
}

// Store выбранной темы ('system' | 'light' | 'dark')
export const themeStore = writable<Theme>(readStoredTheme());

// Запись значения в localStorage при каждом изменении
themeStore.subscribe((theme) => {
    if (typeof window === 'undefined') return;
    try {
        window.localStorage.setItem(STORAGE_KEY, theme);
        // eslint-disable-next-line @typescript-eslint/no-unused-vars
    } catch (err) {
        // localStorage может быть недоступен (private mode и т.п.) — игнорируем
    }
});

// Текущее значение prefers-color-scheme (true = dark)
// VITE_FORCE_DARK_THEME (только dev) заставляет систему считать тему тёмной,
// игнорируя prefers-color-scheme из браузера — удобно для локальной проверки тёмной темы.
const forceDarkInDev = import.meta.env.VITE_FORCE_DARK_THEME === '1';

function initialSystemPrefersDark(): boolean {
    if (forceDarkInDev) return true;

    if (typeof window !== 'undefined' && window.matchMedia) {
        return window.matchMedia('(prefers-color-scheme: dark)').matches;
    }

    return false;
}

const systemPrefersDark = writable<boolean>(initialSystemPrefersDark());

if (!forceDarkInDev && typeof window !== 'undefined' && window.matchMedia) {
    const mql = window.matchMedia('(prefers-color-scheme: dark)');
    const onChange = (e: MediaQueryListEvent) => systemPrefersDark.set(e.matches);
    mql.addEventListener('change', onChange);
}

// Разрешённая тема: учитывает prefers-color-scheme при 'system'
export const resolvedThemeStore = derived<[typeof themeStore, typeof systemPrefersDark], ResolvedTheme>(
    [themeStore, systemPrefersDark],
    ([$theme, $systemPrefersDark]) => {
        if ($theme === 'light') return 'light';
        if ($theme === 'dark') return 'dark';
        return $systemPrefersDark ? 'dark' : 'light';
    },
);
