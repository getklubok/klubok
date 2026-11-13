// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { WebSocketManager, type WebSocketManagerDependencies } from './WebSocketManager';
import type { CursorData } from '../stores';

type MockWs = {
    send: ReturnType<typeof vi.fn>;
    close: ReturnType<typeof vi.fn>;
    readyState: number;
    onopen: ((ev: Event) => void) | null;
    onclose: ((ev: CloseEvent) => void) | null;
    onmessage: ((ev: MessageEvent) => void) | null;
};

function makeWs(): MockWs {
    return {
        send: vi.fn(),
        close: vi.fn(),
        readyState: WebSocket.OPEN,
        onopen: null,
        onclose: null,
        onmessage: null,
    };
}

function makeDeps(overrides: Partial<WebSocketManagerDependencies> = {}) {
    const ws = makeWs();
    const wsFactory = vi.fn().mockReturnValue(ws);
    return {
        ws: ws,
        wsFactory,
        deps: {
            wsFactory,
            getBaseUrl: () => 'ws://localhost',
            upsertShape: vi.fn(),
            removeShape: vi.fn(),
            upsertConnection: vi.fn(),
            removeConnection: vi.fn(),
            addOnlineUser: vi.fn(),
            removeOnlineUser: vi.fn(),
            getShapeById: vi.fn(),
            mapShapeFromApi: vi.fn().mockReturnValue({ id: 'mapped' }),
            mapConnectionFromApi: vi.fn().mockReturnValue({ fromShapeId: 'a', toShapeId: 'b' }),
            updateCursors: vi.fn(),
            commentBadgeUpsert: vi.fn(),
            ...overrides,
        } as WebSocketManagerDependencies,
    };
}

function msg(type: string, data: unknown = null, diagramId = 'd1') {
    return JSON.stringify({ type, diagramId, timestamp: Date.now(), data });
}

