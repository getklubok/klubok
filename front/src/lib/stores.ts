import { Group as FabricGroup } from 'fabric/es';
import { writable, derived } from 'svelte/store';
import type { OnlineUser, DiagramCollaboratorPermissionEnum, UserMeResponse } from './ApiSpec';
import { api } from './Api';

export interface ICoordinates {
    x: number;
    y: number;
    zoom: number;
}

export interface IBaseElement {
    id: string;
    type: ElementType;
    left: number;
    top: number;
    width: number;
    height: number;
    fill: string;
    stroke: string;
    strokeWidth: number;
    scaleX: number;
    scaleY: number;
    text: string;
    order: string;
    fabricGroup?: FabricGroup;
}

export interface ISquareElement extends IBaseElement {
    type: 'square';
}

export interface IEllipseElement extends IBaseElement {
    type: 'ellipse';
    rx: number;
    ry: number;
}

export interface IDiamondElement extends IBaseElement {
    type: 'diamond';
}

export interface IBoxElement extends IBaseElement {
    type: 'box';
}

export interface ITextElement extends IBaseElement {
    type: 'text';
}

export interface IDebugObjectGroup {
    left: number;
    top: number;
    width: number;
    height: number;
    scaleX: number;
    scaleY: number;
}

export interface IDebugObjectShape {
    left: number;
    top: number;
    width: number;
    height: number;
    scaleX: number;
    scaleY: number;
    rx: number | null;
    ry: number | null;
}

export interface IDebugObject {
    name: string;
    type: string;
    order: string;
    object: FabricGroup;
    group: IDebugObjectGroup;
    shape: IDebugObjectShape;
}

// Типы для элементов фигур
export type ElementType = 'square' | 'ellipse' | 'diamond' | 'box' | 'text';

export type ShapeElement =
    IBaseElement | ISquareElement | IEllipseElement | IDiamondElement | IBoxElement | ITextElement;

export type ConnectionElement = {
    fromShapeId: string;
    toShapeId: string;
};

export type Connection = {
    fromElement: FabricGroup;
    toElement: FabricGroup;
    connectionObject: FabricGroup;
};

export type CursorData = {
    user: OnlineUser;
    x: number;
    y: number;
};

export type DrawingMode = 'select' | ElementType | 'arrow';

export type ContextMenuState = {
    visible: boolean;
    x: number;
    y: number;
    target: FabricGroup | null;
};

// Store для текущего режима рисования
export const drawingMode = writable<DrawingMode>('select');

// Store для текущих координат и zoom
export const coordinates = writable<ICoordinates>({
    x: 0,
    y: 0,
    zoom: 1,
});

// Хранилище данных фигур
export const shapes = writable<ShapeElement[]>([]);

// Хранилище связей для отслеживания от каких элементов к каким идут стрелки
export const connections = writable<Connection[]>([]);

// Store для активной диаграммы
export const diagramTitle = writable<string | null>(null);
export const diagramPermission = writable<DiagramCollaboratorPermissionEnum | null>(null);

// Store для комментариев
export const commentsShow = writable<boolean>(false);
export const commentsSelectedGroup = writable<FabricGroup | null>(null);

// Store для статуса синхронизации
export type SyncStatus = 'idle' | 'saving' | 'saved' | 'error';
export const syncStatus = writable<SyncStatus>('idle');

// Store для текущего пользователя
export const currentUser = writable<UserMeResponse | null>(null);

// Store для онлайн пользователей
export const onlineUsers = writable<OnlineUser[]>([]);

// Store для курсоров других пользователей
export const cursors = writable<CursorData[]>([]);

// Store для выбранного объекта
export const selectedObject = writable<IDebugObject | null>(null);

export const contextMenu = writable<ContextMenuState>({
    visible: false,
    x: 0,
    y: 0,
    target: null,
});

// Store фича-флагов: загружается один раз при старте приложения.
export interface FeaturesState {
    captchaEnabled: boolean;
    captchaSiteKey: string;
    authEmailEnabled: boolean;
    authSsoEnabled: boolean;
    loaded: boolean;
}

export const features = writable<FeaturesState>({
    captchaEnabled: false,
    captchaSiteKey: '',
    authEmailEnabled: true,
    authSsoEnabled: false,
    loaded: false,
});

// Store для конфигурации smart-captcha: публичный ключ приходит из API фич,
// чтобы не вшивать его в сборку.
export const smartCaptchaConfig = derived(features, ($features) => ({
    sitekey: $features.captchaSiteKey,
    hl: 'ru',
}));

let featuresLoadPromise: Promise<FeaturesState> | null = null;

export function loadFeaturesOnce(): Promise<FeaturesState> {
    if (featuresLoadPromise) return featuresLoadPromise;
    featuresLoadPromise = api.rootService
        .features()
        .then((f) => {
            const state: FeaturesState = {
                captchaEnabled: Boolean(f?.captcha?.enabled),
                captchaSiteKey: f?.captcha?.siteKey ?? '',
                authEmailEnabled: f?.authEmail?.enabled ?? true,
                authSsoEnabled: f?.authSso?.enabled ?? false,
                loaded: true,
            };
            features.set(state);
            return state;
        })
        .catch(() => {
            const fallback: FeaturesState = {
                captchaEnabled: false,
                captchaSiteKey: '',
                authEmailEnabled: true,
                authSsoEnabled: false,
                loaded: true,
            };
            features.set(fallback);
            return fallback;
        });
    return featuresLoadPromise;
}

export function clearState() {
    drawingMode.set('select');
    coordinates.set({ x: 0, y: 0, zoom: 1 });
    shapes.set([]);
    connections.set([]);
    commentsShow.set(false);
    commentsSelectedGroup.set(null);
    onlineUsers.set([]);
    cursors.set([]);
    selectedObject.set(null);
    contextMenu.set({ visible: false, x: 0, y: 0, target: null });
}
