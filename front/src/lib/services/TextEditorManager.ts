import {
    Canvas as FabricCanvas,
    Group as FabricGroup,
    FabricObject,
    IText as FabricIText,
    type Control,
} from 'fabric/es';
import { type ShapeElement } from '../stores';

import { shapeStoreManager } from './ShapeStoreManager';

import { isShapeGroup } from '../utils/isShapeGroup';
import { calcBoxLine, calcBoxText } from '../utils/calcBox';
import { applyTextHorizontalAlign } from '../utils/textAlign';

export interface TextEditorManagerDependencies {
    getById: (id: string) => ShapeElement | undefined;
    update: (element: ShapeElement) => void;
}

const defaultDeps: TextEditorManagerDependencies = {
    getById: (id) => shapeStoreManager.getById(id),
    update: (element) => shapeStoreManager.update(element),
};

interface IGroupParams {
    left: number;
    top: number;
    width: number;
    height: number;
    scaleX: number;
    scaleY: number;
    controls: Record<string, Control>;
    sourceGroup: FabricGroup;
}

export class TextEditorManager {
    private canvas: FabricCanvas | null = null;

    private groupName: string | null = null;
    private groupItems: FabricObject[] | null = null;
    private groupShape: ShapeElement | null = null;
    private groupParams: IGroupParams | null = null;
    private isEditing: boolean = false;
    private readonly deps: TextEditorManagerDependencies;

    constructor(deps: TextEditorManagerDependencies = defaultDeps) {
        this.deps = deps;
    }

    public setCanvas(canvas: FabricCanvas) {
        this.canvas = canvas;
    }

    public startEditing(group: FabricGroup) {
        if (!this.canvas) return false;
        if (!isShapeGroup(group)) return false;
        if (group.shapeType === 'image') return false;

        const shape = this.deps.getById(group.name || '');
        if (shape) {
            this.groupShape = shape;
        }

        const text = group.getObjects()[1] as FabricIText;

        const shapeWidth = this.groupShape?.width ?? group.width;
        const textAlign = this.groupShape?.textAlign ?? 'center';

        let lastLineCount = (text.text || '').split('\n').length;

        const handleTextChanged = () => {
            const currentLineCount = (text.text || '').split('\n').length;
            if (currentLineCount !== lastLineCount) {
                lastLineCount = currentLineCount;

                if (group.shapeType === 'box') {
                    group.getObjects()[2].set({ points: calcBoxLine(group.width, group.height, text.text || '') });
                    text.set({ top: calcBoxText(group.height, text.height) });
                } else {
                    text.set({ top: 0 });
                }
            }

            applyTextHorizontalAlign(text, shapeWidth, textAlign, group.shapeType);
            text.canvas?.requestRenderAll();
        };

        text.on('changed', handleTextChanged);

        this.canvas.setActiveObject(text);
        text.enterEditing();
        text.selectAll();

        this.groupName = group.name || null;
        this.groupItems = group._objects as FabricObject[];
        this.groupParams = {
            left: group.left,
            top: group.top,
            width: group.width,
            height: group.height,
            scaleX: group.scaleX,
            scaleY: group.scaleY,
            controls: group.controls,
            sourceGroup: group,
        };
        this.isEditing = true;

        this.canvas.remove(group);
        for (let i = 0; i < this.groupItems.length; i++) {
            this.canvas.add(this.groupItems[i]);
        }

        this.canvas.requestRenderAll();

        return true;
    }

    public finishEditing() {
        if (!this.canvas) return;
        if (!this.isEditing || !this.groupItems || !this.groupName) return;

        const items: FabricObject[] = [];
        this.groupItems.forEach((obj) => {
            items.push(obj);
            this.canvas?.remove(obj);
        });

        const newGroup = new FabricGroup(items, {
            name: this.groupName,
        });

        if (this.groupParams) {
            const sourceGroup = this.groupParams.sourceGroup;
            newGroup.controls = this.groupParams.controls;
            // shapeType нужен для масштабирования внутренней фигуры и box-разметки,
            // цвета — для контролов выделения: новая группа создаётся с дефолтными.
            newGroup.set({
                shapeType: sourceGroup.shapeType,
                order: sourceGroup.order,
                borderColor: sourceGroup.borderColor,
                cornerColor: sourceGroup.cornerColor,
                cornerStrokeColor: sourceGroup.cornerStrokeColor,
            });
        }

        if (this.groupShape && this.groupShape.fabricGroup) {
            const textObj = items[1] as FabricIText;
            const updatedShape: ShapeElement = {
                ...this.groupShape,
                left: newGroup.left,
                top: newGroup.top,
                width: newGroup.width,
                height: newGroup.height,
                scaleX: newGroup.scaleX,
                scaleY: newGroup.scaleY,
                text: textObj.text?.trim() || 'Текст',
                fabricGroup: newGroup,
            };
            this.deps.update(updatedShape);
        }

        this.canvas.add(newGroup);
        this.canvas.requestRenderAll();

        this.groupName = null;
        this.groupItems = null;
        this.groupShape = null;
        this.groupParams = null;
        this.isEditing = false;
    }
}

export const textEditorManager = new TextEditorManager();
