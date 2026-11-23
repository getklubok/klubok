<script lang="ts">
    import { imageViewerUrl } from '../stores';

    function close() {
        imageViewerUrl.set(null);
    }
</script>

<svelte:window
    onkeydown={(event) => {
        if (event.key === 'Escape') close();
    }}
/>

{#if $imageViewerUrl}
    <div class="image-viewer">
        <button class="backdrop" type="button" aria-label="Закрыть просмотр" onclick={close}></button>
        <img src={$imageViewerUrl} alt="" />
    </div>
{/if}

<style>
    .image-viewer {
        position: fixed;
        inset: 0;
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 3000;
    }

    .backdrop {
        position: absolute;
        inset: 0;
        border: none;
        padding: 0;
        background: rgb(0 0 0 / 75%);
        cursor: zoom-out;
    }

    img {
        position: relative;
        max-width: 95vw;
        max-height: 95vh;
        object-fit: contain;
        padding: 0;
        border-radius: 4px;
    }
</style>
