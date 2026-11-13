import { type Connection, connections } from '../stores';
import { get } from 'svelte/store';
import { Group as FabricGroup } from 'fabric/es';

class ConnectionStoreManager {
    public add(from: FabricGroup, to: FabricGroup, connection: FabricGroup) {
        connections.update((conn) => [
            ...conn,
            {
                fromElement: from,
                toElement: to,
                connectionObject: connection,
            },
        ]);
    }

    public getByShapeId(shapeId: string): Connection[] {
        return get(connections).filter((c) => c.fromElement.name === shapeId || c.toElement.name === shapeId);
    }

    public getByShapeIds(fromShapeId: string, toShapeId: string): Connection | null {
        const conns = get(connections).filter(
            (c) =>
                (c.fromElement.name === fromShapeId && c.toElement.name === toShapeId) ||
                (c.fromElement.name === toShapeId && c.toElement.name === fromShapeId),
        );
        if (conns.length === 0) return null;

        return conns[0];
    }

    public getByConnectionObject(obj: FabricGroup): Connection | undefined {
        return get(connections).find((c) => c.connectionObject === obj);
    }

    public removeByShapeId(shapeId: string) {
        connections.update((conn) =>
            conn.filter((c) => c.fromElement.name !== shapeId && c.toElement.name !== shapeId),
        );
    }

    public removeByShapeIds(fromShapeId: string, toShapeId: string) {
        connections.update((conn) =>
            conn.filter(
                (c) =>
                    !(
                        (c.fromElement.name === fromShapeId && c.toElement.name === toShapeId) ||
                        (c.fromElement.name === toShapeId && c.toElement.name === fromShapeId)
                    ),
            ),
        );
    }

    public removeByConnectionObject(obj: FabricGroup) {
        connections.update((conn) => conn.filter((c) => c.connectionObject !== obj));
    }
}

export const connectionStoreManager = new ConnectionStoreManager();
