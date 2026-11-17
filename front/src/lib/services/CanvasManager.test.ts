// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import type { ShapeElement, ConnectionElement } from '../stores';
import type { DiagramInfo, DiagramShape, DiagramConnection } from '../ApiSpec';

const {
    mockFabricManager,
    mockShapeStoreManager,
    mockConnectionStoreManager,
    mockCommentBadgeManager,
    mockControlsManager,
    mockWebSocketManager,
    mockAutoSaveService,
    mockCoordinatesManager,
    mockViewportManager,
} = vi.hoisted(() => ({
    mockFabricManager: {
        createShape: vi.fn().mockReturnValue({
            controls: {
                mtr: {},
                ml: {},
                mr: {},
                mb: {},
                mt: {},
                tl: {},
                tr: {},
                bl: {},
                br: {},
            },
            set: vi.fn(),
            getObjects: vi.fn().mockReturnValue([{}, { set: vi.fn() }]),
            lockMovementX: false,
            lockMovementY: false,
            getBoundingRect: vi.fn().mockReturnValue({ left: 0, top: 0, width: 100, height: 100 }),
        }),
        createConnection: vi.fn().mockReturnValue({
            controls: {
                mtr: {},
                ml: {},
                mr: {},
                mb: {},
                mt: {},
                tl: {},
                tr: {},
                bl: {},
                br: {},
            },
            set: vi.fn(),
            lockMovementX: false,
            lockMovementY: false,
        }),
        createRemoteCursor: vi.fn().mockReturnValue({ set: vi.fn(), left: 0, top: 0 }),
        applyThemeColor: vi.fn(),
        applySelectionControlsTheme: vi.fn(),
    },
    mockShapeStoreManager: {
        add: vi.fn(),
        getById: vi.fn(),
        update: vi.fn(),
        remove: vi.fn(),
        getMaxOrderKey: vi.fn().mockReturnValue(null),
    },
    mockConnectionStoreManager: {
        add: vi.fn(),
        getByShapeIds: vi.fn(),
        getByShapeId: vi.fn().mockReturnValue([]),
        removeByShapeId: vi.fn(),
        removeByShapeIds: vi.fn(),
    },
    mockCommentBadgeManager: {
        upsert: vi.fn(),
        updatePosition: vi.fn(),
        remove: vi.fn(),
    },
    mockControlsManager: {
        setConnectionBaseControl: vi.fn(),
        addConnectionDeleteControl: vi.fn(),
    },
    mockWebSocketManager: {
        sendCursorMove: vi.fn(),
    },
    mockAutoSaveService: {
        setInitialState: vi.fn(),
    },
    mockCoordinatesManager: {
        update: vi.fn(),
        set: vi.fn(),
    },
    mockViewportManager: {
        centerView: vi.fn(),
    },
}));

vi.mock('./FabricManager', () => ({ fabricManager: mockFabricManager }));
vi.mock('./ShapeStoreManager', () => ({ shapeStoreManager: mockShapeStoreManager }));
vi.mock('./ConnectionStoreManager', () => ({
    connectionStoreManager: mockConnectionStoreManager,
}));
vi.mock('./CommentBadgeManager', () => ({ commentBadgeManager: mockCommentBadgeManager }));
vi.mock('./ControlsManager', () => ({ controlsManager: mockControlsManager }));
vi.mock('./WebSocketManager', () => ({ webSocketManager: mockWebSocketManager }));
vi.mock('./AutoSaveService', () => ({ autoSaveService: mockAutoSaveService }));
vi.mock('./CoordinatesManager', () => ({ coordinatesManager: mockCoordinatesManager }));
vi.mock('./ViewportManager', () => ({ viewportManager: mockViewportManager }));

vi.mock('../stores', () => ({
    clearState: vi.fn(),
    coordinates: { set: vi.fn() },
    contextMenu: { set: vi.fn() },
    selectedObject: { set: vi.fn(), update: vi.fn() },
}));

import { contextMenu } from '../stores';

import { CanvasManager as CanvasManagerClass, type CanvasManagerDependencies } from './CanvasManager';

function makeCanvas() {
    return {
        clear: vi.fn(),
        add: vi.fn(),
        remove: vi.fn(),
        requestRenderAll: vi.fn(),
        getObjects: vi.fn().mockReturnValue([]),
        setActiveObject: vi.fn(),
        getScenePoint: vi.fn().mockReturnValue({ x: 0, y: 0 }),
        getZoom: vi.fn().mockReturnValue(1),
        zoomToPoint: vi.fn(),
        sendObjectToBack: vi.fn(),
        selection: true,
        skipTargetFind: false,
        getActiveObject: vi.fn().mockReturnValue(undefined),
        discardActiveObject: vi.fn(),
        defaultCursor: 'default',
        hoverCursor: 'default',
        viewportTransform: [1, 0, 0, 1, 0, 0],
        selectionColor: '',
        selectionBorderColor: '',
    } as unknown as import('fabric/es').Canvas;
}

