import { describe, it, expect, vi } from 'vitest';
import { ViewportManager, type ViewportManagerDependencies } from './ViewportManager';
import { FabricObject } from 'fabric/es';

function makeDeps(overrides: Partial<ViewportManagerDependencies> = {}): ViewportManagerDependencies {
    return {
        updateCoordinates: vi.fn(),
        ...overrides,
    };
}

function makeCanvas(
    opts: {
        width?: number;
        height?: number;
        objects?: FabricObject[];
    } = {},
) {
    return {
        getWidth: vi.fn().mockReturnValue(opts.width ?? 800),
        getHeight: vi.fn().mockReturnValue(opts.height ?? 600),
        getObjects: vi.fn().mockReturnValue(opts.objects ?? []),
        setViewportTransform: vi.fn(),
        requestRenderAll: vi.fn(),
    } as unknown as import('fabric/es').Canvas;
}

function makeObject(bounds: { left: number; top: number; width: number; height: number }) {
    return {
        visible: true,
        getBoundingRect: () => bounds,
    } as unknown as FabricObject;
}

const defaultPadding = { left: 0, right: 0, top: 0, bottom: 0 };

describe('ViewportManager', () => {
    describe('fitView', () => {
        it('ничего не делает если canvas не задан', () => {
            const deps = makeDeps();
            const manager = new ViewportManager(deps);

            manager.fitView({ padding: defaultPadding, maxZoom: 5, minZoom: 0.1 });

            expect(deps.updateCoordinates).not.toHaveBeenCalled();
        });

        it('сбрасывает viewport если нет объектов', () => {
            const deps = makeDeps();
            const manager = new ViewportManager(deps);
            const canvas = makeCanvas({ objects: [] });
            manager.setCanvas(canvas);

            manager.fitView({ padding: defaultPadding, maxZoom: 5, minZoom: 0.1 });

            expect(canvas.setViewportTransform).toHaveBeenCalledWith([1, 0, 0, 1, 0, 0]);
        });

        it('сбрасывает viewport если все объекты невидимы', () => {
            const hidden = {
                visible: false,
                getBoundingRect: () => ({ left: 0, top: 0, width: 10, height: 10 }),
            } as unknown as FabricObject;
            const deps = makeDeps();
            const manager = new ViewportManager(deps);
            const canvas = makeCanvas({ objects: [hidden] });
            manager.setCanvas(canvas);

            manager.fitView({ padding: defaultPadding, maxZoom: 5, minZoom: 0.1 });

            expect(canvas.setViewportTransform).toHaveBeenCalledWith([1, 0, 0, 1, 0, 0]);
        });

        it('устанавливает viewportTransform с вычисленным scale', () => {
            const obj = makeObject({ left: 100, top: 100, width: 200, height: 150 });
            const deps = makeDeps();
            const manager = new ViewportManager(deps);
            const canvas = makeCanvas({ width: 800, height: 600, objects: [obj] });
            manager.setCanvas(canvas);

            manager.fitView({ padding: defaultPadding, maxZoom: 5, minZoom: 0.1 });

            const vpt = vi.mocked(canvas.setViewportTransform).mock.calls[0][0];
            expect(vpt[0]).toBeGreaterThan(0); // scale > 0
            expect(canvas.requestRenderAll).toHaveBeenCalled();
            expect(deps.updateCoordinates).toHaveBeenCalled();
        });

        it('ограничивает scale maxZoom', () => {
            // Маленький объект на большом canvas — scale будет большим
            const obj = makeObject({ left: 0, top: 0, width: 10, height: 10 });
            const deps = makeDeps();
            const manager = new ViewportManager(deps);
            const canvas = makeCanvas({ width: 1000, height: 1000, objects: [obj] });
            manager.setCanvas(canvas);

            manager.fitView({ padding: defaultPadding, maxZoom: 2, minZoom: 0.1 });

            const vpt = vi.mocked(canvas.setViewportTransform).mock.calls[0][0];
            expect(vpt[0]).toBeLessThanOrEqual(2);
        });

        it('ограничивает scale minZoom', () => {
            // Большой объект на маленьком canvas — scale будет маленьким
            const obj = makeObject({ left: 0, top: 0, width: 10000, height: 10000 });
            const deps = makeDeps();
            const manager = new ViewportManager(deps);
            const canvas = makeCanvas({ width: 100, height: 100, objects: [obj] });
            manager.setCanvas(canvas);

            manager.fitView({ padding: defaultPadding, maxZoom: 5, minZoom: 0.5 });

            const vpt = vi.mocked(canvas.setViewportTransform).mock.calls[0][0];
            expect(vpt[0]).toBeGreaterThanOrEqual(0.5);
        });

        it('учитывает padding при расчёте scale', () => {
            const obj = makeObject({ left: 0, top: 0, width: 100, height: 100 });
            const deps = makeDeps();
            const manager = new ViewportManager(deps);
            const canvas = makeCanvas({ width: 200, height: 200, objects: [obj] });
            manager.setCanvas(canvas);

            manager.fitView({
                padding: { left: 50, right: 50, top: 50, bottom: 50 },
                maxZoom: 5,
                minZoom: 0.1,
            });

            // С padding=50+50: paddedWidth=100+100=200 → scale=200/200=1
            // Без padding: scale=200/100=2
            const vpt = vi.mocked(canvas.setViewportTransform).mock.calls[0][0];
            expect(vpt[0]).toBe(1);
        });
    });

    describe('centerView', () => {
        it('ничего не делает если canvas не задан', () => {
            const deps = makeDeps();
            const manager = new ViewportManager(deps);

            manager.centerView();

            expect(deps.updateCoordinates).not.toHaveBeenCalled();
        });

        it('сбрасывает viewport если нет объектов', () => {
            const deps = makeDeps();
            const manager = new ViewportManager(deps);
            const canvas = makeCanvas({ objects: [] });
            manager.setCanvas(canvas);

            manager.centerView();

            expect(canvas.setViewportTransform).toHaveBeenCalledWith([1, 0, 0, 1, 0, 0]);
        });

        it('центрирует объекты с zoom=1', () => {
            const obj = makeObject({ left: 100, top: 100, width: 200, height: 150 });
            const deps = makeDeps();
            const manager = new ViewportManager(deps);
            const canvas = makeCanvas({ width: 800, height: 600, objects: [obj] });
            manager.setCanvas(canvas);

            manager.centerView();

            // center объектов: x=100+100=200, y=100+75=175
            // viewport center: 400, 300
            // offset: 400-200=200, 300-175=125
            const vpt = vi.mocked(canvas.setViewportTransform).mock.calls[0][0];
            expect(vpt[0]).toBe(1); // zoom=1
            expect(vpt[3]).toBe(1);
            expect(canvas.requestRenderAll).toHaveBeenCalled();
            expect(deps.updateCoordinates).toHaveBeenCalled();
        });
    });
});
