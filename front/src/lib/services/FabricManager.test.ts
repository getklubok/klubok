import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import type { ShapeElement, IEllipseElement, IImageElement, CursorData } from '../stores';
import type { OnlineUser } from '../ApiSpec';

// Мокаем fabric/es модуль
const mockGetCenterPoint = vi.fn();
const mockGetObjects = vi.fn();
const mockCalcTransformMatrix = vi.fn();
const mockItem = vi.fn();
const mockSet = vi.fn();
const mockGetBoundingRect = vi.fn();
const mockGroupAdd = vi.fn();
const mockImageFromURL = vi.fn();
// Общий mock для set() у классов-фигур — реально обновляет свойство,
// чтобы конвертацию цветов можно было проверить по obj.fill/obj.stroke.
const mockShapeSet = vi.fn(function (
    this: Record<string, unknown>,
    key: string | Record<string, unknown>,
    value?: unknown,
) {
    if (typeof key === 'string') {
        this[key] = value;
    } else if (key && typeof key === 'object') {
        Object.assign(this, key);
    }
    return this;
});

vi.mock('fabric/es', () => {
    class FabricPoint {
        x: number;
        y: number;
        constructor(x: number, y: number) {
            this.x = x;
            this.y = y;
        }
    }

    class FabricGroup {
        type = 'group';
        name: string;
        left: number;
        top: number;
        scaleX: number;
        scaleY: number;
        originX: string;
        originY: string;
        selectable: boolean;
        evented: boolean;
        dirty: boolean;
        createdObjects: unknown[];
        constructor(_objects: unknown[], options: Record<string, unknown> = {}) {
            this.createdObjects = _objects;
            this.name = (options.name as string) || '';
            this.left = (options.left as number) || 0;
            this.top = (options.top as number) || 0;
            this.scaleX = (options.scaleX as number) || 1;
            this.scaleY = (options.scaleY as number) || 1;
            this.originX = (options.originX as string) || '';
            this.originY = (options.originY as string) || '';
            this.selectable = (options.selectable as boolean) ?? true;
            this.evented = (options.evented as boolean) ?? true;
            this.dirty = true;
        }
        add = mockGroupAdd;
        canvas: { requestRenderAll: () => void } | undefined = undefined;
        getCenterPoint = mockGetCenterPoint;
        getObjects = mockGetObjects;
        calcTransformMatrix = mockCalcTransformMatrix;
        item = mockItem;
        set = mockShapeSet;
        forEachObject(cb: (obj: FabricGroup) => void) {
            this.getObjects().forEach(cb);
        }
    }

    class FabricRect {
        width: number;
        height: number;
        fill: string;
        stroke: string;
        strokeWidth: number;
        originX: string;
        originY: string;
        constructor(options: Record<string, unknown>) {
            this.width = (options.width as number) || 0;
            this.height = (options.height as number) || 0;
            this.fill = (options.fill as string) || '';
            this.stroke = (options.stroke as string) || '';
            this.strokeWidth = (options.strokeWidth as number) || 0;
            this.originX = (options.originX as string) || '';
            this.originY = (options.originY as string) || '';
        }
        static type = 'rect';
        type = 'rect';
        getBoundingRect = mockGetBoundingRect;
        set = mockShapeSet;
    }

    class FabricEllipse {
        rx: number;
        ry: number;
        fill: string;
        stroke: string;
        strokeWidth: number;
        originX: string;
        originY: string;
        width: number;
        left: number;
        top: number;
        constructor(options: Record<string, unknown>) {
            this.rx = (options.rx as number) || 0;
            this.ry = (options.ry as number) || 0;
            this.fill = (options.fill as string) || '';
            this.stroke = (options.stroke as string) || '';
            this.strokeWidth = (options.strokeWidth as number) || 0;
            this.originX = (options.originX as string) || '';
            this.originY = (options.originY as string) || '';
            this.width = (options.width as number) || 0;
            this.left = (options.left as number) || 0;
            this.top = (options.top as number) || 0;
        }
        static type = 'ellipse';
        type = 'ellipse';
        set = mockShapeSet;
    }

    class FabricPolygon {
        points: { x: number; y: number }[];
        fill: string;
        stroke: string;
        strokeWidth: number;
        originX: string;
        originY: string;
        left: number;
        top: number;
        constructor(points: { x: number; y: number }[], options: Record<string, unknown>) {
            this.points = points;
            this.fill = (options.fill as string) || '';
            this.stroke = (options.stroke as string) || '';
            this.strokeWidth = (options.strokeWidth as number) || 0;
            this.originX = (options.originX as string) || '';
            this.originY = (options.originY as string) || '';
            this.left = (options.left as number) || 0;
            this.top = (options.top as number) || 0;
        }
        static type = 'polygon';
        type = 'polygon';
        set = mockShapeSet;
    }

    class FabricCircle {
        radius: number;
        fill: string;
        stroke: string;
        strokeWidth: number;
        originX: string;
        originY: string;
        width: number;
        left: number;
        top: number;
        constructor(options: Record<string, unknown>) {
            this.radius = (options.radius as number) || 0;
            this.fill = (options.fill as string) || '';
            this.stroke = (options.stroke as string) || '';
            this.strokeWidth = (options.strokeWidth as number) || 0;
            this.originX = (options.originX as string) || '';
            this.originY = (options.originY as string) || '';
            this.width = (options.width as number) || 0;
            this.left = (options.left as number) || 0;
            this.top = (options.top as number) || 0;
        }
        static type = 'circle';
        type = 'circle';
        set = mockShapeSet;
    }

    class FabricTriangle {
        left: number;
        top: number;
        width: number;
        height: number;
        fill: string;
        angle: number;
        originX: string;
        originY: string;
        selectable: boolean;
        evented: boolean;
        constructor(options: Record<string, unknown>) {
            this.left = (options.left as number) || 0;
            this.top = (options.top as number) || 0;
            this.width = (options.width as number) || 0;
            this.height = (options.height as number) || 0;
            this.fill = (options.fill as string) || '';
            this.angle = (options.angle as number) || 0;
            this.originX = (options.originX as string) || '';
            this.originY = (options.originY as string) || '';
            this.selectable = (options.selectable as boolean) ?? true;
            this.evented = (options.evented as boolean) ?? true;
        }
        static type = 'triangle';
        type = 'triangle';
        set = mockShapeSet;
    }

    class FabricPolyline {
        stroke: string;
        strokeWidth: number;
        selectable: boolean;
        evented: boolean;
        constructor(_points: { x: number; y: number }[], options: Record<string, unknown>) {
            this.stroke = (options.stroke as string) || '';
            this.strokeWidth = (options.strokeWidth as number) || 0;
            this.selectable = (options.selectable as boolean) ?? true;
            this.evented = (options.evented as boolean) ?? true;
        }
        static type = 'polyline';
        type = 'polyline';
        set = mockShapeSet;
    }

    class FabricText {
        text: string;
        fill: string;
        fontSize: number;
        fontFamily: string;
        left: number;
        top: number;
        originX: string;
        originY: string;
        selectable: boolean;
        evented: boolean;
        set: ReturnType<typeof vi.fn>;
        constructor(text: string, options: Record<string, unknown>) {
            this.text = text;
            this.fill = (options.fill as string) || '';
            this.fontSize = (options.fontSize as number) || 16;
            this.fontFamily = (options.fontFamily as string) || '';
            this.left = (options.left as number) || 0;
            this.top = (options.top as number) || 0;
            this.originX = (options.originX as string) || '';
            this.originY = (options.originY as string) || '';
            this.selectable = (options.selectable as boolean) ?? true;
            this.evented = (options.evented as boolean) ?? true;
            this.set = mockShapeSet;
        }
    }
    class FabricImage {
        width: number;
        height: number;
        left: number;
        top: number;
        scaleX: number;
        scaleY: number;
        originX: string;
        originY: string;
        selectable: boolean;
        evented: boolean;
        constructor(options: Record<string, unknown> = {}) {
            this.width = (options.width as number) || 0;
            this.height = (options.height as number) || 0;
            this.left = (options.left as number) || 0;
            this.top = (options.top as number) || 0;
            this.scaleX = (options.scaleX as number) || 1;
            this.scaleY = (options.scaleY as number) || 1;
            this.originX = (options.originX as string) || '';
            this.originY = (options.originY as string) || '';
            this.selectable = (options.selectable as boolean) ?? true;
            this.evented = (options.evented as boolean) ?? true;
        }
        static fromURL(...args: unknown[]) {
            return mockImageFromURL(...args);
        }
        type = 'image';
        set = mockShapeSet;
    }

    class FabricIText extends FabricText {
        constructor(text: string, options: Record<string, unknown>) {
            super(text, options);
        }
    }

    const util = {
        transformPoint: vi.fn((point: FabricPoint, _matrix: unknown) => point),
        radiansToDegrees: (radians: number) => (radians * 180) / Math.PI,
    };

    return {
        Group: FabricGroup,
        Point: FabricPoint,
        Ellipse: FabricEllipse,
        Circle: FabricCircle,
        Rect: FabricRect,
        FabricImage,
        Polygon: FabricPolygon,
        Triangle: FabricTriangle,
        Polyline: FabricPolyline,
        Text: FabricText,
        FabricText,
        IText: FabricIText,
        FabricObject: Object,
        util,
    };
});

