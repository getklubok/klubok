export const PAGE_INIT_WIDTH = 1500;
export const PAGE_INIT_HEIGHT = 1200;
export const PAGE_GRID_STEP = 20;
export const PAGE_PADDING = 40; // должен быть кратен PAGE_GRID_STEP
export const PAGE_EXPAND_STEP = 500;

export const DEFAULT_SIZE = 100;
export const DEFAULT_RADIUS = 50;
export const DEFAULT_SCALE = 1;
export const DEFAULT_FILL: string = '#ffffff';
export const DEFAULT_STROKE_WIDTH: number = 1;
export const DEFAULT_STROKE_SQUARE: string = '#c0392b';
export const DEFAULT_STROKE_ELLIPSE: string = '#2980b9';
export const DEFAULT_STROKE_DIAMOND: string = '#e67e22';
export const DEFAULT_STROKE_BOX: string = '#27ae60';

// Палитра контролов выделения: corner-контролы, рамка вокруг объекта и рамка
// области выделения. В светлой теме — тёмные (#333) на белом фоне, в тёмной —
// светлые (#e0e0e0) на тёмном фоне. Цвета сетки страницы совпадают со светлыми
// значениями, чтобы не разводить лишних сущностей.
export const SELECTION_CORNER_COLOR_LIGHT: string = '#fff';
export const SELECTION_CORNER_COLOR_DARK: string = '#333';
export const SELECTION_CORNER_STROKE_LIGHT: string = '#fff';
export const SELECTION_CORNER_STROKE_DARK: string = '#333';
export const SELECTION_BORDER_LIGHT: string = '#fff';
export const SELECTION_BORDER_DARK: string = '#333';
