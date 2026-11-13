// @vitest-environment jsdom
import { describe, it, expect, vi } from 'vitest';
import { ControlsManager, type ControlsManagerDependencies } from './ControlsManager';
import { FabricObject } from 'fabric/es';

function makeDeps(_overrides: Partial<ControlsManagerDependencies> = {}): ControlsManagerDependencies {
    return {
        removeConnection: vi.fn().mockReturnValue(true),
    };
}

function makeFabricObj() {
    const obj = new FabricObject();
    obj.controls = {} as FabricObject['controls'];
    return obj;
}

describe('ControlsManager', () => {
    describe('addConnectionDeleteControl', () => {
        it('добавляет connectionDeleteControl к объекту', () => {
            const deps = makeDeps();
            const manager = new ControlsManager(deps);
            const obj = makeFabricObj();

            manager.addConnectionDeleteControl(obj);

            expect(obj.controls.connectionDeleteControl).toBeDefined();
            expect(obj.controls.connectionDeleteControl.mouseUpHandler).toBeInstanceOf(Function);
        });
    });
});
