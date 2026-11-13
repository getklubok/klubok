// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { ShapeElement } from '../stores';

const { mockCalculateOrderKey } = vi.hoisted(() => ({
    mockCalculateOrderKey: vi.fn().mockReturnValue('new-order-key'),
}));

vi.mock('../utils/orderKey', () => ({
    calculateOrderKey: mockCalculateOrderKey,
}));

vi.mock('../utils/confirm', () => ({
    confirm: vi.fn(),
}));

vi.mock('../utils/isShapeGroup', () => ({
    isShapeGroup: vi.fn(),
}));

const { mockCanvasManager, mockTextEditorManager, mockShapeStoreManager } = vi.hoisted(() => ({
    mockCanvasManager: {
        upsertShape: vi.fn(),
        removeShape: vi.fn(),
    },
    mockTextEditorManager: {
        startEditing: vi.fn(),
    },
    mockShapeStoreManager: {
        getById: vi.fn(),
    },
}));

vi.mock('../services/CanvasManager', () => ({ canvasManager: mockCanvasManager }));
vi.mock('../services/TextEditorManager', () => ({ textEditorManager: mockTextEditorManager }));
vi.mock('../services/ShapeStoreManager', () => ({ shapeStoreManager: mockShapeStoreManager }));

vi.mock('../stores', () => ({
    commentsSelectedGroup: { set: vi.fn() },
    commentsShow: { set: vi.fn() },
    selectedObject: { update: vi.fn() },
}));

import { GroupActionManager, type GroupActionManagerDependencies } from './GroupActionManager';
import { commentsSelectedGroup, commentsShow } from '../stores';
import { isShapeGroup } from '../utils/isShapeGroup';
import { confirm } from '../utils/confirm';
import { calculateOrderKey } from '../utils/orderKey';

function makeFabricGroup(overrides: Record<string, unknown> = {}): import('fabric/es').Group {
    return {
        type: 'group',
        name: 'shape-1',
        ...overrides,
    } as import('fabric/es').Group;
}

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
        order: 'old-order',
        ...overrides,
    };
}

function makeDeps(overrides: Partial<GroupActionManagerDependencies> = {}): GroupActionManagerDependencies {
    return {
        upsertShape: mockCanvasManager.upsertShape,
        removeShape: mockCanvasManager.removeShape,
        startEditing: mockTextEditorManager.startEditing,
        getShapeById: mockShapeStoreManager.getById,
        createId: () => 'new-id',
        calculateOrderKey,
        isShapeGroup,
        confirm,
        ...overrides,
    };
}

