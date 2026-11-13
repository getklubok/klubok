// @vitest-environment jsdom
import { describe, it, expect, vi } from 'vitest';

vi.mock('fabric/es', () => {
    class MockFabricGroup {
        name?: string;
        constructor(
            public items: unknown[],
            opts: { name?: string } = {},
        ) {
            this.name = opts.name;
        }
    }
    return {
        Canvas: class {},
        Group: MockFabricGroup,
        FabricObject: class {},
        IText: class {
            set() {}
        },
    };
});
import { TextEditorManager, type TextEditorManagerDependencies } from './TextEditorManager';
import type { ShapeElement } from '../stores';

function makeShape(overrides: Partial<ShapeElement> = {}): ShapeElement {
    return {
        id: 'shape-1',
        type: 'square',
        left: 0,
        top: 0,
        width: 100,
        height: 100,
        fill: '#fff',
        stroke: '#000',
        strokeWidth: 1,
        scaleX: 1,
        scaleY: 1,
        text: 'Текст',
        order: '',
        ...overrides,
    };
}

function makeDeps(overrides: Partial<TextEditorManagerDependencies> = {}): TextEditorManagerDependencies {
    return {
        getById: () => undefined,
        update: vi.fn(),
        ...overrides,
    };
}

function makeCanvas() {
    return {
        setActiveObject: vi.fn(),
        remove: vi.fn(),
        add: vi.fn(),
        requestRenderAll: vi.fn(),
    } as unknown as import('fabric/es').Canvas;
}

function makeGroupItems() {
    const rect = { type: 'rect' } as unknown as import('fabric/es').FabricObject;
    const text = {
        type: 'i-text',
        enterEditing: vi.fn(),
        selectAll: vi.fn(),
        text: 'hello',
        set: vi.fn(),
        on: vi.fn(),
    } as unknown as import('fabric/es').IText;
    return [rect, text] as [import('fabric/es').FabricObject, import('fabric/es').IText];
}

function makeGroup(name = 'grp-1', items?: import('fabric/es').FabricObject[]) {
    const groupItems = items ?? makeGroupItems();
    return {
        name,
        type: 'group',
        getObjects: () => groupItems,
        _objects: groupItems,
    } as unknown as import('fabric/es').Group;
}

