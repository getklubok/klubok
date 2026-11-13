import { get } from 'svelte/store';
import {
    type ShapeElement,
    type ConnectionElement,
    type SyncStatus,
    type Connection,
    syncStatus,
    diagramPermission,
    shapes,
    connections,
} from '../stores';

import logger from './logger';
import hasPermissionFn from './hasPermission';
import { mapConnectionToApi as mapConnectionToApiFn, mapShapeToApi as mapShapeToApiFn } from './mapper';

import { api } from '../Api';
import type { DiagramCollaboratorPermissionEnum, DiagramShape, DiagramConnection } from '../ApiSpec';

type ShapeOperationUpsert = {
    type: 'shapeUpsert';
    shape: ShapeElement;
};

type ShapeOperationDelete = {
    type: 'shapeDelete';
    shapeId: string;
};

type ConnectionOperationUpsert = {
    type: 'connectionUpsert';
    connection: ConnectionElement;
};

type ConnectionOperationDelete = {
    type: 'connectionDelete';
    fromShapeId: string;
    toShapeId: string;
};

export type Operation =
    ShapeOperationUpsert | ShapeOperationDelete | ConnectionOperationUpsert | ConnectionOperationDelete;

export interface AutoSaveDependencies {
    getShapes: () => ShapeElement[];
    getConnections: () => Connection[];
    getPermission: () => DiagramCollaboratorPermissionEnum | null;
    setSyncStatus: (status: SyncStatus) => void;
    updateSyncStatus: (fn: (status: SyncStatus) => SyncStatus) => void;
    hasPermission: (permission: DiagramCollaboratorPermissionEnum | null | undefined) => boolean;
    mapShapeToApi: (shape: ShapeElement) => DiagramShape;
    mapConnectionToApi: (conn: ConnectionElement) => DiagramConnection;
    shapeUpsert: (diagramId: string, body: { shape: DiagramShape }) => Promise<unknown>;
    shapeBatchUpsert: (diagramId: string, body: { shapes: DiagramShape[] }) => Promise<unknown>;
    shapeDelete: (diagramId: string, shapeId: string) => Promise<unknown>;
    connectionUpsert: (diagramId: string, body: { connection: DiagramConnection }) => Promise<unknown>;
    connectionBatchUpsert: (diagramId: string, body: { connections: DiagramConnection[] }) => Promise<unknown>;
    connectionDelete: (diagramId: string, fromShapeId: string, toShapeId: string) => Promise<unknown>;
}

const defaultDeps: AutoSaveDependencies = {
    getShapes: () => get(shapes),
    getConnections: () => get(connections),
    getPermission: () => get(diagramPermission),
    setSyncStatus: (status) => syncStatus.set(status),
    updateSyncStatus: (fn) => syncStatus.update(fn),
    hasPermission: hasPermissionFn,
    mapShapeToApi: mapShapeToApiFn,
    mapConnectionToApi: mapConnectionToApiFn,
    shapeUpsert: (diagramId, body) => api.diagramService.shapeUpsert(diagramId, body),
    shapeBatchUpsert: (diagramId, body) => api.diagramService.shapeBatchUpsert(diagramId, body),
    shapeDelete: (diagramId, shapeId) => api.diagramService.shapeDelete(diagramId, shapeId),
    connectionUpsert: (diagramId, body) => api.diagramService.connectionUpsert(diagramId, body),
    connectionBatchUpsert: (diagramId, body) => api.diagramService.connectionBatchUpsert(diagramId, body),
    connectionDelete: (diagramId, fromShapeId, toShapeId) =>
        api.diagramService.connectionDelete(diagramId, fromShapeId, toShapeId),
};

export class AutoSaveService {
    private currentDiagramId: string | null = null;
    private debounceTimer: ReturnType<typeof setTimeout> | null = null;
    private lastSavedShapes: ShapeElement[] = [];
    private lastSavedConnections: ConnectionElement[] = [];

    private readonly debounceDelay = 2000;
    private readonly errorDelay = 5000;
    private readonly requestDelay = 100;

    private readonly deps: AutoSaveDependencies;

    constructor(deps: AutoSaveDependencies = defaultDeps) {
        this.deps = deps;
    }

    setDiagramId(diagramId: string) {
        this.currentDiagramId = diagramId;
    }

    setInitialState(shapes: ShapeElement[], connections: ConnectionElement[]) {
        this.lastSavedShapes = shapes;
        this.lastSavedConnections = connections;
    }

    triggerAutoSave() {
        if (!this.currentDiagramId) return;
        if (!this.deps.hasPermission(this.deps.getPermission())) return;

        if (this.debounceTimer) {
            clearTimeout(this.debounceTimer);
        }

        this.deps.updateSyncStatus((status) => (status === 'saving' ? status : 'saving'));

        this.debounceTimer = setTimeout(() => {
            this.performSave();
        }, this.debounceDelay);
    }

