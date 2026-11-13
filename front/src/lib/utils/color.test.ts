import { describe, it, expect } from 'vitest';
import { darkenColor, lightenColor, resolveColor } from './color';

const HEX_RE = /^#[0-9A-F]{6}$/;

describe('darkenColor', () => {
    it('должен возвращать валидный hex цвет (7 символов, верхний регистр)', () => {
        const result = darkenColor('#FF0000');
        expect(result).toMatch(HEX_RE);
        expect(result).toHaveLength(7);
    });

    it('должен инвертировать светло-серый в тёмно-серый', () => {
        // #808080 → L=50 → 100-50=50; s=0 → без насыщенности
        expect(darkenColor('#808080')).toBe('#7F7F7F');
    });

    it('должен превращать белый (L=100) в тёмно-серый c min-яркостью 8%', () => {
        expect(darkenColor('#FFFFFF')).toBe('#141414');
    });

    it('должен превращать чёрный (L=0) в белый', () => {
        expect(darkenColor('#000000')).toBe('#FFFFFF');
    });

    it('должен гасить насыщенность на 15 пунктов', () => {
        // #00FF00: H=120, S=100 → S=85; L=50 → L=50
        expect(darkenColor('#00FF00')).toBe('#13EC13');
    });

    it('должен возвращать transparent без изменений', () => {
        expect(darkenColor('transparent')).toBe('transparent');
    });

    it('должен возвращать невалидный цвет без изменений', () => {
        expect(darkenColor('not-a-color')).toBe('not-a-color');
        expect(darkenColor('#12345')).toBe('#12345');
        // '#12gggg' — длина 6, но не шестнадцатеричный: раньше parseInt молча
        // обрезал символы и принимал его как валидный цвет
        expect(darkenColor('#12gggg')).toBe('#12gggg');
    });
});

describe('lightenColor', () => {
    it('должен возвращать валидный hex цвет', () => {
        const result = lightenColor('#FF0000');
        expect(result).toMatch(HEX_RE);
    });

    it('должен гасить насыщенность сильнее чем darkenColor', () => {
        // Насыщенность гасится на 10 пунктов (а не на 15), поэтому результат светлее
        expect(lightenColor('#00FF00')).toBe('#0DF20D');
    });

    it('должен превращать белый в тёмно-серый', () => {
        expect(lightenColor('#FFFFFF')).toBe('#141414');
    });

    it('должен превращать чёрный в белый', () => {
        expect(lightenColor('#000000')).toBe('#FFFFFF');
    });

    it('должен возвращать transparent без изменений', () => {
        expect(lightenColor('transparent')).toBe('transparent');
    });
});

describe('resolveColor', () => {
    it('в светлой теме возвращает цвет без изменений', () => {
        expect(resolveColor('#FF0000', false, 'fill')).toBe('#FF0000');
        expect(resolveColor('#00FF00', false, 'stroke')).toBe('#00FF00');
    });

    it('в тёмной теме для fill использует darkenColor', () => {
        expect(resolveColor('#000000', true, 'fill')).toBe(darkenColor('#000000'));
    });

    it('в тёмной теме для stroke использует lightenColor', () => {
        expect(resolveColor('#000000', true, 'stroke')).toBe(lightenColor('#000000'));
    });

    it('fill и stroke в тёмной теме конвертируются по-разному', () => {
        // Для насыщенного цвета разница между darken (S-15) и lighten (S-10) видна
        const fill = resolveColor('#00FF00', true, 'fill');
        const stroke = resolveColor('#00FF00', true, 'stroke');
        expect(fill).not.toBe(stroke);
        expect(fill).toMatch(HEX_RE);
        expect(stroke).toMatch(HEX_RE);
    });

    it('по умолчанию роль — fill', () => {
        expect(resolveColor('#000000', true)).toBe(darkenColor('#000000'));
    });

    it('transparent остаётся transparent в тёмной теме', () => {
        expect(resolveColor('transparent', true, 'fill')).toBe('transparent');
        expect(resolveColor('transparent', true, 'stroke')).toBe('transparent');
    });
});