describe('TextEditorManager', () => {
    describe('startEditing', () => {
        it('возвращает false если canvas не задан', () => {
            const deps = makeDeps();
            const manager = new TextEditorManager(deps);

            const result = manager.startEditing(makeGroup());

            expect(result).toBe(false);
        });

        it('возвращает false если group не group type', () => {
            const deps = makeDeps();
            const manager = new TextEditorManager(deps);
            manager.setCanvas(makeCanvas());

            const group = makeGroup();
            group.type = 'rect';

            expect(manager.startEditing(group)).toBe(false);
        });

        it('вызывает setActiveObject и enterEditing на тексте', () => {
            const deps = makeDeps();
            const manager = new TextEditorManager(deps);
            const canvas = makeCanvas();
            manager.setCanvas(canvas);
            const items = makeGroupItems();
            const group = makeGroup('grp-1', items);

            manager.startEditing(group);

            expect(canvas.setActiveObject).toHaveBeenCalledWith(items[1]);
            expect(items[1].enterEditing).toHaveBeenCalled();
            expect(items[1].selectAll).toHaveBeenCalled();
        });

        it('удаляет группу и добавляет объекты на canvas', () => {
            const deps = makeDeps();
            const manager = new TextEditorManager(deps);
            const canvas = makeCanvas();
            manager.setCanvas(canvas);
            const items = makeGroupItems();
            const group = makeGroup('grp-1', items);

            manager.startEditing(group);

            expect(canvas.remove).toHaveBeenCalledWith(group);
            expect(canvas.add).toHaveBeenCalledTimes(2);
            expect(canvas.add).toHaveBeenCalledWith(items[0]);
            expect(canvas.add).toHaveBeenCalledWith(items[1]);
        });

        it('возвращает true при успехе', () => {
            const deps = makeDeps();
            const manager = new TextEditorManager(deps);
            manager.setCanvas(makeCanvas());

            expect(manager.startEditing(makeGroup())).toBe(true);
        });
    });

    describe('finishEditing', () => {
        it('ничего не делает если canvas не задан', () => {
            const deps = makeDeps();
            const manager = new TextEditorManager(deps);

            manager.finishEditing();

            expect(deps.update).not.toHaveBeenCalled();
        });

        it('ничего не делает если editing не начат', () => {
            const deps = makeDeps();
            const manager = new TextEditorManager(deps);
            manager.setCanvas(makeCanvas());

            manager.finishEditing();

            expect(deps.update).not.toHaveBeenCalled();
        });

        it('ничего не делает если нет groupItems', () => {
            const deps = makeDeps();
            const manager = new TextEditorManager(deps);
            manager.setCanvas(makeCanvas());

            // Без startEditing — groupItems = null
            manager.finishEditing();

            expect(deps.update).not.toHaveBeenCalled();
        });

        it('ничего не делает если нет groupName', () => {
            const deps = makeDeps();
            const manager = new TextEditorManager(deps);
            const canvas = makeCanvas();
            manager.setCanvas(canvas);

            // Группа без name — groupName будет null
            const group = makeGroup('');
            manager.startEditing(group);
            manager.finishEditing();

            // update не вызывается т.к. groupName = null
            expect(deps.update).not.toHaveBeenCalled();
        });

        it('вызывает update для shape если shape и fabricGroup существуют', () => {
            const shape = makeShape({
                id: 's1',
                text: 'old',
                fabricGroup: {} as ShapeElement['fabricGroup'],
            });
            const deps = makeDeps({
                getById: () => shape,
                update: vi.fn(),
            });
            const manager = new TextEditorManager(deps);
            const canvas = makeCanvas();
            manager.setCanvas(canvas);
            const items = makeGroupItems();
            const group = makeGroup('s1', items);

            manager.startEditing(group);
            manager.finishEditing();

            expect(deps.update).toHaveBeenCalledWith(expect.objectContaining({ id: 's1', text: 'hello' }));
        });

        it('подставляет дефолтный текст при пустой строке', () => {
            const shape = makeShape({
                id: 's1',
                text: 'old',
                fabricGroup: {} as ShapeElement['fabricGroup'],
            });
            const deps = makeDeps({
                getById: () => shape,
                update: vi.fn(),
            });
            const manager = new TextEditorManager(deps);
            const canvas = makeCanvas();
            manager.setCanvas(canvas);
            const items = makeGroupItems();
            items[1].text = '';
            const group = makeGroup('s1', items);

            manager.startEditing(group);
            manager.finishEditing();

            expect(deps.update).toHaveBeenCalledWith(expect.objectContaining({ id: 's1', text: 'Текст' }));
        });

        it('подставляет дефолтный текст если текст состоит только из пробелов', () => {
            const shape = makeShape({
                id: 's1',
                text: 'old',
                fabricGroup: {} as ShapeElement['fabricGroup'],
            });
            const deps = makeDeps({
                getById: () => shape,
                update: vi.fn(),
            });
            const manager = new TextEditorManager(deps);
            const canvas = makeCanvas();
            manager.setCanvas(canvas);
            const items = makeGroupItems();
            items[1].text = '   ';
            const group = makeGroup('s1', items);

            manager.startEditing(group);
            manager.finishEditing();

            expect(deps.update).toHaveBeenCalledWith(expect.objectContaining({ id: 's1', text: 'Текст' }));
        });

        it('обрезает пробелы по краям текста', () => {
            const shape = makeShape({
                id: 's1',
                text: 'old',
                fabricGroup: {} as ShapeElement['fabricGroup'],
            });
            const deps = makeDeps({
                getById: () => shape,
                update: vi.fn(),
            });
            const manager = new TextEditorManager(deps);
            const canvas = makeCanvas();
            manager.setCanvas(canvas);
            const items = makeGroupItems();
            items[1].text = '  hello  ';
            const group = makeGroup('s1', items);

            manager.startEditing(group);
            manager.finishEditing();

            expect(deps.update).toHaveBeenCalledWith(expect.objectContaining({ id: 's1', text: 'hello' }));
        });

        it('не вызывает update если fabricGroup отсутствует', () => {
            const shape = makeShape({ id: 's1' }); // нет fabricGroup
            const deps = makeDeps({
                getById: () => shape,
                update: vi.fn(),
            });
            const manager = new TextEditorManager(deps);
            const canvas = makeCanvas();
            manager.setCanvas(canvas);
            const items = makeGroupItems();
            const group = makeGroup('s1', items);

            manager.startEditing(group);
            manager.finishEditing();

            expect(deps.update).not.toHaveBeenCalled();
        });

        it('сбрасывает состояние после finishEditing', () => {
            const deps = makeDeps();
            const manager = new TextEditorManager(deps);
            const canvas = makeCanvas();
            manager.setCanvas(canvas);
            const items = makeGroupItems();
            const group = makeGroup('grp-1', items);

            manager.startEditing(group);
            manager.finishEditing();

            vi.clearAllMocks();
            manager.finishEditing();

            expect(canvas.add).not.toHaveBeenCalled();
        });
    });
});
