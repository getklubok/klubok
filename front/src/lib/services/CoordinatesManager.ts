import { Canvas as FabricCanvas } from 'fabric/es';
import { coordinates, type ICoordinates } from '../stores';

export interface CoordinatesManagerDependencies {
    setCoordinates: (coords: ICoordinates) => void;
}

const defaultDeps: CoordinatesManagerDependencies = {
    setCoordinates: (coords) => coordinates.set(coords),
};

export class CoordinatesManager {
    private canvas: FabricCanvas | null = null;
    private readonly deps: CoordinatesManagerDependencies;

    constructor(deps: CoordinatesManagerDependencies = defaultDeps) {
        this.deps = deps;
    }

    public setCanvas(canvas: FabricCanvas) {
        this.canvas = canvas;
    }

    public set(x: number, y: number, zoom: number) {
        if (!this.canvas) return;

        this.canvas.setZoom(zoom);

        const vpt = this.canvas.viewportTransform;
        if (vpt) {
            vpt[4] = -x;
            vpt[5] = -y;
            this.canvas.setViewportTransform(vpt);
        }
    }

    public update() {
        if (!this.canvas) return;

        const vpt = this.canvas.viewportTransform;
        if (!vpt) return;

        this.deps.setCoordinates({
            x: Math.round(-vpt[4]),
            y: Math.round(-vpt[5]),
            zoom: Math.round(this.canvas.getZoom() * 100) / 100,
        });
    }
}

export const coordinatesManager = new CoordinatesManager();
