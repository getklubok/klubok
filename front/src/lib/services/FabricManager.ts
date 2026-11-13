import {
    Group as FabricGroup,
    Point as FabricPoint,
    Circle as FabricCircle,
    Ellipse as FabricEllipse,
    Rect as FabricRect,
    Polygon as FabricPolygon,
    Triangle as FabricTriangle,
    Polyline as FabricPolyline,
    FabricObject,
    FabricText,
    IText as FabricIText,
    util as fabricUtil,
} from 'fabric/es';

import { get } from 'svelte/store';

import avatarColor from './avatarColor';

import type { CursorData, IEllipseElement, ShapeElement } from '../stores';
import { resolvedThemeStore } from '../stores/theme';

import {
    DEFAULT_SCALE,
    DEFAULT_STROKE_WIDTH,
    SELECTION_BORDER_DARK,
    SELECTION_BORDER_LIGHT,
    SELECTION_CORNER_COLOR_DARK,
    SELECTION_CORNER_COLOR_LIGHT,
    SELECTION_CORNER_STROKE_DARK,
    SELECTION_CORNER_STROKE_LIGHT,
} from '../constants';

import { calcBoxLine, calcBoxText } from '../utils/calcBox';
import { resolveColor } from '../utils/color';

function calculateConnectionPoints(
    fromElement: FabricGroup,
    toElement: FabricGroup,
): {
    startPoint: FabricPoint;
    endPoint: FabricPoint;
} {
    const fromCenter = fromElement.getCenterPoint();
    const toCenter = toElement.getCenterPoint();

    // Находим точки пересечения с границами объектов
    const startPoint = findIntersectionPoint(fromElement, fromCenter, toCenter);
    const endPoint = findIntersectionPoint(toElement, toCenter, fromCenter);

    return { startPoint, endPoint };
}

function findIntersectionPoint(element: FabricGroup, center: FabricPoint, otherCenter: FabricPoint): FabricPoint {
    const firstObjectType = element.getObjects()[0].type;
    switch (firstObjectType) {
        case 'ellipse': // Круг
            return findIntersectionPointWithEllipse(element, center, otherCenter);
        case 'polygon': // Ромб
            return findIntersectionPointWithPolygon(element, center, otherCenter);
        default: // Квадрат или box (оба rect)
            return findIntersectionPointWithRect(element, center, otherCenter);
    }
}

function findIntersectionPointWithEllipse(
    element: FabricGroup,
    center: FabricPoint,
    otherCenter: FabricPoint,
): FabricPoint {
    const ellipseElement = element.getObjects()[0] as FabricEllipse;

    // Получаем эффективные радиусы с учетом масштаба
    const rx = (ellipseElement.rx || 0) * (element.scaleX || DEFAULT_SCALE);
    const ry = (ellipseElement.ry || 0) * (element.scaleY || DEFAULT_SCALE);

    // Направляющий вектор от центра эллипса к другому объекту
    const dx = otherCenter.x - center.x;
    const dy = otherCenter.y - center.y;

    // Находим пересечение луча с эллипсом
    // Параметрическое уравнение эллипса: (x/rx)² + (y/ry)² = 1
    // Луч: x = cx + t*dx, y = cy + t*dy
    // Подставляем: (t*dx/rx)² + (t*dy/ry)² = 1
    // t² * ((dx/rx)² + (dy/ry)²) = 1
    // t = 1 / sqrt((dx/rx)² + (dy/ry)²)
    const t = 1 / Math.sqrt((dx * dx) / (rx * rx) + (dy * dy) / (ry * ry));

    return new FabricPoint(center.x + t * dx, center.y + t * dy);
}

