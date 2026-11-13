import { FabricObject } from 'fabric/es';

export function getBoundingBox(objects: FabricObject[]): {
    left: number;
    top: number;
    width: number;
    height: number;
} {
    let minX = Infinity,
        minY = Infinity,
        maxX = -Infinity,
        maxY = -Infinity;

    for (const obj of objects) {
        const bounds = obj.getBoundingRect();
        minX = Math.min(minX, bounds.left);
        minY = Math.min(minY, bounds.top);
        maxX = Math.max(maxX, bounds.left + bounds.width);
        maxY = Math.max(maxY, bounds.top + bounds.height);
    }

    return {
        left: minX,
        top: minY,
        width: maxX - minX,
        height: maxY - minY,
    };
}
