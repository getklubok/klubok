import { Canvas as FabricCanvas } from 'fabric/es';

export interface ViewportManagerDependencies {
    updateCoordinates: () => void;
}

import { coordinatesManager } from './CoordinatesManager';
import { getBoundingBox } from '../utils/getBoundingBox';

const defaultDeps: ViewportManagerDependencies = {
    updateCoordinates: () => coordinatesManager.update(),
};

interface Padding {
    left: number;
    right: number;
    top: number;
    bottom: number;
}

interface FitViewOptions {
    padding: Padding;
    maxZoom: number;
    minZoom: number;
}

export class ViewportManager {
    private canvas: FabricCanvas | null = null;
    private readonly deps: ViewportManagerDependencies;

    constructor(deps: ViewportManagerDependencies = defaultDeps) {
        this.deps = deps;
    }

    public setCanvas(canvas: FabricCanvas) {
        this.canvas = canvas;
    }

    public fitView(options: FitViewOptions): void {
        if (!this.canvas) return;

        const { padding = { left: 0, right: 0, top: 0, bottom: 0 }, maxZoom = 5, minZoom = 0.1 } = options;

        const objects = this.canvas.getObjects();
        const visibleObjects = objects.filter((obj) => obj.visible !== false);
        if (visibleObjects.length === 0) {
            this.canvas.setViewportTransform([1, 0, 0, 1, 0, 0]);
            return;
        }

        const boundingRect = getBoundingBox(visibleObjects);

        const paddedWidth = boundingRect.width + padding.left + padding.right;
        const paddedHeight = boundingRect.height + padding.top + padding.bottom;

        const viewportWidth = this.canvas.getWidth();
        const viewportHeight = this.canvas.getHeight();

        const scaleX = viewportWidth / paddedWidth;
        const scaleY = viewportHeight / paddedHeight;
        let scale = Math.min(scaleX, scaleY);
        scale = Math.max(minZoom, Math.min(maxZoom, scale));

        const centerX = boundingRect.left + boundingRect.width / 2 - padding.left / 2 + padding.right / 2;
        const centerY = boundingRect.top + boundingRect.height / 2 - padding.top / 2 + padding.bottom / 2;

        const viewportCenterX = viewportWidth / 2;
        const viewportCenterY = viewportHeight / 2;

        const offsetX = viewportCenterX - centerX * scale;
        const offsetY = viewportCenterY - centerY * scale;

        this.canvas.setViewportTransform([scale, 0, 0, scale, offsetX, offsetY]);
        this.canvas.requestRenderAll();

        this.deps.updateCoordinates();
    }

    public centerView(): void {
        if (!this.canvas) return;

        const objects = this.canvas.getObjects();
        const visibleObjects = objects.filter((obj) => obj.visible !== false);
        if (visibleObjects.length === 0) {
            this.canvas.setViewportTransform([1, 0, 0, 1, 0, 0]);
            return;
        }

        const boundingRect = getBoundingBox(visibleObjects);

        const centerX = boundingRect.left + boundingRect.width / 2;
        const centerY = boundingRect.top + boundingRect.height / 2;

        const viewportWidth = this.canvas.getWidth();
        const viewportHeight = this.canvas.getHeight();

        const viewportCenterX = viewportWidth / 2;
        const viewportCenterY = viewportHeight / 2;

        const offsetX = viewportCenterX - centerX;
        const offsetY = viewportCenterY - centerY;

        this.canvas.setViewportTransform([1, 0, 0, 1, offsetX, offsetY]);
        this.canvas.requestRenderAll();

        this.deps.updateCoordinates();
    }
}

export const viewportManager = new ViewportManager();
