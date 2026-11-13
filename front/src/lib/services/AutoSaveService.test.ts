import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { AutoSaveService, type AutoSaveDependencies } from './AutoSaveService';
import type { ShapeElement, ConnectionElement, Connection } from '../stores';

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

function makeDeps(overrides: Partial<AutoSaveDependencies> = {}): AutoSaveDependencies {
    return {
        getShapes: () => [],
        getConnections: () => [],
        getPermission: () => 'OWNER',
        setSyncStatus: vi.fn(),
        updateSyncStatus: vi.fn(),
        hasPermission: () => true,
        mapShapeToApi: (s) => ({
            id: s.id,
            tp: 'S',
            l: s.left,
            t: s.top,
            w: s.width,
            h: s.height,
            sx: s.scaleX,
            sy: s.scaleY,
            fc: s.fill,
            sc: s.stroke,
            txt: s.text,
            o: s.order,
        }),
        mapConnectionToApi: (c) => ({ fid: c.fromShapeId, tid: c.toShapeId }),
        shapeUpsert: vi.fn().mockResolvedValue(undefined),
        shapeDelete: vi.fn().mockResolvedValue(undefined),
        shapeBatchUpsert: vi.fn().mockResolvedValue(undefined),
        connectionUpsert: vi.fn().mockResolvedValue(undefined),
        connectionDelete: vi.fn().mockResolvedValue(undefined),
        connectionBatchUpsert: vi.fn().mockResolvedValue(undefined),
        ...overrides,
    };
}

