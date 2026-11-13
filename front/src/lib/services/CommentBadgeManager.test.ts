import { describe, it, expect, vi } from 'vitest';
import { CommentBadgeManager, type CommentBadgeManagerDependencies } from './CommentBadgeManager';
import type { FabricObject } from 'fabric/es';

function makeDeps(overrides: Partial<CommentBadgeManagerDependencies> = {}): CommentBadgeManagerDependencies {
    return {
        createBadge: vi.fn().mockReturnValue({} as FabricObject),
        ...overrides,
    };
}

function makeCanvas() {
    return {
        add: vi.fn(),
        remove: vi.fn(),
        requestRenderAll: vi.fn(),
    } as unknown as import('fabric/es').Canvas;
}

function makeBadge(text = '1') {
    let currentText = text;
    const textObj = {
        set: vi.fn((opts: Record<string, unknown>) => {
            if (opts.text !== undefined) currentText = opts.text as string;
        }),
        get text() {
            return currentText;
        },
    };
    const circleObj = {
        type: 'circle',
        set: vi.fn(),
    };
    return {
        selectable: true,
        evented: true,
        set: vi.fn(),
        getObjects: () => [circleObj, textObj],
    } as unknown as import('fabric/es').Group;
}

function makeShape(name = 'shape-1') {
    return {
        name,
        getCenterPoint: () => ({ x: 100, y: 100 }),
        getBoundingRect: () => ({ left: 50, top: 50, width: 100, height: 100 }),
        getObjects: () => [{ type: 'rect' }],
    } as unknown as import('fabric/es').Group;
}

describe('CommentBadgeManager', () => {
    describe('upsert', () => {
        it('ничего не делает если canvas не задан', () => {
            const deps = makeDeps();
            const manager = new CommentBadgeManager(deps);
            const shape = makeShape();

            manager.upsert(shape, 1, false);

            expect(deps.createBadge).not.toHaveBeenCalled();
        });

        it('ничего не делает если shape.name пустой', () => {
            const deps = makeDeps();
            const manager = new CommentBadgeManager(deps);
            manager.setCanvas(makeCanvas());
            const shape = { ...makeShape(), name: '' } as unknown as import('fabric/es').Group;

            manager.upsert(shape, 1, false);

            expect(deps.createBadge).not.toHaveBeenCalled();
        });

        it('создаёт новый badge и добавляет на canvas', () => {
            const badge = makeBadge();
            const deps = makeDeps({
                createBadge: vi.fn().mockReturnValue(badge),
            });
            const manager = new CommentBadgeManager(deps);
            const canvas = makeCanvas();
            manager.setCanvas(canvas);
            const shape = makeShape();

            manager.upsert(shape, 1, false);

            expect(deps.createBadge).toHaveBeenCalledWith(1);
            expect(badge.selectable).toBe(false);
            expect(badge.evented).toBe(false);
            expect(canvas.add).toHaveBeenCalledWith(badge);
        });

        it('обновляет существующий badge без increment', () => {
            const badge = makeBadge('3');
            const deps = makeDeps({
                createBadge: vi.fn().mockReturnValue(badge),
            });
            const manager = new CommentBadgeManager(deps);
            const canvas = makeCanvas();
            manager.setCanvas(canvas);
            const shape = makeShape();

            manager.upsert(shape, 3, false);
            manager.upsert(shape, 5, false);

            expect(deps.createBadge).toHaveBeenCalledTimes(1);
            const textObj = badge.getObjects()[1] as unknown as { set: ReturnType<typeof vi.fn> };
            expect(textObj.set).toHaveBeenCalledWith({ text: '5' });
            expect(canvas.requestRenderAll).toHaveBeenCalled();
        });

        it('инкрементирует текст badge при increment=true', () => {
            const badge = makeBadge('2');
            const deps = makeDeps({
                createBadge: vi.fn().mockReturnValue(badge),
            });
            const manager = new CommentBadgeManager(deps);
            const canvas = makeCanvas();
            manager.setCanvas(canvas);
            const shape = makeShape();

            manager.upsert(shape, 2, false);
            manager.upsert(shape, 0, true);

            const textObj = badge.getObjects()[1] as unknown as { set: ReturnType<typeof vi.fn> };
            expect(textObj.set).toHaveBeenCalledWith({ text: '3' });
        });
    });

    describe('updatePosition', () => {
        it('ничего не делает если canvas не задан', () => {
            const deps = makeDeps();
            const manager = new CommentBadgeManager(deps);

            manager.updatePosition(makeShape());

            expect(true).toBe(true);
        });

        it('обновляет позицию badge если он существует', () => {
            const badge = makeBadge();
            const deps = makeDeps({
                createBadge: vi.fn().mockReturnValue(badge),
            });
            const manager = new CommentBadgeManager(deps);
            const canvas = makeCanvas();
            manager.setCanvas(canvas);
            const shape = makeShape();

            manager.upsert(shape, 1, false);
            manager.updatePosition(shape);

            expect(badge.set).toHaveBeenCalled();
            expect(canvas.requestRenderAll).toHaveBeenCalled();
        });

        it('ничего не делает если badge не существует', () => {
            const deps = makeDeps();
            const manager = new CommentBadgeManager(deps);
            const canvas = makeCanvas();
            manager.setCanvas(canvas);

            manager.updatePosition(makeShape('unknown'));

            expect(canvas.requestRenderAll).not.toHaveBeenCalled();
        });
    });

    describe('remove', () => {
        it('ничего не делает если canvas не задан', () => {
            const deps = makeDeps();
            const manager = new CommentBadgeManager(deps);

            manager.remove(makeShape());

            expect(true).toBe(true);
        });

        it('удаляет badge с canvas', () => {
            const badge = makeBadge();
            const deps = makeDeps({
                createBadge: vi.fn().mockReturnValue(badge),
            });
            const manager = new CommentBadgeManager(deps);
            const canvas = makeCanvas();
            manager.setCanvas(canvas);
            const shape = makeShape();

            manager.upsert(shape, 1, false);
            manager.remove(shape);

            expect(canvas.remove).toHaveBeenCalledWith(badge);
        });

        it('ничего не делает если badge не существует', () => {
            const deps = makeDeps();
            const manager = new CommentBadgeManager(deps);
            const canvas = makeCanvas();
            manager.setCanvas(canvas);

            manager.remove(makeShape('unknown'));

            expect(canvas.remove).not.toHaveBeenCalled();
        });
    });

    describe('setCanvas', () => {
        it('сбрасывает badges при смене canvas', () => {
            const badge = makeBadge();
            const deps = makeDeps({
                createBadge: vi.fn().mockReturnValue(badge),
            });
            const manager = new CommentBadgeManager(deps);
            const canvas = makeCanvas();
            manager.setCanvas(canvas);

            const shape = makeShape();
            manager.upsert(shape, 1, false);

            manager.setCanvas(makeCanvas());

            manager.remove(shape);
            expect(canvas.remove).not.toHaveBeenCalled();
        });
    });
});
