import { describe, it, expect } from 'vitest';
import {
    mapShapeToApi,
    mapShapeFromApi,
    mapConnectionToApi,
    mapConnectionFromApi,
    mapTypeToApi,
    mapShapeJsonToApi,
} from './mapper';
import type { ShapeElement, ConnectionElement } from '../stores';
import type { DiagramShape, DiagramConnection } from '../ApiSpec';
import type { ShapeJson } from '../types';

describe('mapShapeToApi', () => {
    it('должен маппить square в S', () => {
        const shape: ShapeElement = {
            id: '01ABC',
            type: 'square',
            left: 10,
            top: 20,
            width: 100,
            height: 100,
            fill: '#FF0000',
            stroke: '#000000',
            strokeWidth: 1,
            scaleX: 1,
            scaleY: 1,
            text: 'hello',
            order: '',
        };

        const result = mapShapeToApi(shape);
        expect(result).toEqual({
            id: '01ABC',
            tp: 'S',
            l: 10,
            t: 20,
            w: 100,
            h: 100,
            sx: 1,
            sy: 1,
            fc: '#FF0000',
            sc: '#000000',
            txt: 'hello',
            o: '',
        });
    });

    it('должен маппить ellipse в E', () => {
        const shape: ShapeElement = {
            id: '02DEF',
            type: 'ellipse',
            left: 30,
            top: 40,
            width: 100,
            height: 100,
            fill: '#00FF00',
            stroke: '#111111',
            strokeWidth: 1,
            scaleX: 2,
            scaleY: 2,
            rx: 50,
            ry: 50,
            text: 'world',
            order: '',
        };

        const result = mapShapeToApi(shape);
        expect(result).toEqual({
            id: '02DEF',
            tp: 'E',
            l: 30,
            t: 40,
            w: 100,
            h: 100,
            sx: 2,
            sy: 2,
            fc: '#00FF00',
            sc: '#111111',
            txt: 'world',
            o: '',
        });
    });

    it('должен маппить diamond в D', () => {
        const shape: ShapeElement = {
            id: '03GHI',
            type: 'diamond',
            left: 50,
            top: 60,
            width: 100,
            height: 100,
            fill: '#0000FF',
            stroke: '#222222',
            strokeWidth: 1,
            scaleX: 1.5,
            scaleY: 1.5,
            text: 'test',
            order: '',
        };

        const result = mapShapeToApi(shape);
        expect(result).toEqual({
            id: '03GHI',
            tp: 'D',
            l: 50,
            t: 60,
            w: 100,
            h: 100,
            sx: 1.5,
            sy: 1.5,
            fc: '#0000FF',
            sc: '#222222',
            txt: 'test',
            o: '',
        });
    });

    it('должен маппить text в T', () => {
        const shape: ShapeElement = {
            id: '05TXT',
            type: 'text',
            left: 70,
            top: 80,
            width: 100,
            height: 100,
            fill: 'transparent',
            stroke: 'transparent',
            strokeWidth: 1,
            scaleX: 1,
            scaleY: 1,
            text: 'Текст',
            order: '',
        };

        const result = mapShapeToApi(shape);
        expect(result).toEqual({
            id: '05TXT',
            tp: 'T',
            l: 70,
            t: 80,
            w: 100,
            h: 100,
            sx: 1,
            sy: 1,
            fc: 'transparent',
            sc: 'transparent',
            txt: 'Текст',
            o: '',
        });
    });

    it('должен выбрасывать ошибку для неизвестного типа', () => {
        const shape = {
            id: '04XYZ',
            type: 'unknown',
            left: 0,
            top: 0,
            width: 100,
            height: 100,
            fill: '#000000',
            stroke: '#000000',
            strokeWidth: 1,
            scaleX: 1,
            scaleY: 1,
            text: '',
        } as unknown as ShapeElement;

        expect(() => mapShapeToApi(shape)).toThrow('Unknown shape type');
    });
});