vi.mock('./avatarColor', () => ({
    default: vi.fn((name: string) => '#' + name.charCodeAt(0).toString(16).padStart(6, '0')),
}));

import { fabricManager } from './FabricManager';
import { Group as FabricGroup, Point as FabricPoint, Rect as FabricRect } from 'fabric/es';
import { themeStore } from '../stores/theme';
import { darkenColor, lightenColor } from '../utils/color';

function makeShapeElement(overrides: Partial<ShapeElement> = {}): ShapeElement {
    return {
        id: 'shape-1',
        type: 'square',
        left: 100,
        top: 100,
        width: 80,
        height: 60,
        fill: '#3498db',
        stroke: '#2980b9',
        strokeWidth: 2,
        scaleX: 1,
        scaleY: 1,
        text: 'Test',
        order: '',
        ...overrides,
    };
}

function makeEllipseElement(overrides: Partial<IEllipseElement> = {}): IEllipseElement {
    return {
        id: 'ellipse-1',
        type: 'ellipse',
        left: 200,
        top: 200,
        width: 100,
        height: 100,
        rx: 50,
        ry: 50,
        fill: '#e74c3c',
        stroke: '#c0392b',
        strokeWidth: 2,
        scaleX: 1,
        scaleY: 1,
        text: 'Ellipse',
        order: '',
        ...overrides,
    };
}

