import { PAGE_GRID_STEP } from '../constants';

export function calcBoxLine(
    shapeWidth: number,
    shapeHeight: number,
    text: string,
): [{ x: number; y: number }, { x: number; y: number }] {
    const x = shapeWidth / 2;
    const y = shapeHeight / 2;

    const lineCount = text.split('\n').length;
    const lineOffset = lineCount * PAGE_GRID_STEP;

    const points: [{ x: number; y: number }, { x: number; y: number }] = [
        { x: -x, y: -y + lineOffset },
        { x: x, y: -y + lineOffset },
    ];

    return points;
}

export function calcBoxText(shapeHeight: number, textHeight: number): number {
    const y = shapeHeight / 2;
    const textOffset = textHeight / 2;

    const top = -y + textOffset;

    return top;
}
