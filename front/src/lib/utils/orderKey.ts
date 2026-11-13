import { generateKeyBetween } from 'fractional-indexing';

export function calculateOrderKey(before: string | null, after: string | null): string {
    if (!before && !after) {
        return generateKeyBetween(null, null);
    }
    if (!before) {
        return generateKeyBetween(null, after);
    }
    if (!after) {
        return generateKeyBetween(before, null);
    }
    return generateKeyBetween(before, after);
}

export function sortByOrderKey<T extends { order: string }>(items: T[]): T[] {
    return [...items].sort((a, b) => {
        if (a.order < b.order) return -1;
        if (a.order > b.order) return 1;
        return 0;
    });
}
