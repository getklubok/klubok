import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

vi.mock('../stores', () => ({}));

describe('logger', () => {
    let consoleLog: ReturnType<typeof vi.spyOn>;
    let consoleError: ReturnType<typeof vi.spyOn>;
    let consoleInfo: ReturnType<typeof vi.spyOn>;
    let consoleWarn: ReturnType<typeof vi.spyOn>;

    beforeEach(() => {
        consoleLog = vi.spyOn(console, 'log').mockImplementation(() => {});
        consoleError = vi.spyOn(console, 'error').mockImplementation(() => {});
        consoleInfo = vi.spyOn(console, 'info').mockImplementation(() => {});
        consoleWarn = vi.spyOn(console, 'warn').mockImplementation(() => {});
    });

    afterEach(() => {
        consoleLog.mockRestore();
        consoleError.mockRestore();
        consoleInfo.mockRestore();
        consoleWarn.mockRestore();
        vi.resetModules();
    });

    it('logger.log вызывает console.log', async () => {
        import.meta.env.DEV = true;
        const logger = (await import('./logger')).default;
        logger.log('test');

        expect(consoleLog).toHaveBeenCalledTimes(1);
        expect(consoleLog).toHaveBeenCalledWith(expect.stringMatching(/\[.*\]/), 'test');
    });

    it('logger.log не вызывает console.log когда DEV=false', async () => {
        import.meta.env.DEV = false;
        const logger = (await import('./logger')).default;
        logger.log('test');

        expect(consoleLog).not.toHaveBeenCalled();
    });

    it('logger.error вызывает console.error', async () => {
        import.meta.env.DEV = true;
        const logger = (await import('./logger')).default;
        logger.error('error');

        expect(consoleError).toHaveBeenCalledTimes(1);
        expect(consoleError).toHaveBeenCalledWith(expect.stringMatching(/\[.*\]/), 'error');
    });

    it('logger.error не вызывает console.error когда DEV=false', async () => {
        import.meta.env.DEV = false;
        const logger = (await import('./logger')).default;
        logger.error('error');

        expect(consoleError).not.toHaveBeenCalled();
    });

    it('logger.info вызывает console.info', async () => {
        import.meta.env.DEV = true;
        const logger = (await import('./logger')).default;
        logger.info('info');

        expect(consoleInfo).toHaveBeenCalledTimes(1);
        expect(consoleInfo).toHaveBeenCalledWith(expect.stringMatching(/\[.*\]/), 'info');
    });

    it('logger.info не вызывает console.info когда DEV=false', async () => {
        import.meta.env.DEV = false;
        const logger = (await import('./logger')).default;
        logger.info('info');

        expect(consoleInfo).not.toHaveBeenCalled();
    });

    it('logger.warn вызывает console.warn', async () => {
        import.meta.env.DEV = true;
        const logger = (await import('./logger')).default;
        logger.warn('warn');

        expect(consoleWarn).toHaveBeenCalledTimes(1);
        expect(consoleWarn).toHaveBeenCalledWith(expect.stringMatching(/\[.*\]/), 'warn');
    });

    it('logger.warn не вызывает console.warn когда DEV=false', async () => {
        import.meta.env.DEV = false;
        const logger = (await import('./logger')).default;
        logger.warn('warn');

        expect(consoleWarn).not.toHaveBeenCalled();
    });

    it('передаёт несколько аргументов', async () => {
        import.meta.env.DEV = true;
        const logger = (await import('./logger')).default;
        logger.log('a', { b: 1 }, [1, 2, 3]);

        expect(consoleLog).toHaveBeenCalledWith(expect.stringMatching(/\[.*\]/), 'a', { b: 1 }, [1, 2, 3]);
    });
});
