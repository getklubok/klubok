import { Canvas as FabricCanvas, Group as FabricGroup } from 'fabric/es';
import { ulid } from 'ulid';

import { canvasManager } from '../services/CanvasManager';
import { textEditorManager } from '../services/TextEditorManager';
import { shapeStoreManager } from '../services/ShapeStoreManager';

import { PAGE_GRID_STEP } from '../constants';
import { commentsSelectedGroup, commentsShow, selectedObject, type ShapeElement } from '../stores';

import { isShapeGroup } from '../utils/isShapeGroup';
import { confirm } from '../utils/confirm';
import { calculateOrderKey } from '../utils/orderKey';

export interface GroupActionManagerDependencies {
    upsertShape: (obj: ShapeElement) => FabricGroup | null;
    removeShape: (obj: FabricGroup) => boolean;
    startEditing: (obj: FabricGroup) => boolean;
    getShapeById: (id: string) => ShapeElement | undefined;
    createId: () => string;
    calculateOrderKey: (before: string | null, after: string | null) => string;
    isShapeGroup: (obj: FabricGroup) => boolean;
    confirm: typeof confirm;
}

const defaultDeps: GroupActionManagerDependencies = {
    upsertShape: (obj) => canvasManager.upsertShape(obj),
    removeShape: (obj) => canvasManager.removeShape(obj),
    startEditing: (obj) => textEditorManager.startEditing(obj),
    getShapeById: (id) => shapeStoreManager.getById(id),
    createId: () => ulid().toString(),
    calculateOrderKey,
    isShapeGroup,
    confirm,
};

export class GroupActionManager {
    private readonly deps: GroupActionManagerDependencies;

    constructor(deps: GroupActionManagerDependencies = defaultDeps) {
        this.deps = deps;
    }

    public startEditing(obj: FabricGroup) {
        if (!isShapeGroup(obj)) return;

        this.deps.startEditing(obj);
    }

    public showComments(obj: FabricGroup) {
        if (!isShapeGroup(obj)) return;

        commentsSelectedGroup.set(obj);
        commentsShow.set(true);
    }

    public async cloneObject(original: FabricGroup) {
        if (!isShapeGroup(original)) return;

        const confirmed = await confirm({
            message: `Клонировать объект?`,
        });
        if (!confirmed) return;

        const originalShape = this.deps.getShapeById(original.name || '');
        if (originalShape) {
            const canvas = original.canvas as FabricCanvas;

            const userObjects = canvas.getObjects().filter((o) => isShapeGroup(o));
            const afterObj = userObjects[userObjects.length - 1] as FabricGroup;
            const before = afterObj.order || null;
            const newOrder = calculateOrderKey(before, null);

            const clonedShape = {
                ...originalShape,
                id: this.deps.createId(),
                left: originalShape.left + PAGE_GRID_STEP,
                top: originalShape.top + PAGE_GRID_STEP,
                order: newOrder,
            };

            const cloned = this.deps.upsertShape(clonedShape);
            if (cloned) {
                cloned.order = newOrder;
                selectedObject.update((el) => {
                    if (!el) return el;
                    el.order = newOrder;
                    return el;
                });

                canvas.bringObjectToFront(cloned);
                canvas.setActiveObject(cloned);
                canvas.requestRenderAll();
            }
        }
    }

    public async deleteObject(obj: FabricGroup) {
        const confirmed = await confirm({
            message: `Вы уверены, что хотите удалить объект?`,
        });
        if (!confirmed) return;

        this.deps.removeShape(obj);
    }
}

export const groupActionManager = new GroupActionManager();
