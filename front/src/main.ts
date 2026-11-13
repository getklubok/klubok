import 'bootstrap/dist/css/bootstrap.min.css';
import 'bootstrap/dist/js/bootstrap.bundle.min.js';
import './styles.css';

import { mount } from 'svelte';

import App from './App.svelte';

import type { ElementType } from './lib/stores';

declare module 'fabric' {
    // to have the properties recognized on the instance and in the constructor
    interface FabricObject {
        shapeType?: ElementType;
        name?: string;
        order?: string;
        isConnection?: boolean;
        sourceFill?: string;
        sourceStroke?: string;
    }
    // to have the properties typed in the exported object
    interface SerializedObjectProps {
        shapeType?: ElementType;
        name?: string;
        order?: string;
        isConnection?: boolean;
    }
}

mount(App, { target: document.getElementById('app')! });
