<script lang="ts">
    import { type DiagramCollaboratorPermissionEnum } from '../ApiSpec';

    interface Props {
        permission: DiagramCollaboratorPermissionEnum | null | undefined;
    }
    const { permission }: Props = $props();
</script>

{#if permission === 'OWNER'}
    <span class="badge text-bg-success"> Полный доступ </span>
{:else if permission === 'VIEW'}
    <span class="badge text-bg-secondary"> Просмотр </span>
{:else if permission === 'EDIT'}
    <span class="badge text-bg-primary"> Редактирование </span>
{/if}
