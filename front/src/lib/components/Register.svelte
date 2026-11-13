<script lang="ts">
    import Privacy from './Privacy.svelte';
    import { api } from '../Api';
    import type { RespDataMsg } from '../types';
    import { smartCaptchaConfig, features } from '../stores';
    import { getCaptchaToken, renderSmartCaptcha } from '../services/SmartCaptcha';

    let { onSuccess } = $props();

    let email = $state('');
    let username = $state('');
    let loading = $state(false);
    let error = $state('');
    let isAcceptPrivacy = $state(false);
    let showPrivacy = $state(false);
    let captchaEnabled = $derived($features.captchaEnabled);
    let authEmailEnabled = $derived($features.authEmailEnabled);

    const registerCaptchaContainerId = 'js-register-captcha-container';
    let captchaRendered = $state(false);

    $effect(() => {
        if (captchaEnabled && authEmailEnabled && !captchaRendered) {
            captchaRendered = true;
            renderSmartCaptcha(registerCaptchaContainerId, $smartCaptchaConfig);
        }
    });

    function openPrivacy(event: Event) {
        event.preventDefault();

        showPrivacy = true;
    }

    async function handleRegister(event: Event) {
        event.preventDefault();

        if (!email || !username) {
            error = 'Заполните все поля';
            return;
        }

        loading = true;
        error = '';

        try {
            await api.userService.userRegister({
                email,
                username,
                captchaToken: captchaEnabled ? getCaptchaToken(registerCaptchaContainerId) : '',
            });
            // После успешной регистрации диспатчим событие success
            onSuccess();
            // eslint-disable-next-line require-atomic-updates
            email = '';
            // eslint-disable-next-line require-atomic-updates
            username = '';
        } catch (err) {
            error = (err as RespDataMsg)?.response?.data?.message || 'Ошибка регистрации';
        } finally {
            loading = false;
        }
    }
</script>

<div class="container">
    <div class="row justify-content-center">
        <div class="col-md-11 pt-3 pb-3">
            {#if authEmailEnabled}
                <form onsubmit={handleRegister}>
                    <div class="mb-3">
                        <label for="email" class="form-label">Email</label>
                        <input
                            type="email"
                            class="form-control"
                            id="email"
                            bind:value={email}
                            required
                            placeholder="user@example.com"
                        />
                    </div>
                    <div class="mb-3">
                        <label for="username" class="form-label">Имя пользователя</label>
                        <input
                            type="text"
                            class="form-control"
                            id="username"
                            bind:value={username}
                            required
                            pattern="^[a-z0-9]+$"
                            maxlength="50"
                            placeholder="username"
                        />
                        <div class="form-text">
                            Только латинские буквы в нижнем регистре и цифры,<br />до 50 символов
                        </div>
                    </div>
                    {#if captchaEnabled}
                        <div class="mb-3">
                            <div style="height: 100px" id={registerCaptchaContainerId} class="smart-captcha"></div>
                        </div>
                    {/if}
                    {#if error}
                        <div class="alert alert-danger" role="alert">
                            {error}
                        </div>
                    {/if}
                    <button type="submit" class="btn btn-primary w-100 mb-3" disabled={!isAcceptPrivacy || loading}>
                        {#if loading}
                            <span class="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true"></span>
                            Регистрация...
                        {:else}
                            Зарегистрироваться
                        {/if}
                    </button>
                    <div class="form-check">
                        <input
                            class="form-check-input"
                            type="checkbox"
                            id="acceptPrivacy"
                            name="acceptPrivacy"
                            bind:checked={isAcceptPrivacy}
                        />
                        <label class="form-check-label" for="acceptPrivacy">
                            <small>
                                Я даю согласие на обработку персональных данных и соглашаюсь с
                                <!-- svelte-ignore a11y_invalid_attribute -->
                                <a href="#" onclick={openPrivacy}>политикой конфиденциальности</a>.
                            </small>
                        </label>
                    </div>
                </form>
            {:else}
                <div class="alert alert-warning" role="alert">Регистрация по email недоступна</div>
            {/if}
        </div>
    </div>
</div>

<Privacy bind:show={showPrivacy} />
