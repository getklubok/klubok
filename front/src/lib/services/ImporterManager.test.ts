// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { ImporterManager } from './ImporterManager';
import type { DiagramJson, ShapeJson } from '../types';

const { mockUlid, mockApi, mockMapShapeJsonToApi, mockMapConnectionToApi, mockLogger, mockAlert } = vi.hoisted(() => ({
    mockUlid: vi.fn(),
    mockApi: {
        diagramService: {
            diagramImport: vi.fn(),
        },
    },
    mockMapShapeJsonToApi: vi.fn(),
    mockMapConnectionToApi: vi.fn(),
    mockLogger: {
        mustError: vi.fn(),
    },
    mockAlert: vi.fn(),
}));

vi.mock('ulid', () => ({
    ulid: mockUlid,
}));

vi.mock('../Api', () => ({
    api: mockApi,
}));

vi.mock('./mapper', () => ({
    mapShapeJsonToApi: mockMapShapeJsonToApi,
    mapConnectionToApi: mockMapConnectionToApi,
}));

vi.mock('./logger', () => ({
    default: mockLogger,
}));

vi.mock('../utils/alert', () => ({
    alert: mockAlert,
}));

function makeFile(content: string): File {
    return new File([content], 'test.klubok', { type: 'application/json' });
}

function makeValidDiagramJson(overrides: Partial<DiagramJson> = {}): DiagramJson {
    return {
        source: 'https://github.com/getklubok/klubok',
        type: 'klubok',
        version: 1,
        title: 'Test Diagram',
        shapes: [
            {
                id: 1,
                type: 'square',
                left: 0,
                top: 0,
                width: 100,
                height: 100,
                fill: '#fff',
                stroke: '#000',
                text: '',
                order: '',
            },
            {
                id: 2,
                type: 'ellipse',
                left: 200,
                top: 0,
                width: 80,
                height: 80,
                fill: '#000',
                stroke: '#fff',
                text: 'Hello',
                order: '1',
            },
        ],
        connections: [[1, 2]],
        ...overrides,
    };
}

function makeShapeJson(overrides: Partial<ShapeJson> = {}): ShapeJson {
    return {
        id: 1,
        type: 'square',
        left: 0,
        top: 0,
        width: 100,
        height: 100,
        fill: '#fff',
        stroke: '#000',
        text: '',
        order: '',
        ...overrides,
    };
}