describe('mapShapeFromApi', () => {
    it('должен маппить S в square', () => {
        const shape: DiagramShape = {
            id: '01ABC',
            tp: 'S',
            l: 10,
            t: 20,
            w: 100,
            h: 100,
            sx: 1,
            sy: 1,
            fc: '#FF0000',
            sc: '#000000',
            txt: 'hello',
            o: '',
        };

        const result = mapShapeFromApi(shape);
        expect(result).toEqual({
            id: '01ABC',
            type: 'square',
            left: 10,
            top: 20,
            width: 100,
            height: 100,
            fill: '#FF0000',
            stroke: '#000000',
            strokeWidth: 1,
            scaleX: 1,
            scaleY: 1,
            text: 'hello',
            order: '',
        });
    });

    it('должен маппить E в ellipse с radius', () => {
        const shape: DiagramShape = {
            id: '02DEF',
            tp: 'E',
            l: 30,
            t: 40,
            w: 100,
            h: 100,
            sx: 2,
            sy: 2,
            fc: '#00FF00',
            sc: '#111111',
            txt: 'world',
            o: '',
        };

        const result = mapShapeFromApi(shape);
        expect(result).toEqual({
            id: '02DEF',
            type: 'ellipse',
            left: 30,
            top: 40,
            width: 100,
            height: 100,
            fill: '#00FF00',
            stroke: '#111111',
            strokeWidth: 1,
            scaleX: 2,
            scaleY: 2,
            text: 'world',
            rx: 50,
            ry: 50,
            order: '',
        });
    });

    it('должен маппить D в diamond', () => {
        const shape: DiagramShape = {
            id: '03GHI',
            tp: 'D',
            l: 50,
            t: 60,
            w: 100,
            h: 100,
            sx: 1.5,
            sy: 1.5,
            fc: '#0000FF',
            sc: '#222222',
            txt: 'test',
            o: '',
        };

        const result = mapShapeFromApi(shape);
        expect(result).toEqual({
            id: '03GHI',
            type: 'diamond',
            left: 50,
            top: 60,
            width: 100,
            height: 100,
            fill: '#0000FF',
            stroke: '#222222',
            strokeWidth: 1,
            scaleX: 1.5,
            scaleY: 1.5,
            text: 'test',
            order: '',
        });
    });

    it('должен маппить T в text', () => {
        const shape: DiagramShape = {
            id: '05TXT',
            tp: 'T',
            l: 70,
            t: 80,
            w: 100,
            h: 100,
            sx: 1,
            sy: 1,
            fc: 'transparent',
            sc: 'transparent',
            txt: 'Текст',
            o: '',
        };

        const result = mapShapeFromApi(shape);
        expect(result).toEqual({
            id: '05TXT',
            type: 'text',
            left: 70,
            top: 80,
            width: 100,
            height: 100,
            fill: 'transparent',
            stroke: 'transparent',
            strokeWidth: 1,
            scaleX: 1,
            scaleY: 1,
            text: 'Текст',
            order: '',
        });
    });

    it('должен выбрасывать ошибку для неизвестного tp', () => {
        const shape = {
            id: '04XYZ',
            tp: 'X',
            l: 0,
            t: 0,
            sx: 1,
            sy: 1,
            fc: '#000000',
            sc: '#000000',
            txt: '',
        } as unknown as DiagramShape;

        expect(() => mapShapeFromApi(shape)).toThrow('Unknown shape type');
    });
});

describe('mapConnectionToApi', () => {
    it('должен маппить fromShapeId/tid в fid/tid', () => {
        const conn: ConnectionElement = {
            fromShapeId: 'AAA',
            toShapeId: 'BBB',
        };

        const result = mapConnectionToApi(conn);
        expect(result).toEqual({ fid: 'AAA', tid: 'BBB' });
    });
});