function findIntersectionPointWithPolygon(
    element: FabricGroup,
    center: FabricPoint,
    otherCenter: FabricPoint,
): FabricPoint {
    const diamondElement = element.getObjects()[0] as FabricPolygon;

    // Предполагаем, что points ромба заданы в порядке: верх, правый, низ, левый
    const points = diamondElement.points || [
        { x: 0, y: -50 }, // верхняя точка
        { x: 50, y: 0 }, // правая точка
        { x: 0, y: 50 }, // нижняя точка
        { x: -50, y: 0 }, // левая точка
    ];

    // Преобразуем точки в глобальные координаты
    const globalPoints = points.map((point) => {
        const transformed = fabricUtil.transformPoint(new FabricPoint(point.x, point.y), element.calcTransformMatrix());
        return new FabricPoint(transformed.x, transformed.y);
    });

    // Создаем стороны ромба
    const sides = [
        { start: globalPoints[0], end: globalPoints[1] }, // верх-правый
        { start: globalPoints[1], end: globalPoints[2] }, // правый-низ
        { start: globalPoints[2], end: globalPoints[3] }, // низ-левый
        { start: globalPoints[3], end: globalPoints[0] }, // левый-верх
    ];

    // Луч от центра ромба к другому объекту
    const rayStart = center;
    const rayEnd = otherCenter;

    // Ищем пересечение с ближайшей стороной
    let closestIntersection: FabricPoint | null = null;
    let minDistance = Infinity;

    for (const side of sides) {
        const intersection = lineIntersection(rayStart, rayEnd, side.start, side.end);

        if (intersection) {
            const dx = intersection.x - rayStart.x;
            const dy = intersection.y - rayStart.y;
            const distance = Math.sqrt(dx * dx + dy * dy);
            if (distance < minDistance) {
                minDistance = distance;
                closestIntersection = intersection;
            }
        }
    }

    // Если нашли пересечение, возвращаем его, иначе используем fallback
    return (
        closestIntersection ||
        new FabricPoint(center.x + (otherCenter.x - center.x) * 0.5, center.y + (otherCenter.y - center.y) * 0.5)
    );
}

// Вспомогательная функция для нахождения пересечения двух отрезков
function lineIntersection(p1: FabricPoint, p2: FabricPoint, p3: FabricPoint, p4: FabricPoint): FabricPoint | null {
    const denominator = (p4.y - p3.y) * (p2.x - p1.x) - (p4.x - p3.x) * (p2.y - p1.y);

    if (denominator === 0) return null; // Параллельные линии

    const ua = ((p4.x - p3.x) * (p1.y - p3.y) - (p4.y - p3.y) * (p1.x - p3.x)) / denominator;
    const ub = ((p2.x - p1.x) * (p1.y - p3.y) - (p2.y - p1.y) * (p1.x - p3.x)) / denominator;

    if (ua >= 0 && ua <= 1 && ub >= 0 && ub <= 1) {
        return new FabricPoint(p1.x + ua * (p2.x - p1.x), p1.y + ua * (p2.y - p1.y));
    }

    return null;
}

function findIntersectionPointWithRect(
    element: FabricGroup,
    center: FabricPoint,
    otherCenter: FabricPoint,
): FabricPoint {
    const rectElement = element.getObjects()[0] as FabricRect;

    const width = rectElement.width || 0;
    const height = rectElement.height || 0;
    const left = rectElement.left || 0;
    const top = rectElement.top || 0;

    // Локальные координаты углов прямоугольника (originX/originY = 'center')
    const localPoints = [
        { x: left - width / 2, y: top - height / 2 },
        { x: left + width / 2, y: top - height / 2 },
        { x: left + width / 2, y: top + height / 2 },
        { x: left - width / 2, y: top + height / 2 },
    ];

    // Преобразуем точки в глобальные координаты
    const globalPoints = localPoints.map((point) => {
        const transformed = fabricUtil.transformPoint(new FabricPoint(point.x, point.y), element.calcTransformMatrix());
        return new FabricPoint(transformed.x, transformed.y);
    });

    // Создаем стороны прямоугольника
    const sides = [
        { start: globalPoints[0], end: globalPoints[1] }, // верх-правый
        { start: globalPoints[1], end: globalPoints[2] }, // правый-низ
        { start: globalPoints[2], end: globalPoints[3] }, // низ-левый
        { start: globalPoints[3], end: globalPoints[0] }, // левый-верх
    ];

    // Луч от центра прямоугольника к другому объекту
    const rayStart = center;
    const rayEnd = otherCenter;

    // Ищем пересечение с ближайшей стороной
    let closestIntersection: FabricPoint | null = null;
    let minDistance = Infinity;

    for (const side of sides) {
        const intersection = lineIntersection(rayStart, rayEnd, side.start, side.end);

        if (intersection) {
            const dx = intersection.x - rayStart.x;
            const dy = intersection.y - rayStart.y;
            const distance = Math.sqrt(dx * dx + dy * dy);
            if (distance < minDistance) {
                minDistance = distance;
                closestIntersection = intersection;
            }
        }
    }

    // Если нашли пересечение, возвращаем его, иначе используем fallback
    return (
        closestIntersection ||
        new FabricPoint(center.x + (otherCenter.x - center.x) * 0.5, center.y + (otherCenter.y - center.y) * 0.5)
    );
}

