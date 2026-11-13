import { describe, it, expect, vi } from 'vitest';
import { CoordinatesManager, type CoordinatesManagerDependencies } from './CoordinatesManager';

function makeDeps(overrides: Partial<CoordinatesManagerDependencies> = {}): CoordinatesManagerDependencies {
    return {
        setCoordinates: vi.fn(),
        ...overrides,
    };
}

function makeCanvas(
    overrides: {
        viewportTransform?: number[] | null;
        zoom?: number;
    } = {},
) {
    const zoom = overrides.zoom ?? 1;

    return {
        setZoom: vi.fn(),
        setViewportTransform: vi.fn(),
        getZoom: vi.fn().mockReturnValue(zoom),
        viewportTransform: overrides.viewportTransform === undefined ? [1, 0, 0, 1, 0, 0] : overrides.viewportTransform,
    } as unknown as import('fabric/es').Canvas;
}

describe('CoordinatesManager', () => {
    describe('set', () => {
        it('ничего не делает если canvas не задан', () => {
            const deps = makeDeps();
            const manager = new CoordinatesManager(deps);

            manager.set(10, 20, 2);

            expect(deps.setCoordinates).not.toHaveBeenCalled();
        });

        it('вызывает setZoom с переданным zoom', () => {
            const deps = makeDeps();
            const manager = new CoordinatesManager(deps);
            const canvas = makeCanvas();
            manager.setCanvas(canvas);

            manager.set(10, 20, 2);

            expect(canvas.setZoom).toHaveBeenCalledWith(2);
        });

        it('устанавливает viewportTransform через -x, -y', () => {
            const deps = makeDeps();
            const manager = new CoordinatesManager(deps);
            const canvas = makeCanvas();
            manager.setCanvas(canvas);

            manager.set(100, 200, 1);

            expect(canvas.setViewportTransform).toHaveBeenCalledWith([1, 0, 0, 1, -100, -200]);
        });

        it('не вызывает setViewportTransform если vpt равен null', () => {
            const deps = makeDeps();
            const manager = new CoordinatesManager(deps);
            const canvas = makeCanvas({ viewportTransform: null });
            manager.setCanvas(canvas);

            manager.set(10, 20, 1);

            expect(canvas.setZoom).toHaveBeenCalledWith(1);
            expect(canvas.setViewportTransform).not.toHaveBeenCalled();
        });
    });

    describe('update', () => {
        it('ничего не делает если canvas не задан', () => {
            const deps = makeDeps();
            const manager = new CoordinatesManager(deps);

            manager.update();

            expect(deps.setCoordinates).not.toHaveBeenCalled();
        });

        it('записывает в store координаты из viewportTransform', () => {
            const deps = makeDeps();
            const manager = new CoordinatesManager(deps);
            const canvas = makeCanvas({
                viewportTransform: [1, 0, 0, 1, -50, -30],
                zoom: 1.5,
            });
            manager.setCanvas(canvas);

            manager.update();

            expect(deps.setCoordinates).toHaveBeenCalledWith({
                x: 50,
                y: 30,
                zoom: 1.5,
            });
        });

        it('округляет x и y до целых чисел', () => {
            const deps = makeDeps();
            const manager = new CoordinatesManager(deps);
            const canvas = makeCanvas({
                viewportTransform: [1, 0, 0, 1, -10.7, -20.3],
                zoom: 1,
            });
            manager.setCanvas(canvas);

            manager.update();

            expect(deps.setCoordinates).toHaveBeenCalledWith({
                x: 11,
                y: 20,
                zoom: 1,
            });
        });

        it('округляет zoom до двух знаков', () => {
            const deps = makeDeps();
            const manager = new CoordinatesManager(deps);
            const canvas = makeCanvas({
                viewportTransform: [1, 0, 0, 1, -1, -1],
                zoom: 1.33333,
            });
            manager.setCanvas(canvas);

            manager.update();

            expect(deps.setCoordinates).toHaveBeenCalledWith({
                x: 1,
                y: 1,
                zoom: 1.33,
            });
        });

        it('ничего не делает если vpt равен null', () => {
            const deps = makeDeps();
            const manager = new CoordinatesManager(deps);
            const canvas = makeCanvas({ viewportTransform: null });
            manager.setCanvas(canvas);

            manager.update();

            expect(deps.setCoordinates).not.toHaveBeenCalled();
        });
    });
});
