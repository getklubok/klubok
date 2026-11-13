import {
    Group as FabricGroup,
    Control as FabricControl,
    FabricObject,
    util as fabricUtil,
    type TPointerEvent,
    type Transform,
} from 'fabric/es';

import { canvasManager } from './CanvasManager';

// Docs
// Icons: https://fontawesome.com/icons
// SVG (Not Full SVG)
// fill="#333333"

export interface ControlsManagerDependencies {
    removeConnection: (group: FabricGroup) => boolean;
}

const defaultDeps: ControlsManagerDependencies = {
    removeConnection: (group) => canvasManager.removeConnectionByObject(group),
};

const sizeIcon = 24;

function renderIcon(icon: HTMLImageElement) {
    // eslint-disable-next-line max-params
    return function (
        ctx: CanvasRenderingContext2D,
        left: number,
        top: number,
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        _styleOverride: any,
        obj: FabricObject,
    ) {
        ctx.save();
        ctx.translate(left, top);
        ctx.rotate(fabricUtil.degreesToRadians(obj.angle || 0));
        ctx.drawImage(icon, -sizeIcon / 2, -sizeIcon / 2, sizeIcon, sizeIcon);
        ctx.restore();
    };
}

export class ControlsManager {
    private readonly deps: ControlsManagerDependencies;

    constructor(deps: ControlsManagerDependencies = defaultDeps) {
        this.deps = deps;
    }

    public setConnectionBaseControl(fabricObj: FabricObject) {
        // Поворот
        delete fabricObj.controls.mtr;

        // Центры сторон
        delete fabricObj.controls.ml; // top-left scalingXOrSkewingY
        delete fabricObj.controls.mr; // top-right scalingXOrSkewingY
        delete fabricObj.controls.mb; // bottom-left scalingXOrSkewingY
        delete fabricObj.controls.mt; // bottom-right scalingXOrSkewingY

        // Углы
        delete fabricObj.controls.tl; // top-left scaling
        delete fabricObj.controls.tr; // top-right scaling
        delete fabricObj.controls.bl; // bottom-left scaling
        delete fabricObj.controls.br; // bottom-right scaling

        fabricObj.lockMovementX = true;
        fabricObj.lockMovementY = true;
        fabricObj.hasBorders = true; // TODO: set to false if needed
    }

    public addConnectionDeleteControl(fabricObj: FabricObject) {
        const deleteIcon = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512"><!--!Font Awesome Free v7.1.0 by @fontawesome - https://fontawesome.com License - https://fontawesome.com/license/free Copyright 2025 Fonticons, Inc.--><path fill="#ff0000" d="M256 48a208 208 0 1 1 0 416 208 208 0 1 1 0-416zm0 464a256 256 0 1 0 0-512 256 256 0 1 0 0 512zM167 167c-9.4 9.4-9.4 24.6 0 33.9l55 55-55 55c-9.4 9.4-9.4 24.6 0 33.9s24.6 9.4 33.9 0l55-55 55 55c9.4 9.4 24.6 9.4 33.9 0s9.4-24.6 0-33.9l-55-55 55-55c9.4-9.4 9.4-24.6 0-33.9s-24.6-9.4-33.9 0l-55 55-55-55c-9.4-9.4-24.6-9.4-33.9 0z"/></svg>`;

        const deleteImg = document.createElement('img');
        deleteImg.src = `data:image/svg+xml,${encodeURIComponent(deleteIcon)}`;

        fabricObj.controls.connectionDeleteControl = new FabricControl({
            x: 0.5,
            y: -0.5,
            offsetX: sizeIcon,
            offsetY: -sizeIcon,
            cursorStyle: 'pointer',
            render: renderIcon(deleteImg),
            mouseUpHandler: (_eventData: TPointerEvent, transform: Transform, _x: number, _y: number) => {
                return this.deps.removeConnection(transform.target as FabricGroup);
            },
        });
    }
}

export const controlsManager = new ControlsManager();