class FabricManager {
    /**
     * Запоминает светлый исходный цвет объекта и применяет конвертацию под тему.
     * sourceFill/sourceStroke хранят цвет в светлой теме (как в модели данных),
     * чтобы recolorAll() мог конвертировать его туда-обратно при смене темы.
     * Также задаёт цвета corner-контролов и рамки выделения под тему, чтобы они
     * были различимы на тёмном фоне (canvas-уровневые selectionColor/
     * selectionBorderColor задаёт CanvasManager.recolorAll).
     */
    public applyThemeColor(obj: FabricObject, isDark: boolean): void {
        if (typeof obj.fill === 'string' && obj.fill !== 'transparent') {
            obj.sourceFill = obj.sourceFill ?? obj.fill;
            obj.set('fill', resolveColor(obj.sourceFill, isDark));
        }
        if (typeof obj.stroke === 'string' && obj.stroke !== 'transparent') {
            obj.sourceStroke = obj.sourceStroke ?? obj.stroke;
            obj.set('stroke', resolveColor(obj.sourceStroke, isDark, 'stroke'));
        }
        obj.set({
            cornerColor: isDark ? SELECTION_CORNER_COLOR_LIGHT : SELECTION_CORNER_COLOR_DARK,
            cornerStrokeColor: isDark ? SELECTION_CORNER_STROKE_LIGHT : SELECTION_CORNER_STROKE_DARK,
            borderColor: isDark ? SELECTION_BORDER_LIGHT : SELECTION_BORDER_DARK,
        });
    }

    /**
     * Применяет цвета контролов выделения (corner-контролы, рамка вокруг объекта)
     * к активному объекту и, если это группа, к ней самой — рамка вокруг группы
     * рисуется активным объектом, а не его детьми.
     */
    public applySelectionControlsTheme(obj: FabricObject, isDark: boolean): void {
        this.applySelectionControlsThemeToObject(obj, isDark);
        if (obj.type === 'group') {
            (obj as FabricGroup).forEachObject((child) => this.applySelectionControlsThemeToObject(child, isDark));
        }
    }

    private applySelectionControlsThemeToObject(obj: FabricObject, isDark: boolean): void {
        obj.set({
            cornerColor: isDark ? SELECTION_CORNER_COLOR_LIGHT : SELECTION_CORNER_COLOR_DARK,
            cornerStrokeColor: isDark ? SELECTION_CORNER_STROKE_LIGHT : SELECTION_CORNER_STROKE_DARK,
            borderColor: isDark ? SELECTION_BORDER_LIGHT : SELECTION_BORDER_DARK,
        });
        // cornerColor/borderColor не входят в stateProperties/cacheProperties Fabric,
        // поэтому set() не сбрасывает кэш отрисовки — поднимаем dirty вручную, чтобы
        // контролы гарантированно перерисовались с актуальным цветом при смене темы.
        obj.dirty = true;
    }

    /** Текущая разрешённая тема: true — тёмная. */
    private isDarkTheme(): boolean {
        return get(resolvedThemeStore) === 'dark';
    }

