import { driver, type DriveStep, type Driver } from 'driver.js';
import 'driver.js/dist/driver.css';

import { get, type Readable } from 'svelte/store';

import { commentsShow, connections, contextMenu, drawingMode, shapes } from '../stores';

const STORAGE_KEY = 'klubok-tutorial-completed';

let driverObj: Driver | null = null;

function isCompleted(): boolean {
    try {
        return window.localStorage.getItem(STORAGE_KEY) !== null;
    } catch {
        return false;
    }
}

function markCompleted(): void {
    try {
        window.localStorage.setItem(STORAGE_KEY, '1');
    } catch {
        // localStorage недоступен (приватный режим) — туториал просто покажется снова
    }
}

type StepWatch = Pick<DriveStep, 'onHighlighted' | 'onDeselected'>;

/**
 * Автоматически переходит к следующему шагу, когда условие выполнено.
 * onStart вызывается при показе шага — удобно запомнить исходное состояние.
 */
function waitFor(stores: Readable<unknown>[], isDone: () => boolean, onStart?: () => void): StepWatch {
    let cleanup: (() => void) | null = null;

    return {
        onHighlighted: (_element, _step, { index }) => {
            let unsubscribers: Array<() => void> = [];
            let timer: ReturnType<typeof setTimeout> | null = null;

            const stop = () => {
                unsubscribers.forEach((unsubscribe) => unsubscribe());
                unsubscribers = [];
                if (timer !== null) clearTimeout(timer);
                timer = null;
                cleanup = null;
            };

            const next = () => {
                if (timer !== null) return;
                stop();
                // Ждём тик, чтобы Svelte отрисовал следующий шаг (меню, панель комментариев).
                timer = setTimeout(() => {
                    timer = null;
                    if (driverObj?.isActive() && driverObj.getActiveIndex() === index) {
                        driverObj.moveNext();
                    }
                }, 0);
            };

            const check = () => {
                if (isDone()) next();
            };

            onStart?.();
            unsubscribers = stores.map((store) => store.subscribe(check));
            if (isDone()) {
                unsubscribers.forEach((unsubscribe) => unsubscribe());
                unsubscribers = [];
                next();
            }

            cleanup = stop;
        },
        onDeselected: () => cleanup?.(),
    };
}

function createIntroStep(): DriveStep {
    return {
        popover: {
            title: 'Добро пожаловать в «Клубок»!',
            description:
                'Это совместная бесконечная доска для блок-схем. Создавайте фигуры и связи, ' +
                'оставляйте комментарии к элементам, экспортируйте схему в PNG или файл. ' +
                'Давайте попробуем вместе — это займёт минуту.',
        },
    };
}

function createShapeSteps(): DriveStep[] {
    let squareIds = new Set<string>();
    let ellipseIds = new Set<string>();
    let connectionCount = 0;

    return [
        {
            element: 'button[title="Квадрат"]',
            popover: {
                title: 'Квадрат',
                description: 'Начнём с блока. Выберите инструмент «Квадрат» на панели слева.',
            },
            ...waitFor([drawingMode], () => get(drawingMode) === 'square'),
        },
        {
            allowPageInteraction: true,
            popover: {
                title: 'Поставьте квадрат',
                description: 'Кликните по холсту в том месте, где должен появиться квадрат.',
            },
            ...waitFor(
                [shapes],
                () => get(shapes).some((shape) => shape.type === 'square' && !squareIds.has(shape.id)),
                () => {
                    squareIds = new Set(
                        get(shapes)
                            .filter((shape) => shape.type === 'square')
                            .map((shape) => shape.id),
                    );
                },
            ),
        },
        {
            element: 'button[title="Круг"]',
            popover: {
                title: 'Круг',
                description: 'Теперь выберите инструмент «Круг».',
            },
            ...waitFor([drawingMode], () => get(drawingMode) === 'ellipse'),
        },
        {
            allowPageInteraction: true,
            popover: {
                title: 'Поставьте круг',
                description: 'Кликните по холсту в другом месте.',
            },
            ...waitFor(
                [shapes],
                () => get(shapes).some((shape) => shape.type === 'ellipse' && !ellipseIds.has(shape.id)),
                () => {
                    ellipseIds = new Set(
                        get(shapes)
                            .filter((shape) => shape.type === 'ellipse')
                            .map((shape) => shape.id),
                    );
                },
            ),
        },
        {
            element: 'button[title="Стрелка"]',
            allowPageInteraction: true,
            popover: {
                title: 'Свяжите фигуры',
                description: 'Выберите инструмент «Стрелка», затем кликните сначала по квадрату, а потом по кругу.',
            },
            ...waitFor(
                [connections],
                () => get(connections).length > connectionCount,
                () => {
                    connectionCount = get(connections).length;
                },
            ),
        },
        {
            element: 'button[title="Выбор"]',
            allowPageInteraction: true,
            popover: {
                title: 'Выбор элемента',
                description: 'Переключитесь на инструмент «Выбор» и кликните по квадрату или кругу.',
            },
        },
    ];
}

function createMenuSteps(): DriveStep[] {
    return [
        {
            element: '.group-panel',
            popover: {
                title: 'Меню элемента',
                description:
                    'У выбранного элемента появилась панель действий: комментарии, редактирование, ' +
                    'клонирование, удаление и работа со слоями.',
            },
        },
        {
            allowPageInteraction: true,
            popover: {
                title: 'Контекстное меню',
                description:
                    'То же меню доступно по правой кнопке мыши на элементе. На телефоне — долгое нажатие. ' +
                    'Откройте его, чтобы продолжить.',
            },
            ...waitFor([contextMenu], () => get(contextMenu).visible),
        },
        {
            element: '.context-menu .menu-item',
            popover: {
                title: 'Комментарии',
                description: 'В контекстном меню выберите «Открыть комментарии».',
            },
            ...waitFor([commentsShow], () => get(commentsShow)),
        },
        {
            element: '.new-comment',
            advanceOn: {
                event: 'click',
                when: (_element, _step, { event }) =>
                    Boolean((event.target as Element | null)?.closest('.new-comment button')),
            },
            popover: {
                title: 'Новый комментарий',
                description: 'Напишите комментарий в поле и нажмите «Добавить».',
            },
        },
    ];
}

function createFinalStep(): DriveStep {
    return {
        popover: {
            title: 'Готово!',
            description:
                'Вы создали квадрат, круг и связь, а также научились обсуждать элементы. ' +
                'Приятной работы в «Клубке»!',
        },
    };
}

function createSteps(): DriveStep[] {
    return [createIntroStep(), ...createShapeSteps(), ...createMenuSteps(), createFinalStep()];
}

function start(): void {
    if (driverObj?.isActive()) return;

    driverObj = driver({
        showProgress: true,
        progressText: 'Шаг {{current}} из {{total}}',
        nextBtnText: 'Далее',
        prevBtnText: 'Назад',
        doneBtnText: 'Готово',
        closeBtnLabel: 'Закрыть',
        overlayOpacity: 0.6,
        stagePadding: 4,
        stageRadius: 6,
        onDestroyed: () => {
            markCompleted();
            driverObj = null;
        },
        steps: createSteps(),
    });

    driverObj.drive();
}

export const tutorialService = {
    start,
    stop: () => driverObj?.destroy(),
    maybeStart: () => {
        if (!isCompleted()) start();
    },
};
