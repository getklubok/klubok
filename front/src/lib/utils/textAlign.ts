import type { IText as FabricIText } from 'fabric/es';
import type { ElementType, TextAlign } from '../stores';
import { TEXT_PADDING } from '../constants';

export function applyTextHorizontalAlign(
    text: FabricIText,
    shapeWidth: number,
    textAlign: TextAlign,
    shapeType: ElementType | undefined,
): void {
    const textScale = text.scaleX ?? 1;
    const padding = shapeType === 'box' ? TEXT_PADDING * textScale : 0;
    const offset = (shapeWidth - text.width * textScale) / 2 - padding;
    let left = 0;
    if (textAlign === 'left') left = -offset;
    else if (textAlign === 'right') left = offset;

    text.set({ textAlign, left });
}