    // eslint-disable-next-line max-lines-per-function
    public createShape(element: ShapeElement): FabricGroup {
        // Создаем фигуру - координаты внутри группы должны быть (0, 0)
        let shape: FabricObject;
        switch (element.type) {
            case 'square': {
                shape = new FabricRect({
                    width: element.width,
                    height: element.height,
                    fill: element.fill,
                    stroke: element.stroke,
                    strokeWidth: element.strokeWidth,
                    originX: 'center',
                    originY: 'center',
                });
                break;
            }

            case 'ellipse': {
                shape = new FabricEllipse({
                    rx: (element as IEllipseElement).rx,
                    ry: (element as IEllipseElement).ry,
                    fill: element.fill,
                    stroke: element.stroke,
                    strokeWidth: element.strokeWidth,
                    originX: 'center',
                    originY: 'center',
                });
                break;
            }

            case 'diamond': {
                // Создаем ромб как полигон для правильного позиционирования controls
                const diamondPoints = [
                    { x: 0, y: -element.height / 2 }, // верхняя точка
                    { x: element.width / 2, y: 0 }, // правая точка
                    { x: 0, y: element.height / 2 }, // нижняя точка
                    { x: -element.width / 2, y: 0 }, // левая точка
                ];
                shape = new FabricPolygon(diamondPoints, {
                    fill: element.fill,
                    stroke: element.stroke,
                    strokeWidth: element.strokeWidth,
                    originX: 'center',
                    originY: 'center',
                });
                break;
            }

            case 'box': {
                shape = new FabricRect({
                    width: element.width,
                    height: element.height,
                    fill: 'transparent',
                    stroke: element.stroke,
                    strokeWidth: element.strokeWidth,
                    originX: 'center',
                    originY: 'center',
                });
                break;
            }

            case 'text': {
                shape = new FabricRect({
                    width: element.width,
                    height: element.height,
                    fill: 'transparent',
                    stroke: 'transparent',
                    strokeWidth: element.strokeWidth,
                    originX: 'center',
                    originY: 'center',
                });
                break;
            }

            default:
                throw new Error(`Неизвестный тип элемента: ${(element as ShapeElement).type}`);
        }

        // Создаем текст - координаты внутри группы должны быть (0, 0)
        element.text = element.text || 'Текст';
        const text = new FabricIText(element.text, {
            fontFamily: 'PT Sans',
            fontSize: 16,
            fill: '#000',
            originX: 'center',
            originY: 'center',
            textAlign: 'center',
        });

        const isBox = element.type === 'box';

        let group: FabricGroup;
        let separatorLine: FabricPolyline | undefined;
        if (isBox) {
            separatorLine = new FabricPolyline(calcBoxLine(element.width, element.height, text.text || ''), {
                stroke: element.stroke,
                strokeWidth: DEFAULT_STROKE_WIDTH,
            });

            group = new FabricGroup([shape, text, separatorLine], {
                shapeType: element.type,
                name: element.id,
                order: element.order,
                left: element.left,
                top: element.top,
                scaleX: element.scaleX,
                scaleY: element.scaleY,
                originX: 'center',
                originY: 'center',
            });

            group.item(1).set({
                top: calcBoxText(element.height, text.height),
                scaleX: DEFAULT_SCALE / (group.scaleX || DEFAULT_SCALE),
                scaleY: DEFAULT_SCALE / (group.scaleY || DEFAULT_SCALE),
            });

            group.item(2).set({
                scaleX: DEFAULT_SCALE / (group.scaleX || DEFAULT_SCALE),
                scaleY: DEFAULT_SCALE / (group.scaleY || DEFAULT_SCALE),
            });
        } else {
            group = new FabricGroup([shape, text], {
                shapeType: element.type,
                name: element.id,
                order: element.order,
                left: element.left,
                top: element.top,
                scaleX: element.scaleX,
                scaleY: element.scaleY,
                originX: 'center',
                originY: 'center',
            });

            group.item(1).set({
                scaleX: DEFAULT_SCALE / (group.scaleX || DEFAULT_SCALE),
                scaleY: DEFAULT_SCALE / (group.scaleY || DEFAULT_SCALE),
            });
        }

        // Конвертируем fill/stroke фигуры, текста и разделителя под текущую тему
        const isDark = this.isDarkTheme();
        this.applyThemeColor(shape, isDark);
        this.applyThemeColor(text, isDark);
        if (separatorLine) {
            this.applyThemeColor(separatorLine, isDark);
        }

        // Сразу задаём цвета контролов выделения самой группе — рамка вокруг
        // объекта и corner-контролы рисуются группой как активным объектом,
        // а не её детьми. Без этого у нового объекта контролы были бы дефолтными
        // до следующего recolorAll (сломано при старте и при добавлении объекта).
        this.applySelectionControlsTheme(group, isDark);

        return group;
    }

