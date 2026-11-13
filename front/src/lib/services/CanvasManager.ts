import {
    Canvas as FabricCanvas,
    Rect as FabricRect,
    Group as FabricGroup,
    Point as FabricPoint,
    Polyline as FabricPolyline,
    FabricObject,
    FabricText,
    IText as FabricIText,
} from 'fabric/es';
import { ulid } from 'ulid';

import {
    type ShapeElement,
    type ConnectionElement,
    type CursorData,
    type DrawingMode,
    clearState,
    coordinates,
    selectedObject,
    contextMenu,
} from '../stores';
import {
    PAGE_GRID_STEP,
    PAGE_INIT_HEIGHT,
    PAGE_INIT_WIDTH,
    PAGE_PADDING,
    DEFAULT_RADIUS,
    DEFAULT_SCALE,
    DEFAULT_SIZE,
    DEFAULT_STROKE_WIDTH,
    PAGE_EXPAND_STEP,
    DEFAULT_FILL,
    DEFAULT_STROKE_SQUARE,
    DEFAULT_STROKE_ELLIPSE,
    DEFAULT_STROKE_DIAMOND,
    DEFAULT_STROKE_BOX,
    SELECTION_BORDER_DARK,
    SELECTION_BORDER_LIGHT,
} from '../constants';
import { calculateOrderKey, sortByOrderKey } from '../utils/orderKey';

import logger from './logger';
import hasPermission from './hasPermission';
import { mapConnectionFromApi, mapShapeFromApi } from './mapper';

import { fabricManager } from './FabricManager';
import { coordinatesManager } from './CoordinatesManager';
import { viewportManager } from './ViewportManager';
import { shapeStoreManager } from './ShapeStoreManager';
import { connectionStoreManager } from './ConnectionStoreManager';
import { commentBadgeManager } from './CommentBadgeManager';
import { controlsManager } from './ControlsManager';
import { webSocketManager } from './WebSocketManager';
import { autoSaveService } from './AutoSaveService';

import type { DiagramCollaboratorPermissionEnum, DiagramInfo, OnlineUser } from '../ApiSpec';
import { getBoundingBox } from '../utils/getBoundingBox';
import { isShapeGroup } from '../utils/isShapeGroup';
import { calcBoxLine, calcBoxText } from '../utils/calcBox';

type UserID = string;

type RemoteCursor = {
    user: OnlineUser;
    x: number;
    y: number;
    fabricObj: FabricGroup;
};

type PageSize = {
    width: number;
    height: number;
    offsetX: number;
    offsetY: number;
};

export interface CanvasManagerDependencies {
    createId: () => string;
    hasPermission: (permission: DiagramCollaboratorPermissionEnum | null | undefined) => boolean;
    clearState: () => void;
    setCoordinates: (x: number, y: number, zoom: number) => void;
    mapShapeFromApi: (shape: import('../ApiSpec').DiagramShape) => ShapeElement;
    mapConnectionFromApi: (conn: import('../ApiSpec').DiagramConnection) => ConnectionElement;
}

const defaultDeps: CanvasManagerDependencies = {
    createId: () => ulid().toString(),
    hasPermission,
    clearState: () => clearState(),
    setCoordinates: (x, y, zoom) => coordinatesManager.set(x, y, zoom),
    mapShapeFromApi,
    mapConnectionFromApi,
};

export class CanvasManager {
    private canvas: FabricCanvas | null = null;
    private diagramInfo: DiagramInfo | null = null;

    // Текущий режим рисования
    private currentDrawingMode: DrawingMode = 'select';

    // Переменные для панорамирования
    private isPanning: boolean = false;
    private lastPosX: number = 0;
    private lastPosY: number = 0;

    // Переменные для отслеживания курсора
    private lastCursorSendTime: number = 0;
    private cursorThrottleMs: number = 50; // отправлять не чаще 50ms TODO: подобрать оптимальное значение

    // Переменные для режима рисования стрелок
    private fromElement: FabricGroup | null = null;

    private remoteCursors: Record<UserID, RemoteCursor> = {};

    private pageBorder: FabricObject | null = null;
    private gridLines: FabricPolyline[] = [];

    // Текущая тема (true — тёмная). Обновляется в recolorAll, используется при
    // создании фона/сетки, чтобы они сразу рендерились под нужную тему.
    private isDark: boolean = false;

    private readonly deps: CanvasManagerDependencies;

    constructor(deps: CanvasManagerDependencies = defaultDeps) {
        this.deps = deps;
    }

    public init(canvas: FabricCanvas) {
        logger.log('CanvasManager init');
        this.canvas = canvas;
    }

    public reset() {
        if (!this.canvas) return;

        this.canvas = null;
        this.diagramInfo = null;
        this.currentDrawingMode = 'select';
        this.remoteCursors = {};

        contextMenu.set({
            visible: false,
            x: 0,
            y: 0,
            target: null,
        });
    }

    /**
     * Текущая тема канваса (true — тёмная). Обновляется в recolorAll(), поэтому
     * всегда отражает фактически отрисованную тему — и в основном приложении, и
     * в embed (где тема локальная и в resolvedThemeStore не пишется). Используется
     * при экспорте, чтобы восстановить тему после временной перекраски в светлую.
     */
    public get isDarkTheme(): boolean {
        return this.isDark;
    }