describe('GroupActionManager', () => {
    beforeEach(() => {
        vi.resetAllMocks();
        vi.mocked(isShapeGroup).mockImplementation((obj) => {
            return obj && typeof obj === 'object' && obj.type === 'group' && Boolean(obj.name);
        });
        vi.mocked(calculateOrderKey).mockReturnValue('new-order-key');
    });

    describe('startEditing', () => {
        it('вызывает startEditing если объект валиден', () => {
            const deps = makeDeps();
            const manager = new GroupActionManager(deps);
            const group = makeFabricGroup();

            manager.startEditing(group);

            expect(mockTextEditorManager.startEditing).toHaveBeenCalledWith(group);
        });

        it('не делает ничего если объект невалиден', () => {
            const deps = makeDeps();
            const manager = new GroupActionManager(deps);

            vi.mocked(isShapeGroup).mockReturnValue(false);

            manager.startEditing({} as import('fabric/es').Group);

            expect(mockTextEditorManager.startEditing).not.toHaveBeenCalled();
        });

        it('не делает ничего если isShapeGroup возвращает false', () => {
            const deps = makeDeps();
            const manager = new GroupActionManager(deps);

            vi.mocked(isShapeGroup).mockReturnValue(false);

            manager.startEditing(makeFabricGroup());

            expect(mockTextEditorManager.startEditing).not.toHaveBeenCalled();
        });
    });

    describe('showComments', () => {
        it('устанавливает commentsSelectedGroup и показывает комментарии', () => {
            const deps = makeDeps();
            const manager = new GroupActionManager(deps);
            const group = makeFabricGroup();

            manager.showComments(group);

            expect(commentsSelectedGroup.set).toHaveBeenCalledWith(group);
            expect(commentsShow.set).toHaveBeenCalledWith(true);
        });

        it('не делает ничего если объект невалиден', () => {
            const deps = makeDeps();
            const manager = new GroupActionManager(deps);

            vi.mocked(isShapeGroup).mockReturnValue(false);

            manager.showComments({} as import('fabric/es').Group);

            expect(commentsSelectedGroup.set).not.toHaveBeenCalled();
            expect(commentsShow.set).not.toHaveBeenCalled();
        });
    });

    describe('cloneObject', () => {
        it('клонирует объект со смещением', async () => {
            const deps = makeDeps({ createId: () => 'cloned-id' });
            const manager = new GroupActionManager(deps);

            const originalGroup = makeFabricGroup({ name: 'shape-1', left: 10, top: 20 });
            const shape = makeShape({
                id: 'shape-1',
                left: 10,
                top: 20,
                order: 'original-order',
            });
            mockShapeStoreManager.getById.mockReturnValue(shape);

            const mockCanvas = {
                getObjects: vi.fn().mockReturnValue([makeFabricGroup({ name: 'other' })]),
                bringObjectToFront: vi.fn(),
                setActiveObject: vi.fn(),
                requestRenderAll: vi.fn(),
            };
            originalGroup.canvas = mockCanvas as unknown as import('fabric/es').Canvas;

            vi.mocked(confirm).mockResolvedValue(true);
            mockCanvasManager.upsertShape.mockReturnValue(makeFabricGroup({ name: 'cloned-id' }) as never);

            await manager.cloneObject(originalGroup);

            expect(confirm).toHaveBeenCalledWith({ message: 'Клонировать объект?' });
            expect(mockCanvasManager.upsertShape).toHaveBeenCalledWith(
                expect.objectContaining({
                    id: 'cloned-id',
                    left: 30,
                    top: 40,
                    order: 'new-order-key',
                }),
            );
            expect(mockCanvas.bringObjectToFront).toHaveBeenCalled();
            expect(mockCanvas.setActiveObject).toHaveBeenCalled();
            expect(mockCanvas.requestRenderAll).toHaveBeenCalled();
        });

        it('не клонирует если пользователь отменил', async () => {
            const deps = makeDeps();
            const manager = new GroupActionManager(deps);
            const group = makeFabricGroup();

            vi.mocked(confirm).mockResolvedValue(false);

            await manager.cloneObject(group);

            expect(mockCanvasManager.upsertShape).not.toHaveBeenCalled();
        });

        it('ничего не делает если объект невалиден', async () => {
            const deps = makeDeps();
            const manager = new GroupActionManager(deps);

            vi.mocked(isShapeGroup).mockReturnValue(false);

            await manager.cloneObject({} as import('fabric/es').Group);

            expect(confirm).not.toHaveBeenCalled();
        });

        it('ничего не делает если shape не найден в store', async () => {
            const deps = makeDeps();
            const manager = new GroupActionManager(deps);
            const group = makeFabricGroup({ name: 'missing' });
            group.canvas = {
                getObjects: vi.fn().mockReturnValue([]),
            } as unknown as import('fabric/es').Canvas;

            mockShapeStoreManager.getById.mockReturnValue(undefined);

            await manager.cloneObject(group);

            expect(confirm).toHaveBeenCalled();
            expect(mockCanvasManager.upsertShape).not.toHaveBeenCalled();
        });

        it('не делает ничего если upsertShape вернул null', async () => {
            const deps = makeDeps();
            const manager = new GroupActionManager(deps);

            const group = makeFabricGroup({ name: 'shape-1' });
            mockShapeStoreManager.getById.mockReturnValue(makeShape());

            const mockCanvas = {
                getObjects: vi.fn().mockReturnValue([makeFabricGroup()]),
                bringObjectToFront: vi.fn(),
                setActiveObject: vi.fn(),
                requestRenderAll: vi.fn(),
            };
            group.canvas = mockCanvas as unknown as import('fabric/es').Canvas;

            vi.mocked(confirm).mockResolvedValue(true);
            mockCanvasManager.upsertShape.mockReturnValue(null);

            await manager.cloneObject(group);

            expect(mockCanvas.bringObjectToFront).not.toHaveBeenCalled();
            expect(mockCanvas.setActiveObject).not.toHaveBeenCalled();
            expect(mockCanvas.requestRenderAll).not.toHaveBeenCalled();
        });
    });

    describe('deleteObject', () => {
        it('удаляет объект после подтверждения', async () => {
            const deps = makeDeps();
            const manager = new GroupActionManager(deps);
            const group = makeFabricGroup();

            vi.mocked(confirm).mockResolvedValue(true);
            mockCanvasManager.removeShape.mockReturnValue(true);

            await manager.deleteObject(group);

            expect(confirm).toHaveBeenCalledWith({
                message: 'Вы уверены, что хотите удалить объект?',
            });
            expect(mockCanvasManager.removeShape).toHaveBeenCalledWith(group);
        });

        it('не удаляет если пользователь отменил', async () => {
            const deps = makeDeps();
            const manager = new GroupActionManager(deps);
            const group = makeFabricGroup();

            vi.mocked(confirm).mockResolvedValue(false);

            await manager.deleteObject(group);

            expect(mockCanvasManager.removeShape).not.toHaveBeenCalled();
        });
    });
});