    public createConnection(fromElement: FabricGroup, toElement: FabricGroup): FabricGroup {
        const { startPoint, endPoint } = calculateConnectionPoints(fromElement, toElement);

        const arrowSize = 10; // Размер стрелки
        const arrowLength = arrowSize / 2; // Половина стрелки
        const angle = Math.atan2(endPoint.y - startPoint.y, endPoint.x - startPoint.x);

        // Корректируем конечную точку линии для стрелки
        const lineEndPoint = new FabricPoint(
            endPoint.x - Math.cos(angle) * arrowLength,
            endPoint.y - Math.sin(angle) * arrowLength,
        );

        // Создаем линию между центрами объектов
        const line = new FabricPolyline(
            [
                { x: startPoint.x, y: startPoint.y },
                { x: lineEndPoint.x, y: lineEndPoint.y },
            ],
            {
                stroke: '#000',
                strokeWidth: 2,
            },
        );

        // Создаем наконечник стрелки
        const arrowHead = new FabricTriangle({
            left: lineEndPoint.x,
            top: lineEndPoint.y,
            width: arrowSize,
            height: arrowSize,
            fill: '#000',
            angle: fabricUtil.radiansToDegrees(angle) + 90,
            originX: 'center',
            originY: 'center',
        });

        // Конвертируем линию и наконечник под текущую тему
        const isDark = this.isDarkTheme();
        this.applyThemeColor(line, isDark);
        this.applyThemeColor(arrowHead, isDark);

        // Объединяем линию и наконечник в группу
        const connection = new FabricGroup([line, arrowHead], {
            isConnection: true,
        });

        // Сразу задаём цвета контролов выделения группе соединения
        this.applySelectionControlsTheme(connection, isDark);

        return connection;
    }

    public createRemoteCursor(cursor: CursorData) {
        const userName = cursor.user.userName || '?';
        const color = avatarColor(userName);

        const radius = 12; // Радиус круга
        const strokeWidth = 2; // Толщина линии круга
        const pointerSize = 6; // Размер "уголка"

        // 1. Создаем указатель (треугольник в левом верхнем углу)
        // Точки: (0,0) - кончик, (pointerSize, 0) - право, (0, pointerSize) - низ
        const pointer = new FabricPolygon(
            [
                { x: 0, y: 0 },
                { x: pointerSize, y: 0 },
                { x: 0, y: pointerSize },
            ],
            {
                fill: '#000',
                left: 0,
                top: 0,
                originX: 'left',
                originY: 'top',
                selectable: false,
                evented: false,
            },
        );

        // 2. Создаем круг
        // Смещаем circle немного вправо-вниз, чтобы он перекрывал основание указателя,
        // но оставлял кончик (0,0) свободным.
        const circle = new FabricCircle({
            radius: radius,
            fill: color,
            stroke: '#fff',
            strokeWidth: strokeWidth,
            left: 0, // Круг начинается там же, где группа (но визуально сдвинут центром)
            top: 0,
            originX: 'left',
            originY: 'top',
            selectable: false,
            evented: false,
        });

        // 3. Создаем текст (первая буква)
        const text = new FabricText(userName.charAt(0).toUpperCase(), {
            fill: '#fff',
            fontSize: 14,
            fontFamily: 'Arial',
            left: pointerSize / 2 + radius - strokeWidth, // Центр по горизонтали
            top: pointerSize / 2 + radius - strokeWidth, // Центр по вертикали
            originX: 'center',
            originY: 'center',
            selectable: false,
            evented: false,
        });

        // 4. Конвертируем цвета под текущую тему
        const isDark = this.isDarkTheme();
        this.applyThemeColor(circle, isDark);
        this.applyThemeColor(text, isDark);
        this.applyThemeColor(pointer, isDark);

        // 5. Объединяем в группу
        return new FabricGroup([circle, text, pointer], {
            left: cursor.x,
            top: cursor.y,
            originX: 'left', // Важно: левый верхний угол группы = точка (0,0) указателя
            originY: 'top',
            selectable: false, // Группу нельзя выделять
            evented: false, // Группа не перехватывает события
        });
    }
}

export const fabricManager = new FabricManager();