function makeMockGroup(options: { left?: number; top?: number; name?: string } = {}): FabricGroup {
    const group = new FabricGroup([], {
        left: options.left || 0,
        top: options.top || 0,
        name: options.name || '',
    });
    return group;
}

function makeCursorData(overrides: Partial<CursorData> = {}): CursorData {
    return {
        user: { id: 'user-1', userName: 'alice' } as OnlineUser,
        x: 150,
        y: 250,
        ...overrides,
    };
}

function makeImageMock(width: number, height: number) {
    return { width, height, left: 0, top: 0, scaleX: 1, scaleY: 1, set: mockShapeSet };
}

function makeRectMock() {
    return {
        type: 'rect',
        getBoundingRect: mockGetBoundingRect,
        set: mockShapeSet,
    };
}

function makeEllipseMock(overrides: Record<string, unknown> = {}) {
    return {
        type: 'ellipse',
        rx: 50,
        ry: 50,
        width: 100,
        ...overrides,
    };
}

function makePolygonMock(overrides: Record<string, unknown> = {}) {
    return {
        type: 'polygon',
        points: [
            { x: 0, y: -50 },
            { x: 50, y: 0 },
            { x: 0, y: 50 },
            { x: -50, y: 0 },
        ],
        ...overrides,
    };
}