describe('AutoSaveService', () => {
    beforeEach(() => {
        vi.useFakeTimers();
    });

    afterEach(() => {
        vi.useRealTimers();
    });

    describe('generateOperations', () => {
        it('возвращает пустой массив если нет изменений', () => {
            const service = new AutoSaveService(makeDeps());
            const shape = makeShape();
            service.setInitialState([shape], []);

            const ops = service.generateOperations([shape], []);
            expect(ops).toEqual([]);
        });

        it('генерирует shapeUpsert для нового shape', () => {
            const service = new AutoSaveService(makeDeps());
            service.setInitialState([], []);

            const shape = makeShape({ id: 'new-shape' });
            const ops = service.generateOperations([shape], []);

            expect(ops).toEqual([{ type: 'shapeUpsert', shape }]);
        });

        it('генерирует shapeUpsert для изменённого shape', () => {
            const service = new AutoSaveService(makeDeps());
            const original = makeShape({ id: 's1', text: 'old' });
            service.setInitialState([original], []);

            const updated = makeShape({ id: 's1', text: 'new' });
            const ops = service.generateOperations([updated], []);

            expect(ops).toEqual([{ type: 'shapeUpsert', shape: updated }]);
        });

        it('не генерирует shapeUpsert для неизменённого shape', () => {
            const service = new AutoSaveService(makeDeps());
            const shape = makeShape({ id: 's1' });
            service.setInitialState([shape], []);

            const ops = service.generateOperations([shape], []);
            expect(ops).toEqual([]);
        });

        it('генерирует shapeDelete для удалённого shape', () => {
            const service = new AutoSaveService(makeDeps());
            const shape = makeShape({ id: 'deleted' });
            service.setInitialState([shape], []);

            const ops = service.generateOperations([], []);

            expect(ops).toEqual([{ type: 'shapeDelete', shapeId: 'deleted' }]);
        });

        it('генерирует connectionUpsert для новой связи', () => {
            const service = new AutoSaveService(makeDeps());
            service.setInitialState([], []);

            const conn: ConnectionElement = { fromShapeId: 'a', toShapeId: 'b' };
            const ops = service.generateOperations([], [conn]);

            expect(ops).toEqual([{ type: 'connectionUpsert', connection: conn }]);
        });

        it('не генерирует connectionUpsert для существующей связи', () => {
            const service = new AutoSaveService(makeDeps());
            const conn: ConnectionElement = { fromShapeId: 'a', toShapeId: 'b' };
            service.setInitialState([], [conn]);

            const ops = service.generateOperations([], [conn]);
            expect(ops).toEqual([]);
        });

        it('генерирует connectionDelete для удалённой связи', () => {
            const service = new AutoSaveService(makeDeps());
            const conn: ConnectionElement = { fromShapeId: 'a', toShapeId: 'b' };
            service.setInitialState([], [conn]);

            const ops = service.generateOperations([], []);

            expect(ops).toEqual([{ type: 'connectionDelete', fromShapeId: 'a', toShapeId: 'b' }]);
        });

        it('генерирует несколько операций одновременно', () => {
            const service = new AutoSaveService(makeDeps());
            const savedShape = makeShape({ id: 'old' });
            const savedConn: ConnectionElement = { fromShapeId: 'x', toShapeId: 'y' };
            service.setInitialState([savedShape], [savedConn]);

            const newShape = makeShape({ id: 'new' });
            const newConn: ConnectionElement = { fromShapeId: 'a', toShapeId: 'b' };
            const ops = service.generateOperations([newShape], [newConn]);

            expect(ops).toHaveLength(4);
            expect(ops).toContainEqual({ type: 'shapeUpsert', shape: newShape });
            expect(ops).toContainEqual({ type: 'shapeDelete', shapeId: 'old' });
            expect(ops).toContainEqual({ type: 'connectionUpsert', connection: newConn });
            expect(ops).toContainEqual({
                type: 'connectionDelete',
                fromShapeId: 'x',
                toShapeId: 'y',
            });
        });
    });

    describe('triggerAutoSave', () => {
        it('ничего не делает без diagramId', () => {
            const deps = makeDeps();
            const service = new AutoSaveService(deps);

            service.triggerAutoSave();

            expect(deps.updateSyncStatus).not.toHaveBeenCalled();
        });

        it('ничего не делает без прав', () => {
            const deps = makeDeps({ hasPermission: () => false });
            const service = new AutoSaveService(deps);
            service.setDiagramId('d1');

            service.triggerAutoSave();

            expect(deps.updateSyncStatus).not.toHaveBeenCalled();
        });

        it('устанавливает статус saving и запускает таймер', () => {
            const deps = makeDeps();
            const service = new AutoSaveService(deps);
            service.setDiagramId('d1');

            service.triggerAutoSave();

            expect(deps.updateSyncStatus).toHaveBeenCalledTimes(1);
        });

        it('сбрасывает предыдущий таймер при повторном вызове', () => {
            const deps = makeDeps({
                getShapes: () => [makeShape({ id: 's1' })],
                getConnections: () => [],
            });
            const service = new AutoSaveService(deps);
            service.setDiagramId('d1');

            service.triggerAutoSave();
            service.triggerAutoSave();

            // Таймер сброшен и установлен заново — updateSyncStatus вызывается дважды
            expect(deps.updateSyncStatus).toHaveBeenCalledTimes(2);
        });
    });

    describe('performSave (через triggerAutoSave + advanceTimers)', () => {
        it('вызывает API для новых shapes', async () => {
            const consoleLog = vi.spyOn(console, 'log').mockImplementation(() => {});
            const shape = makeShape({ id: 's1' });
            const deps = makeDeps({
                getShapes: () => [shape],
                getConnections: () => [],
            });
            const service = new AutoSaveService(deps);
            service.setDiagramId('d1');
            service.setInitialState([], []);

            service.triggerAutoSave();
            await vi.runAllTimersAsync();

            expect(deps.shapeBatchUpsert).toHaveBeenCalledWith('d1', {
                shapes: [expect.objectContaining({ id: 's1' })],
            });
            expect(deps.setSyncStatus).toHaveBeenCalledWith('saved');
            consoleLog.mockRestore();
        });

        it('вызывает shapeDelete для удалённых shapes', async () => {
            const deps = makeDeps({
                getShapes: () => [],
                getConnections: () => [],
            });
            const service = new AutoSaveService(deps);
            service.setDiagramId('d1');
            service.setInitialState([makeShape({ id: 'deleted' })], []);

            service.triggerAutoSave();
            await vi.advanceTimersByTimeAsync(2000);

            expect(deps.shapeDelete).toHaveBeenCalledWith('d1', 'deleted');
        });

        it('вызывает connectionUpsert для новых связей', async () => {
            const consoleLog = vi.spyOn(console, 'log').mockImplementation(() => {});
            const fabricConn = {
                fromElement: { name: 'a' } as unknown as Connection['fromElement'],
                toElement: { name: 'b' } as unknown as Connection['toElement'],
                connectionObject: {} as Connection['connectionObject'],
            };
            const deps = makeDeps({
                getShapes: () => [],
                getConnections: () => [fabricConn],
            });
            const service = new AutoSaveService(deps);
            service.setDiagramId('d1');
            service.setInitialState([], []);

            service.triggerAutoSave();
            await vi.runAllTimersAsync();

            expect(deps.connectionBatchUpsert).toHaveBeenCalledWith('d1', {
                connections: [{ fid: 'a', tid: 'b' }],
            });
            consoleLog.mockRestore();
        });

        it('вызывает connectionDelete для удалённых связей', async () => {
            const deps = makeDeps({
                getShapes: () => [],
                getConnections: () => [],
            });
            const service = new AutoSaveService(deps);
            service.setDiagramId('d1');
            service.setInitialState([], [{ fromShapeId: 'x', toShapeId: 'y' }]);

            service.triggerAutoSave();
            await vi.advanceTimersByTimeAsync(2000);

            expect(deps.connectionDelete).toHaveBeenCalledWith('d1', 'x', 'y');
        });

        it('не вызывает API если нет изменений', async () => {
            const shape = makeShape({ id: 's1' });
            const deps = makeDeps({
                getShapes: () => [shape],
                getConnections: () => [],
            });
            const service = new AutoSaveService(deps);
            service.setDiagramId('d1');
            service.setInitialState([shape], []);

            service.triggerAutoSave();
            await vi.advanceTimersByTimeAsync(2000);

            expect(deps.shapeUpsert).not.toHaveBeenCalled();
            expect(deps.setSyncStatus).toHaveBeenCalledWith('saved');
        });

        it('устанавливает статус error при ошибке API', async () => {
            const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {});
            const deps = makeDeps({
                getShapes: () => [makeShape({ id: 's1' })],
                getConnections: () => [],
                shapeBatchUpsert: vi.fn().mockRejectedValue(new Error('fail')),
            });
            const service = new AutoSaveService(deps);
            service.setDiagramId('d1');
            service.setInitialState([], []);

            service.triggerAutoSave();
            await vi.runAllTimersAsync();

            expect(deps.setSyncStatus).toHaveBeenCalledWith('error');
            consoleError.mockRestore();
        });

        it('игнорирует connections без name у fromElement/toElement', async () => {
            const fabricConn = {
                fromElement: { name: '' } as unknown as Connection['fromElement'],
                toElement: { name: 'b' } as unknown as Connection['toElement'],
                connectionObject: {} as Connection['connectionObject'],
            };
            const deps = makeDeps({
                getShapes: () => [],
                getConnections: () => [fabricConn],
            });
            const service = new AutoSaveService(deps);
            service.setDiagramId('d1');
            service.setInitialState([], []);

            service.triggerAutoSave();
            await vi.advanceTimersByTimeAsync(2000);

            // Нет валидных connections → нет операций → saved без API вызовов
            expect(deps.connectionUpsert).not.toHaveBeenCalled();
            expect(deps.setSyncStatus).toHaveBeenCalledWith('saved');
        });
    });

    describe('dispose', () => {
        it('очищает таймер', () => {
            const deps = makeDeps();
            const service = new AutoSaveService(deps);
            service.setDiagramId('d1');

            service.triggerAutoSave();
            service.dispose();

            // После dispose таймер не должен сработать
            vi.advanceTimersByTime(2000);
            // setSyncStatus был вызван только в triggerAutoSave (updateSyncStatus), не в performSave
            expect(deps.setSyncStatus).not.toHaveBeenCalled();
        });
    });

    describe('setDiagramId / setInitialState', () => {
        it('setDiagramId позволяет вызывать triggerAutoSave', () => {
            const deps = makeDeps();
            const service = new AutoSaveService(deps);

            // Без diagramId
            service.triggerAutoSave();
            expect(deps.updateSyncStatus).not.toHaveBeenCalled();

            // После setDiagramId
            service.setDiagramId('d1');
            service.triggerAutoSave();
            expect(deps.updateSyncStatus).toHaveBeenCalled();
        });

        it('setInitialState задаёт базу для diff', () => {
            const service = new AutoSaveService(makeDeps());
            const shape = makeShape({ id: 's1' });
            service.setInitialState([shape], []);

            const ops = service.generateOperations([shape], []);
            expect(ops).toEqual([]);
        });
    });
});
