<script lang="ts">
    import avatarColor from '../services/avatarColor';
    import { type OnlineUser } from '../ApiSpec';

    interface Props {
        user: OnlineUser;
        index: number;
    }

    const { user, index }: Props = $props();
</script>

<div
    class="user-avatar"
    style="background-color: {avatarColor(user.userName)}; margin-left: {index === 0 ? '0px' : '-15px'};"
    title={user.userName}
>
    {user.userName.charAt(0).toUpperCase()}
</div>

<style>
    .user-avatar {
        width: 30px;
        height: 30px;
        border-radius: 50%;
        color: white;
        display: flex;
        align-items: center;
        justify-content: center;
        font-weight: bold;
        font-size: 14px;
        border: 2px solid white;
        transition: margin-left 0.2s ease;
        cursor: pointer;
    }
</style>
