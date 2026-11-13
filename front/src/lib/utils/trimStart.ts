export function trimStart(str: string, maxLength: number): string {
    if (str.length <= maxLength) return str;
    // eslint-disable-next-line prefer-template
    return '...' + str.slice(str.length - maxLength + 3);
}