    /**
     * Перекрашивает все объекты канваса под заданную тему, не пересоздавая их.
     * Цвета берутся из сохранённых светлых исходников (sourceFill/sourceStroke),
     * поэтому операция обратима и идемпотентна. Собственно конвертацию выполняет
     * FabricManager.applyThemeColor. Также перекрашивает контролы выделения
     * (рамку и corner-контролы), чтобы они были видны на фоне темы.
     */
    public recolorAll(isDark: boolean): void {
        if (!this.canvas) return;

        this.isDark = isDark;

        this.applySelectionControlsTheme(this.canvas, isDark);

        this.canvas.getObjects().forEach((obj) => {
            // Группы-коннекции — перекрашиваем children (line.stroke, arrowHead.fill)
            if (obj.isConnection) {
                (obj as FabricGroup).forEachObject((child) => fabricManager.applyThemeColor(child, isDark));
                // Рамка выделения соединения рисуется активным объектом — группой,
                // поэтому красив её тоже
                fabricManager.applySelectionControlsTheme(obj, isDark);
            }
            // Обычные группы (фигура + текст + разделитель) — перекрашиваем children
            else if (obj.type === 'group') {
                (obj as FabricGroup).forEachObject((child) => fabricManager.applyThemeColor(child, isDark));
                fabricManager.applySelectionControlsTheme(obj, isDark);
            }
            // Одиночные объекты (commentBadge, remoteCursor и т.д.)
            else {
                fabricManager.applyThemeColor(obj, isDark);
            }
        });

        this.canvas.requestRenderAll();
    }

    /**
     * Перекрашивает canvas-уровневые контролы выделения под текущую тему
     * (заливка области выделения и её рамка). Corner-контролы и рамка вокруг
     * объекта — per-object, их перекрашивает FabricManager.applySelectionControlsTheme
     * в цикле по объектам ниже.
     */
    private applySelectionControlsTheme(canvas: FabricCanvas, isDark: boolean): void {
        canvas.selectionColor = isDark ? 'rgba(255,255,255,0.2)' : 'rgba(0,0,0,0.1)';
        canvas.selectionBorderColor = isDark ? SELECTION_BORDER_LIGHT : SELECTION_BORDER_DARK;
    }

    public setDiagramInfo(diagramInfo: DiagramInfo) {
        logger.log('CanvasManager setDiagramInfo');

        this.diagramInfo = diagramInfo;
        this.currentDrawingMode = 'select';
        this.remoteCursors = {};

        this.updateCanvasCursor();
        this.restoreState();
    }

    public setCurrentDrawingMode(mode: DrawingMode) {
        this.currentDrawingMode = mode;
        this.updateCanvasCursor();
    }

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    public onMouseDown = (options: any) => {
        if (!this.canvas) return;
        const event = options.e;

        // Обрабатываем панорамирование (CTRL + левая кнопка)
        if (event.ctrlKey) {
            this.isPanning = true;
            this.lastPosX = event.clientX;
            this.lastPosY = event.clientY;
            this.canvas.selection = false;
            this.canvas.defaultCursor = 'grab';
            this.canvas.hoverCursor = 'grab';
            return;
        }

        switch (this.currentDrawingMode) {
            case 'select': {
                this.canvas.selection = true;
                break;
            }

            case 'arrow': {
                if (!this.deps.hasPermission(this.diagramInfo?.permission)) return;

                // Режим рисования стрелок
                const target = options.target;
                if (!target) return;

                // Проверяем, что кликнули на блок (группу с формой и текстом)
                if (!isShapeGroup(target)) return;

                this.appendNewConnection(target as FabricGroup);
                break;
            }

            default: {
                if (!this.deps.hasPermission(this.diagramInfo?.permission)) return;

                // Получаем координаты клика относительно канваса
                const pointer = this.canvas.getScenePoint(event);

                this.appendNewShape(pointer.x, pointer.y);
                break;
            }
        }
    };

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    public onMouseMove = (options: any) => {
        if (!this.canvas) return;

        const event = options.e;

        if (this.isPanning) {
            const deltaX = event.clientX - this.lastPosX;
            const deltaY = event.clientY - this.lastPosY;

            const vpt = this.canvas.viewportTransform;
            if (!vpt) return;

            vpt[4] += deltaX;
            vpt[5] += deltaY;

            this.canvas.requestRenderAll();

            this.lastPosX = event.clientX;
            this.lastPosY = event.clientY;

            coordinatesManager.update();
        } else {
            // Отправка позиции курсора
            this.sendCursorPosition(event);
        }
    };

