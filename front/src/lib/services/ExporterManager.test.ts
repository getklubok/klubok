// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { ExporterManager } from './ExporterManager';
import type { FabricObject } from 'fabric/es';
import type { ShapeElement } from '../stores';

const {
    mockGetBoundingBox,
    mockShapesSubscribe,
    mockConnectionsSubscribe,
    mockDiagramTitleSubscribe,
    mockCanvasManagerRecolorAll,
    mockCanvasManagerIsDarkTheme,
} = vi.hoisted(() => ({
    mockGetBoundingBox: vi.fn(),
    mockShapesSubscribe: vi.fn(),
    mockConnectionsSubscribe: vi.fn(),
    mockDiagramTitleSubscribe: vi.fn(),
    mockCanvasManagerRecolorAll: vi.fn(),
    mockCanvasManagerIsDarkTheme: vi.fn(),
}));

vi.mock('../utils/getBoundingBox', () => ({
    getBoundingBox: mockGetBoundingBox,
}));

vi.mock('../stores', () => ({
    shapes: { subscribe: mockShapesSubscribe },
    connections: { subscribe: mockConnectionsSubscribe },
    diagramTitle: { subscribe: mockDiagramTitleSubscribe },
}));

// Мокаем canvasManager, чтобы контролировать тему и перекраску без тяжёлой
// цепочки зависимостей (WebSocket, AutoSave, Api).
vi.mock('./CanvasManager', () => ({
    canvasManager: {
        get isDarkTheme() {
            return mockCanvasManagerIsDarkTheme();
        },
        recolorAll: mockCanvasManagerRecolorAll,
    },
}));

function makeCanvas() {
    return {
        getObjects: vi.fn().mockReturnValue([]),
        requestRenderAll: vi.fn(),
        width: 800,
        height: 600,
        viewportTransform: [1, 0, 0, 1, 0, 0] as number[],
        backgroundColor: '#f8f9fa',
        toDataURL: vi.fn().mockReturnValue('data:image/png;base64,fake'),
    } as unknown as import('fabric/es').Canvas;
}

function makeFabricObject(overrides: Partial<FabricObject> = {}): FabricObject {
    return {
        excludeFromExport: false,
        visible: true,
        getBoundingRect: vi.fn().mockReturnValue({ left: 0, top: 0, width: 100, height: 100 }),
        ...overrides,
    } as unknown as FabricObject;
}

