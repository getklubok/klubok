import { describe, it, expect } from 'vitest';
import hasPermission from './hasPermission';

describe('hasPermission', () => {
    it('должен возвращать false для null', () => {
        expect(hasPermission(null)).toBe(false);
    });

    it('должен возвращать false для undefined', () => {
        expect(hasPermission(undefined)).toBe(false);
    });

    it('должен возвращать true для OWNER', () => {
        expect(hasPermission('OWNER')).toBe(true);
    });

    it('должен возвращать true для EDIT', () => {
        expect(hasPermission('EDIT')).toBe(true);
    });

    it('должен возвращать false для VIEW', () => {
        expect(hasPermission('VIEW')).toBe(false);
    });
});
