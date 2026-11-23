import { HttpClient, ApiSpec } from './ApiSpec';

export function getBaseUrl() {
    if (import.meta.env.DEV) {
        return import.meta.env.VITE_API_BASE_URL;
    }

    return `${location.protocol}//${location.host}/api`;
}

export function getSsoLoginUrl() {
    return `${getBaseUrl()}/auth/sso/login`;
}

// Прямая ссылка на вариант изображения: браузер сам скачает файл (в том числе по
// Content-Disposition) и отправит cookie сессии, поэтому fetch не нужен.
export function getImageUrl(diagramId: string, imageId: string, variant: 'full' | 'download') {
    return `${getBaseUrl()}/diagrams/${diagramId}/images/${imageId}/${variant}`;
}

const httpClient = new HttpClient({
    baseURL: getBaseUrl(),
    withCredentials: true,
});

export const api = new ApiSpec(httpClient);
