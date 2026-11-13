import type { ElementType } from './stores';

export interface RespDataMsg {
    response?: {
        data?: {
            message?: string;
        };
    };
}

export interface RespStatus {
    response?: {
        status?: number;
    };
}

export interface ShapeJson {
    id: number;
    type: ElementType;
    left: number;
    top: number;
    width: number;
    height: number;
    fill: string;
    stroke: string;
    text: string;
    order: string;
    rx?: number;
    ry?: number;
}

export interface DiagramJson {
    source: string;
    type: string;
    version: number;
    title: string;
    shapes: ShapeJson[];
    connections: [number, number][];
}