describe('ImporterManager', () => {
    let manager: ImporterManager;
    let originalLocation: Location;

    beforeEach(() => {
        vi.clearAllMocks();
        manager = new ImporterManager();

        originalLocation = window.location;
        delete (window as unknown as { location: unknown }).location;
        (window as unknown as { location: { hash: string } }).location = { hash: '' } as Location;

        mockUlid.mockReturnValue('new-id-1');
        mockMapShapeJsonToApi.mockImplementation((id, shape) => ({
            id,
            tp: 'S',
            l: shape.left,
            t: shape.top,
            w: shape.width,
            h: shape.height,
            fc: shape.fill,
            sc: shape.stroke,
            txt: shape.text,
            o: shape.order,
        }));
        mockMapConnectionToApi.mockImplementation((conn) => ({
            fid: conn.fromShapeId,
            tid: conn.toShapeId,
        }));
    });

    afterEach(() => {
        (window as unknown as { location: Location }).location = originalLocation;
    });

    describe('fromFile', () => {
        it('успешно импортирует валидный файл', async () => {
            const diagramData = makeValidDiagramJson();
            const file = makeFile(JSON.stringify(diagramData));

            mockApi.diagramService.diagramImport.mockResolvedValue({
                diagram: { id: 'diag-1' },
            });

            await manager.fromFile(file);

            expect(mockApi.diagramService.diagramImport).toHaveBeenCalledWith({
                title: 'Test Diagram',
                shapes: expect.any(Array),
                connections: expect.any(Array),
            });
            expect(window.location.hash).toBe('/diagrams/diag-1');
        });

        it('выбрасывает ошибку при невалидном source', async () => {
            const diagramData = makeValidDiagramJson({ source: 'invalid' });
            const file = makeFile(JSON.stringify(diagramData));

            await expect(manager.fromFile(file)).rejects.toThrow('Invalid file format');
            expect(mockLogger.mustError).toHaveBeenCalled();
        });

        it('выбрасывает ошибку при невалидном type', async () => {
            const diagramData = makeValidDiagramJson({ type: 'invalid' });
            const file = makeFile(JSON.stringify(diagramData));

            await expect(manager.fromFile(file)).rejects.toThrow('Invalid file format');
            expect(mockLogger.mustError).toHaveBeenCalled();
        });

        it('выбрасывает ошибку при невалидном JSON', async () => {
            const file = makeFile('invalid json');

            await expect(manager.fromFile(file)).rejects.toThrow();
            expect(mockLogger.mustError).toHaveBeenCalled();
        });
    });

    describe('importData', () => {
        it('создаёт новые ID для фигур через ulid', async () => {
            const diagramData = makeValidDiagramJson({
                shapes: [makeShapeJson({ id: 1 }), makeShapeJson({ id: 2 })],
            });
            const file = makeFile(JSON.stringify(diagramData));

            mockUlid.mockReturnValueOnce('ulid-1').mockReturnValueOnce('ulid-2');
            mockApi.diagramService.diagramImport.mockResolvedValue({
                diagram: { id: 'diag-1' },
            });

            await manager.fromFile(file);

            expect(mockUlid).toHaveBeenCalledTimes(2);
            expect(mockMapShapeJsonToApi).toHaveBeenCalledWith('ulid-1', expect.any(Object));
            expect(mockMapShapeJsonToApi).toHaveBeenCalledWith('ulid-2', expect.any(Object));
        });

        it('маппит соединения с использованием ID map', async () => {
            const diagramData = makeValidDiagramJson({
                shapes: [makeShapeJson({ id: 10 }), makeShapeJson({ id: 20 })],
                connections: [[10, 20]],
            });
            const file = makeFile(JSON.stringify(diagramData));

            mockUlid.mockReturnValueOnce('new-10').mockReturnValueOnce('new-20');
            mockApi.diagramService.diagramImport.mockResolvedValue({
                diagram: { id: 'diag-1' },
            });

            await manager.fromFile(file);

            expect(mockMapConnectionToApi).toHaveBeenCalledWith({
                fromShapeId: 'new-10',
                toShapeId: 'new-20',
            });
        });

        it('пропускает соединения с отсутствующими ID в карте', async () => {
            const diagramData = makeValidDiagramJson({
                shapes: [makeShapeJson({ id: 1 })],
                connections: [[1, 999]],
            });
            const file = makeFile(JSON.stringify(diagramData));

            mockApi.diagramService.diagramImport.mockResolvedValue({
                diagram: { id: 'diag-1' },
            });

            await manager.fromFile(file);

            expect(mockMapConnectionToApi).not.toHaveBeenCalled();
        });

        it('работает без фигур', async () => {
            const diagramData = makeValidDiagramJson({ shapes: [], connections: [] });
            const file = makeFile(JSON.stringify(diagramData));

            mockApi.diagramService.diagramImport.mockResolvedValue({
                diagram: { id: 'diag-1' },
            });

            await manager.fromFile(file);

            expect(mockMapShapeJsonToApi).not.toHaveBeenCalled();
            expect(mockMapConnectionToApi).not.toHaveBeenCalled();
            expect(mockApi.diagramService.diagramImport).toHaveBeenCalledWith({
                title: 'Test Diagram',
                shapes: [],
                connections: [],
            });
        });

        it('работает без соединений', async () => {
            const diagramData = makeValidDiagramJson({ connections: [] });
            const file = makeFile(JSON.stringify(diagramData));

            mockApi.diagramService.diagramImport.mockResolvedValue({
                diagram: { id: 'diag-1' },
            });

            await manager.fromFile(file);

            expect(mockMapConnectionToApi).not.toHaveBeenCalled();
            expect(mockApi.diagramService.diagramImport).toHaveBeenCalled();
        });

        it('показывает alert при ошибке API', async () => {
            const diagramData = makeValidDiagramJson();
            const file = makeFile(JSON.stringify(diagramData));

            mockApi.diagramService.diagramImport.mockRejectedValue({
                response: { data: { message: 'Import failed' } },
            });

            await manager.fromFile(file);

            expect(mockAlert).toHaveBeenCalledWith({
                title: 'Ошибка импортирования диаграммы',
                message: 'Import failed',
                type: 'error',
            });
        });

        it('использует message из Error при ошибке API без response.data', async () => {
            const diagramData = makeValidDiagramJson();
            const file = makeFile(JSON.stringify(diagramData));

            mockApi.diagramService.diagramImport.mockRejectedValue(new Error('Network error'));

            await manager.fromFile(file);

            expect(mockAlert).toHaveBeenCalledWith({
                title: 'Ошибка импортирования диаграммы',
                message: 'Network error',
                type: 'error',
            });
        });

        it('не меняет hash при ошибке API', async () => {
            const diagramData = makeValidDiagramJson();
            const file = makeFile(JSON.stringify(diagramData));

            mockApi.diagramService.diagramImport.mockRejectedValue(new Error('error'));

            await manager.fromFile(file);

            expect(window.location.hash).toBe('');
        });
    });
});
