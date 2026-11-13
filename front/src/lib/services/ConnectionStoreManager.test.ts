import { describe, it, expect, beforeEach } from 'vitest';
import { get } from 'svelte/store';
import { type Connection, connections } from '../stores';
import { connectionStoreManager } from './ConnectionStoreManager';
import { Group as FabricGroup } from 'fabric/es';

function mockGroup(name: string): FabricGroup {
    return { name } as unknown as FabricGroup;
}

function makeConnection(fromName: string, toName: string): Connection {
    return {
        fromElement: mockGroup(fromName),
        toElement: mockGroup(toName),
        connectionObject: mockGroup(`conn-${fromName}-${toName}`),
    };
}

describe('ConnectionStoreManager', () => {
    beforeEach(() => {
        connections.set([]);
    });

    describe('add', () => {
        it('должен добавлять соединение в пустой store', () => {
            connectionStoreManager.add(mockGroup('A'), mockGroup('B'), mockGroup('c1'));
            const result = get(connections);
            expect(result).toHaveLength(1);
            expect(result[0].fromElement.name).toBe('A');
            expect(result[0].toElement.name).toBe('B');
        });

        it('должен добавлять несколько соединений', () => {
            connectionStoreManager.add(mockGroup('A'), mockGroup('B'), mockGroup('c1'));
            connectionStoreManager.add(mockGroup('C'), mockGroup('D'), mockGroup('c2'));
            expect(get(connections)).toHaveLength(2);
        });
    });

    describe('getByShapeId', () => {
        it('должен находить соединение по fromElement.name', () => {
            connections.set([makeConnection('A', 'B')]);
            const result = connectionStoreManager.getByShapeId('A');
            expect(result).toHaveLength(1);
            expect(result[0].fromElement.name).toBe('A');
        });

        it('должен находить соединение по toElement.name', () => {
            connections.set([makeConnection('A', 'B')]);
            const result = connectionStoreManager.getByShapeId('B');
            expect(result).toHaveLength(1);
            expect(result[0].toElement.name).toBe('B');
        });

        it('должен возвращать несколько совпадений', () => {
            connections.set([makeConnection('A', 'B'), makeConnection('A', 'C')]);
            const result = connectionStoreManager.getByShapeId('A');
            expect(result).toHaveLength(2);
        });

        it('должен возвращать пустой массив если не найдено', () => {
            connections.set([makeConnection('A', 'B')]);
            const result = connectionStoreManager.getByShapeId('Z');
            expect(result).toEqual([]);
        });
    });

    describe('getByShapeIds', () => {
        it('должен находить соединение по fromShapeId и toShapeId', () => {
            connections.set([makeConnection('A', 'B')]);
            const result = connectionStoreManager.getByShapeIds('A', 'B');
            expect(result).not.toBeNull();
            expect(result!.fromElement.name).toBe('A');
            expect(result!.toElement.name).toBe('B');
        });

        it('должен находить соединение в обратном порядке (to, from)', () => {
            connections.set([makeConnection('A', 'B')]);
            const result = connectionStoreManager.getByShapeIds('B', 'A');
            expect(result).not.toBeNull();
            expect(result!.fromElement.name).toBe('A');
            expect(result!.toElement.name).toBe('B');
        });

        it('должен возвращать null если не найдено', () => {
            connections.set([makeConnection('A', 'B')]);
            const result = connectionStoreManager.getByShapeIds('X', 'Y');
            expect(result).toBeNull();
        });

        it('должен возвращать конкретное соединение', () => {
            connections.set([makeConnection('A', 'B'), makeConnection('A', 'C')]);
            const result = connectionStoreManager.getByShapeIds('A', 'C');
            expect(result).not.toBeNull();
            expect(result!.fromElement.name).toBe('A');
            expect(result!.toElement.name).toBe('C');
        });
    });

    describe('removeByShapeId', () => {
        it('должен удалять когда совпадает fromElement.name', () => {
            connections.set([makeConnection('A', 'B'), makeConnection('C', 'D')]);
            connectionStoreManager.removeByShapeId('A');
            expect(get(connections)).toHaveLength(1);
            expect(get(connections)[0].fromElement.name).toBe('C');
        });

        it('должен удалять когда совпадает toElement.name', () => {
            connections.set([makeConnection('A', 'B'), makeConnection('C', 'D')]);
            connectionStoreManager.removeByShapeId('B');
            expect(get(connections)).toHaveLength(1);
            expect(get(connections)[0].fromElement.name).toBe('C');
        });

        it('должен удалять несколько соединений с одним shapeId', () => {
            connections.set([makeConnection('A', 'B'), makeConnection('A', 'C'), makeConnection('D', 'E')]);
            connectionStoreManager.removeByShapeId('A');
            expect(get(connections)).toHaveLength(1);
            expect(get(connections)[0].fromElement.name).toBe('D');
        });

        it('не должен ничего удалять если id не найден', () => {
            connections.set([makeConnection('A', 'B')]);
            connectionStoreManager.removeByShapeId('Z');
            expect(get(connections)).toHaveLength(1);
        });

        it('должен работать с пустым store', () => {
            connectionStoreManager.removeByShapeId('A');
            expect(get(connections)).toEqual([]);
        });
    });
});