describe('FabricManager', () => {
    afterEach(() => {
        // Возвращаем тему в исходное состояние
        themeStore.set('light');
    });

    beforeEach(() => {
        vi.clearAllMocks();

        // Настройка моков по умолчанию
        mockGetCenterPoint.mockReturnValue(new FabricPoint(100, 100));
        mockGetObjects.mockReturnValue([makeRectMock()]);
        mockCalcTransformMatrix.mockReturnValue([1, 0, 0, 1, 0, 0]);
        mockItem.mockReturnValue({ set: mockSet });
        mockGetBoundingRect.mockReturnValue({ left: 0, top: 0, width: 80, height: 60 });
    });

    describe('createShape', () => {
        it('должен создавать square с правильными параметрами', () => {
            const element = makeShapeElement({ type: 'square' });
            const group = fabricManager.createShape(element);

            expect(group).toBeInstanceOf(FabricGroup);
            expect(group.left).toBe(100);
            expect(group.top).toBe(100);
            expect(group.scaleX).toBe(1);
            expect(group.scaleY).toBe(1);
            expect(group.name).toBe('shape-1');
            expect(group.originX).toBe('center');
            expect(group.originY).toBe('center');
        });

        it('должен создавать ellipse с правильными параметрами', () => {
            const element = makeEllipseElement();
            const group = fabricManager.createShape(element);

            expect(group).toBeInstanceOf(FabricGroup);
            expect(group.left).toBe(200);
            expect(group.top).toBe(200);
            expect(group.name).toBe('ellipse-1');
        });

        it('должен создавать diamond с правильными параметрами', () => {
            const element = makeShapeElement({ type: 'diamond' });
            const group = fabricManager.createShape(element);

            expect(group).toBeInstanceOf(FabricGroup);
            expect(group.name).toBe('shape-1');
        });

        it('должен создавать text с правильными параметрами', () => {
            const element = makeShapeElement({ type: 'text' });
            const group = fabricManager.createShape(element);

            expect(group).toBeInstanceOf(FabricGroup);
            expect(group.left).toBe(100);
            expect(group.top).toBe(100);
            expect(group.name).toBe('shape-1');
        });

        it('должен создавать image без pv как placeholder без изображения', () => {
            const element = makeShapeElement({ type: 'image', img: '01IMAGE' }) as IImageElement;

            const group = fabricManager.createShape(element);

            expect(group).toBeInstanceOf(FabricGroup);
            expect(mockImageFromURL).not.toHaveBeenCalled();
            expect(mockGroupAdd).not.toHaveBeenCalled();
        });

        it('должен рисовать preview по принципу contain и центрировать его', async () => {
            mockGetCenterPoint.mockReturnValue(new FabricPoint(250, 150));
            mockImageFromURL.mockResolvedValue(makeImageMock(1600, 900));
            const element = makeShapeElement({
                type: 'image',
                width: 320,
                height: 320,
                img: '01IMAGE',
                pv: 'data:image/jpeg;base64,AAAA',
            }) as IImageElement;

            fabricManager.createShape(element);
            await Promise.resolve();

            expect(mockImageFromURL).toHaveBeenCalledWith(element.pv);
            const image = mockGroupAdd.mock.calls[0][0] as {
                scaleX: number;
                scaleY: number;
                left: number;
                top: number;
            };
            // 16:9 вписывается по ширине: 320/1600 = 0.2, высота 900*0.2 = 180 < 320.
            expect(image.scaleX).toBeCloseTo(0.2);
            expect(image.scaleY).toBeCloseTo(0.2);
            // add() сам пересчитает канвасные координаты в локальные, поэтому preview
            // передаётся в центр группы, а не в (0, 0).
            expect(image.left).toBe(250);
            expect(image.top).toBe(150);
        });

        it('должен вписывать высокое preview по высоте', async () => {
            mockImageFromURL.mockResolvedValue(makeImageMock(500, 1000));
            const element = makeShapeElement({
                type: 'image',
                width: 320,
                height: 320,
                img: '01IMAGE',
                pv: 'data:image/jpeg;base64,AAAA',
            }) as IImageElement;

            fabricManager.createShape(element);
            await Promise.resolve();

            const image = mockGroupAdd.mock.calls[0][0] as { scaleX: number; scaleY: number };
            expect(image.scaleX).toBeCloseTo(0.32);
            expect(image.scaleY).toBeCloseTo(0.32);
        });

        it('должен оставлять placeholder при ошибке загрузки preview', async () => {
            mockImageFromURL.mockRejectedValue(new Error('network'));
            const element = makeShapeElement({
                type: 'image',
                img: '01IMAGE',
                pv: 'data:image/jpeg;base64,AAAA',
            }) as IImageElement;

            fabricManager.createShape(element);
            await Promise.resolve();
            await Promise.resolve();

            expect(mockGroupAdd).not.toHaveBeenCalled();
        });

        it('должен делать рамку image прозрачной', () => {
            const element = makeShapeElement({
                type: 'image',
                img: '01IMAGE',
                fill: '#ffffff',
                stroke: '#c0392b',
            }) as IImageElement;

            const group = fabricManager.createShape(element);
            const frame = (group as unknown as { createdObjects: Array<{ fill: string; stroke: string }> })
                .createdObjects[0];

            expect(frame).toMatchObject({ fill: 'transparent', stroke: 'transparent' });
        });

        it('должен скрывать текст у image-шейпа', () => {
            const element = makeShapeElement({ type: 'image', img: '01IMAGE' }) as IImageElement;

            fabricManager.createShape(element);

            expect(mockShapeSet).toHaveBeenCalledWith({ visible: false });
        });

        it('должен выбрасывать ошибку для неизвестного типа элемента', () => {
            const element = makeShapeElement({ type: 'unknown' as ShapeElement['type'] });

            expect(() => fabricManager.createShape(element)).toThrow('Неизвестный тип элемента');
        });

        it('должен использовать масштаб элемента', () => {
            const element = makeShapeElement({ scaleX: 2, scaleY: 1.5 });
            const group = fabricManager.createShape(element);

            expect(group.scaleX).toBe(2);
            expect(group.scaleY).toBe(1.5);
        });

        it('должен устанавливать текст по умолчанию если не указан', () => {
            const element = makeShapeElement({ text: '' });
            const group = fabricManager.createShape(element);

            expect(group).toBeInstanceOf(FabricGroup);
        });

        it('должен устанавливать масштаб текста компенсирующим групповой', () => {
            const element = makeShapeElement({ scaleX: 2, scaleY: 2 });
            fabricManager.createShape(element);

            expect(mockItem).toHaveBeenCalledWith(1);
            expect(mockSet).toHaveBeenCalledWith({
                scaleX: 0.5,
                scaleY: 0.5,
            });
        });

        it('должен использовать defaultScale когда scaleX равен 0', () => {
            const element = makeShapeElement({ scaleX: 0, scaleY: 0 });
            fabricManager.createShape(element);

            expect(mockSet).toHaveBeenCalledWith({
                scaleX: 1,
                scaleY: 1,
            });
        });
    });

    describe('createConnection', () => {
        it('должен создавать соединение между двумя элементами', () => {
            const fromElement = makeMockGroup({ left: 0, top: 0 });
            const toElement = makeMockGroup({ left: 200, top: 0 });

            mockGetCenterPoint.mockReturnValueOnce(new FabricPoint(0, 0)).mockReturnValueOnce(new FabricPoint(200, 0));

            mockGetObjects.mockReturnValueOnce([makeRectMock()]).mockReturnValueOnce([makeRectMock()]);

            mockGetBoundingRect.mockReturnValue({ left: -40, top: -30, width: 80, height: 60 });

            const connection = fabricManager.createConnection(fromElement, toElement);

            expect(connection).toBeInstanceOf(FabricGroup);
            expect(connection.selectable).toBe(true);
            expect(connection.evented).toBe(true);
        });

        it('должен создавать соединение для кругов', () => {
            const fromElement = makeMockGroup({ left: 0, top: 0 });
            const toElement = makeMockGroup({ left: 200, top: 0 });

            mockGetCenterPoint.mockReturnValueOnce(new FabricPoint(0, 0)).mockReturnValueOnce(new FabricPoint(200, 0));

            mockGetObjects.mockReturnValueOnce([makeEllipseMock()]).mockReturnValueOnce([makeEllipseMock()]);

            const connection = fabricManager.createConnection(fromElement, toElement);

            expect(connection).toBeInstanceOf(FabricGroup);
        });

        it('должен создавать соединение для полигонов (ромбов)', () => {
            const fromElement = makeMockGroup({ left: 0, top: 0 });
            const toElement = makeMockGroup({ left: 200, top: 0 });

            mockGetCenterPoint.mockReturnValueOnce(new FabricPoint(0, 0)).mockReturnValueOnce(new FabricPoint(200, 0));

            mockGetObjects.mockReturnValueOnce([makePolygonMock()]).mockReturnValueOnce([makePolygonMock()]);

            const connection = fabricManager.createConnection(fromElement, toElement);

            expect(connection).toBeInstanceOf(FabricGroup);
        });

        it('должен использовать fallback если нет пересечения с полигоном', () => {
            const fromElement = makeMockGroup({ left: 0, top: 0 });
            const toElement = makeMockGroup({ left: 0, top: 0 });

            mockGetCenterPoint.mockReturnValueOnce(new FabricPoint(0, 0)).mockReturnValueOnce(new FabricPoint(0, 0));

            mockGetObjects.mockReturnValueOnce([makePolygonMock()]).mockReturnValueOnce([makePolygonMock()]);

            const connection = fabricManager.createConnection(fromElement, toElement);

            expect(connection).toBeInstanceOf(FabricGroup);
        });
    });

    describe('createRemoteCursor', () => {
        it('должен создавать курсор с правильной позицией', () => {
            const cursor = makeCursorData({ x: 150, y: 250 });
            const group = fabricManager.createRemoteCursor(cursor);

            expect(group).toBeInstanceOf(FabricGroup);
            expect(group.left).toBe(150);
            expect(group.top).toBe(250);
            expect(group.originX).toBe('left');
            expect(group.originY).toBe('top');
            expect(group.selectable).toBe(false);
            expect(group.evented).toBe(false);
        });

        it('должен использовать первую букву имени пользователя', () => {
            const cursor = makeCursorData({
                user: { id: 'user-1', userName: 'bob' } as OnlineUser,
            });

            const group = fabricManager.createRemoteCursor(cursor);

            expect(group).toBeInstanceOf(FabricGroup);
        });

        it('должен использовать "?" если имя пользователя отсутствует', () => {
            const cursor = makeCursorData({
                user: { id: 'user-1', userName: '' } as OnlineUser,
            });

            const group = fabricManager.createRemoteCursor(cursor);

            expect(group).toBeInstanceOf(FabricGroup);
        });

        it('должен работать с undefined userName', () => {
            const cursor = makeCursorData({
                user: { id: 'user-1', userName: undefined as unknown as string } as OnlineUser,
            });

            const group = fabricManager.createRemoteCursor(cursor);

            expect(group).toBeInstanceOf(FabricGroup);
        });
    });

    describe('lineIntersection', () => {
        it('должен находить пересечение двух пересекающихся отрезков через createConnection', () => {
            const fromElement = makeMockGroup({ left: 0, top: 0 });
            const toElement = makeMockGroup({ left: 100, top: 0 });

            mockGetCenterPoint.mockReturnValueOnce(new FabricPoint(0, 0)).mockReturnValueOnce(new FabricPoint(100, 0));

            mockGetObjects.mockReturnValueOnce([makeRectMock()]).mockReturnValueOnce([makeRectMock()]);

            mockGetBoundingRect.mockReturnValue({ left: -40, top: -30, width: 80, height: 60 });

            const connection = fabricManager.createConnection(fromElement, toElement);

            expect(connection).toBeInstanceOf(FabricGroup);
        });
    });

    describe('findIntersectionPointWithRect', () => {
        it('должен работать когда луч направлен вправо', () => {
            const fromElement = makeMockGroup({ left: 0, top: 0 });
            const toElement = makeMockGroup({ left: 200, top: 0 });

            mockGetCenterPoint.mockReturnValueOnce(new FabricPoint(0, 0)).mockReturnValueOnce(new FabricPoint(200, 0));

            mockGetObjects.mockReturnValueOnce([makeRectMock()]).mockReturnValueOnce([makeRectMock()]);

            mockGetBoundingRect.mockReturnValue({ left: -40, top: -30, width: 80, height: 60 });

            const connection = fabricManager.createConnection(fromElement, toElement);

            expect(connection).toBeInstanceOf(FabricGroup);
        });

        it('должен работать когда луч направлен вниз', () => {
            const fromElement = makeMockGroup({ left: 0, top: 0 });
            const toElement = makeMockGroup({ left: 0, top: 200 });

            mockGetCenterPoint.mockReturnValueOnce(new FabricPoint(0, 0)).mockReturnValueOnce(new FabricPoint(0, 200));

            mockGetObjects.mockReturnValueOnce([makeRectMock()]).mockReturnValueOnce([makeRectMock()]);

            mockGetBoundingRect.mockReturnValue({ left: -40, top: -30, width: 80, height: 60 });

            const connection = fabricManager.createConnection(fromElement, toElement);

            expect(connection).toBeInstanceOf(FabricGroup);
        });

        it('должен использовать fallback когда нет пересечения', () => {
            const fromElement = makeMockGroup({ left: 0, top: 0 });
            const toElement = makeMockGroup({ left: 0, top: 0 });

            mockGetCenterPoint.mockReturnValueOnce(new FabricPoint(0, 0)).mockReturnValueOnce(new FabricPoint(0, 0));

            mockGetObjects.mockReturnValueOnce([makeRectMock()]).mockReturnValueOnce([makeRectMock()]);

            mockGetBoundingRect.mockReturnValue({ left: -40, top: -30, width: 80, height: 60 });

            const connection = fabricManager.createConnection(fromElement, toElement);

            expect(connection).toBeInstanceOf(FabricGroup);
        });
    });

    describe('findIntersectionPointWithEllipse', () => {
        it('должен находить пересечение с границей круга', () => {
            const fromElement = makeMockGroup({ left: 0, top: 0 });
            const toElement = makeMockGroup({ left: 200, top: 0 });

            mockGetCenterPoint.mockReturnValueOnce(new FabricPoint(0, 0)).mockReturnValueOnce(new FabricPoint(200, 0));

            mockGetObjects.mockReturnValueOnce([makeEllipseMock()]).mockReturnValueOnce([makeEllipseMock()]);

            const connection = fabricManager.createConnection(fromElement, toElement);

            expect(connection).toBeInstanceOf(FabricGroup);
        });

        it('должен использовать width/2 если radius не задан', () => {
            const fromElement = makeMockGroup({ left: 0, top: 0 });
            const toElement = makeMockGroup({ left: 200, top: 0 });

            mockGetCenterPoint.mockReturnValueOnce(new FabricPoint(0, 0)).mockReturnValueOnce(new FabricPoint(200, 0));

            mockGetObjects
                .mockReturnValueOnce([makeEllipseMock({ rx: undefined, ry: undefined, width: 100 })])
                .mockReturnValueOnce([makeEllipseMock({ rx: undefined, ry: undefined, width: 100 })]);

            const connection = fabricManager.createConnection(fromElement, toElement);

            expect(connection).toBeInstanceOf(FabricGroup);
        });

        it('должен учитывать масштаб круга', () => {
            const fromElement = makeMockGroup({ left: 0, top: 0 });
            const toElement = makeMockGroup({ left: 200, top: 0 });

            Object.defineProperty(fromElement, 'scaleX', { value: 2 });
            Object.defineProperty(fromElement, 'scaleY', { value: 2 });

            mockGetCenterPoint.mockReturnValueOnce(new FabricPoint(0, 0)).mockReturnValueOnce(new FabricPoint(200, 0));

            mockGetObjects.mockReturnValueOnce([makeEllipseMock()]).mockReturnValueOnce([makeEllipseMock()]);

            const connection = fabricManager.createConnection(fromElement, toElement);

            expect(connection).toBeInstanceOf(FabricGroup);
        });
    });

    describe('createShape с edge cases', () => {
        it('должен использовать defaultScale когда scaleX/ scaleY не определены', () => {
            const element = makeShapeElement();
            delete (element as Partial<ShapeElement>).scaleX;
            delete (element as Partial<ShapeElement>).scaleY;

            fabricManager.createShape(element);

            expect(mockSet).toHaveBeenCalledWith({
                scaleX: 1,
                scaleY: 1,
            });
        });

        it('должен обрабатывать элемент с отрицательными координатами', () => {
            const element = makeShapeElement({ left: -100, top: -50 });
            const group = fabricManager.createShape(element);

            expect(group.left).toBe(-100);
            expect(group.top).toBe(-50);
        });

        it('должен обрабатывать элемент с нулевыми размерами', () => {
            const element = makeShapeElement({ width: 0, height: 0 });
            const group = fabricManager.createShape(element);

            expect(group).toBeInstanceOf(FabricGroup);
        });

        it('должен обрабатывать элемент с большой толщиной обводки', () => {
            const element = makeShapeElement({ strokeWidth: 10 });
            const group = fabricManager.createShape(element);

            expect(group).toBeInstanceOf(FabricGroup);
        });
    });

    describe('applyThemeColor', () => {
        it('в светлой теме оставляет цвета как есть и сохраняет исходники', () => {
            const obj = new FabricRect({ fill: '#3498db', stroke: '#2980b9' });

            fabricManager.applyThemeColor(obj, false);

            expect(obj.fill).toBe('#3498db');
            expect(obj.stroke).toBe('#2980b9');
            expect(obj.sourceFill).toBe('#3498db');
            expect(obj.sourceStroke).toBe('#2980b9');
        });

        it('в тёмной теме конвертирует fill через darkenColor, stroke — через lightenColor', () => {
            const obj = new FabricRect({ fill: '#3498db', stroke: '#2980b9' });

            fabricManager.applyThemeColor(obj, true);

            expect(obj.fill).toBe(darkenColor('#3498db'));
            expect(obj.stroke).toBe(lightenColor('#2980b9'));
        });

        it('не перекрашивает transparent', () => {
            const obj = new FabricRect({ fill: 'transparent', stroke: 'transparent' });

            fabricManager.applyThemeColor(obj, true);

            expect(obj.fill).toBe('transparent');
            expect(obj.stroke).toBe('transparent');
        });

        it('идемпотентен: повторный вызов с той же темой не меняет цвет', () => {
            const obj = new FabricRect({ fill: '#3498db', stroke: '#2980b9' });

            fabricManager.applyThemeColor(obj, true);
            const darkFill = obj.fill;
            const darkStroke = obj.stroke;
            fabricManager.applyThemeColor(obj, true);

            expect(obj.fill).toBe(darkFill);
            expect(obj.stroke).toBe(darkStroke);
        });

        it('обратим: после dark → light возвращает исходные светлые цвета', () => {
            const obj = new FabricRect({ fill: '#3498db', stroke: '#2980b9' });

            fabricManager.applyThemeColor(obj, true);
            fabricManager.applyThemeColor(obj, false);

            expect(obj.fill).toBe('#3498db');
            expect(obj.stroke).toBe('#2980b9');
        });

        it('в тёмной теме задаёт светлые corner-контролы и рамку выделения', () => {
            const obj = new FabricRect({ fill: '#3498db', stroke: '#2980b9' });

            fabricManager.applyThemeColor(obj, true);

            expect(obj.cornerColor).toBe('#fff');
            expect(obj.cornerStrokeColor).toBe('#fff');
            expect(obj.borderColor).toBe('#fff');
        });

        it('в светлой теме задаёт стандартные corner-контролы и рамку выделения', () => {
            const obj = new FabricRect({ fill: '#3498db', stroke: '#2980b9' });

            fabricManager.applyThemeColor(obj, false);

            expect(obj.cornerColor).toBe('#333');
            expect(obj.cornerStrokeColor).toBe('#333');
            expect(obj.borderColor).toBe('#333');
        });
    });

    describe('applySelectionControlsTheme', () => {
        it('применяет цвета контролов к самой группе и её детям', () => {
            const obj = new FabricRect({ fill: '#3498db', stroke: '#2980b9' });
            const child = new FabricRect({ fill: '#fff', stroke: '#000' });
            const group = new FabricGroup([obj, child], {});
            group.getObjects = vi.fn().mockReturnValue([obj, child]);

            fabricManager.applySelectionControlsTheme(group, true);

            expect(group.cornerColor).toBe('#fff');
            expect(group.borderColor).toBe('#fff');
            expect(obj.cornerColor).toBe('#fff');
            expect(child.cornerColor).toBe('#fff');
        });

        it('поднимает dirty на группе и детях, чтобы инвалидировать кэш отрисовки контролов', () => {
            const obj = new FabricRect({ fill: '#3498db', stroke: '#2980b9' });
            const child = new FabricRect({ fill: '#fff', stroke: '#000' });
            const group = new FabricGroup([obj, child], {});
            group.getObjects = vi.fn().mockReturnValue([obj, child]);
            group.dirty = false;

            fabricManager.applySelectionControlsTheme(group, true);

            expect(group.dirty).toBe(true);
            expect(obj.dirty).toBe(true);
            expect(child.dirty).toBe(true);
        });

        it('применяет цвета контролов к одиночному объекту', () => {
            const obj = new FabricRect({ fill: '#3498db', stroke: '#2980b9' });

            fabricManager.applySelectionControlsTheme(obj, true);

            expect(obj.cornerColor).toBe('#fff');
            expect(obj.borderColor).toBe('#fff');
            expect(obj.dirty).toBe(true);
        });
    });

    describe('createShape применяет контролы выделения к группе', () => {
        it('в тёмной теме задаёт светлые контролы самой группе', () => {
            themeStore.set('dark');

            const group = fabricManager.createShape(makeShapeElement());

            expect(group.cornerColor).toBe('#fff');
            expect(group.borderColor).toBe('#fff');
        });

        it('в светлой теме задаёт тёмные контролы самой группе', () => {
            themeStore.set('light');

            const group = fabricManager.createShape(makeShapeElement());

            expect(group.cornerColor).toBe('#333');
            expect(group.borderColor).toBe('#333');
        });
    });

    describe('createShape с учётом темы', () => {
        it('в тёмной теме конвертирует fill и stroke фигуры', () => {
            themeStore.set('dark');
            mockShapeSet.mockClear();

            fabricManager.createShape(makeShapeElement({ fill: '#ffffff', stroke: '#c0392b' }));

            // applyThemeColor вызывает set('fill', ...) и set('stroke', ...)
            const fillCall = mockShapeSet.mock.calls.find(([key]) => key === 'fill');
            const strokeCall = mockShapeSet.mock.calls.find(([key]) => key === 'stroke');
            expect(fillCall?.[1]).toBe(darkenColor('#ffffff'));
            expect(strokeCall?.[1]).toBe(lightenColor('#c0392b'));
        });

        it('в светлой теме оставляет цвета как есть', () => {
            themeStore.set('light');
            mockShapeSet.mockClear();

            fabricManager.createShape(makeShapeElement({ fill: '#ffffff', stroke: '#c0392b' }));

            const fillCall = mockShapeSet.mock.calls.find(([key]) => key === 'fill');
            const strokeCall = mockShapeSet.mock.calls.find(([key]) => key === 'stroke');
            expect(fillCall?.[1]).toBe('#ffffff');
            expect(strokeCall?.[1]).toBe('#c0392b');
        });
    });
});
