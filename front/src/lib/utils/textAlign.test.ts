import { describe, it, expect, vi } from 'vitest';
import type { IText as FabricIText } from 'fabric/es';
import type { ElementType, TextAlign } from '../stores';
import { applyTextHorizontalAlign } from './textAlign';

function makeText(width: number, scaleX = 1) {
    const set = vi.fn();
    return { text: { width, scaleX, set } as unknown as FabricIText, set };
}

describe('applyTextHorizontalAlign', () => {
    it.each<{ align: TextAlign; type: ElementType; width: number; scaleX?: number; expectedLeft: number }>([
        { align: 'left', type: 'box', width: 40, expectedLeft: -20 },
        { align: 'center', type: 'box', width: 40, expectedLeft: 0 },
        { align: 'right', type: 'box', width: 40, expectedLeft: 20 },
        { align: 'left', type: 'text', width: 40, expectedLeft: -30 },
        { align: 'center', type: 'text', width: 40, expectedLeft: 0 },
        { align: 'right', type: 'text', width: 40, expectedLeft: 30 },
        { align: 'left', type: 'box', width: 40, scaleX: 0.5, expectedLeft: -35 },
        { align: 'right', type: 'box', width: 40, scaleX: 0.5, expectedLeft: 35 },
        { align: 'left', type: 'text', width: 40, scaleX: 0.5, expectedLeft: -40 },
        { align: 'right', type: 'text', width: 40, scaleX: 0.5, expectedLeft: 40 },
    ])(
        'выравнивает текст "$align" в фигуре "$type" шириной 100 (scaleX=$scaleX)',
        ({ align, type, width, scaleX, expectedLeft }) => {
            const { text, set } = makeText(width, scaleX);

            applyTextHorizontalAlign(text, 100, align, type);

            expect(set).toHaveBeenCalledWith({ textAlign: align, left: expectedLeft });
        },
    );
});