    public onMouseUp = () => {
        if (!this.canvas) return;

        this.isPanning = false;

        // Восстанавливаем выделение в зависимости от режима
        this.canvas.selection = this.currentDrawingMode === 'select';

        // Обновляем курсоры в зависимости от режима
        this.updateCanvasCursor();
    };

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    public onMouseWheel = (options: any) => {
        if (!this.canvas) return;

        const event = options.e;
        let zoom = this.canvas.getZoom();

        // Определяем шаг масштабирования (10%)
        const zoomStep = 0.1;

        // Определяем направление: положительный delta - уменьшение, отрицательный - увеличение
        if (event.deltaY > 0) {
            zoom -= zoomStep;
        } else {
            zoom += zoomStep;
        }

        // Ограничиваем zoom в разумных пределах
        zoom = Math.max(0.1, Math.min(5, zoom));

        // Масштабируем относительно позиции курсора
        const point = new FabricPoint(event.offsetX, event.offsetY);
        this.canvas.zoomToPoint(point, zoom);

        // Обновляем координаты после масштабирования
        coordinatesManager.update();

        // Предотвращаем прокрутку страницы
        event.preventDefault();
        event.stopPropagation();
    };

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    public onObjectMoving = (options: any) => {
        if (!this.canvas) return;
        if (!isShapeGroup(options.target)) return;

        const obj = options.target as FabricGroup;
        this.snapToGridObj(obj, 'moving');
        this.updateDebug(obj);
        this.updateConnectionByShapeId(obj.name || '');
        commentBadgeManager.updatePosition(obj);
    };

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    public onObjectScaling = (options: any) => {
        if (!this.canvas) return;
        if (!isShapeGroup(options.target)) return;

        const obj = options.target as FabricGroup;
        this.snapToGridObj(obj, 'scaling');
        this.updateDebug(obj);
        this.updateConnectionByShapeId(obj.name || '');
        commentBadgeManager.updatePosition(obj);
    };

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    public onObjectModified = (options: any) => {
        if (!this.canvas) return;
        if (!isShapeGroup(options.target)) return;

        const obj = options.target as FabricGroup;
        this.updateDebug(obj);
        this.updateConnectionByShapeId(obj.name || '');
        commentBadgeManager.updatePosition(obj);
        this.syncLayers();

        // Обновляем данные в store
        const shape = shapeStoreManager.getById(obj.name || '');
        if (shape) {
            shapeStoreManager.update({
                ...shape,
                left: obj.left || shape.left,
                top: obj.top || shape.top,
                width: obj.width || shape.width,
                height: obj.height || shape.height,
                scaleX: obj.scaleX || shape.scaleX,
                scaleY: obj.scaleY || shape.scaleY,
            });
        }
    };

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    public onContextMenu = (options: any) => {
        if (!this.canvas) return;
        if (!isShapeGroup(options.target)) return;

        this.updateDebug(options.target);
        this.canvas.setActiveObject(options.target);

        const event = options.e as MouseEvent;
        contextMenu.set({
            visible: true,
            x: event.clientX,
            y: event.clientY,
            target: options.target,
        });

        this.canvas.requestRenderAll();
    };

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    public onSelectionCreated = (options: any) => {
        selectedObject.set(null);
        if (!isShapeGroup(options.selected?.[0])) return;
        this.updateDebug(options.selected?.[0]);
    };

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    public onSelectionUpdated = (options: any) => {
        selectedObject.set(null);
        if (!isShapeGroup(options.selected?.[0])) return;
        this.updateDebug(options.selected?.[0]);
    };

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    public onSelectionCleared = (_options: any) => {
        selectedObject.set(null);
    };

    public upsertShape(upsertedShape: ShapeElement): FabricGroup | null {
        const existShape = shapeStoreManager.getById(upsertedShape.id);
        if (existShape) {
            const fabricObj = existShape.fabricGroup as FabricGroup;
            fabricObj.set({
                order: upsertedShape.order,
                left: upsertedShape.left,
                top: upsertedShape.top,
                scaleX: upsertedShape.scaleX,
                scaleY: upsertedShape.scaleY,
                strokeWidth: DEFAULT_STROKE_WIDTH,
            });

            const textObj = fabricObj.getObjects()[1] as FabricText;
            textObj.set({
                text: upsertedShape.text,
                scaleX: DEFAULT_SCALE / (fabricObj.scaleX || DEFAULT_SCALE),
                scaleY: DEFAULT_SCALE / (fabricObj.scaleY || DEFAULT_SCALE),
            });

            this.updateConnectionByShapeId(existShape.id);
            commentBadgeManager.updatePosition(fabricObj);

            shapeStoreManager.update({
                ...existShape,
                order: upsertedShape.order,
                left: upsertedShape.left,
                top: upsertedShape.top,
                scaleX: upsertedShape.scaleX,
                scaleY: upsertedShape.scaleY,
                text: upsertedShape.text,
            });

            return fabricObj;
        }

        return this.addShape(upsertedShape);
    }

    public removeShape(shape: FabricGroup): boolean {
        if (!this.canvas) return false;

        commentBadgeManager.remove(shape);
        shapeStoreManager.remove(shape);
        this.removeConnectionByShapeId(shape.name || '');
        this.canvas.remove(shape);
        this.canvas.requestRenderAll();

        return true;
    }

    public upsertConnection(conn: ConnectionElement) {
        const existConn = connectionStoreManager.getByShapeIds(conn.fromShapeId, conn.toShapeId);
        if (existConn) {
            // Connection уже есть, обновлять нечего
            return;
        }

        const fromShape = shapeStoreManager.getById(conn.fromShapeId);
        const toShape = shapeStoreManager.getById(conn.toShapeId);
        if (fromShape?.fabricGroup && toShape?.fabricGroup) {
            this.addConnection(fromShape.fabricGroup, toShape.fabricGroup);
        }
    }

    public removeConnection(conn: ConnectionElement) {
        if (!this.canvas) return;

        const targetConn = connectionStoreManager.getByShapeIds(conn.fromShapeId, conn.toShapeId);
        if (targetConn) {
            this.canvas.remove(targetConn.connectionObject);
            connectionStoreManager.removeByShapeIds(conn.fromShapeId, conn.toShapeId);
            this.canvas.requestRenderAll();
        }
    }

