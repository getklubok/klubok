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

const httpClient = new HttpClient({
    baseURL: getBaseUrl(),
    withCredentials: true,
});

export const api = new ApiSpec(httpClient);