    generateOperations(currentShapes: ShapeElement[], currentConnections: ConnectionElement[]): Operation[] {
        const operations: Operation[] = [];
        const shapeMap = new Map(this.lastSavedShapes.map((s) => [s.id, s]));
        const connectionMap = new Map(this.lastSavedConnections.map((c) => [`${c.fromShapeId}-${c.toShapeId}`, c]));

        const currentShapeIds = new Set(currentShapes.map((s) => s.id));
        for (const shape of currentShapes) {
            const existing = shapeMap.get(shape.id);
            if (!existing || JSON.stringify(existing) !== JSON.stringify(shape)) {
                operations.push({ type: 'shapeUpsert', shape });
            }
        }
        for (const savedShape of this.lastSavedShapes) {
            if (!currentShapeIds.has(savedShape.id)) {
                operations.push({ type: 'shapeDelete', shapeId: savedShape.id });
            }
        }

        const currentConnectionsSet = new Set(currentConnections.map((c) => `${c.fromShapeId}-${c.toShapeId}`));
        for (const conn of currentConnections) {
            if (!connectionMap.has(`${conn.fromShapeId}-${conn.toShapeId}`)) {
                operations.push({ type: 'connectionUpsert', connection: conn });
            }
        }
        for (const savedConn of this.lastSavedConnections) {
            const key = `${savedConn.fromShapeId}-${savedConn.toShapeId}`;
            if (!currentConnectionsSet.has(key)) {
                operations.push({
                    type: 'connectionDelete',
                    fromShapeId: savedConn.fromShapeId,
                    toShapeId: savedConn.toShapeId,
                });
            }
        }

        return operations;
    }

    private async performSave() {
        if (!this.currentDiagramId) return;
        if (!this.deps.hasPermission(this.deps.getPermission())) return;

        try {
            const currentShapes = this.deps.getShapes();

            const connectionsValue = this.deps.getConnections();
            const currentConnections: ConnectionElement[] = [];
            for (const conn of connectionsValue) {
                if (conn.fromElement.name && conn.toElement.name) {
                    currentConnections.push({
                        fromShapeId: conn.fromElement.name,
                        toShapeId: conn.toElement.name,
                    });
                }
            }

            const operations = this.generateOperations(currentShapes, currentConnections);

            if (operations.length === 0) {
                this.deps.setSyncStatus('saved');
                setTimeout(() => {
                    this.deps.setSyncStatus('idle');
                }, this.debounceDelay);
                return;
            }

            // Group operations by type
            const shapesToUpsert = operations
                .filter((op): op is ShapeOperationUpsert => op.type === 'shapeUpsert')
                .map((op) => this.deps.mapShapeToApi(op.shape));

            const connectionsToUpsert = operations
                .filter((op): op is ConnectionOperationUpsert => op.type === 'connectionUpsert')
                .map((op) => this.deps.mapConnectionToApi(op.connection));

            const shapesToDelete = operations
                .filter((op): op is ShapeOperationDelete => op.type === 'shapeDelete')
                .map((op) => op.shapeId);

            const connectionsToDelete = operations
                .filter((op): op is ConnectionOperationDelete => op.type === 'connectionDelete')
                .map((op) => ({ fromShapeId: op.fromShapeId, toShapeId: op.toShapeId }));

            // Batch upsert for shapes
            if (shapesToUpsert.length > 0) {
                await this.deps.shapeBatchUpsert(this.currentDiagramId, { shapes: shapesToUpsert });
            }

            // Batch upsert for connections
            if (connectionsToUpsert.length > 0) {
                await this.deps.connectionBatchUpsert(this.currentDiagramId, {
                    connections: connectionsToUpsert,
                });
            }

            // Individual deletes (for now without batch delete)
            for (const shapeId of shapesToDelete) {
                await this.deps.shapeDelete(this.currentDiagramId, shapeId);
                await new Promise((resolve) => {
                    setTimeout(resolve, this.requestDelay);
                });
            }
            for (const conn of connectionsToDelete) {
                await this.deps.connectionDelete(this.currentDiagramId, conn.fromShapeId, conn.toShapeId);
                await new Promise((resolve) => {
                    setTimeout(resolve, this.requestDelay);
                });
            }

            this.lastSavedShapes = currentShapes;
            this.lastSavedConnections = currentConnections;

            this.deps.setSyncStatus('saved');
            logger.log('Auto save successful');

            setTimeout(() => {
                this.deps.setSyncStatus('idle');
            }, this.debounceDelay);
        } catch (error) {
            logger.mustError('Auto save failed:', error);
            this.deps.setSyncStatus('error');

            setTimeout(() => {
                this.deps.setSyncStatus('idle');
            }, this.errorDelay);
        }
    }

    dispose() {
        if (this.debounceTimer) {
            clearTimeout(this.debounceTimer);
            this.debounceTimer = null;
        }
    }
}

export const autoSaveService = new AutoSaveService();