    public removeConnectionByObject(connectionObject: FabricGroup): boolean {
        if (!this.canvas) return false;

        const targetConn = connectionStoreManager.getByConnectionObject(connectionObject);
        if (!targetConn) return false;

        this.canvas.remove(connectionObject);
        connectionStoreManager.removeByConnectionObject(connectionObject);
        this.canvas.discardActiveObject();
        this.canvas.requestRenderAll();

        return true;
    }

    public upsertCursors(cursors: CursorData[]) {
        if (!this.canvas) return;

        cursors.forEach((cursor) => {
            if (!this.canvas) return;

            const existCursor = this.remoteCursors[cursor.user.id];
            if (existCursor) {
                const fabricObj = existCursor.fabricObj;
                if (fabricObj.left !== cursor.x || fabricObj.top !== cursor.y) {
                    fabricObj.set({
                        left: cursor.x,
                        top: cursor.y,
                    });
                }
            } else {
                const fabricObj = fabricManager.createRemoteCursor(cursor);

                this.remoteCursors[cursor.user.id] = {
                    user: cursor.user,
                    x: cursor.x,
                    y: cursor.y,
                    fabricObj: fabricObj,
                };

                this.canvas.add(fabricObj);
            }
        });

        this.canvas.requestRenderAll();
    }

    public expandPage(direction: 'left' | 'right' | 'top' | 'bottom') {
        if (!this.canvas) return;
        if (!this.pageBorder) return;

        let left: number = this.pageBorder.left;
        let top: number = this.pageBorder.top;

        // Создаем невидимый объект-якорь в нужном месте
        switch (direction) {
            case 'left':
                left = this.pageBorder.left - this.pageBorder.width / 2 - PAGE_EXPAND_STEP;
                break;
            case 'right':
                left = this.pageBorder.left + this.pageBorder.width / 2 + PAGE_EXPAND_STEP;
                break;
            case 'top':
                top = this.pageBorder.top - this.pageBorder.height / 2 - PAGE_EXPAND_STEP;
                break;
            case 'bottom':
                top = this.pageBorder.top + this.pageBorder.height / 2 + PAGE_EXPAND_STEP;
                break;
        }

        const anchor = new FabricRect({
            left: left,
            top: top,
            width: 1,
            height: 1,
            fill: 'transparent',
            stroke: 'transparent',
            selectable: false,
            evented: false,
            hasControls: false,
            excludeFromExport: true,
        });

        this.canvas.add(anchor);
        this.canvas.sendObjectToBack(anchor);

        // Перерисуем страницу и сетку
        this.canvas.remove(this.pageBorder);
        this.gridLines.forEach((line) => this.canvas?.remove(line));
        this.gridLines = [];
        this.prepareCanvas(this.canvas.getObjects());
        viewportManager.centerView();
    }

    // eslint-disable-next-line complexity
    public moveObject(name: string, direction: 'up' | 'down' | 'front' | 'back') {
        if (!this.canvas) return;

        const userObjects = this.canvas.getObjects().filter((o) => isShapeGroup(o));

        const currentIndex = userObjects.findIndex((o) => o.name === name);
        if (currentIndex === -1) return;

        const shape = shapeStoreManager.getById(name);
        if (!shape) return;

        const fabricObj = shape.fabricGroup as FabricGroup;

        switch (direction) {
            // Переместить вперёд
            case 'up': {
                if (currentIndex < userObjects.length - 1) {
                    const beforeIndex = currentIndex + 1;
                    const afterIndex = currentIndex + 2 < userObjects.length ? currentIndex + 2 : -1;

                    const beforeObj = userObjects[beforeIndex] as FabricGroup;
                    const afterObj = afterIndex >= 0 ? (userObjects[afterIndex] as FabricGroup) : null;

                    const before = beforeObj.order || null;
                    const after = afterObj?.order || null;

                    const newOrder = calculateOrderKey(before, after);

                    shapeStoreManager.update({
                        ...shape,
                        order: newOrder,
                    });
                    fabricObj.order = newOrder;
                    selectedObject.update((el) => {
                        if (!el) return el;
                        el.order = newOrder;
                        return el;
                    });

                    this.canvas.bringObjectForward(fabricObj);
                    this.syncLayers();
                }
                break;
            }

            // Переместить назад
            case 'down': {
                if (currentIndex > 0) {
                    const afterIndex = currentIndex - 1;
                    const beforeIndex = currentIndex - 2 >= 0 ? currentIndex - 2 : -1;

                    const afterObj = userObjects[afterIndex] as FabricGroup;
                    const beforeObj = beforeIndex >= 0 ? (userObjects[beforeIndex] as FabricGroup) : null;

                    const before = beforeObj?.order || null;
                    const after = afterObj.order || null;

                    const newOrder = calculateOrderKey(before, after);

                    shapeStoreManager.update({
                        ...shape,
                        order: newOrder,
                    });
                    fabricObj.order = newOrder;
                    selectedObject.update((el) => {
                        if (!el) return el;
                        el.order = newOrder;
                        return el;
                    });

                    this.canvas.sendObjectBackwards(fabricObj);
                    this.syncLayers();
                }
                break;
            }

            // На передний план
            case 'front': {
                const afterObj = userObjects[userObjects.length - 1] as FabricGroup;
                const before = afterObj.order || null;

                const newOrder = calculateOrderKey(before, null);

                shapeStoreManager.update({
                    ...shape,
                    order: newOrder,
                });
                fabricObj.order = newOrder;
                selectedObject.update((el) => {
                    if (!el) return el;
                    el.order = newOrder;
                    return el;
                });

                this.canvas.bringObjectToFront(fabricObj);
                this.syncLayers();
                break;
            }

            // На задний план
            case 'back': {
                const beforeObj = userObjects[0] as FabricGroup;
                const after = beforeObj.order || null;

                const newOrder = calculateOrderKey(null, after);

                shapeStoreManager.update({
                    ...shape,
                    order: newOrder,
                });
                fabricObj.order = newOrder;
                selectedObject.update((el) => {
                    if (!el) return el;
                    el.order = newOrder;
                    return el;
                });

                this.canvas.sendObjectToBack(fabricObj);
                this.syncLayers();
                break;
            }
        }
    }