describe('mapConnectionFromApi', () => {
    it('должен маппить fid/tid в fromShapeId/toShapeId', () => {
        const conn: DiagramConnection = {
            fid: 'CCC',
            tid: 'DDD',
        };

        const result = mapConnectionFromApi(conn);
        expect(result).toEqual({ fromShapeId: 'CCC', toShapeId: 'DDD' });
    });
});

describe('mapTypeToApi', () => {
    it('должен маппить square в S', () => {
        expect(mapTypeToApi('square')).toBe('S');
    });

    it('должен маппить ellipse в E', () => {
        expect(mapTypeToApi('ellipse')).toBe('E');
    });

    it('должен маппить diamond в D', () => {
        expect(mapTypeToApi('diamond')).toBe('D');
    });

    it('должен маппить text в T', () => {
        expect(mapTypeToApi('text')).toBe('T');
    });

    it('должен выбрасывать ошибку для неизвестного типа', () => {
        expect(() => mapTypeToApi('unknown')).toThrow('Unknown shape type');
    });
});

describe('mapShapeJsonToApi', () => {
    it('должен маппить square ShapeJson в DiagramShape', () => {
        const shapeJson: ShapeJson = {
            id: 1,
            type: 'square',
            left: 10,
            top: 20,
            width: 100,
            height: 100,
            fill: '#FF0000',
            stroke: '#000000',
            text: 'hello',
            order: '',
        };

        const result = mapShapeJsonToApi('newId123', shapeJson);
        expect(result).toEqual({
            id: 'newId123',
            tp: 'S',
            l: 10,
            t: 20,
            w: 100,
            h: 100,
            sx: 1,
            sy: 1,
            fc: '#FF0000',
            sc: '#000000',
            txt: 'hello',
            o: '',
        });
    });

    it('должен маппить ellipse ShapeJson в DiagramShape', () => {
        const shapeJson: ShapeJson = {
            id: 2,
            type: 'ellipse',
            left: 30,
            top: 40,
            width: 100,
            height: 100,
            fill: '#00FF00',
            stroke: '#111111',
            text: 'world',
            order: '',
            rx: 50,
            ry: 50,
        };

        const result = mapShapeJsonToApi('newId456', shapeJson);
        expect(result).toEqual({
            id: 'newId456',
            tp: 'E',
            l: 30,
            t: 40,
            w: 100,
            h: 100,
            sx: 1,
            sy: 1,
            fc: '#00FF00',
            sc: '#111111',
            txt: 'world',
            o: '',
        });
    });

    it('должен маппить diamond ShapeJson в DiagramShape', () => {
        const shapeJson: ShapeJson = {
            id: 3,
            type: 'diamond',
            left: 50,
            top: 60,
            width: 100,
            height: 100,
            fill: '#0000FF',
            stroke: '#222222',
            text: 'test',
            order: '',
        };

        const result = mapShapeJsonToApi('newId789', shapeJson);
        expect(result).toEqual({
            id: 'newId789',
            tp: 'D',
            l: 50,
            t: 60,
            w: 100,
            h: 100,
            sx: 1,
            sy: 1,
            fc: '#0000FF',
            sc: '#222222',
            txt: 'test',
            o: '',
        });
    });

    it('должен маппить text ShapeJson в DiagramShape', () => {
        const shapeJson: ShapeJson = {
            id: 4,
            type: 'text',
            left: 70,
            top: 80,
            width: 100,
            height: 100,
            fill: 'transparent',
            stroke: 'transparent',
            text: 'Текст',
            order: '',
        };

        const result = mapShapeJsonToApi('newId012', shapeJson);
        expect(result).toEqual({
            id: 'newId012',
            tp: 'T',
            l: 70,
            t: 80,
            w: 100,
            h: 100,
            sx: 1,
            sy: 1,
            fc: 'transparent',
            sc: 'transparent',
            txt: 'Текст',
            o: '',
        });
    });
});
