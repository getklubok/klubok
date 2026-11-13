<script lang="ts">
    import Register from './Register.svelte';
    import Login from './Login.svelte';
    import { features } from '../stores';

    let activeTab = $state('login'); // 'login' or 'register'
    let authEmailEnabled = $derived($features.authEmailEnabled);

    let { show = $bindable(), onSuccess } = $props();

    $effect(() => {
        if (!authEmailEnabled && activeTab === 'register') {
            activeTab = 'login';
        }
    });

    function handleSuccess(isRegister: boolean = false) {
        show = false;
        onSuccess(isRegister);
    }
</script>

{#if show}
    <div class="modal fade show d-block" tabindex="-1" style="background-color: rgba(0,0,0,0.5);">
        <div class="modal-dialog modal-dialog-centered">
            <div class="modal-content">
                <div class="modal-body">
                    <ul class="nav nav-tabs" role="tablist">
                        <li class="nav-item" role="presentation">
                            <button
                                class="nav-link {activeTab === 'login' ? 'active' : ''}"
                                type="button"
                                onclick={() => (activeTab = 'login')}
                            >
                                Вход
                            </button>
                        </li>
                        {#if authEmailEnabled}
                            <li class="nav-item" role="presentation">
                                <button
                                    class="nav-link {activeTab === 'register' ? 'active' : ''}"
                                    type="button"
                                    onclick={() => (activeTab = 'register')}
                                >
                                    Регистрация
                                </button>
                            </li>
                        {/if}
                    </ul>
                    <div class="border border-top-0">
                        {#if activeTab === 'login'}
                            <Login onSuccess={handleSuccess} />
                        {:else}
                            <Register onSuccess={() => handleSuccess(true)} />
                        {/if}
                    </div>
                </div>
            </div>
        </div>
    </div>
{/if}