    private restoreState() {
        if (!this.canvas) return;
        if (!this.diagramInfo) return;

        logger.log('CanvasManager restoreState');

        this.deps.clearState();
        this.canvas.clear();

        let shapesData: ShapeElement[] = [];
        let connectionsData: ConnectionElement[] = [];
        const coordinatesData: { x: number; y: number; zoom: number } = { x: 0, y: 0, zoom: 1 };

        // Маппинг shapes из API
        if (this.diagramInfo.shapes) {
            shapesData = this.diagramInfo.shapes.map((shape) => this.deps.mapShapeFromApi(shape));
        }

        // Маппинг connections
        if (this.diagramInfo.connections) {
            connectionsData = this.diagramInfo.connections.map((conn) => this.deps.mapConnectionFromApi(conn));
        }

        const shapeCommentStats: Record<string, number> = this.diagramInfo.shapeCommentStats.reduce(
            (items, v) => {
                items[v.shapeId] = v.commentCount;
                return items;
            },
            {} as Record<string, number>,
        );

        // Подготавливаем canvas: размеры, сетка, рамка
        const objects = shapesData.map((shape) => fabricManager.createShape(shape));
        this.prepareCanvas(objects);

        // Сортируем пользовательские объекты по order
        const orderedShapesData = sortByOrderKey(shapesData);

        // Загружаем shapes
        orderedShapesData.forEach((shape: ShapeElement) => {
            const commentCount = shapeCommentStats[shape.id] || 0;
            logger.log(shape.id, commentCount);
            this.addShape(shape, commentCount);
        });

        // Загружаем connections
        connectionsData.forEach((conn: ConnectionElement) => {
            const fromShape = shapeStoreManager.getById(conn.fromShapeId);
            const toShape = shapeStoreManager.getById(conn.toShapeId);
            if (fromShape?.fabricGroup && toShape?.fabricGroup) {
                this.addConnection(fromShape.fabricGroup, toShape.fabricGroup);
            }
        });

        // Загружаем coordinates
        coordinates.set(coordinatesData);

        this.deps.setCoordinates(coordinatesData.x, coordinatesData.y, coordinatesData.zoom);

        this.canvas.requestRenderAll();
        viewportManager.centerView();

        // Устанавливаем начальное состояние для AutoSave
        autoSaveService.setInitialState(shapesData || [], connectionsData || []);
    }

    // eslint-disable-next-line max-lines-per-function
    private appendNewShape(left: number, top: number) {
        if (!this.canvas) return;

        // Создаем элемент в зависимости от режима
        let newElement: ShapeElement;

        switch (this.currentDrawingMode) {
            case 'square': {
                newElement = {
                    id: this.deps.createId(),
                    type: 'square',
                    left: left,
                    top: top,
                    width: DEFAULT_SIZE,
                    height: DEFAULT_SIZE,
                    fill: DEFAULT_FILL,
                    stroke: DEFAULT_STROKE_SQUARE,
                    strokeWidth: DEFAULT_STROKE_WIDTH,
                    scaleX: DEFAULT_SCALE,
                    scaleY: DEFAULT_SCALE,
                    text: 'Квадрат',
                    order: '',
                };
                break;
            }

            case 'ellipse': {
                newElement = {
                    id: this.deps.createId(),
                    type: 'ellipse',
                    left: left,
                    top: top,
                    width: DEFAULT_SIZE,
                    height: DEFAULT_SIZE,
                    rx: DEFAULT_RADIUS,
                    ry: DEFAULT_RADIUS,
                    fill: DEFAULT_FILL,
                    stroke: DEFAULT_STROKE_ELLIPSE,
                    strokeWidth: DEFAULT_STROKE_WIDTH,
                    scaleX: DEFAULT_SCALE,
                    scaleY: DEFAULT_SCALE,
                    text: 'Круг',
                    order: '',
                };
                break;
            }

            case 'diamond': {
                newElement = {
                    id: this.deps.createId(),
                    type: 'diamond',
                    left: left,
                    top: top,
                    width: DEFAULT_SIZE,
                    height: DEFAULT_SIZE,
                    fill: DEFAULT_FILL,
                    stroke: DEFAULT_STROKE_DIAMOND,
                    strokeWidth: DEFAULT_STROKE_WIDTH,
                    scaleX: DEFAULT_SCALE,
                    scaleY: DEFAULT_SCALE,
                    text: 'Ромб',
                    order: '',
                };
                break;
            }

            case 'box': {
                newElement = {
                    id: this.deps.createId(),
                    type: 'box',
                    left: left,
                    top: top,
                    width: DEFAULT_SIZE,
                    height: DEFAULT_SIZE,
                    fill: 'transparent',
                    stroke: DEFAULT_STROKE_BOX,
                    strokeWidth: DEFAULT_STROKE_WIDTH,
                    scaleX: DEFAULT_SCALE,
                    scaleY: DEFAULT_SCALE,
                    text: 'Контейнер',
                    order: '',
                };
                break;
            }

            case 'text': {
                newElement = {
                    id: this.deps.createId(),
                    type: 'text',
                    left: left,
                    top: top,
                    width: DEFAULT_SIZE,
                    height: DEFAULT_SIZE,
                    fill: 'transparent',
                    stroke: 'transparent',
                    strokeWidth: DEFAULT_STROKE_WIDTH,
                    scaleX: DEFAULT_SCALE,
                    scaleY: DEFAULT_SCALE,
                    text: 'Текст',
                    order: '',
                };
                break;
            }

            default:
                return;
        }

        // Размещаем элемент на сетке
        const snapped = this.snapPositionToGrid(newElement.left, newElement.top, newElement.width, newElement.height);
        newElement.left = snapped.left;
        newElement.top = snapped.top;

        this.addShape(newElement);
        this.canvas.requestRenderAll();
    }

