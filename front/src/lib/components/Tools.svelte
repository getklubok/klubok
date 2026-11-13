<script lang="ts">
    import Fa from 'svelte-fa';
    import {
        faCircleHalfStroke,
        faCopy,
        faList,
        faMoon,
        faSignOut,
        faSun,
        faUser,
    } from '@fortawesome/free-solid-svg-icons';
    import DiagramListModal from './DiagramListModal.svelte';
    import { api } from '../Api';
    import { currentUser } from '../stores';
    import { resolvedThemeStore, themeStore, type Theme } from '../stores/theme';

    let showDiagramListModal = $state(false);
    let copySuccess = $state(false);

    const MAX_BUTTON_LENGTH = 20;

    function truncate(text: string, maxLength: number): string {
        if (text.length <= maxLength) return text;
        // eslint-disable-next-line prefer-template
        return text.slice(0, maxLength - 3) + '...';
    }

    function setTheme(theme: Theme) {
        themeStore.set(theme);
    }

    async function copyUsername() {
        if ($currentUser?.username) {
            await navigator.clipboard.writeText($currentUser.username);
            copySuccess = true;
            setTimeout(() => (copySuccess = false), 2000);
        }
    }
</script>

<div class="tools">
    <div class="tools-item">
        <button class="btn btn-sm btn-outline-primary" type="button" onclick={() => (showDiagramListModal = true)}>
            <Fa class="mt-1 me-2" icon={faList} />Мои диаграммы
        </button>
    </div>
    {#if $currentUser}
        <div class="tools-item">
            <div class="dropdown">
                <button class="btn btn-sm btn-outline-primary dropdown-toggle" type="button" data-bs-toggle="dropdown">
                    <Fa class="mt-1 me-2" icon={faUser} />{truncate($currentUser.username, MAX_BUTTON_LENGTH)}
                </button>
                <ul class="dropdown-menu dropdown-menu-end">
                    <li>
                        <div class="dropdown-item-text">
                            <Fa icon={faUser} /><span class="username-text">{$currentUser.username}</span>
                            <button
                                class="copy-btn"
                                type="button"
                                onclick={(e) => {
                                    e.stopPropagation();
                                    copyUsername();
                                }}
                            >
                                <Fa icon={faCopy} />
                                {#if copySuccess}
                                    <span class="copy-success">Скопировано!</span>
                                {/if}
                            </button>
                        </div>
                    </li>
                    {#if !$currentUser?.isFake}
                        <li><hr class="dropdown-divider" /></li>
                        <li>
                            <button
                                class="dropdown-item"
                                type="button"
                                onclick={async () => {
                                    await api.userService.userLogout();
                                    window.location.reload();
                                }}
                            >
                                <Fa class="me-2" icon={faSignOut} />Выход
                            </button>
                        </li>
                    {/if}
                </ul>
            </div>
        </div>
    {/if}
    <div class="tools-item">
        <div class="dropdown">
            <button class="btn btn-sm btn-outline-primary dropdown-toggle" type="button" data-bs-toggle="dropdown">
                <Fa class="mt-1 me-2" icon={$resolvedThemeStore === 'dark' ? faMoon : faSun} />Тема
            </button>
            <ul class="dropdown-menu dropdown-menu-end">
                <li>
                    <button
                        class:active={$themeStore === 'system'}
                        class="dropdown-item"
                        type="button"
                        onclick={() => setTheme('system')}
                    >
                        <Fa class="me-2" icon={faCircleHalfStroke} />Системная
                    </button>
                </li>
                <li>
                    <button
                        class:active={$themeStore === 'light'}
                        class="dropdown-item"
                        type="button"
                        onclick={() => setTheme('light')}
                    >
                        <Fa class="me-2" icon={faSun} />Светлая
                    </button>
                </li>
                <li>
                    <button
                        class:active={$themeStore === 'dark'}
                        class="dropdown-item"
                        type="button"
                        onclick={() => setTheme('dark')}
                    >
                        <Fa class="me-2" icon={faMoon} />Тёмная
                    </button>
                </li>
            </ul>
        </div>
    </div>
</div>

{#if showDiagramListModal}
    <DiagramListModal bind:show={showDiagramListModal} />
{/if}

<style>
    .tools {
        font-family: monospace;
        font-size: 15px;
        display: flex;
        gap: 15px;
    }

    .tools-item {
        display: flex;
        align-items: center;
    }

    .dropdown-item-text {
        display: flex;
        align-items: center;
        padding: 0.25rem 1rem;
        color: inherit;
        text-decoration: none;
    }

    .username-text {
        margin-left: 8px;
        margin-right: 8px;
        max-width: 200px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    .copy-btn {
        background: none;
        border: none;
        padding: 0;
        color: #6c757d;
        cursor: pointer;
        display: flex;
        align-items: center;
        gap: 4px;
    }

    .copy-btn:hover {
        color: #0d6efd;
    }

    .copy-success {
        color: #198754;
        font-size: 12px;
        margin-left: 4px;
    }
</style>
