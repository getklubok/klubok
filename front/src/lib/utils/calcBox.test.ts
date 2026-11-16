import { describe, it, expect } from 'vitest';
import { calcBoxLine, calcBoxText } from './calcBox';

describe('calcBoxText', () => {
    it.each<{ textHeight: number; expectedTop: number }>([
        { textHeight: 20, expectedTop: -30 },
        { textHeight: 40, expectedTop: -20 },
    ])('оставляет отступ 10 сверху при высоте текста $textHeight', ({ textHeight, expectedTop }) => {
        // originY: 'center', поэтому верх текста = top - textHeight / 2.
        expect(calcBoxText(100, textHeight) - textHeight / 2).toBe(-40);
        expect(calcBoxText(100, textHeight)).toBe(expectedTop);
    });
});

describe('calcBoxLine', () => {
    it.each<{ text: string; expectedY: number }>([
        { text: 'одна строка', expectedY: -10 },
        { text: 'первая\nвторая', expectedY: 10 },
    ])('оставляет отступ 10 снизу от текста для "$text"', ({ text, expectedY }) => {
        const [left, right] = calcBoxLine(100, 100, text);

        expect(left).toEqual({ x: -50, y: expectedY });
        expect(right).toEqual({ x: 50, y: expectedY });

        const lines = text.split('\n').length;
        const textBottom = calcBoxText(100, lines * 20) + (lines * 20) / 2;
        expect(expectedY - textBottom).toBe(10);
    });
});