    private addShape(element: ShapeElement, commentCount: number = 0): FabricGroup | null {
        if (!this.canvas) return null;
        if (!this.diagramInfo) return null;

        if (element.order === '') {
            const before = shapeStoreManager.getMaxOrderKey();
            element.order = calculateOrderKey(before, null);
        }

        const shape = fabricManager.createShape(element);

        // Поворот
        delete shape.controls.mtr;

        // Углы
        delete shape.controls.tl; // top-left scaling
        delete shape.controls.tr; // top-right scaling
        delete shape.controls.bl; // bottom-left scaling
        delete shape.controls.br; // bottom-right scaling

        if (!this.deps.hasPermission(this.diagramInfo.permission)) {
            // Центры сторон
            delete shape.controls.ml; // top-left scalingXOrSkewingY
            delete shape.controls.mr; // top-right scalingXOrSkewingY
            delete shape.controls.mb; // bottom-left scalingXOrSkewingY
            delete shape.controls.mt; // bottom-right scalingXOrSkewingY

            shape.lockMovementX = true;
            shape.lockMovementY = true;
        }

        element.fabricGroup = shape;
        shapeStoreManager.add(element);

        if (commentCount > 0) {
            commentBadgeManager.upsert(shape, commentCount, false);
        }

        this.canvas.add(shape);

        return shape;
    }

    private appendNewConnection(target: FabricGroup) {
        if (!this.canvas) return;

        // Первый клик - запоминаем элемент
        if (!this.fromElement) {
            this.fromElement = target;
            return;
        }

        // Второй клик - создаем связь
        if (this.fromElement !== target) {
            this.addConnection(this.fromElement, target);
            this.canvas.requestRenderAll();
        }

        // Сбрасываем fromElement
        this.fromElement = null;
    }

    private addConnection(from: FabricGroup, target: FabricGroup) {
        if (!this.canvas) return;

        const connection = fabricManager.createConnection(from, target);
        connectionStoreManager.add(from, target, connection);
        controlsManager.setConnectionBaseControl(connection);
        if (this.deps.hasPermission(this.diagramInfo?.permission)) {
            controlsManager.addConnectionDeleteControl(connection);
        }
        this.canvas.add(connection);
        this.syncLayers();
    }

    private updateConnectionByShapeId(shapeId: string) {
        if (!this.canvas) return;

        connectionStoreManager.getByShapeId(shapeId).forEach((conn) => {
            this.canvas?.remove(conn.connectionObject);
            const newConn = fabricManager.createConnection(conn.fromElement, conn.toElement);
            controlsManager.setConnectionBaseControl(newConn);
            if (this.deps.hasPermission(this.diagramInfo?.permission)) {
                controlsManager.addConnectionDeleteControl(newConn);
            }
            conn.connectionObject = newConn;
            this.canvas?.add(newConn);
        });

        this.canvas.requestRenderAll();
    }

    private removeConnectionByShapeId(shapeId: string) {
        if (!this.canvas) return;

        connectionStoreManager.getByShapeId(shapeId).forEach((conn) => {
            this.canvas?.remove(conn.connectionObject);
        });

        connectionStoreManager.removeByShapeId(shapeId);

        this.canvas.requestRenderAll();
    }

    private updateCanvasCursor() {
        if (!this.canvas) return;

        switch (this.currentDrawingMode) {
            case 'select': {
                this.canvas.defaultCursor = 'default';
                if (this.deps.hasPermission(this.diagramInfo?.permission)) {
                    this.canvas.hoverCursor = 'move';
                } else {
                    this.canvas.hoverCursor = 'default';
                }
                break;
            }

            case 'square':
            case 'ellipse':
            case 'diamond':
            case 'box':
            case 'text': {
                this.canvas.defaultCursor = 'crosshair';
                this.canvas.hoverCursor = 'crosshair';
                break;
            }

            case 'arrow': {
                this.canvas.defaultCursor = 'default';
                this.canvas.hoverCursor = 'pointer';
                break;
            }
        }
    }

