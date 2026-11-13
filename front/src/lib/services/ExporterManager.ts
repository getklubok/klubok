import { Canvas as FabricCanvas, FabricObject, type TFiller } from 'fabric/es';
import { get } from 'svelte/store';

import { canvasManager } from './CanvasManager';

import { getBoundingBox } from '../utils/getBoundingBox';
import { shapes, connections, diagramTitle, type IEllipseElement } from '../stores';
import type { DiagramJson, ShapeJson } from '../types';

export class ExporterManager {
    private canvas: FabricCanvas | null = null;
    private hiddenStates: { obj: FabricObject; visible: boolean }[] = [];

    public setCanvas(canvas: FabricCanvas) {
        this.canvas = canvas;
    }

    public toPNG(): void {
        if (!this.canvas) return;

        // Запоминаем текущую тему и фон канваса. PNG всегда экспортируется в
        // светлой теме с белым фоном — результат консистентен независимо от
        // текущей темы приложения (и embed). recolorAll() обратим (цвета берутся
        // из sourceFill/sourceStroke), поэтому после экспорта всё возвращаем.
        const originalIsDark = canvasManager.isDarkTheme;
        const originalBackgroundColor = this.canvas.backgroundColor;

        // Временно перекрашиваем объекты в светлую тему (recolorAll сам вызывает
        // requestRenderAll) и ставим белый фон.
        this.canvas.backgroundColor = '#ffffff';
        canvasManager.recolorAll(false);

        // Скрываем исключаемые объекты
        this.hiddenStates = [];
        this.canvas.getObjects().forEach((obj) => {
            if (obj.excludeFromExport) {
                this.hiddenStates.push({ obj, visible: obj.visible });
                obj.visible = false;
            }
        });
        this.canvas.requestRenderAll();

        const visibleObjects = this.canvas.getObjects().filter((obj) => obj.visible);
        if (visibleObjects.length === 0) {
            this.restoreVisibility();
            this.restoreTheme(originalIsDark, originalBackgroundColor);
            return;
        }

        // Сохраняем оригинальные размеры canvas и viewport
        const originalWidth = this.canvas.width;
        const originalHeight = this.canvas.height;
        const originalViewportTransform = this.canvas.viewportTransform;

        // Всё, что временно меняет канвас (размеры, viewport, видимость, тема,
        // фон), возвращаем в finally — даже если getBoundingBox, toDataURL или
        // клик бросят исключение, приложение не останется в светлой теме.
        try {
            const boundingRect = getBoundingBox(visibleObjects);
            const padding = 10;

            // Временно меняем размер canvas и смещаем viewport
            this.canvas.width = boundingRect.width + padding * 2;
            this.canvas.height = boundingRect.height + padding * 2;
            this.canvas.viewportTransform = [1, 0, 0, 1, -boundingRect.left + padding, -boundingRect.top + padding];
            this.canvas.requestRenderAll();

            const dataUrl = this.canvas.toDataURL({
                format: 'png',
                multiplier: 2,
            });

            const link = document.createElement('a');
            link.download = `klubok_${this.nowTimestamp()}.png`;
            link.href = dataUrl;
            link.click();
        } finally {
            // Возвращаем оригинальные параметры, видимость, тему и фон
            // (backgroundColor ставим до recolorAll, чтобы его внутренний
            // requestRenderAll отрисовал и фон, и объекты).
            this.canvas.width = originalWidth;
            this.canvas.height = originalHeight;
            this.canvas.viewportTransform = originalViewportTransform;
            this.restoreVisibility();
            this.restoreTheme(originalIsDark, originalBackgroundColor);
        }
    }

    public toFile(): void {
        const shapesData = get(shapes);
        const connectionsData = get(connections);
        const title = get(diagramTitle);

        const idMap = new Map<string, number>();
        shapesData.forEach((s, i) => idMap.set(s.id, i + 1));

        const shapeArr: ShapeJson[] = [];
        for (const s of shapesData) {
            const id = idMap.get(s.id);
            if (!id) continue;

            let o: ShapeJson = {
                id: id,
                type: s.type,
                left: s.left,
                top: s.top,
                width: s.width,
                height: s.height,
                fill: s.fill,
                stroke: s.stroke,
                text: s.text,
                order: s.order,
            };

            if (s.type === 'ellipse') {
                const ellipse = s as IEllipseElement;
                o = {
                    ...o,
                    rx: ellipse.rx,
                    ry: ellipse.ry,
                };
            }

            shapeArr.push(o);
        }

        const connArr: [number, number][] = [];
        for (const c of connectionsData) {
            if (c.fromElement.name && c.toElement.name) {
                const fromId = idMap.get(c.fromElement.name);
                const toId = idMap.get(c.toElement.name);
                if (fromId && toId) {
                    connArr.push([fromId, toId]);
                }
            }
        }

        const data = JSON.stringify(
            {
                source: 'https://github.com/getklubok/klubok',
                type: 'klubok',
                version: 1,
                title: title,
                shapes: shapeArr,
                connections: connArr,
            } as DiagramJson,
            null,
            2,
        );

        const blob = new Blob([data], { type: 'application/json' });
        const link = document.createElement('a');
        link.download = `${this.nowTimestamp()}.klubok`;
        link.href = URL.createObjectURL(blob);
        link.click();
        URL.revokeObjectURL(link.href);
    }

    private restoreVisibility(): void {
        if (!this.canvas) return;
        this.hiddenStates.forEach((item) => (item.obj.visible = item.visible));
        this.canvas.requestRenderAll();
    }

    private restoreTheme(isDark: boolean, backgroundColor: TFiller | string): void {
        if (!this.canvas) return;
        // backgroundColor ставим до recolorAll, чтобы его внутренний
        // requestRenderAll отрисовал и фон, и объекты.
        this.canvas.backgroundColor = backgroundColor;
        canvasManager.recolorAll(isDark);
    }

    private nowTimestamp(): string {
        const now = new Date();
        const y = now.getFullYear();
        const m = String(now.getMonth() + 1).padStart(2, '0');
        const d = String(now.getDate()).padStart(2, '0');
        const th = String(now.getHours()).padStart(2, '0');
        const tm = String(now.getMinutes()).padStart(2, '0');

        return `${y}-${m}-${d}-${th}${tm}`;
    }
}

export const exporterManager = new ExporterManager();
