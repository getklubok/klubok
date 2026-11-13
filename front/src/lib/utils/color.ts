// Конвертация цветов из светлой темы в тёмную (и обратно).
// Данные фигур (fill/stroke) всегда хранятся в светлой теме; при отрисовке
// в тёмной теме цвета алгоритмически конвертируются в тёмные аналоги.

type Rgb = { r: number; g: number; b: number };

/** Разбирает #RRGGBB (или #RGB) в RGB-каналы 0..255. Для 'transparent' и невалидных — null. */
function parseHex(hex: string): Rgb | null {
    if (hex === 'transparent') return null;
    let h = hex.trim();
    if (h.startsWith('#')) h = h.slice(1);
    if (h.length === 3) {
        h = h
            .split('')
            .map((c) => c + c)
            .join('');
    }
    if (h.length !== 6) return null;
    if (!/^[0-9a-fA-F]{6}$/u.test(h)) return null;
    const n = parseInt(h, 16);
    return { r: (n >> 16) & 0xff, g: (n >> 8) & 0xff, b: n & 0xff };
}

/** #RRGGBB → [h, s, l], h∈[0,360), s,l∈[0,100]. */
function rgbToHsl({ r, g, b }: Rgb): [number, number, number] {
    const rn = r / 255;
    const gn = g / 255;
    const bn = b / 255;
    const max = Math.max(rn, gn, bn);
    const min = Math.min(rn, gn, bn);
    const l = (max + min) / 2;

    if (max === min) return [0, 0, l * 100];

    const d = max - min;
    const s = l > 0.5 ? d / (2 - max - min) : d / (max + min);
    let h: number;
    if (max === rn) {
        h = (gn - bn) / d + (gn < bn ? 6 : 0);
    } else if (max === gn) {
        h = (bn - rn) / d + 2;
    } else {
        h = (rn - gn) / d + 4;
    }
    h *= 60;

    return [h, s * 100, l * 100];
}

function hue2rgb(p: number, q: number, t: number): number {
    let c = t;
    if (c < 0) c += 1;
    if (c > 1) c -= 1;
    if (c < 1 / 6) return p + (q - p) * 6 * c;
    if (c < 1 / 2) return q;
    if (c < 2 / 3) return p + (q - p) * (2 / 3 - c) * 6;
    return p;
}

/** [h, s, l], h∈[0,360), s,l∈[0,100] → #RRGGBB. */
function hslToHex(h: number, s: number, l: number): string {
    const sn = s / 100;
    const ln = l / 100;
    const hn = (((h % 360) + 360) % 360) / 360;

    let r: number;
    let g: number;
    let b: number;
    if (sn === 0) {
        r = ln;
        g = ln;
        b = ln;
    } else {
        const q = ln < 0.5 ? ln * (1 + sn) : ln + sn - ln * sn;
        const p = 2 * ln - q;
        r = hue2rgb(p, q, hn + 1 / 3);
        g = hue2rgb(p, q, hn);
        b = hue2rgb(p, q, hn - 1 / 3);
    }

    return toHex({ r: Math.round(r * 255), g: Math.round(g * 255), b: Math.round(b * 255) });
}

function toHex({ r, g, b }: Rgb): string {
    return `#${((1 << 24) | (r << 16) | (g << 8) | b).toString(16).slice(1).toUpperCase()}`;
}

/**
 * Инвертирует яркость цвета (light → dark): L = 100 - L, насыщенность слегка
 * гасится. Возвращает исходный цвет как есть для 'transparent' и невалидных.
 *
 * @param saturateBy на сколько процентов уменьшить насыщенность
 */
function convertColor(color: string, saturateBy: number): string {
    const rgb = parseHex(color);
    if (!rgb) return color;

    const [h, s, l] = rgbToHsl(rgb);
    let newL = 100 - l;
    const newS = Math.max(0, s - saturateBy);

    // Очень светлый исходный цвет (L > 90) инвертируется в почти чёрный —
    // поднимаем до тёмно-серого, чтобы оставался различимый цвет.
    if (newL < 8) newL = 8;

    return hslToHex(h, newS, newL);
}

/** Светлая тема → тёмная для заливок (fill). */
export function darkenColor(hex: string): string {
    return convertColor(hex, 15);
}

/** Светлая тема → тёмная для обводок (stroke) и текста. */
export function lightenColor(hex: string): string {
    return convertColor(hex, 10);
}

export type ColorRole = 'fill' | 'stroke';

/**
 * Светлые исходные цвета Fabric-объекта (до конвертации под тему).
 * Хранятся на объекте при создании, чтобы recolorAll() мог обратимо
 * конвертировать цвет при смене темы туда и обратно.
 */
export interface ThemeColorSource {
    sourceFill?: string;
    sourceStroke?: string;
}

/**
 * Возвращает цвет для рендеринга с учётом темы.
 * В светлой теме — как есть; в тёмной — конвертируется в тёмный аналог.
 */
export function resolveColor(color: string, isDark: boolean, role: ColorRole = 'fill'): string {
    if (!isDark) return color;
    return role === 'stroke' ? lightenColor(color) : darkenColor(color);
}