    private sendCursorPosition(event: MouseEvent) {
        if (!this.canvas) return;
        if (!this.diagramInfo) return;

        const now = Date.now();
        if (now - this.lastCursorSendTime < this.cursorThrottleMs) return;
        this.lastCursorSendTime = now;

        const position = this.canvas.getScenePoint(event);

        webSocketManager.sendCursorMove(this.diagramInfo.id, {
            userId: '', // Будет заполнено на сервере
            username: '', // Будет заполнено на сервере
            x: position.x,
            y: position.y,
        });
    }

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    private updateDebug(target: any) {
        if (!isShapeGroup(target)) {
            selectedObject.set(null);
            return;
        }

        const shape = target.getObjects()[0];
        if (!shape) return;

        let rx: number | null = null;
        let ry: number | null = null;
        if (shape.type === 'ellipse') {
            rx = shape.rx;
            ry = shape.ry;
        }

        selectedObject.set({
            name: target.name,
            type: target.shapeType,
            order: target.order,
            object: target,
            group: {
                left: target.left,
                top: target.top,
                width: target.width,
                height: target.height,
                scaleX: target.scaleX,
                scaleY: target.scaleY,
            },
            shape: {
                left: shape.left,
                top: shape.top,
                width: shape.width,
                height: shape.height,
                scaleX: shape.scaleX,
                scaleY: shape.scaleY,
                rx: rx,
                ry: ry,
            },
        });
    }

    private prepareCanvas(objects: FabricObject[]) {
        const pageBounds = this.calculatePageBounds(objects);
        this.drawPage(pageBounds);
        this.drawGrid(pageBounds);
    }

    /**
     * Привязка значения к сетке
     */
    private snapValue(value: number): number {
        return Math.round(value / PAGE_GRID_STEP) * PAGE_GRID_STEP;
    }

    /**
     * Привязка позиций краёв объекта к сетке.
     * Работает корректно при originX/Y = 'center'/'center'.
     * Возвращает новые left/top для центра объекта.
     */
    private snapPositionToGrid(
        left: number,
        top: number,
        width: number,
        height: number,
    ): { left: number; top: number } {
        // Вычисляем позиции краёв
        const leftEdge = left - width / 2;
        const topEdge = top - height / 2;

        // Привязываем края к сетке
        const snappedLeftEdge = this.snapValue(leftEdge);
        const snappedTopEdge = this.snapValue(topEdge);

        // Возвращаем координаты центра из привязанных краёв
        return {
            left: snappedLeftEdge + width / 2,
            top: snappedTopEdge + height / 2,
        };
    }

    private calculatePageBounds(objects: FabricObject[]): PageSize {
        if (objects.length === 0) {
            return {
                width: PAGE_INIT_WIDTH,
                height: PAGE_INIT_HEIGHT,
                offsetX: 0,
                offsetY: 0,
            };
        }

        const bbox = getBoundingBox(objects);

        // Добавляем отступы со всех сторон
        const tmpWidth = bbox.width + PAGE_PADDING * 2;
        const tmpHeight = bbox.height + PAGE_PADDING * 2;

        // Привязываемся к сетке
        const gridWidth = Math.ceil(tmpWidth / PAGE_GRID_STEP) * PAGE_GRID_STEP;
        const gridHeight = Math.ceil(tmpHeight / PAGE_GRID_STEP) * PAGE_GRID_STEP;

        const pageSize = {
            width: gridWidth,
            height: gridHeight,
            offsetX: this.snapValue(bbox.left - PAGE_PADDING),
            offsetY: this.snapValue(bbox.top - PAGE_PADDING),
        };

        return pageSize;
    }

    private drawPage(size: PageSize) {
        if (!this.canvas) return;

        this.pageBorder = new FabricRect({
            left: size.width / 2 + size.offsetX,
            top: size.height / 2 + size.offsetY,
            width: size.width,
            height: size.height,
            fill: '#ffffff',
            stroke: '#cccccc',
            strokeWidth: 1,
            selectable: false,
            evented: false,
            hasControls: false,
            lockMovementX: true,
            lockMovementY: true,
            excludeFromExport: true, // Исключаем из экспорта
        });

        // Сохраняем светлые исходники и применяем текущую тему сразу при создании
        this.pageBorder.sourceFill = '#ffffff';
        this.pageBorder.sourceStroke = '#cccccc';
        fabricManager.applyThemeColor(this.pageBorder, this.isDark);

        this.canvas.add(this.pageBorder);
        this.canvas.requestRenderAll();
    }

    private drawGrid(size: PageSize) {
        if (!this.canvas) return;

        // Удаляем старые линии
        this.gridLines.forEach((line) => this.canvas?.remove(line));
        this.gridLines = [];

        const color = '#e0e0e0';

        // Вертикальные линии
        for (let i = 0; i <= size.width; i += PAGE_GRID_STEP) {
            const line = new FabricPolyline(
                [
                    { x: i + size.offsetX, y: 0 + size.offsetY },
                    { x: i + size.offsetX, y: size.height + size.offsetY },
                ],
                {
                    stroke: color,
                    selectable: false,
                    evented: false,
                    hoverCursor: 'default',
                    excludeFromExport: true,
                },
            );
            this.gridLines.push(line);
            this.applyGridLineTheme(line);
            this.canvas.add(line);
        }

        // Горизонтальные линии
        for (let i = 0; i <= size.height; i += PAGE_GRID_STEP) {
            const line = new FabricPolyline(
                [
                    { x: 0 + size.offsetX, y: i + size.offsetY },
                    { x: size.width + size.offsetX, y: i + size.offsetY },
                ],
                {
                    stroke: color,
                    selectable: false,
                    evented: false,
                    hoverCursor: 'default',
                    excludeFromExport: true,
                },
            );
            this.gridLines.push(line);
            this.applyGridLineTheme(line);
            this.canvas.add(line);
        }

        // Отправляем сетку на задний план
        this.gridLines.forEach((line) => this.canvas?.sendObjectToBack(line));

        // Отправляем фон на задний план
        if (this.pageBorder) {
            this.canvas.sendObjectToBack(this.pageBorder);
        }

        this.canvas.requestRenderAll();
    }

