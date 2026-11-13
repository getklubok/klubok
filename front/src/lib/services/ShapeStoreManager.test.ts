import { describe, it, expect, vi } from 'vitest';
import { ShapeStoreManager, type ShapeStoreManagerDependencies } from './ShapeStoreManager';
import type { ShapeElement } from '../stores';

function makeShape(overrides: Partial<ShapeElement> = {}): ShapeElement {
    return {
        id: 'shape-1',
        type: 'square',
        left: 0,
        top: 0,
        width: 100,
        height: 100,
        fill: '#fff',
        stroke: '#000',
        strokeWidth: 1,
        scaleX: 1,
        scaleY: 1,
        text: '',
        order: '',
        ...overrides,
    };
}

function makeDeps(overrides: Partial<ShapeStoreManagerDependencies> = {}): ShapeStoreManagerDependencies {
    return {
        updateShapes: vi.fn(),
        getShapes: () => [],
        ...overrides,
    };
}

describe('ShapeStoreManager', () => {
    describe('add', () => {
        it('добавляет элемент в store', () => {
            const deps = makeDeps();
            const manager = new ShapeStoreManager(deps);
            const shape = makeShape({ id: 's1' });

            manager.add(shape);

            expect(deps.updateShapes).toHaveBeenCalledTimes(1);
            const updater = vi.mocked(deps.updateShapes).mock.calls[0][0];
            expect(updater([])).toEqual([shape]);
        });

        it('добавляет к существующим элементам', () => {
            const deps = makeDeps();
            const manager = new ShapeStoreManager(deps);
            const existing = makeShape({ id: 'old' });
            const newShape = makeShape({ id: 'new' });

            manager.add(newShape);

            const updater = vi.mocked(deps.updateShapes).mock.calls[0][0];
            expect(updater([existing])).toEqual([existing, newShape]);
        });
    });

    describe('getById', () => {
        it('возвращает элемент по id', () => {
            const shape = makeShape({ id: 's1' });
            const deps = makeDeps({ getShapes: () => [shape] });
            const manager = new ShapeStoreManager(deps);

            expect(manager.getById('s1')).toBe(shape);
        });

        it('возвращает undefined если элемент не найден', () => {
            const deps = makeDeps({ getShapes: () => [] });
            const manager = new ShapeStoreManager(deps);

            expect(manager.getById('missing')).toBeUndefined();
        });

        it('возвращает undefined для пустого store', () => {
            const deps = makeDeps();
            const manager = new ShapeStoreManager(deps);

            expect(manager.getById('any')).toBeUndefined();
        });
    });

    describe('update', () => {
        it('обновляет существующий элемент', () => {
            const original = makeShape({ id: 's1', text: 'old' });
            const deps = makeDeps({ getShapes: () => [original] });
            const manager = new ShapeStoreManager(deps);

            const updated = makeShape({ id: 's1', text: 'new' });
            manager.update(updated);

            expect(deps.updateShapes).toHaveBeenCalledTimes(1);
            const updater = vi.mocked(deps.updateShapes).mock.calls[0][0];
            expect(updater([original])).toEqual([updated]);
        });

        it('ничего не делает если элемент не найден', () => {
            const deps = makeDeps({ getShapes: () => [] });
            const manager = new ShapeStoreManager(deps);

            const shape = makeShape({ id: 'missing' });
            manager.update(shape);

            expect(deps.updateShapes).not.toHaveBeenCalled();
        });

        it('не изменяет другие элементы', () => {
            const s1 = makeShape({ id: 's1' });
            const s2 = makeShape({ id: 's2' });
            const deps = makeDeps({ getShapes: () => [s1, s2] });
            const manager = new ShapeStoreManager(deps);

            const updated = makeShape({ id: 's1', text: 'changed' });
            manager.update(updated);

            const updater = vi.mocked(deps.updateShapes).mock.calls[0][0];
            const result = updater([s1, s2]);
            expect(result[0]).toEqual(updated);
            expect(result[1]).toBe(s2);
        });
    });

    describe('remove', () => {
        it('удаляет элемент по имени fabricGroup', () => {
            const shape = makeShape({
                id: 's1',
                fabricGroup: { name: 'grp-1' } as ShapeElement['fabricGroup'],
            });
            const deps = makeDeps();
            const manager = new ShapeStoreManager(deps);

            manager.remove({ name: 'grp-1' } as import('fabric/es').Group);

            const updater = vi.mocked(deps.updateShapes).mock.calls[0][0];
            expect(updater([shape])).toEqual([]);
        });

        it('не удаляет элементы с другим именем', () => {
            const shape = makeShape({
                id: 's1',
                fabricGroup: { name: 'grp-1' } as import('fabric/es').Group,
            });
            const deps = makeDeps();
            const manager = new ShapeStoreManager(deps);

            manager.remove({ name: 'grp-other' } as import('fabric/es').Group);

            const updater = vi.mocked(deps.updateShapes).mock.calls[0][0];
            expect(updater([shape])).toEqual([shape]);
        });
    });
});
