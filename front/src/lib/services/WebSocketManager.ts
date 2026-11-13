import { Group as FabricGroup } from 'fabric/es';

import { canvasManager } from './CanvasManager';
import { commentBadgeManager } from './CommentBadgeManager';
import { onlineUserStoreManager } from './OnlineUserStoreManager';
import { shapeStoreManager } from './ShapeStoreManager';

import logger from './logger';
import { mapConnectionFromApi, mapShapeFromApi } from './mapper';

import { cursors, type ShapeElement, type ConnectionElement, type CursorData } from '../stores';
import { type DiagramShape, type DiagramConnection, type OnlineUser } from '../ApiSpec';

type shapeId = string;

type User = {
    id: string;
    username: string;
};

type CursorPosition = {
    userId: string;
    username: string;
    x: number;
    y: number;
};

type MessageType =
    | 'ws:connected'
    | 'ws:ping'
    | 'ws:pong'
    | 'shape:upserted'
    | 'shape:deleted'
    | 'connection:upserted'
    | 'connection:deleted'
    | 'comment:added'
    | 'cursor:move'
    | 'user:connected'
    | 'user:disconnected';

type MessageData = DiagramShape | DiagramConnection | shapeId | User | CursorPosition | null;

interface Message {
    type: MessageType;
    diagramId: string;
    timestamp: number;
    data: MessageData;
}

export interface WebSocketManagerDependencies {
    wsFactory: (url: string) => WebSocket;
    getBaseUrl: () => string;
    upsertShape: (shape: ShapeElement) => FabricGroup | null;
    removeShape: (group: FabricGroup) => boolean;
    upsertConnection: (conn: ConnectionElement) => void;
    removeConnection: (conn: ConnectionElement) => void;
    addOnlineUser: (user: OnlineUser) => void;
    removeOnlineUser: (id: string) => void;
    getShapeById: (id: string) => ShapeElement | undefined;
    mapShapeFromApi: (shape: DiagramShape) => ShapeElement;
    mapConnectionFromApi: (conn: DiagramConnection) => ConnectionElement;
    updateCursors: (fn: (cursors: CursorData[]) => CursorData[]) => void;
    commentBadgeUpsert: (group: FabricGroup, count: number, increment: boolean) => void;
}

const defaultDeps: WebSocketManagerDependencies = {
    wsFactory: (url) => new WebSocket(url),
    getBaseUrl: () => {
        if (import.meta.env.DEV) {
            return import.meta.env.VITE_WEBSOCKET_BASE_URL;
        }

        return `wss://${location.host}/ws`;
    },
    upsertShape: (shape) => canvasManager.upsertShape(shape),
    removeShape: (group) => canvasManager.removeShape(group),
    upsertConnection: (conn) => canvasManager.upsertConnection(conn),
    removeConnection: (conn) => canvasManager.removeConnection(conn),
    addOnlineUser: (user) => onlineUserStoreManager.add(user),
    removeOnlineUser: (id) => onlineUserStoreManager.remove(id),
    getShapeById: (id) => shapeStoreManager.getById(id),
    mapShapeFromApi: mapShapeFromApi,
    mapConnectionFromApi: mapConnectionFromApi,
    updateCursors: (fn) => cursors.update(fn),
    commentBadgeUpsert: (group, count, increment) => commentBadgeManager.upsert(group, count, increment),
};

export class WebSocketManager {
    private currentWebSocket: WebSocket | null = null;

    private isManualClose = false;

    private logPrefix = '[WebSocket]';

    // Reconnect settings
    private attemptCount = 0;
    private maxAttempts = 10;
    private initialDelay = 1000;
    private maxDelay = 30000;
    private backoffFactor = 2;

    // Heartbeat settings
    private heartbeatInterval = 60000; // 60 seconds
    private heartbeatTimer: number | null = null;
    private pongTimeout = 10000; // 10 seconds
    private pongTimer: number | null = null;

    private readonly deps: WebSocketManagerDependencies;

    constructor(deps: WebSocketManagerDependencies = defaultDeps) {
        this.deps = deps;
    }

    public connect(diagramID: string | null) {
        logger.info(`${this.logPrefix} Connecting to diagram ${diagramID}`);

        // Закрываем предыдущее соединение
        if (this.currentWebSocket) {
            this.disconnect();
        }

        if (!diagramID) {
            logger.error(`${this.logPrefix} No diagram ID provided`);
            return;
        }

        this.dial(diagramID);
    }

    public disconnect() {
        logger.info(`${this.logPrefix} Disconnecting from diagram`);

        this.stopHeartbeat();

        if (this.currentWebSocket) {
            this.isManualClose = true;
            this.currentWebSocket.close(1000, 'Manual disconnect');
            this.currentWebSocket = null;
        }
    }

    public sendCursorMove(diagramId: string, position: CursorPosition) {
        this.sendMessage('cursor:move', diagramId, Date.now(), position);
    }

    private dial(diagramID: string) {
        const ws = this.deps.wsFactory(`${this.deps.getBaseUrl()}?diagramID=${diagramID}`);
        this.currentWebSocket = ws;

        ws.onopen = (_ev) => {
            this.wsOnOpen(diagramID);
        };

        ws.onclose = (ev) => {
            this.wsOnClose(ev, diagramID);
        };

        ws.onmessage = (ev) => {
            this.wsOnMessage(ev, diagramID);
        };
    }

    private wsOnOpen = (diagramID: string) => {
        logger.info(`${this.logPrefix} OnOpen to diagram ${diagramID}`);
        this.isManualClose = false;
        this.attemptCount = 0;
        this.startHeartbeat(diagramID);
    };

