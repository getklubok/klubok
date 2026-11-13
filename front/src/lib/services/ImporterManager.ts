import { ulid } from 'ulid';

import { api } from '../Api';
import type { DiagramShape, DiagramConnection } from '../ApiSpec';

import type { DiagramJson, RespDataMsg, ShapeJson } from '../types';
import { mapConnectionToApi, mapShapeJsonToApi } from './mapper';
import logger from './logger';
import { alert } from '../utils/alert';

export class ImporterManager {
    public fromFile(file: File): Promise<void> {
        return new Promise((resolve, reject) => {
            const reader = new FileReader();

            reader.onload = (e) => {
                try {
                    const content = e.target?.result as string;
                    const data = JSON.parse(content);
                    if (data.source !== 'https://github.com/getklubok/klubok' || data.type !== 'klubok') {
                        throw new Error('Invalid file format');
                    }

                    this.importData(data);

                    resolve();
                } catch (error) {
                    logger.mustError('Failed to import file:', error);
                    reject(error);
                }
            };

            reader.onerror = () => {
                reject(new Error('Failed to read file'));
            };

            reader.readAsText(file);
        });
    }

    private async importData(data: DiagramJson) {
        const diagramShapes: DiagramShape[] = [];
        const diagramConnections: DiagramConnection[] = [];

        // Create ID mapping (numeric ID from file -> new string ID)
        const idMap = new Map<number, string>();

        // First pass: create shape elements with new IDs
        if (data.shapes && Array.isArray(data.shapes)) {
            data.shapes.forEach((shape: ShapeJson, _index: number) => {
                const newId = ulid().toString();
                idMap.set(shape.id, newId);
                diagramShapes.push(mapShapeJsonToApi(newId, shape));
            });
        }

        // Load connections
        if (data.connections && Array.isArray(data.connections)) {
            data.connections.forEach((conn: [number, number]) => {
                const fromId = idMap.get(conn[0]);
                const toId = idMap.get(conn[1]);
                if (fromId && toId) {
                    diagramConnections.push(
                        mapConnectionToApi({
                            fromShapeId: fromId,
                            toShapeId: toId,
                        }),
                    );
                }
            });
        }

        try {
            const response = await api.diagramService.diagramImport({
                title: data.title,
                shapes: diagramShapes,
                connections: diagramConnections,
            });
            window.location.hash = `/diagrams/${response.diagram.id}`;
        } catch (err) {
            alert({
                title: 'Ошибка импортирования диаграммы',
                message: (err as RespDataMsg)?.response?.data?.message || (err as Error).message,
                type: 'error',
            });
        }
    }
}

export const importerManager = new ImporterManager();
