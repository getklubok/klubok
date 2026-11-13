<script lang="ts">
    import { api, getSsoLoginUrl } from '../Api';
    import type { RespDataMsg } from '../types';
    import { smartCaptchaConfig, features } from '../stores';
    import { getCaptchaToken, renderSmartCaptcha } from '../services/SmartCaptcha';

    let { onSuccess } = $props();

    let step = $state('email'); // 'email' or 'code'
    let email = $state('');
    let code = $state('');
    let loading = $state(false);
    let error = $state('');
    let captchaEnabled = $derived($features.captchaEnabled);
    let authEmailEnabled = $derived($features.authEmailEnabled);
    let authSsoEnabled = $derived($features.authSsoEnabled);
    const ssoLoginUrl = getSsoLoginUrl();

    const loginCaptchaContainerId = 'js-login-captcha-container';
    let captchaRendered = $state(false);

    $effect(() => {
        if (captchaEnabled && authEmailEnabled && !captchaRendered) {
            captchaRendered = true;
            renderSmartCaptcha(loginCaptchaContainerId, $smartCaptchaConfig);
        }
    });

    async function handleSendCode(event: Event) {
        event.preventDefault();

        if (!email) {
            error = 'Введите email';
            return;
        }

        loading = true;
        error = '';

        try {
            await api.userService.userLogin({
                email,
                captchaToken: captchaEnabled ? getCaptchaToken(loginCaptchaContainerId) : '',
            });
            step = 'code';
        } catch (err) {
            error = (err as RespDataMsg)?.response?.data?.message || 'Ошибка отправки кода';
        } finally {
            loading = false;
        }
    }

    async function handleVerifyCode(event: Event) {
        event.preventDefault();

        if (!code) {
            error = 'Введите код';
            return;
        }

        loading = true;
        error = '';

        try {
            await api.userService.userLoginVerify({ email, code });
            // После успешной верификации диспатчим событие success
            onSuccess();
        } catch (err) {
            error = (err as RespDataMsg)?.response?.data?.message || 'Ошибка верификации кода';
        } finally {
            loading = false;
        }
    }

    function backToEmail() {
        step = 'email';
        code = '';
        error = '';
    }
</script>

<div class="container">
    <div class="row justify-content-center">
        <div class="col-md-11 pt-3 pb-3">
            {#if authSsoEnabled}
                <a href={ssoLoginUrl} class="btn btn-outline-primary w-100 mb-3"> Войти через SSO </a>
            {/if}
            {#if authEmailEnabled}
                {#if step === 'email'}
                    <form onsubmit={handleSendCode}>
                        <div class="mb-3">
                            <label for="login-email" class="form-label">Email</label>
                            <input
                                type="email"
                                class="form-control"
                                id="login-email"
                                bind:value={email}
                                required
                                placeholder="user@example.com"
                            />
                        </div>
                        {#if captchaEnabled}
                            <div class="mb-3">
                                <div style="height: 100px" id={loginCaptchaContainerId} class="smart-captcha"></div>
                            </div>
                        {/if}
                        {#if error}
                            <div class="alert alert-danger" role="alert">
                                {error}
                            </div>
                        {/if}
                        <button type="submit" class="btn btn-primary w-100" disabled={loading}>
                            {#if loading}
                                <span class="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true"
                                ></span>
                                Отправка...
                            {:else}
                                Отправить код
                            {/if}
                        </button>
                    </form>
                {:else}
                    <form onsubmit={handleVerifyCode}>
                        <div class="mb-3">
                            <label for="login-code" class="form-label">Код из email</label>
                            <input
                                type="text"
                                class="form-control"
                                id="login-code"
                                bind:value={code}
                                required
                                placeholder="ХХХХХХ"
                            />
                        </div>
                        {#if error}
                            <div class="alert alert-danger" role="alert">
                                {error}
                            </div>
                        {/if}
                        <div class="d-grid gap-2">
                            <button type="submit" class="btn btn-success w-100" disabled={loading}>
                                {#if loading}
                                    <span class="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true"
                                    ></span>
                                    Вход...
                                {:else}
                                    Войти
                                {/if}
                            </button>
                            <button type="button" class="btn btn-outline-secondary w-100" onclick={backToEmail}>
                                Назад
                            </button>
                        </div>
                    </form>
                {/if}
            {:else if !authSsoEnabled}
                <div class="alert alert-warning" role="alert">Нет доступных способов входа</div>
            {/if}
        </div>
    </div>
</div>
