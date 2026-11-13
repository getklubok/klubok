import { DEFAULT_SCALE, DEFAULT_STROKE_WIDTH } from '../constants';
import type { ShapeElement, ConnectionElement } from '../stores';
import type { DiagramShape, DiagramConnection } from '../ApiSpec';
import type { ShapeJson } from '../types';

export function mapTypeToApi(type: string): 'S' | 'E' | 'D' | 'B' | 'T' {
    switch (type) {
        case 'square': {
            return 'S';
        }

        case 'ellipse': {
            return 'E';
        }

        case 'diamond': {
            return 'D';
        }

        case 'box': {
            return 'B';
        }

        case 'text': {
            return 'T';
        }

        default:
            throw new Error(`Unknown shape type: ${type}`);
    }
}

export function mapShapeJsonToApi(newId: string, src: ShapeJson): DiagramShape {
    return {
        id: newId,
        tp: mapTypeToApi(src.type),
        l: src.left,
        t: src.top,
        sx: DEFAULT_SCALE,
        sy: DEFAULT_SCALE,
        w: src.width,
        h: src.height,
        fc: src.fill,
        sc: src.stroke,
        txt: src.text,
        o: src.order,
    };
}

export function mapShapeToApi(src: ShapeElement): DiagramShape {
    return {
        id: src.id,
        tp: mapTypeToApi(src.type),
        l: src.left,
        t: src.top,
        sx: src.scaleX,
        sy: src.scaleY,
        w: src.width,
        h: src.height,
        fc: src.fill,
        sc: src.stroke,
        txt: src.text,
        o: src.order,
    };
}

export function mapShapeFromApi(src: DiagramShape): ShapeElement {
    const base = {
        id: src.id,
        left: src.l,
        top: src.t,
        width: src.w,
        height: src.h,
        fill: src.fc,
        stroke: src.sc,
        strokeWidth: DEFAULT_STROKE_WIDTH,
        scaleX: src.sx,
        scaleY: src.sy,
        text: src.txt,
        order: src.o,
    };

    switch (src.tp) {
        case 'S': {
            return {
                ...base,
                type: 'square' as const,
            };
        }

        case 'E': {
            return {
                ...base,
                type: 'ellipse' as const,
                rx: src.w / 2,
                ry: src.h / 2,
            };
        }

        case 'D': {
            return {
                ...base,
                type: 'diamond' as const,
            };
        }

        case 'B': {
            return {
                ...base,
                type: 'box' as const,
            };
        }

        case 'T': {
            return {
                ...base,
                type: 'text' as const,
            };
        }

        default:
            throw new Error(`Unknown shape type: ${src.tp}`);
    }
}

export function mapConnectionToApi(conn: ConnectionElement): DiagramConnection {
    return {
        fid: conn.fromShapeId,
        tid: conn.toShapeId,
    };
}

export function mapConnectionFromApi(conn: DiagramConnection): ConnectionElement {
    return {
        fromShapeId: conn.fid,
        toShapeId: conn.tid,
    };
}