    /** Сохраняет светлый цвет линии сетки и применяет текущую тему. */
    private applyGridLineTheme(line: FabricPolyline) {
        line.sourceStroke = '#e0e0e0';
        fabricManager.applyThemeColor(line, this.isDark);
    }

    private snapToGridObj(obj: FabricGroup, action: 'moving' | 'scaling') {
        switch (action) {
            case 'moving':
                {
                    const realW = obj.width * obj.scaleX;
                    const realH = obj.height * obj.scaleY;

                    const snapped = this.snapPositionToGrid(obj.left, obj.top, realW, realH);
                    obj.left = snapped.left;
                    obj.top = snapped.top;
                }
                break;
            case 'scaling':
                {
                    const corner = obj.__corner; // 'tl', 'tr', 'bl', 'br', 'ml', 'mr', 'mt', 'mb'

                    const realW = obj.width * obj.scaleX;
                    const realH = obj.height * obj.scaleY;

                    // Привязываем размеры к сетке
                    const snappedW = this.snapValue(realW);
                    const snappedH = this.snapValue(realH);

                    // Вычисляем все четыре края ПЕРЕД изменением размеров
                    const leftEdge = obj.left - realW / 2;
                    const rightEdge = obj.left + realW / 2;
                    const topEdge = obj.top - realH / 2;
                    const bottomEdge = obj.top + realH / 2;

                    let newLeft: number;
                    let newTop: number;

                    // ─── Горизонталь ───
                    // Если тяним левый край → фиксируем правый край
                    if (corner === 'tl' || corner === 'ml' || corner === 'bl') {
                        const snappedRightEdge = this.snapValue(rightEdge);
                        newLeft = snappedRightEdge - snappedW / 2;
                    } else {
                        // Тянем правый край → фиксируем левый край
                        const snappedLeftEdge = this.snapValue(leftEdge);
                        newLeft = snappedLeftEdge + snappedW / 2;
                    }

                    // ─── Вертикаль ───
                    // Если тяним верхний край → фиксируем нижний край
                    if (corner === 'tl' || corner === 'mt' || corner === 'tr') {
                        const snappedBottomEdge = this.snapValue(bottomEdge);
                        newTop = snappedBottomEdge - snappedH / 2;
                    } else {
                        // Тянем нижний край → фиксируем верхний край
                        const snappedTopEdge = this.snapValue(topEdge);
                        newTop = snappedTopEdge + snappedH / 2;
                    }

                    // Применяем
                    obj.set({
                        width: snappedW,
                        height: snappedH,
                        scaleX: DEFAULT_SCALE,
                        scaleY: DEFAULT_SCALE,
                        left: newLeft,
                        top: newTop,
                    });

                    // Обновляем внутренние объекты: фигуру подгоняем под размер группы, текст сохраняем масштаб
                    const shape = obj.item(0);
                    const text = obj.item(1) as FabricIText;

                    if (text) {
                        // Сохраняем постоянный размер текста, компенсируя масштаб группы
                        text.set({
                            scaleX: DEFAULT_SCALE,
                            scaleY: DEFAULT_SCALE,
                        });
                    }

                    obj.setCoords();

                    if (shape) {
                        switch (obj.shapeType) {
                            case 'square':
                            case 'text':
                                {
                                    shape.set({ width: snappedW, height: snappedH });
                                }
                                break;
                            case 'ellipse':
                                {
                                    shape.set({ rx: snappedW / 2, ry: snappedH / 2 });
                                }
                                break;
                            case 'diamond':
                                {
                                    // diamond: points are based on width/height
                                    const points = [
                                        { x: 0, y: -snappedH / 2 },
                                        { x: snappedW / 2, y: 0 },
                                        { x: 0, y: snappedH / 2 },
                                        { x: -snappedW / 2, y: 0 },
                                    ];
                                    shape.set({ points });
                                }
                                break;
                            case 'box':
                                {
                                    shape.set({ width: snappedW, height: snappedH });
                                    obj.setCoords();
                                    obj.item(2).set({
                                        points: calcBoxLine(snappedW, snappedH, text.text || ''),
                                    });
                                    text.set({ top: calcBoxText(snappedH, text.height) });
                                }
                                break;
                            default:
                                break;
                        }
                    }

                    obj.dirty = true;
                    obj.setCoords();
                }
                break;
        }
    }

    private syncLayers() {
        if (!this.canvas) return;

        // Отправляем соединения на задний план (поверх сетки и фона)
        this.canvas
            .getObjects()
            .filter((obj) => obj.isConnection)
            .forEach((conn) => this.canvas?.sendObjectToBack(conn));

        // Отправляем сетку на задний план
        if (this.gridLines.length > 0) {
            this.gridLines.forEach((line) => this.canvas?.sendObjectToBack(line));
        }

        // Отправляем фон на задний план
        if (this.pageBorder) {
            this.canvas.sendObjectToBack(this.pageBorder);
        }
    }
}

export const canvasManager = new CanvasManager();
