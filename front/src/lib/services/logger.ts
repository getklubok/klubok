/*eslint no-console: "off"*/

const isDev = import.meta.env.DEV;

function getCallerLocation(): string {
    const error = new Error();
    const stack = error.stack?.split('\n') || [];

    // Find the first stack frame that's NOT from logger.ts
    for (let i = 0; i < stack.length; i++) {
        const line = stack[i];
        if (!line.includes('logger.ts') && !line.includes('getCallerLocation')) {
            // Extract filename and line number from stack trace
            // Format: "    at functionName (http://localhost:5173/src/path/file.ts?t=1776052630537:10:5)"
            // or: "    at http://localhost:5173/src/path/file.ts?t=1776052630537:10:5"
            const match = line.match(/at\s+(?:([\w.]+)\s+)?\(?([^)]+):\d+:\d+\)?/u);
            if (match) {
                const [, func, path] = match;
                const fileName = path.split('/').pop()?.split('?')[0] || path;
                const funcName = func ? `:${func}` : '';

                return `[${fileName}${funcName}]`;
            }
        }
    }

    return '';
}

const logger = {
    log: (...args: unknown[]) => {
        if (isDev) {
            const location = getCallerLocation();
            console.log(location, ...args);
        }
    },
    error: (...args: unknown[]) => {
        if (isDev) {
            const location = getCallerLocation();
            console.error(location, ...args);
        }
    },
    info: (...args: unknown[]) => {
        if (isDev) {
            const location = getCallerLocation();
            console.info(location, ...args);
        }
    },
    warn: (...args: unknown[]) => {
        if (isDev) {
            const location = getCallerLocation();
            console.warn(location, ...args);
        }
    },
    mustLog: (...args: unknown[]) => {
        const location = getCallerLocation();
        console.log(location, ...args);
    },
    mustError: (...args: unknown[]) => {
        const location = getCallerLocation();
        console.error(location, ...args);
    },
    mustInfo: (...args: unknown[]) => {
        const location = getCallerLocation();
        console.info(location, ...args);
    },
    mustWarn: (...args: unknown[]) => {
        const location = getCallerLocation();
        console.warn(location, ...args);
    },
};

export default logger;
