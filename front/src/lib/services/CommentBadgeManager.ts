import { Canvas as FabricCanvas, Group as FabricGroup, Circle as FabricCircle, FabricText } from 'fabric/es';
import { get } from 'svelte/store';
import { fabricManager } from './FabricManager';
import { resolvedThemeStore } from '../stores/theme';

export interface CommentBadgeManagerDependencies {
    createBadge: (count: number) => FabricGroup;
}

const defaultDeps: CommentBadgeManagerDependencies = {
    createBadge(count: number) {
        const radius = count > 9 ? 10 : 8;

        return new FabricGroup(
            [
                new FabricCircle({
                    radius: radius,
                    fill: '#333333',
                    stroke: 'white',
                    strokeWidth: 1.5,
                }),
                new FabricText(count.toString(), {
                    fontSize: count > 9 ? 8 : 10,
                    fill: 'white',
                    fontWeight: 'bold',
                    fontFamily: 'Arial',
                }),
            ],
            {
                originX: 'center',
                originY: 'center',
                selectable: false,
                evented: false,
            },
        );
    },
};

export class CommentBadgeManager {
    private canvas: FabricCanvas | null = null;
    private badges: Map<string, FabricGroup> = new Map();
    private readonly deps: CommentBadgeManagerDependencies;

    constructor(deps: CommentBadgeManagerDependencies = defaultDeps) {
        this.deps = deps;
    }

    public setCanvas(canvas: FabricCanvas) {
        this.canvas = canvas;
        this.badges = new Map(); // shapeId -> badge
    }

    public upsert(shape: FabricGroup, commentCount: number, increment: boolean) {
        if (!this.canvas) return;
        if (!shape.name) return;

        let badge = this.badges.get(shape.name);
        if (!badge) {
            badge = this.deps.createBadge(commentCount);

            badge.selectable = false;
            badge.evented = false;

            // Применяем тему к кругу и тексту сразу при создании (как к остальным объектам канваса)
            const isDark = get(resolvedThemeStore) === 'dark';
            badge.getObjects().forEach((obj) => fabricManager.applyThemeColor(obj, isDark));

            this.updateBadgePosition(shape, badge);

            this.canvas.add(badge);
            this.badges.set(shape.name, badge);

            return;
        }

        const textObj = badge.getObjects()[1] as FabricText;
        textObj.set({
            text: increment ? (parseInt(textObj.text || '0', 10) + 1).toString() : commentCount.toString(),
        });
        this.canvas.requestRenderAll();
    }

    public updatePosition(shape: FabricGroup) {
        if (!this.canvas) return;
        if (!shape.name) return;

        const badge = this.badges.get(shape.name);

        if (badge) {
            this.updateBadgePosition(shape, badge);
            this.canvas.requestRenderAll();
        }
    }

    public remove(shape: FabricGroup) {
        if (!this.canvas) return;
        if (!shape.name) return;

        const badge = this.badges.get(shape.name);
        if (badge) {
            this.canvas.remove(badge);
            this.badges.delete(shape.name);
        }
    }

    private updateBadgePosition(group: FabricGroup, badge: FabricGroup) {
        const center = group.getCenterPoint();
        const boundingRect = group.getBoundingRect();

        let offsetX: number = boundingRect.width / 2 - 10;
        let offsetY: number = -boundingRect.height / 2 + 10;

        if (group.getObjects()[0].type === 'rect') {
            offsetX += 20;
            offsetY -= 20;
        }

        badge.set({
            left: center.x + offsetX,
            top: center.y + offsetY,
        });
    }
}

export const commentBadgeManager = new CommentBadgeManager();
