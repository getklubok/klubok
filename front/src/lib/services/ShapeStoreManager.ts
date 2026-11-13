import { type ShapeElement, shapes } from '../stores';
import { get } from 'svelte/store';
import { Group as FabricGroup } from 'fabric/es';

export interface ShapeStoreManagerDependencies {
    updateShapes: (fn: (shapes: ShapeElement[]) => ShapeElement[]) => void;
    getShapes: () => ShapeElement[];
}

const defaultDeps: ShapeStoreManagerDependencies = {
    updateShapes: (fn) => shapes.update(fn),
    getShapes: () => get(shapes),
};

export class ShapeStoreManager {
    private readonly deps: ShapeStoreManagerDependencies;

    constructor(deps: ShapeStoreManagerDependencies = defaultDeps) {
        this.deps = deps;
    }

    public add(element: ShapeElement) {
        this.deps.updateShapes((s) => [...s, element]);
    }

    public getById(id: string): ShapeElement | undefined {
        return this.deps.getShapes().find((el) => el.id === id);
    }

    public update(element: ShapeElement): void {
        const index = this.deps.getShapes().findIndex((el) => el.id === element.id);
        if (index === -1) return;

        this.deps.updateShapes((s) => {
            const newShapes = [...s];
            newShapes[index] = element;
            return newShapes;
        });
    }

    public remove(group: FabricGroup) {
        this.deps.updateShapes((s) => s.filter((el) => el.fabricGroup?.name !== group.name));
    }

    public getMaxOrderKey(): string | null {
        if (this.deps.getShapes().length === 0) {
            return null;
        }

        const last = this.deps.getShapes().reduce((prev, curr) => {
            return curr.order > prev.order ? curr : prev;
        });

        return last.order;
    }
}

export const shapeStoreManager = new ShapeStoreManager();
