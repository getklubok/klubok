// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function isShapeGroup(target: any): boolean {
    return target && target.type === 'group' && target.name && target.name !== '';
}