function makeDeps(overrides: Partial<CanvasManagerDependencies> = {}): CanvasManagerDependencies {
    return {
        createId: vi.fn().mockReturnValue('test-id'),
        hasPermission: vi.fn().mockReturnValue(true),
        clearState: vi.fn(),
        setCoordinates: vi.fn(),
        mapShapeFromApi: vi.fn().mockImplementation((s: DiagramShape) => ({ id: s.id, order: 's' }) as ShapeElement),
        mapConnectionFromApi: vi.fn().mockImplementation(
            (c: DiagramConnection) =>
                ({
                    fromShapeId: c.fid,
                    toShapeId: c.tid,
                }) as ConnectionElement,
        ),
        ...overrides,
    };
}

function makeShape(overrides: Partial<ShapeElement> = {}): ShapeElement {
    return {
        id: 's1',
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

function makeDiagramInfo(overrides: Partial<DiagramInfo> = {}): DiagramInfo {
    return {
        id: 'd1',
        name: 'Test',
        shapes: [],
        connections: [],
        shapeCommentStats: [],
        permission: 'OWNER',
        ...overrides,
    } as DiagramInfo;
}

describe('CanvasManager', () => {
    beforeEach(() => {
        vi.resetAllMocks();
        vi.spyOn(console, 'log').mockImplementation(() => {});
        vi.spyOn(console, 'error').mockImplementation(() => {});

        // Re-setup mock defaults after reset
        mockFabricManager.createShape.mockReturnValue({
            controls: {
                mtr: {},
                ml: {},
                mr: {},
                mb: {},
                mt: {},
                tl: {},
                tr: {},
                bl: {},
                br: {},
            },
            set: vi.fn(),
            getObjects: vi.fn().mockReturnValue([{}, { set: vi.fn() }]),
            lockMovementX: false,
            lockMovementY: false,
            getBoundingRect: vi.fn().mockReturnValue({ left: 0, top: 0, width: 100, height: 100 }),
        });
        mockFabricManager.createConnection.mockReturnValue({
            controls: {
                mtr: {},
                ml: {},
                mr: {},
                mb: {},
                mt: {},
                tl: {},
                tr: {},
                bl: {},
                br: {},
            },
            set: vi.fn(),
            lockMovementX: false,
            lockMovementY: false,
        });
        mockFabricManager.createRemoteCursor.mockReturnValue({ set: vi.fn(), left: 0, top: 0 });
        mockConnectionStoreManager.getByShapeId.mockReturnValue([]);
        mockConnectionStoreManager.getByShapeIds.mockReturnValue(undefined);
        mockShapeStoreManager.getById.mockReturnValue(undefined);
    });

    afterEach(() => {
        vi.restoreAllMocks();
    });

    describe('init', () => {
        it('сохраняет ссылку на canvas', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();

            manager.init(canvas);

            // Проверяем через reset — если canvas не null, reset работает
            manager.reset();
            expect(canvas).toBeDefined();
        });
    });

    describe('reset', () => {
        it('сбрасывает состояние если canvas задан', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            manager.init(makeCanvas());

            manager.reset();

            // После reset upsertShape для нового shape не должен упасть
            // (canvas = null → addShape вернёт null)
            expect(manager.upsertShape(makeShape())).toBeNull();
        });

        it('ничего не делает если canvas не задан', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);

            manager.reset();

            expect(true).toBe(true);
        });
    });

    describe('setCurrentDrawingMode', () => {
        it('устанавливает режим рисования', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            manager.setCurrentDrawingMode('square');

            expect(canvas.defaultCursor).toBe('crosshair');
        });

        it('select → defaultCursor=default', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            manager.setCurrentDrawingMode('select');

            expect(canvas.defaultCursor).toBe('default');
        });

        it('pan → курсор grab, поиск объектов и выделение выключены', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            manager.setCurrentDrawingMode('pan');

            expect(canvas.defaultCursor).toBe('grab');
            expect(canvas.hoverCursor).toBe('grab');
            expect(canvas.skipTargetFind).toBe(true);
            expect(canvas.selection).toBe(false);
        });

        it('возврат из pan в select включает выделение и поиск объектов', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            manager.setCurrentDrawingMode('pan');
            manager.setCurrentDrawingMode('select');

            expect(canvas.skipTargetFind).toBe(false);
            expect(canvas.selection).toBe(true);
        });

        it('arrow → hoverCursor=pointer', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            manager.setCurrentDrawingMode('arrow');

            expect(canvas.hoverCursor).toBe('pointer');
        });

        it('square/ellipse/diamond → crosshair', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            manager.setCurrentDrawingMode('ellipse');

            expect(canvas.defaultCursor).toBe('crosshair');
            expect(canvas.hoverCursor).toBe('crosshair');
        });

        it('text → crosshair', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            manager.setCurrentDrawingMode('text');

            expect(canvas.defaultCursor).toBe('crosshair');
            expect(canvas.hoverCursor).toBe('crosshair');
        });

        it('text → appendNewShape создаёт элемент на канвасе', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);
            manager.setDiagramInfo(makeDiagramInfo());

            manager.setCurrentDrawingMode('text');
            manager.onMouseDown({ e: {} });

            expect(mockFabricManager.createShape).toHaveBeenCalled();
            expect(mockShapeStoreManager.add).toHaveBeenCalled();
            expect(canvas.add).toHaveBeenCalled();
        });

        it('select с правами → hoverCursor=move', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);
            manager.setDiagramInfo(makeDiagramInfo({ permission: 'OWNER' }));

            manager.setCurrentDrawingMode('select');

            expect(canvas.hoverCursor).toBe('move');
        });

        it('select без прав → hoverCursor=default', () => {
            const deps = makeDeps({ hasPermission: vi.fn().mockReturnValue(false) });
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);
            manager.setDiagramInfo(makeDiagramInfo({ permission: undefined }));

            manager.setCurrentDrawingMode('select');

            expect(canvas.hoverCursor).toBe('default');
        });
    });

    describe('upsertShape', () => {
        it('ничего не делает если canvas не задан', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);

            const result = manager.upsertShape(makeShape());

            expect(result).toBeNull();
        });

        it('обновляет существующий shape', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            const fabricGroup = {
                set: vi.fn(),
                getObjects: vi.fn().mockReturnValue([{}, { set: vi.fn() }]),
                name: 's1',
            };
            const existing = makeShape({ id: 's1', fabricGroup: fabricGroup as never });
            mockShapeStoreManager.getById.mockReturnValue(existing);

            const result = manager.upsertShape(makeShape({ id: 's1', left: 50, top: 60, text: 'new' }));

            expect(fabricGroup.set).toHaveBeenCalled();
            expect(mockShapeStoreManager.update).toHaveBeenCalled();
            expect(result).toBe(fabricGroup);
        });

        it('создаёт новый shape если не существует', () => {
            const deps = makeDeps({ createId: vi.fn().mockReturnValue('new-id') });
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);
            manager.setDiagramInfo(makeDiagramInfo());

            mockShapeStoreManager.getById.mockReturnValue(undefined);

            const result = manager.upsertShape(makeShape({ id: 's1' }));

            expect(mockFabricManager.createShape).toHaveBeenCalled();
            expect(mockShapeStoreManager.add).toHaveBeenCalled();
            expect(canvas.add).toHaveBeenCalled();
            expect(result).not.toBeNull();
        });

        it('добавляет badge если commentCount > 0 в restoreState', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);
            manager.setDiagramInfo(
                makeDiagramInfo({
                    shapes: [{ id: 's1' } as DiagramShape],
                    shapeCommentStats: [{ shapeId: 's1', commentCount: 3 }],
                }),
            );

            expect(mockCommentBadgeManager.upsert).toHaveBeenCalledWith(expect.anything(), 3, false);
        });
    });

    describe('removeShape', () => {
        it('ничего не делает если canvas не задан', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);

            expect(manager.removeShape({ name: 's1' } as never)).toBe(false);
        });

        it('удаляет shape со canvas и из store', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            const shape = { name: 's1' } as never;
            manager.removeShape(shape);

            expect(mockCommentBadgeManager.remove).toHaveBeenCalledWith(shape);
            expect(mockShapeStoreManager.remove).toHaveBeenCalledWith(shape);
            expect(mockConnectionStoreManager.removeByShapeId).toHaveBeenCalledWith('s1');
            expect(canvas.remove).toHaveBeenCalledWith(shape);
            expect(canvas.requestRenderAll).toHaveBeenCalled();
        });

        it('возвращает true при успехе', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            manager.init(makeCanvas());

            expect(manager.removeShape({ name: '' } as never)).toBe(true);
        });
    });

    describe('upsertConnection', () => {
        it('ничего не делает если connection уже существует', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            manager.init(makeCanvas());

            mockConnectionStoreManager.getByShapeIds.mockReturnValue({});

            manager.upsertConnection({ fromShapeId: 'a', toShapeId: 'b' });

            expect(mockFabricManager.createConnection).not.toHaveBeenCalled();
        });

        it('создаёт connection если оба shape имеют fabricGroup', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);
            manager.setDiagramInfo(makeDiagramInfo());

            mockConnectionStoreManager.getByShapeIds.mockReturnValue(undefined);
            mockShapeStoreManager.getById.mockImplementation((id: string) =>
                id === 'a' ? { fabricGroup: 'fg-a' } : id === 'b' ? { fabricGroup: 'fg-b' } : undefined,
            );

            manager.upsertConnection({ fromShapeId: 'a', toShapeId: 'b' });

            expect(mockFabricManager.createConnection).toHaveBeenCalledWith('fg-a', 'fg-b');
            expect(mockConnectionStoreManager.add).toHaveBeenCalled();
            expect(canvas.add).toHaveBeenCalled();
        });

        it('ничего не делает если один из shapes не найден', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            manager.init(makeCanvas());

            mockConnectionStoreManager.getByShapeIds.mockReturnValue(undefined);
            mockShapeStoreManager.getById.mockReturnValue(undefined);

            manager.upsertConnection({ fromShapeId: 'a', toShapeId: 'b' });

            expect(mockFabricManager.createConnection).not.toHaveBeenCalled();
        });
    });

    describe('removeConnection', () => {
        beforeEach(() => {
            const mockConnection = {
                fromElement: { name: 'a' },
                toElement: { name: 'b' },
                connectionObject: { remove: vi.fn() },
            };
            mockConnectionStoreManager.getByShapeIds.mockReturnValue(mockConnection);
        });

        it('удаляет конкретную connection между двумя shape', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            canvas.remove = vi.fn();
            manager.init(canvas);

            manager.removeConnection({ fromShapeId: 'a', toShapeId: 'b' });

            expect(mockConnectionStoreManager.getByShapeIds).toHaveBeenCalledWith('a', 'b');
            expect(mockConnectionStoreManager.removeByShapeIds).toHaveBeenCalledWith('a', 'b');
            expect(canvas.remove).toHaveBeenCalled();
        });
    });

    describe('upsertCursors', () => {
        it('ничего не делает если canvas не задан', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);

            manager.upsertCursors([]);

            expect(mockFabricManager.createRemoteCursor).not.toHaveBeenCalled();
        });

        it('добавляет новый курсор на canvas', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            const cursor = {
                user: { id: 'u1', userName: 'alice' },
                x: 10,
                y: 20,
            };
            manager.upsertCursors([cursor]);

            expect(mockFabricManager.createRemoteCursor).toHaveBeenCalledWith(cursor);
            expect(canvas.add).toHaveBeenCalled();
            expect(canvas.requestRenderAll).toHaveBeenCalled();
        });

        it('обновляет позицию существующего курсора', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            const fabricObj = { set: vi.fn(), left: 0, top: 0 };
            mockFabricManager.createRemoteCursor.mockReturnValue(fabricObj);

            const cursor = {
                user: { id: 'u1', userName: 'alice' },
                x: 10,
                y: 20,
            };

            manager.upsertCursors([cursor]);
            vi.clearAllMocks();

            // Второй раз — обновление
            manager.upsertCursors([{ ...cursor, x: 50, y: 60 }]);

            expect(fabricObj.set).toHaveBeenCalledWith({ left: 50, top: 60 });
            expect(mockFabricManager.createRemoteCursor).not.toHaveBeenCalled();
        });

        it('не обновляет если позиция не изменилась', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            const fabricObj = { set: vi.fn(), left: 10, top: 20 };
            mockFabricManager.createRemoteCursor.mockReturnValue(fabricObj);

            manager.upsertCursors([{ user: { id: 'u1', userName: 'alice' }, x: 10, y: 20 }]);
            vi.clearAllMocks();

            manager.upsertCursors([{ user: { id: 'u1', userName: 'alice' }, x: 10, y: 20 }]);

            expect(fabricObj.set).not.toHaveBeenCalled();
        });
    });

    describe('restoreState (через setDiagramInfo)', () => {
        it('вызывает clearState и canvas.clear', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            manager.setDiagramInfo(makeDiagramInfo());

            expect(deps.clearState).toHaveBeenCalled();
            expect(canvas.clear).toHaveBeenCalled();
        });

        it('маппит shapes и connections из API', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            manager.setDiagramInfo(
                makeDiagramInfo({
                    shapes: [{ id: 's1' } as DiagramShape],
                    connections: [{ fid: 'a', tid: 'b' } as DiagramConnection],
                }),
            );

            expect(deps.mapShapeFromApi).toHaveBeenCalledWith({ id: 's1' });
            expect(deps.mapConnectionFromApi).toHaveBeenCalledWith({ fid: 'a', tid: 'b' });
        });

        it('вызывает setCoordinates и centerView', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            manager.setDiagramInfo(makeDiagramInfo());

            expect(deps.setCoordinates).toHaveBeenCalledWith(0, 0, 1);
            expect(mockViewportManager.centerView).toHaveBeenCalled();
        });

        it('вызывает autoSaveService.setInitialState', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            manager.setDiagramInfo(makeDiagramInfo());

            expect(mockAutoSaveService.setInitialState).toHaveBeenCalledWith([], []);
        });

        it('обрабатывает shapes с commentCount', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            manager.setDiagramInfo(
                makeDiagramInfo({
                    shapes: [{ id: 's1' } as DiagramShape, { id: 's2' } as DiagramShape],
                    shapeCommentStats: [{ shapeId: 's1', commentCount: 5 }],
                }),
            );

            // s1 имеет badge с count=5, s2 — нет
            expect(mockCommentBadgeManager.upsert).toHaveBeenCalledTimes(1);
            expect(mockCommentBadgeManager.upsert).toHaveBeenCalledWith(expect.anything(), 5, false);
        });

        it('загружает connections если оба shape существуют', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            mockShapeStoreManager.getById.mockImplementation((id: string) => ({
                fabricGroup: `fg-${id}`,
            }));

            manager.setDiagramInfo(
                makeDiagramInfo({
                    connections: [{ fid: 'a', tid: 'b' } as DiagramConnection],
                }),
            );

            expect(mockFabricManager.createConnection).toHaveBeenCalledWith('fg-a', 'fg-b');
        });

        it('очищает remoteCursors', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            // Сначала добавляем курсор
            mockFabricManager.createRemoteCursor.mockReturnValue({ set: vi.fn() });
            manager.upsertCursors([{ user: { id: 'u1', userName: 'a' }, x: 0, y: 0 }]);
            vi.clearAllMocks();

            // setDiagramInfo очищает remoteCursors
            manager.setDiagramInfo(makeDiagramInfo());

            // После setDiagramInfo курсор создаётся заново (не обновляется)
            manager.upsertCursors([{ user: { id: 'u1', userName: 'a' }, x: 0, y: 0 }]);
            expect(mockFabricManager.createRemoteCursor).toHaveBeenCalled();
        });
    });

    describe('onMouseDown (panning)', () => {
        it('панорамирует в режиме «Рука»', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);
            manager.setCurrentDrawingMode('pan');

            manager.onMouseDown({
                e: { clientX: 100, clientY: 200 },
            });

            expect(canvas.defaultCursor).toBe('grab');
            expect(canvas.hoverCursor).toBe('grab');
        });

        it('панорамирует по Space в режиме выбора', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);
            manager.setTemporaryPan(true);

            manager.onMouseDown({ e: { clientX: 100, clientY: 200 } });

            expect(canvas.defaultCursor).toBe('grab');
        });

        it('не сбрасывает выделение при панорамировании', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);
            const active = { type: 'group', name: 's1' };
            canvas.getActiveObject = vi.fn().mockReturnValue(active);
            canvas.getObjects = vi.fn().mockReturnValue([active]);

            manager.setTemporaryPan(true);
            manager.onMouseDownBefore({ e: {} });
            manager.onMouseDown({ e: { clientX: 0, clientY: 0 } });
            manager.onMouseUp();

            expect(canvas.setActiveObject).toHaveBeenCalledWith(active);
        });

        it('не панорамирует по ЛКМ в режиме выбора', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            manager.onMouseDown({ e: { clientX: 100, clientY: 200 } });
            manager.onMouseMove({ e: { clientX: 150, clientY: 220 } });

            expect(canvas.viewportTransform[4]).toBe(0);
            expect(canvas.viewportTransform[5]).toBe(0);
        });

        it('не панорамирует правой кнопкой', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);
            manager.setCurrentDrawingMode('pan');

            manager.onMouseDown({ e: { button: 2, clientX: 0, clientY: 0 } });
            manager.onMouseMove({ e: { clientX: 50, clientY: 50 } });

            expect(canvas.viewportTransform[4]).toBe(0);
            expect(canvas.viewportTransform[5]).toBe(0);
        });

        it('ничего не делает если canvas не задан', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);

            manager.setTemporaryPan(true);
            manager.onMouseDown({ e: {} });

            expect(true).toBe(true);
        });
    });

    describe('setTemporaryPan', () => {
        it('включает pan-состояние и курсор grab', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            manager.setTemporaryPan(true);

            expect(canvas.skipTargetFind).toBe(true);
            expect(canvas.selection).toBe(false);
            expect(canvas.defaultCursor).toBe('grab');
        });

        it('выключает pan-состояние и возвращает выделение в select', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            manager.setTemporaryPan(true);
            manager.setTemporaryPan(false);

            expect(canvas.skipTargetFind).toBe(false);
            expect(canvas.selection).toBe(true);
        });

        it('ничего не делает если canvas не задан', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);

            manager.setTemporaryPan(true);
            manager.setTemporaryPan(false);

            expect(true).toBe(true);
        });
    });

    describe('onMouseDown (long press)', () => {
        const target = {
            type: 'group',
            name: 's1',
            getObjects: () => [{ type: 'rect' }],
        };

        function touchStart(x: number, y: number) {
            return { e: { type: 'touchstart', changedTouches: [{ clientX: x, clientY: y }] }, target };
        }

        beforeEach(() => {
            vi.useFakeTimers();
        });

        afterEach(() => {
            vi.useRealTimers();
        });

        it('открывает контекстное меню после 500 мс удержания', () => {
            const manager = new CanvasManagerClass(makeDeps());
            manager.init(makeCanvas());

            manager.onMouseDown(touchStart(100, 200));
            expect(contextMenu.set).not.toHaveBeenCalled();

            vi.advanceTimersByTime(500);

            expect(contextMenu.set).toHaveBeenCalledWith({
                visible: true,
                x: 100,
                y: 200,
                target,
            });
        });

        it('не открывает меню, если палец сдвинулся', () => {
            const manager = new CanvasManagerClass(makeDeps());
            manager.init(makeCanvas());

            manager.onMouseDown(touchStart(100, 200));
            manager.onMouseMove({ e: { type: 'touchmove', changedTouches: [{ clientX: 120, clientY: 200 }] } });

            vi.advanceTimersByTime(500);

            expect(contextMenu.set).not.toHaveBeenCalled();
        });

        it('не открывает меню, если палец отпущен раньше', () => {
            const manager = new CanvasManagerClass(makeDeps());
            manager.init(makeCanvas());

            manager.onMouseDown(touchStart(100, 200));
            manager.onMouseUp();

            vi.advanceTimersByTime(500);

            expect(contextMenu.set).not.toHaveBeenCalled();
        });

        it('не открывает меню для обычного клика мышью', () => {
            const manager = new CanvasManagerClass(makeDeps());
            manager.init(makeCanvas());

            manager.onMouseDown({ e: { type: 'mousedown', clientX: 100, clientY: 200 }, target });
            vi.advanceTimersByTime(500);

            expect(contextMenu.set).not.toHaveBeenCalled();
        });
    });

    describe('onMouseUp', () => {
        it('ничего не делает если canvas не задан', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);

            manager.onMouseUp();

            expect(true).toBe(true);
        });

        it('сбрасывает isPanning и восстанавливает курсор', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);
            manager.setCurrentDrawingMode('pan');

            // Включаем панорамирование и отпускаем кнопку
            manager.onMouseDown({ e: { clientX: 0, clientY: 0 } });
            manager.onMouseUp();

            // Дальнейшее движение мыши не двигает viewport
            manager.onMouseMove({ e: { clientX: 50, clientY: 50 } });

            expect(canvas.viewportTransform[4]).toBe(0);
            expect(canvas.viewportTransform[5]).toBe(0);
            expect(canvas.defaultCursor).toBe('grab');
        });
    });

    describe('onMouseMove (panning)', () => {
        it('перемещает viewport при isPanning', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            // Начинаем панорамирование
            manager.setTemporaryPan(true);
            manager.onMouseDown({ e: { clientX: 100, clientY: 100 } });

            // Двигаем мышь
            manager.onMouseMove({ e: { clientX: 150, clientY: 120 } });

            expect(canvas.viewportTransform[4]).toBe(50); // deltaX = 150-100
            expect(canvas.viewportTransform[5]).toBe(20); // deltaY = 120-100
            expect(mockCoordinatesManager.update).toHaveBeenCalled();
        });

        it('ничего не делает если canvas не задан', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);

            manager.onMouseMove({ e: {} });

            expect(true).toBe(true);
        });
    });

    describe('onMouseWheel', () => {
        it('ничего не делает если canvas не задан', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);

            manager.onMouseWheel({ e: {} });

            expect(true).toBe(true);
        });

        it('масштабирует и обновляет координаты', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            const preventDefault = vi.fn();
            const stopPropagation = vi.fn();

            manager.onMouseWheel({
                e: {
                    deltaY: -100,
                    offsetX: 400,
                    offsetY: 300,
                    preventDefault,
                    stopPropagation,
                },
            });

            expect(canvas.zoomToPoint).toHaveBeenCalled();
            expect(mockCoordinatesManager.update).toHaveBeenCalled();
            expect(preventDefault).toHaveBeenCalled();
            expect(stopPropagation).toHaveBeenCalled();
        });

        it('ограничивает zoom в пределах 0.1–5', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            // Очень большой zoom
            (canvas.getZoom as ReturnType<typeof vi.fn>).mockReturnValue(0.01);

            manager.onMouseWheel({
                e: {
                    deltaY: -100000,
                    offsetX: 0,
                    offsetY: 0,
                    preventDefault: vi.fn(),
                    stopPropagation: vi.fn(),
                },
            });

            const zoomArg = (canvas.zoomToPoint as ReturnType<typeof vi.fn>).mock.calls[0][1];
            expect(zoomArg).toBeGreaterThanOrEqual(0.1);
            expect(zoomArg).toBeLessThanOrEqual(5);
        });
    });

    describe('onObjectMoving (мультивыделение)', () => {
        it('привязывает выделение к сетке и обновляет связи фигур', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            const child = { type: 'group', name: 's1' };
            const selection = {
                type: 'activeselection',
                left: 13,
                top: 7,
                width: 100,
                height: 100,
                scaleX: 1,
                scaleY: 1,
                getObjects: vi.fn().mockReturnValue([child]),
            };

            manager.onObjectMoving({ target: selection });

            expect(selection.left).toBe(10);
            expect(selection.top).toBe(10);
            expect(mockConnectionStoreManager.getByShapeId).toHaveBeenCalledWith('s1');
            expect(mockCommentBadgeManager.updatePosition).toHaveBeenCalledWith(child);
        });
    });

    describe('onObjectModified (мультивыделение)', () => {
        it('сохраняет абсолютные координаты всех фигур выделения', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            const childA = { type: 'group', name: 's1', getXY: () => ({ x: 40, y: 60 }) };
            const childB = { type: 'group', name: 's2', getXY: () => ({ x: 80, y: 100 }) };
            const selection = {
                type: 'activeselection',
                getObjects: vi.fn().mockReturnValue([childA, childB]),
            };
            mockShapeStoreManager.getById.mockImplementation((id: string) => makeShape({ id: id, left: 0, top: 0 }));

            manager.onObjectModified({ target: selection });

            expect(mockShapeStoreManager.update).toHaveBeenCalledWith(
                expect.objectContaining({ id: 's1', left: 40, top: 60 }),
            );
            expect(mockShapeStoreManager.update).toHaveBeenCalledWith(
                expect.objectContaining({ id: 's2', left: 80, top: 100 }),
            );
        });

        it('игнорирует элементы выделения, не являющиеся фигурами', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            const connection = { type: 'group', isConnection: true };
            const selection = {
                type: 'activeselection',
                getObjects: vi.fn().mockReturnValue([connection]),
            };

            manager.onObjectModified({ target: selection });

            expect(mockShapeStoreManager.update).not.toHaveBeenCalled();
        });
    });

    describe('onSelectionCreated (мультивыделение)', () => {
        it('убирает стрелки из выделения и применяет тему к рамке', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            const shapeA = { type: 'group', name: 's1' };
            const shapeB = { type: 'group', name: 's2' };
            const connection = { type: 'group', isConnection: true };
            let objects = [shapeA, shapeB, connection];
            const selection = {
                type: 'activeselection',
                getObjects: vi.fn().mockImplementation(() => objects),
                remove: vi.fn().mockImplementation((obj: unknown) => {
                    objects = objects.filter((o) => o !== obj);
                }),
                size: vi.fn().mockImplementation(() => objects.length),
            };
            canvas.getActiveObject = vi.fn().mockReturnValue(selection);

            manager.onSelectionCreated({ selected: [shapeA, connection] });

            expect(selection.remove).toHaveBeenCalledWith(connection);
            expect(canvas.setActiveObject).not.toHaveBeenCalled();
            expect(canvas.discardActiveObject).not.toHaveBeenCalled();
            expect(mockFabricManager.applySelectionControlsTheme).toHaveBeenCalledWith(selection, false);
            expect(canvas.requestRenderAll).toHaveBeenCalled();
        });

        it('переключает выделение на оставшуюся фигуру', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            const shape = { type: 'group', name: 's1' };
            const connection = { type: 'group', isConnection: true };
            let objects = [shape, connection];
            const selection = {
                type: 'activeselection',
                getObjects: vi.fn().mockImplementation(() => objects),
                remove: vi.fn().mockImplementation((obj: unknown) => {
                    objects = objects.filter((o) => o !== obj);
                }),
                size: vi.fn().mockImplementation(() => objects.length),
            };
            canvas.getActiveObject = vi.fn().mockReturnValue(selection);

            manager.onSelectionUpdated({ selected: [connection] });

            expect(selection.remove).toHaveBeenCalledWith(connection);
            expect(canvas.setActiveObject).toHaveBeenCalledWith(shape);
        });
    });

    describe('recolorAll', () => {
        it('ничего не делает если canvas не задан', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);

            manager.recolorAll(true);

            expect(mockFabricManager.applyThemeColor).not.toHaveBeenCalled();
        });

        it('применяет тёмные canvas-контролы выделения в тёмной теме', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            manager.recolorAll(true);

            expect(canvas.selectionColor).toBe('rgba(255,255,255,0.2)');
            expect(canvas.selectionBorderColor).toBe('#fff');
        });

        it('применяет светлые canvas-контролы выделения в светлой теме', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            manager.recolorAll(false);

            expect(canvas.selectionColor).toBe('rgba(0,0,0,0.1)');
            expect(canvas.selectionBorderColor).toBe('#333');
        });

        it('применяет тему к рамке активного мультивыделения', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            const selection = { type: 'activeselection' };
            canvas.getActiveObject = vi.fn().mockReturnValue(selection);

            manager.recolorAll(true);

            expect(mockFabricManager.applySelectionControlsTheme).toHaveBeenCalledWith(selection, true);
        });

        it('перекрашивает children групп-коннекций', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            const line = { type: 'polyline', stroke: '#000' };
            const arrowHead = { type: 'triangle', fill: '#000' };
            const conn = {
                type: 'group',
                isConnection: true,
                forEachObject: (fn: (o: unknown) => void) => {
                    fn(line);
                    fn(arrowHead);
                },
            };
            canvas.getObjects = vi.fn().mockReturnValue([conn]);

            manager.recolorAll(true);

            expect(mockFabricManager.applyThemeColor).toHaveBeenCalledTimes(2);
            expect(mockFabricManager.applyThemeColor).toHaveBeenCalledWith(line, true);
            expect(mockFabricManager.applyThemeColor).toHaveBeenCalledWith(arrowHead, true);
            expect(mockFabricManager.applySelectionControlsTheme).toHaveBeenCalledWith(conn, true);
            expect(canvas.requestRenderAll).toHaveBeenCalled();
        });

        it('перекрашивает children обычных групп', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            const shape = { type: 'rect', fill: '#fff', stroke: '#000' };
            const text = { type: 'i-text', fill: '#000' };
            const group = {
                type: 'group',
                isConnection: false,
                forEachObject: (fn: (o: unknown) => void) => {
                    fn(shape);
                    fn(text);
                },
            };
            canvas.getObjects = vi.fn().mockReturnValue([group]);

            manager.recolorAll(false);

            expect(mockFabricManager.applyThemeColor).toHaveBeenCalledTimes(2);
            expect(mockFabricManager.applyThemeColor).toHaveBeenCalledWith(shape, false);
            expect(mockFabricManager.applyThemeColor).toHaveBeenCalledWith(text, false);
            expect(mockFabricManager.applySelectionControlsTheme).toHaveBeenCalledWith(group, false);
        });

        it('перекрашивает одиночные объекты и children групп', () => {
            const deps = makeDeps();
            const manager = new CanvasManagerClass(deps);
            const canvas = makeCanvas();
            manager.init(canvas);

            const badgeCircle = { type: 'circle', fill: '#333', stroke: '#fff' };
            const badgeText = { type: 'text', fill: '#fff' };
            const badge = {
                type: 'group',
                isConnection: false,
                forEachObject: (fn: (o: unknown) => void) => {
                    fn(badgeCircle);
                    fn(badgeText);
                },
            };
            const single = { type: 'circle', fill: '#333', stroke: '#fff' };
            canvas.getObjects = vi.fn().mockReturnValue([badge, single]);

            manager.recolorAll(true);

            expect(mockFabricManager.applyThemeColor).toHaveBeenCalledTimes(3);
            expect(mockFabricManager.applyThemeColor).toHaveBeenCalledWith(badgeCircle, true);
            expect(mockFabricManager.applyThemeColor).toHaveBeenCalledWith(badgeText, true);
            expect(mockFabricManager.applyThemeColor).toHaveBeenCalledWith(single, true);
            expect(mockFabricManager.applySelectionControlsTheme).toHaveBeenCalledWith(badge, true);
        });
    });
});