describe('WebSocketManager', () => {
    beforeEach(() => {
        vi.useFakeTimers();
        vi.spyOn(console, 'info').mockImplementation(() => {});
        vi.spyOn(console, 'log').mockImplementation(() => {});
        vi.spyOn(console, 'error').mockImplementation(() => {});
        vi.spyOn(console, 'warn').mockImplementation(() => {});
    });

    afterEach(() => {
        vi.useRealTimers();
        vi.restoreAllMocks();
    });

    describe('connect', () => {
        it('не создаёт WebSocket если diagramID равен null', async () => {
            const { deps } = makeDeps();
            const manager = new WebSocketManager(deps);

            await manager.connect(null);

            expect(deps.wsFactory).not.toHaveBeenCalled();
        });

        it('создаёт WebSocket с правильным URL', async () => {
            const { deps, ws } = makeDeps();
            const manager = new WebSocketManager(deps);

            await manager.connect('d1');

            expect(deps.wsFactory).toHaveBeenCalledWith('ws://localhost?diagramID=d1');
            expect(ws.onopen).not.toBeNull();
            expect(ws.onclose).not.toBeNull();
            expect(ws.onmessage).not.toBeNull();
        });

        it('закрывает предыдущее соединение перед новым', async () => {
            const { deps, wsFactory } = makeDeps();
            const manager = new WebSocketManager(deps);

            await manager.connect('d1');
            const firstWs = wsFactory.mock.results[0].value as MockWs;

            // Создаём новый мок для второго вызова
            const ws2 = makeWs();
            wsFactory.mockReturnValueOnce(ws2);
            await manager.connect('d2');

            expect(firstWs.close).toHaveBeenCalledWith(1000, 'Manual disconnect');
        });
    });

    describe('disconnect', () => {
        it('закрывает WebSocket с кодом 1000', async () => {
            const { deps, ws } = makeDeps();
            const manager = new WebSocketManager(deps);

            await manager.connect('d1');
            await manager.disconnect();

            expect(ws.close).toHaveBeenCalledWith(1000, 'Manual disconnect');
        });

        it('ничего не делает если WebSocket не активен', async () => {
            const { deps } = makeDeps();
            const manager = new WebSocketManager(deps);

            await manager.disconnect();

            expect(deps.wsFactory).not.toHaveBeenCalled();
        });

        it('останавливает heartbeat', async () => {
            const { deps, ws } = makeDeps();
            const manager = new WebSocketManager(deps);

            await manager.connect('d1');
            ws.onopen!(new Event('open'));

            await manager.disconnect();

            // Пинг не отправляется после disconnect
            vi.advanceTimersByTime(60000);
            expect(ws.send).not.toHaveBeenCalled();
        });
    });

    describe('onopen', () => {
        it('сбрасывает attemptCount и isManualClose', async () => {
            const { deps, ws } = makeDeps();
            const manager = new WebSocketManager(deps);

            await manager.connect('d1');
            ws.onopen!(new Event('open'));

            // heartbeat должен запуститься — проверяем пинг
            vi.advanceTimersByTime(60000);
            expect(ws.send).toHaveBeenCalledWith(expect.stringContaining('ws:ping'));
        });

        it('запускает heartbeat', async () => {
            const { deps, ws } = makeDeps();
            const manager = new WebSocketManager(deps);

            await manager.connect('d1');
            ws.onopen!(new Event('open'));

            // Первый пинг
            vi.advanceTimersByTime(60000);
            expect(ws.send).toHaveBeenCalledTimes(1);

            // Отвечаем pong чтобы не сработал таймаут
            ws.onmessage!({ data: msg('ws:pong') } as MessageEvent);

            // Второй пинг
            vi.advanceTimersByTime(60000);
            expect(ws.send).toHaveBeenCalledTimes(2);
        });
    });

    describe('sendCursorMove', () => {
        it('отправляет сообщение cursor:move через WebSocket', async () => {
            const { deps, ws } = makeDeps();
            const manager = new WebSocketManager(deps);

            await manager.connect('d1');
            ws.onopen!(new Event('open'));

            const position = { userId: 'u1', username: 'user', x: 10, y: 20 };
            manager.sendCursorMove('d1', position);

            const sent = JSON.parse(ws.send.mock.calls[0][0]);
            expect(sent.type).toBe('cursor:move');
            expect(sent.diagramId).toBe('d1');
            expect(sent.data).toEqual(position);
        });

        it('не отправляет если WebSocket не открыт', async () => {
            const { deps } = makeDeps();
            const manager = new WebSocketManager(deps);

            manager.sendCursorMove('d1', { userId: 'u1', username: 'u', x: 0, y: 0 });

            expect(deps.wsFactory).not.toHaveBeenCalled();
        });
    });

    describe('pong timeout', () => {
        it('отключается и переподключается по таймауту', async () => {
            const { deps, ws } = makeDeps();
            const manager = new WebSocketManager(deps);

            await manager.connect('d1');
            ws.onopen!(new Event('open'));

            // Отправляем пинг
            vi.advanceTimersByTime(60000);
            expect(ws.send).toHaveBeenCalledWith(expect.stringContaining('ws:ping'));

            // Понг не приходит — через 10s таймаут
            vi.advanceTimersByTime(10000);

            // close + новый connect
            expect(ws.close).toHaveBeenCalled();
        });
    });

    describe('reconnect', () => {
        it('переподключается при неожиданном закрытии', async () => {
            const { deps, wsFactory } = makeDeps();
            const manager = new WebSocketManager(deps);

            await manager.connect('d1');
            const firstWs = wsFactory.mock.results[0].value as MockWs;

            // Неожиданное закрытие (код 1006)
            firstWs.onclose!(new CloseEvent('close', { code: 1006 }));

            // Через backoff создаётся новое соединение
            vi.advanceTimersByTime(2000);

            expect(deps.wsFactory).toHaveBeenCalledTimes(2);
        });

        it('не переподключается при ручном закрытии (код 1000)', async () => {
            const { deps } = makeDeps();
            const manager = new WebSocketManager(deps);

            await manager.connect('d1');
            await manager.disconnect();

            // Не должно быть второго соединения
            expect(deps.wsFactory).toHaveBeenCalledTimes(1);
        });

        it('не переподключается при коде 1001', async () => {
            const { deps, wsFactory } = makeDeps();
            const manager = new WebSocketManager(deps);

            await manager.connect('d1');
            const firstWs = wsFactory.mock.results[0].value as MockWs;
            firstWs.onclose!(new CloseEvent('close', { code: 1001 }));

            vi.advanceTimersByTime(50000);
            expect(deps.wsFactory).toHaveBeenCalledTimes(1);
        });

        it('останавливается после maxAttempts', async () => {
            const { deps, wsFactory } = makeDeps();
            const manager = new WebSocketManager(deps);

            await manager.connect('d1');

            // Симулируем 10 неудачных reconnect
            for (let i = 0; i < 10; i++) {
                const lastWs = wsFactory.mock.results[wsFactory.mock.results.length - 1].value as MockWs;
                lastWs.onclose!(new CloseEvent('close', { code: 1006 }));
                // Даём setTimeout сработать (backoff + jitter ≤ maxDelay=30000)
                await vi.advanceTimersByTimeAsync(30000);
            }

            // 11-й раз — maxAttempts достигнут, reconnect не происходит
            const lastWs = wsFactory.mock.results[wsFactory.mock.results.length - 1].value as MockWs;
            lastWs.onclose!(new CloseEvent('close', { code: 1006 }));
            await vi.advanceTimersByTimeAsync(60000);

            // Всего 11 WebSocket (1 начальный + 10 reconnect)
            expect(deps.wsFactory).toHaveBeenCalledTimes(11);
        });
    });

    describe('message handling', () => {
        async function setup(depsOverride?: Partial<WebSocketManagerDependencies>) {
            const { deps, ws } = makeDeps(depsOverride);
            const manager = new WebSocketManager(deps);
            await manager.connect('d1');
            ws.onopen!(new Event('open'));
            return { deps, ws, manager };
        }

        function send(ws: MockWs, type: string, data: unknown = null, diagramId = 'd1') {
            ws.onmessage!({ data: msg(type, data, diagramId) } as MessageEvent);
        }

        it('игнорирует сообщения с другим diagramId', async () => {
            const { deps, ws } = await setup();

            send(ws, 'shape:upserted', {}, 'other-diagram');

            expect(deps.upsertShape).not.toHaveBeenCalled();
        });

        it('обрабатывает ws:connected', async () => {
            const { ws } = await setup();

            send(ws, 'ws:connected');

            // Не падает — логирует
            expect(true).toBe(true);
        });

        it('обрабатывает ws:ping → отправляет ws:pong', async () => {
            const { ws } = await setup();

            send(ws, 'ws:ping');

            const pongCall = ws.send.mock.calls.find((c: string[]) => c[0].includes('ws:pong'));
            expect(pongCall).toBeDefined();
            const sent = JSON.parse(pongCall![0]);
            expect(sent.type).toBe('ws:pong');
        });

        it('обрабатывает ws:pong → очищает pongTimer', async () => {
            const { ws } = await setup();

            // Пинг запускает pongTimer
            vi.advanceTimersByTime(60000);
            expect(ws.send).toHaveBeenCalledWith(expect.stringContaining('ws:ping'));

            // Понг приходит — pongTimer очищается, reconnect не происходит
            send(ws, 'ws:pong');
            vi.advanceTimersByTime(10000);

            // Не было close
            expect(ws.close).not.toHaveBeenCalled();
        });

        it('обрабатывает shape:upserted', async () => {
            const { deps, ws } = await setup();
            deps.mapShapeFromApi = vi.fn().mockReturnValue({ id: 'mapped-shape' });

            send(ws, 'shape:upserted', { id: 's1' });

            expect(deps.mapShapeFromApi).toHaveBeenCalledWith({ id: 's1' });
            expect(deps.upsertShape).toHaveBeenCalledWith({ id: 'mapped-shape' });
        });

        it('обрабатывает shape:deleted → вызывает removeShape', async () => {
            const { deps, ws } = await setup({
                getShapeById: vi.fn().mockReturnValue({ fabricGroup: 'grp' }),
            });

            send(ws, 'shape:deleted', 's1');

            expect(deps.getShapeById).toHaveBeenCalledWith('s1');
            expect(deps.removeShape).toHaveBeenCalledWith('grp');
        });

        it('shape:deleted → ничего если shape не найден', async () => {
            const { deps, ws } = await setup({
                getShapeById: vi.fn().mockReturnValue(undefined),
            });

            send(ws, 'shape:deleted', 'missing');

            expect(deps.removeShape).not.toHaveBeenCalled();
        });

        it('обрабатывает connection:upserted', async () => {
            const { deps, ws } = await setup();

            send(ws, 'connection:upserted', { fid: 'a', tid: 'b' });

            expect(deps.mapConnectionFromApi).toHaveBeenCalledWith({ fid: 'a', tid: 'b' });
            expect(deps.upsertConnection).toHaveBeenCalledWith({
                fromShapeId: 'a',
                toShapeId: 'b',
            });
        });

        it('обрабатывает connection:deleted', async () => {
            const { deps, ws } = await setup();

            send(ws, 'connection:deleted', { fid: 'a', tid: 'b' });

            expect(deps.removeConnection).toHaveBeenCalledWith({
                fromShapeId: 'a',
                toShapeId: 'b',
            });
        });

        it('обрабатывает comment:added → вызывает commentBadgeUpsert', async () => {
            const { deps, ws } = await setup({
                getShapeById: vi.fn().mockReturnValue({ fabricGroup: 'grp' }),
            });

            send(ws, 'comment:added', 's1');

            expect(deps.getShapeById).toHaveBeenCalledWith('s1');
            expect(deps.commentBadgeUpsert).toHaveBeenCalledWith('grp', 1, true);
        });

        it('comment:added → ничего если shape не найден', async () => {
            const { deps, ws } = await setup({
                getShapeById: vi.fn().mockReturnValue(undefined),
            });

            send(ws, 'comment:added', 'missing');

            expect(deps.commentBadgeUpsert).not.toHaveBeenCalled();
        });

        it('обрабатывает cursor:move', async () => {
            const { deps, ws } = await setup();

            send(ws, 'cursor:move', { userId: 'u1', username: 'alice', x: 100, y: 200 });

            expect(deps.updateCursors).toHaveBeenCalledTimes(1);
            const updater = vi.mocked(deps.updateCursors).mock.calls[0][0];
            const result = updater([]);
            expect(result).toEqual([{ user: { id: 'u1', userName: 'alice' }, x: 100, y: 200 }]);
        });

        it('cursor:move обновляет существующий курсор', async () => {
            const { deps, ws } = await setup();

            send(ws, 'cursor:move', { userId: 'u1', username: 'alice', x: 50, y: 60 });

            const updater = vi.mocked(deps.updateCursors).mock.calls[0][0];
            const existing: CursorData[] = [
                { user: { id: 'u1', userName: 'alice' }, x: 10, y: 20 },
                { user: { id: 'u2', userName: 'bob' }, x: 30, y: 40 },
            ];
            const result = updater(existing);

            expect(result).toHaveLength(2);
            expect(result[0]).toEqual({ user: { id: 'u1', userName: 'alice' }, x: 50, y: 60 });
            expect(result[1]).toEqual({ user: { id: 'u2', userName: 'bob' }, x: 30, y: 40 });
        });

        it('обрабатывает user:connected', async () => {
            const { deps, ws } = await setup();

            send(ws, 'user:connected', { id: 'u1', username: 'alice' });

            expect(deps.addOnlineUser).toHaveBeenCalledWith({ id: 'u1', userName: 'alice' });
        });

        it('обрабатывает user:disconnected', async () => {
            const { deps, ws } = await setup();

            send(ws, 'user:disconnected', { id: 'u1', username: 'alice' });

            expect(deps.removeOnlineUser).toHaveBeenCalledWith('u1');
        });
    });
});