describe('ExporterManager', () => {
    let manager: ExporterManager;
    let canvas: ReturnType<typeof makeCanvas>;

    beforeEach(() => {
        vi.useFakeTimers();
        vi.setSystemTime(new Date(2026, 4, 8, 14, 30));
        vi.clearAllMocks();
        mockCanvasManagerIsDarkTheme.mockReturnValue(false);
        canvas = makeCanvas();
        manager = new ExporterManager();
        mockGetBoundingBox.mockReturnValue({ left: 0, top: 0, width: 100, height: 100 });
    });

    afterEach(() => {
        vi.useRealTimers();
    });

    describe('setCanvas', () => {
        it('сохраняет ссылку на canvas', () => {
            manager.setCanvas(canvas);
            canvas.getObjects = vi.fn().mockReturnValue([]);
            manager.toPNG();
            expect(canvas.getObjects).toHaveBeenCalled();
        });
    });

    describe('toPNG', () => {
        it('ничего не делает если canvas не задан', () => {
            manager.toPNG();
            expect(canvas.getObjects).not.toHaveBeenCalled();
        });

        it('исключает объекты с excludeFromExport из экспорта', () => {
            manager.setCanvas(canvas);
            const obj1 = makeFabricObject({ excludeFromExport: true, visible: true });
            const obj2 = makeFabricObject({ excludeFromExport: false, visible: true });
            canvas.getObjects = vi.fn().mockReturnValue([obj1, obj2]);

            const mockLink = {
                download: '',
                href: '',
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                click: vi.fn().mockImplementation(function (this: any) {
                    // Предотвращаем навигацию
                    if (this.href && !this.href.startsWith('data:')) {
                        // eslint-disable-next-line no-useless-return
                        return;
                    }
                }),
            } as unknown as HTMLAnchorElement;

            const clickSpy = vi.spyOn(document, 'createElement').mockImplementation((tagName) => {
                if (tagName === 'a') {
                    return mockLink;
                }
                return document.createElement(tagName);
            });

            manager.toPNG();

            expect(mockGetBoundingBox).toHaveBeenCalledWith([obj2]);

            expect(mockLink.download).toBe('klubok_2026-05-08-1430.png');
            expect(mockLink.href).toBe('data:image/png;base64,fake');
            expect(mockLink.click).toHaveBeenCalled();

            clickSpy.mockRestore();
        });

        it('ничего не делает если нет видимых объектов', () => {
            manager.setCanvas(canvas);
            const obj = makeFabricObject({ visible: false });
            canvas.getObjects = vi.fn().mockReturnValue([obj]);

            manager.toPNG();

            expect(canvas.toDataURL).not.toHaveBeenCalled();
        });

        it('восстанавливает видимость после экспорта', () => {
            manager.setCanvas(canvas);
            const obj = makeFabricObject({ excludeFromExport: true, visible: true });
            canvas.getObjects = vi.fn().mockReturnValue([obj]);

            manager.toPNG();

            expect(obj.visible).toBe(true);
        });

        it('вычисляет bounding box видимых объектов', () => {
            manager.setCanvas(canvas);
            const obj = makeFabricObject({ visible: true });
            canvas.getObjects = vi.fn().mockReturnValue([obj]);
            mockGetBoundingBox.mockReturnValue({ left: 50, top: 30, width: 200, height: 150 });

            const mockLink = {
                download: '',
                href: '',
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                click: vi.fn().mockImplementation(function (this: any) {
                    // Предотвращаем навигацию
                    if (this.href && !this.href.startsWith('data:')) {
                        // eslint-disable-next-line no-useless-return
                        return;
                    }
                }),
            } as unknown as HTMLAnchorElement;

            const clickSpy = vi.spyOn(document, 'createElement').mockImplementation((tagName) => {
                if (tagName === 'a') {
                    return mockLink;
                }
                return document.createElement(tagName);
            });

            manager.toPNG();

            expect(mockGetBoundingBox).toHaveBeenCalledWith([obj]);

            expect(mockLink.download).toBe('klubok_2026-05-08-1430.png');
            expect(mockLink.href).toBe('data:image/png;base64,fake');
            expect(mockLink.click).toHaveBeenCalled();

            clickSpy.mockRestore();
        });

        it('восстанавливает размеры canvas после экспорта', () => {
            manager.setCanvas(canvas);
            const obj = makeFabricObject({ visible: true });
            canvas.getObjects = vi.fn().mockReturnValue([obj]);
            mockGetBoundingBox.mockReturnValue({ left: 50, top: 30, width: 200, height: 150 });

            const originalWidth = canvas.width;
            const originalHeight = canvas.height;
            const originalViewportTransform = [...canvas.viewportTransform];

            const mockLink = {
                download: '',
                href: '',
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                click: vi.fn().mockImplementation(function (this: any) {
                    // Предотвращаем навигацию
                    if (this.href && !this.href.startsWith('data:')) {
                        // eslint-disable-next-line no-useless-return
                        return;
                    }
                }),
            } as unknown as HTMLAnchorElement;

            const clickSpy = vi.spyOn(document, 'createElement').mockImplementation((tagName) => {
                if (tagName === 'a') {
                    return mockLink;
                }
                return document.createElement(tagName);
            });

            manager.toPNG();

            expect(canvas.width).toBe(originalWidth);
            expect(canvas.height).toBe(originalHeight);
            expect(canvas.viewportTransform).toEqual(originalViewportTransform);

            expect(mockLink.download).toBe('klubok_2026-05-08-1430.png');
            expect(mockLink.href).toBe('data:image/png;base64,fake');
            expect(mockLink.click).toHaveBeenCalled();

            clickSpy.mockRestore();
        });

        it('создаёт ссылку для скачивания', () => {
            manager.setCanvas(canvas);
            const obj = makeFabricObject({ visible: true });
            canvas.getObjects = vi.fn().mockReturnValue([obj]);

            const clickSpy = vi.spyOn(document, 'createElement').mockImplementation(() => {
                return {
                    download: '',
                    href: '',
                    click: vi.fn(),
                } as unknown as HTMLAnchorElement;
            });

            manager.toPNG();

            const link = (document.createElement as ReturnType<typeof vi.fn>).mock.results[0].value;
            expect(link.download).toBe('klubok_2026-05-08-1430.png');
            expect(link.href).toBe('data:image/png;base64,fake');
            expect(link.click).toHaveBeenCalled();

            clickSpy.mockRestore();
        });

        it('вызывает requestRenderAll трижды', () => {
            manager.setCanvas(canvas);
            const obj = makeFabricObject({ visible: true });
            canvas.getObjects = vi.fn().mockReturnValue([obj]);

            const mockLink = {
                download: '',
                href: '',
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                click: vi.fn().mockImplementation(function (this: any) {
                    // Предотвращаем навигацию
                    if (this.href && !this.href.startsWith('data:')) {
                        // eslint-disable-next-line no-useless-return
                        return;
                    }
                }),
            } as unknown as HTMLAnchorElement;

            const clickSpy = vi.spyOn(document, 'createElement').mockImplementation((tagName) => {
                if (tagName === 'a') {
                    return mockLink;
                }
                return document.createElement(tagName);
            });

            manager.toPNG();

            expect(canvas.requestRenderAll).toHaveBeenCalledTimes(3);

            expect(mockLink.download).toBe('klubok_2026-05-08-1430.png');
            expect(mockLink.href).toBe('data:image/png;base64,fake');
            expect(mockLink.click).toHaveBeenCalled();

            clickSpy.mockRestore();
        });

        it('генерирует PNG с multiplier=2', () => {
            manager.setCanvas(canvas);
            const obj = makeFabricObject({ visible: true });
            canvas.getObjects = vi.fn().mockReturnValue([obj]);

            const mockLink = {
                download: '',
                href: '',
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                click: vi.fn().mockImplementation(function (this: any) {
                    // Предотвращаем навигацию
                    if (this.href && !this.href.startsWith('data:')) {
                        // eslint-disable-next-line no-useless-return
                        return;
                    }
                }),
            } as unknown as HTMLAnchorElement;

            const clickSpy = vi.spyOn(document, 'createElement').mockImplementation((tagName) => {
                if (tagName === 'a') {
                    return mockLink;
                }
                return document.createElement(tagName);
            });

            manager.toPNG();

            expect(canvas.toDataURL).toHaveBeenCalledWith({
                format: 'png',
                multiplier: 2,
            });

            expect(mockLink.download).toBe('klubok_2026-05-08-1430.png');
            expect(mockLink.href).toBe('data:image/png;base64,fake');
            expect(mockLink.click).toHaveBeenCalled();

            clickSpy.mockRestore();
        });

        it('экспортирует в светлой теме независимо от текущей', () => {
            manager.setCanvas(canvas);
            const obj = makeFabricObject({ visible: true });
            canvas.getObjects = vi.fn().mockReturnValue([obj]);
            // Текущая тема — тёмная
            mockCanvasManagerIsDarkTheme.mockReturnValue(true);

            // Фиксируем фон в момент вызова toDataURL (во время экспорта)
            const backgroundsAtExport: string[] = [];
            canvas.toDataURL = vi.fn().mockImplementation(() => {
                backgroundsAtExport.push(String(canvas.backgroundColor));
                return 'data:image/png;base64,fake';
            });

            const mockLink = {
                download: '',
                href: '',
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                click: vi.fn().mockImplementation(function (this: any) {
                    // Предотвращаем навигацию
                    if (this.href && !this.href.startsWith('data:')) {
                        // eslint-disable-next-line no-useless-return
                        return;
                    }
                }),
            } as unknown as HTMLAnchorElement;

            const clickSpy = vi.spyOn(document, 'createElement').mockImplementation((tagName) => {
                if (tagName === 'a') {
                    return mockLink;
                }
                return document.createElement(tagName);
            });

            manager.toPNG();

            // Перед экспортом перекрасили в светлую тему
            expect(mockCanvasManagerRecolorAll).toHaveBeenCalledWith(false);
            // Во время экспорта фон белый
            expect(backgroundsAtExport).toEqual(['#ffffff']);
            // После экспорта вернули исходную тёмную тему и фон
            expect(mockCanvasManagerRecolorAll).toHaveBeenCalledWith(true);
            expect(canvas.backgroundColor).toBe('#f8f9fa');

            clickSpy.mockRestore();
        });

        it('возвращает тёмную тему после экспорта', () => {
            manager.setCanvas(canvas);
            const obj = makeFabricObject({ visible: true });
            canvas.getObjects = vi.fn().mockReturnValue([obj]);
            mockCanvasManagerIsDarkTheme.mockReturnValue(true);

            const mockLink = {
                download: '',
                href: '',
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                click: vi.fn().mockImplementation(function (this: any) {
                    // Предотвращаем навигацию
                    if (this.href && !this.href.startsWith('data:')) {
                        // eslint-disable-next-line no-useless-return
                        return;
                    }
                }),
            } as unknown as HTMLAnchorElement;

            const clickSpy = vi.spyOn(document, 'createElement').mockImplementation((tagName) => {
                if (tagName === 'a') {
                    return mockLink;
                }
                return document.createElement(tagName);
            });

            manager.toPNG();

            expect(mockCanvasManagerRecolorAll).toHaveBeenNthCalledWith(1, false);
            expect(mockCanvasManagerRecolorAll).toHaveBeenNthCalledWith(2, true);
            expect(canvas.backgroundColor).toBe('#f8f9fa');

            clickSpy.mockRestore();
        });

        it('возвращает тему и фон если нет видимых объектов', () => {
            manager.setCanvas(canvas);
            const obj = makeFabricObject({ visible: false });
            canvas.getObjects = vi.fn().mockReturnValue([obj]);
            mockCanvasManagerIsDarkTheme.mockReturnValue(true);

            manager.toPNG();

            expect(canvas.toDataURL).not.toHaveBeenCalled();
            expect(canvas.backgroundColor).toBe('#f8f9fa');
            expect(mockCanvasManagerRecolorAll).toHaveBeenNthCalledWith(1, false);
            expect(mockCanvasManagerRecolorAll).toHaveBeenNthCalledWith(2, true);
        });

        it('не трогает тему при светлой исходной теме', () => {
            manager.setCanvas(canvas);
            const obj = makeFabricObject({ visible: true });
            canvas.getObjects = vi.fn().mockReturnValue([obj]);
            mockCanvasManagerIsDarkTheme.mockReturnValue(false);

            const mockLink = {
                download: '',
                href: '',
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                click: vi.fn().mockImplementation(function (this: any) {
                    // Предотвращаем навигацию
                    if (this.href && !this.href.startsWith('data:')) {
                        // eslint-disable-next-line no-useless-return
                        return;
                    }
                }),
            } as unknown as HTMLAnchorElement;

            const clickSpy = vi.spyOn(document, 'createElement').mockImplementation((tagName) => {
                if (tagName === 'a') {
                    return mockLink;
                }
                return document.createElement(tagName);
            });

            manager.toPNG();

            expect(mockCanvasManagerRecolorAll).toHaveBeenNthCalledWith(1, false);
            expect(mockCanvasManagerRecolorAll).toHaveBeenNthCalledWith(2, false);
            expect(canvas.backgroundColor).toBe('#f8f9fa');

            clickSpy.mockRestore();
        });
    });

    describe('toFile', () => {
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

        function makeConnection(
            overrides: Partial<{ fromElement: { name: string }; toElement: { name: string } }> = {},
        ) {
            return {
                fromElement: { name: 'shape-1' },
                toElement: { name: 'shape-2' },
                ...overrides,
            };
        }

        function setStores(shapesData: ShapeElement[] = [], connectionsData: unknown[] = [], title = '') {
            mockShapesSubscribe.mockImplementation((cb: (v: ShapeElement[]) => void) => {
                cb(shapesData);
                return () => {};
            });
            mockConnectionsSubscribe.mockImplementation((cb: (v: unknown[]) => void) => {
                cb(connectionsData);
                return () => {};
            });
            mockDiagramTitleSubscribe.mockImplementation((cb: (v: string) => void) => {
                cb(title);
                return () => {};
            });
        }

        beforeEach(() => {
            setStores();
        });

        it('создаёт файл с правильной структурой', () => {
            const shape1 = makeShape({
                id: 's1',
                type: 'square',
                left: 10,
                top: 20,
                width: 100,
                height: 50,
                fill: '#fff',
                stroke: '#000',
                text: 'Hello',
                order: '1',
            });
            const shape2 = makeShape({
                id: 's2',
                type: 'ellipse',
                left: 30,
                top: 40,
                width: 80,
                height: 80,
                rx: 40,
                ry: 40,
                fill: '#000',
                stroke: '#fff',
                text: 'World',
                order: '2',
            });
            setStores([shape1, shape2], [], 'Test Diagram');

            const mockLink = {
                download: '',
                href: '',
                click: vi.fn(),
            } as unknown as HTMLAnchorElement;
            const clickSpy = vi.spyOn(document, 'createElement').mockImplementation((tagName) => {
                if (tagName === 'a') return mockLink;
                return document.createElement(tagName);
            });
            const revokeSpy = vi.spyOn(URL, 'revokeObjectURL').mockImplementation(() => {});

            manager.toFile();

            expect(mockLink.download).toMatch(/^\d{4}-\d{2}-\d{2}-\d{4}\.klubok$/);
            expect(mockLink.click).toHaveBeenCalled();
            expect(revokeSpy).toHaveBeenCalled();

            clickSpy.mockRestore();
            revokeSpy.mockRestore();
        });

        it('учитывает соединения между фигурами', () => {
            const shape1 = makeShape({ id: 's1' });
            const shape2 = makeShape({ id: 's2' });
            const conn = makeConnection({ fromElement: { name: 's1' }, toElement: { name: 's2' } });
            setStores([shape1, shape2], [conn]);

            const mockLink = {
                download: '',
                href: '',
                click: vi.fn(),
            } as unknown as HTMLAnchorElement;
            const clickSpy = vi.spyOn(document, 'createElement').mockImplementation((tagName) => {
                if (tagName === 'a') return mockLink;
                return document.createElement(tagName);
            });

            manager.toFile();

            expect(mockLink.download).toMatch(/^\d{4}-\d{2}-\d{2}-\d{4}\.klubok$/);
            expect(mockLink.click).toHaveBeenCalled();

            clickSpy.mockRestore();
        });

        it('пропускает соединения с отсутствующими фигурами', () => {
            const shape1 = makeShape({ id: 's1' });
            const conn = makeConnection({
                fromElement: { name: 's1' },
                toElement: { name: 'missing' },
            });
            setStores([shape1], [conn]);

            const mockLink = {
                download: '',
                href: '',
                click: vi.fn(),
            } as unknown as HTMLAnchorElement;
            const clickSpy = vi.spyOn(document, 'createElement').mockImplementation((tagName) => {
                if (tagName === 'a') return mockLink;
                return document.createElement(tagName);
            });

            manager.toFile();

            expect(mockLink.click).toHaveBeenCalled();

            clickSpy.mockRestore();
        });

        it('пропускает соединения без имен фигур', () => {
            const shape1 = makeShape({ id: 's1' });
            const shape2 = makeShape({ id: 's2' });
            const conn = makeConnection({ fromElement: { name: '' }, toElement: { name: 's2' } });
            setStores([shape1, shape2], [conn]);

            const mockLink = {
                download: '',
                href: '',
                click: vi.fn(),
            } as unknown as HTMLAnchorElement;
            const clickSpy = vi.spyOn(document, 'createElement').mockImplementation((tagName) => {
                if (tagName === 'a') return mockLink;
                return document.createElement(tagName);
            });

            manager.toFile();

            expect(mockLink.click).toHaveBeenCalled();

            clickSpy.mockRestore();
        });

        it('использует текущее время в имени файла', () => {
            const mockLink = {
                download: '',
                href: '',
                click: vi.fn(),
            } as unknown as HTMLAnchorElement;
            const clickSpy = vi.spyOn(document, 'createElement').mockImplementation((tagName) => {
                if (tagName === 'a') return mockLink;
                return document.createElement(tagName);
            });

            manager.toFile();

            expect(mockLink.download).toBe('2026-05-08-1430.klubok');

            clickSpy.mockRestore();
        });

        it('создаёт Blob с правильным типом', () => {
            const mockLink = {
                download: '',
                href: '',
                click: vi.fn(),
            } as unknown as HTMLAnchorElement;
            const clickSpy = vi.spyOn(document, 'createElement').mockImplementation((tagName) => {
                if (tagName === 'a') return mockLink;
                return document.createElement(tagName);
            });

            const createObjectURLSpy = vi.spyOn(URL, 'createObjectURL').mockReturnValue('blob:test');

            manager.toFile();

            expect(createObjectURLSpy).toHaveBeenCalledWith(expect.any(Blob));
            const blob = createObjectURLSpy.mock.calls[0][0] as Blob;
            expect(blob.type).toBe('application/json');

            clickSpy.mockRestore();
            createObjectURLSpy.mockRestore();
        });

        it('вызывает revokeObjectURL после скачивания', () => {
            const mockLink = {
                download: '',
                href: '',
                click: vi.fn(),
            } as unknown as HTMLAnchorElement;
            const clickSpy = vi.spyOn(document, 'createElement').mockImplementation((tagName) => {
                if (tagName === 'a') return mockLink;
                return document.createElement(tagName);
            });
            const revokeSpy = vi.spyOn(URL, 'revokeObjectURL').mockImplementation(() => {});

            manager.toFile();

            expect(revokeSpy).toHaveBeenCalledWith(mockLink.href);

            clickSpy.mockRestore();
            revokeSpy.mockRestore();
        });

        it('работает с пустым списком фигур', () => {
            setStores([]);

            const mockLink = {
                download: '',
                href: '',
                click: vi.fn(),
            } as unknown as HTMLAnchorElement;
            const clickSpy = vi.spyOn(document, 'createElement').mockImplementation((tagName) => {
                if (tagName === 'a') return mockLink;
                return document.createElement(tagName);
            });

            manager.toFile();

            expect(mockLink.click).toHaveBeenCalled();
            expect(mockLink.download).toMatch(/^\d{4}-\d{2}-\d{2}-\d{4}\.klubok$/);

            clickSpy.mockRestore();
        });

        it('работает с пустым заголовком', () => {
            setStores([], [], '');

            const mockLink = {
                download: '',
                href: '',
                click: vi.fn(),
            } as unknown as HTMLAnchorElement;
            const clickSpy = vi.spyOn(document, 'createElement').mockImplementation((tagName) => {
                if (tagName === 'a') return mockLink;
                return document.createElement(tagName);
            });

            manager.toFile();

            expect(mockLink.click).toHaveBeenCalled();

            clickSpy.mockRestore();
        });
    });
});