    private wsOnClose = (ev: CloseEvent, diagramID: string) => {
        this.currentWebSocket = null;
        this.stopHeartbeat();
        logger.info(`${this.logPrefix} Disconnected from diagram ${diagramID}, code: ${ev.code}, reason: ${ev.reason}`);

        if (!this.isManualClose && ev.code !== 1000 && ev.code !== 1001) {
            if (this.attemptCount >= this.maxAttempts) {
                logger.error(`${this.logPrefix} Max attempts (${this.maxAttempts}) reached for diagram ${diagramID}`);
                return;
            }
            this.attemptCount++;

            // Вычисляем задержку с экспоненциальным откатом
            let delay = Math.min(
                this.initialDelay * Math.pow(this.backoffFactor, this.attemptCount - 1),
                this.maxDelay,
            );

            // Добавляем случайность от -20% до +20%
            delay = Math.round(delay * (0.8 + Math.random() * 0.4));

            logger.info(
                `Reconnecting to diagram ${diagramID} with attempt ${this.attemptCount} in ${Math.round(delay / 1000)}s`,
            );
            setTimeout(() => {
                this.dial(diagramID);
            }, delay);
        }
    };

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    private wsOnMessage = (ev: MessageEvent<any>, diagramID: string) => {
        const msg = JSON.parse(ev.data) as Message;

        if (msg.diagramId !== diagramID) {
            return;
        }

        switch (msg.type) {
            case 'ws:connected': {
                logger.info(`${this.logPrefix} Connected to diagram ${msg.diagramId}`);
                break;
            }

            case 'ws:ping': {
                this.sendMessage('ws:pong', msg.diagramId, Date.now(), null);
                break;
            }

            case 'ws:pong': {
                if (this.pongTimer) {
                    clearTimeout(this.pongTimer);
                    this.pongTimer = null;
                }
                break;
            }

            case 'shape:upserted': {
                const shape = msg.data as DiagramShape;
                logger.log(`${this.logPrefix} Shape upserted`, shape);
                this.deps.upsertShape(this.deps.mapShapeFromApi(shape));
                break;
            }

            case 'shape:deleted': {
                const shapeId = msg.data as shapeId;
                logger.log(`${this.logPrefix} Shape deleted`, shapeId);
                this.shapeDeleted(shapeId);
                break;
            }

            case 'connection:upserted': {
                const connection = msg.data as DiagramConnection;
                logger.log(`${this.logPrefix} Connection upserted`, connection);
                this.deps.upsertConnection(this.deps.mapConnectionFromApi(connection));
                break;
            }
            case 'connection:deleted': {
                const connection = msg.data as DiagramConnection;
                logger.log(`${this.logPrefix} Connection deleted`, connection);
                this.deps.removeConnection(this.deps.mapConnectionFromApi(connection));
                break;
            }

            case 'comment:added': {
                const shapeId = msg.data as shapeId;
                logger.log(`${this.logPrefix} Comment added`, shapeId);
                this.commentAdded(shapeId);
                break;
            }

            case 'cursor:move': {
                const cursorData = msg.data as CursorPosition;
                this.cursorMove(cursorData);
                break;
            }

            case 'user:connected': {
                const user = msg.data as User;
                this.deps.addOnlineUser({
                    id: user.id,
                    userName: user.username,
                });
                break;
            }

            case 'user:disconnected': {
                const user = msg.data as User;
                logger.info(`${this.logPrefix} User disconnected`, user);
                this.deps.removeOnlineUser(user.id);
                break;
            }

            default: {
                logger.error(`${this.logPrefix} Unknown message type: ${msg.type}`);
            }
        }
    };

    private shapeDeleted(shapeId: string) {
        const shape = this.deps.getShapeById(shapeId);
        if (!shape) {
            return;
        }

        this.deps.removeShape(shape.fabricGroup as FabricGroup);
    }

    private commentAdded(shapeId: string) {
        const shape = this.deps.getShapeById(shapeId);
        if (!shape) {
            return;
        }

        this.deps.commentBadgeUpsert(shape.fabricGroup as FabricGroup, 1, true);
    }

    private cursorMove(cursorData: CursorPosition) {
        this.deps.updateCursors((currentCursors) => {
            const existingIndex = currentCursors.findIndex((c) => c.user.id === cursorData.userId);
            const cursorInfo = {
                user: { id: cursorData.userId, userName: cursorData.username },
                x: cursorData.x,
                y: cursorData.y,
            };

            if (existingIndex >= 0) {
                currentCursors[existingIndex] = cursorInfo;
            } else {
                currentCursors.push(cursorInfo);
            }

            return [...currentCursors];
        });
    }

    private sendMessage(type: MessageType, diagramId: string, timestamp: number, data: MessageData) {
        if (!this.currentWebSocket || this.currentWebSocket.readyState !== WebSocket.OPEN) {
            return;
        }

        const msg: Message = {
            type,
            diagramId,
            timestamp,
            data,
        };

        this.currentWebSocket.send(JSON.stringify(msg));
    }

    private startHeartbeat(diagramId: string) {
        this.stopHeartbeat();

        this.heartbeatTimer = window.setInterval(() => {
            this.sendMessage('ws:ping', diagramId, Date.now(), null);

            // Set timeout for pong
            this.pongTimer = window.setTimeout(() => {
                logger.warn(`${this.logPrefix} Pong timeout, reconnecting`);
                this.disconnect();
                this.connect(diagramId);
            }, this.pongTimeout);
        }, this.heartbeatInterval);
    }

    private stopHeartbeat() {
        if (this.heartbeatTimer) {
            clearInterval(this.heartbeatTimer);
            this.heartbeatTimer = null;
        }
        if (this.pongTimer) {
            clearTimeout(this.pongTimer);
            this.pongTimer = null;
        }
    }
}

export const webSocketManager = new WebSocketManager();
