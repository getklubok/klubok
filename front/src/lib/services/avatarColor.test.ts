import { describe, it, expect } from 'vitest';
import avatarColor from './avatarColor';

describe('avatarColor', () => {
    it('должен возвращать строку начинающуюся с #', () => {
        const result = avatarColor('user');
        expect(result.startsWith('#')).toBe(true);
    });

    it('должен возвращать строку длиной 7 символов (# + 6 hex)', () => {
        const result = avatarColor('user');
        expect(result).toHaveLength(7);
    });

    it('должен возвращать валидный hex цвет', () => {
        const result = avatarColor('user');
        expect(result).toMatch(/^#[0-9A-F]{6}$/);
    });

    it('должен возвращать одинаковый цвет для одинакового имени', () => {
        const result1 = avatarColor('alice');
        const result2 = avatarColor('alice');
        expect(result1).toBe(result2);
    });

    it('должен возвращать разные цвета для разных имён', () => {
        const result1 = avatarColor('alice');
        const result2 = avatarColor('bob');
        expect(result1).not.toBe(result2);
    });

    it('должен обрабатывать пустую строку', () => {
        const result = avatarColor('');
        expect(result).toMatch(/^#[0-9A-F]{6}$/);
    });

    it('должен обрабатывать строку с пробелами', () => {
        const result = avatarColor('john doe');
        expect(result).toMatch(/^#[0-9A-F]{6}$/);
    });

    it('должен обрабатывать спецсимволы', () => {
        const result = avatarColor('user@domain.com');
        expect(result).toMatch(/^#[0-9A-F]{6}$/);
    });

    it('должен обрабатывать unicode символы', () => {
        const result = avatarColor('пользователь');
        expect(result).toMatch(/^#[0-9A-F]{6}$/);
    });

    it('должен быть чувствителен к регистру', () => {
        const result1 = avatarColor('User');
        const result2 = avatarColor('user');
        expect(result1).not.toBe(result2);
    });
});
