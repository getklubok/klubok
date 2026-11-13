<script lang="ts">
    import Fa from 'svelte-fa';
    import { faTimes } from '@fortawesome/free-solid-svg-icons';
    import { onMount } from 'svelte';

    import { commentBadgeManager } from '../services/CommentBadgeManager';

    import { commentsShow, commentsSelectedGroup, currentUser } from '../stores';
    import { type DiagramComment } from '../ApiSpec';
    import { api } from '../Api';
    import logger from '../services/logger';

    export let diagramId: string;

    let comments: DiagramComment[] = [];
    let newCommentText = '';

    async function handleLoadComments() {
        if (!$commentsSelectedGroup) return;
        if (!$commentsSelectedGroup.name) return;

        try {
            const response = await api.diagramService.commentList(diagramId, $commentsSelectedGroup.name);
            comments = response.comments;
        } catch (error) {
            logger.mustError('Ошибка загрузки комментариев:', error);
            comments = [];
        }
    }

    async function handleCreateComment() {
        if (!$commentsSelectedGroup) return;
        if (!$commentsSelectedGroup.name) return;

        const text = newCommentText.trim();
        if (text) {
            try {
                await api.diagramService.commentCreate(diagramId, $commentsSelectedGroup.name, {
                    text,
                });
                // eslint-disable-next-line require-atomic-updates
                newCommentText = '';
                commentBadgeManager.upsert($commentsSelectedGroup, 1, true);
                await handleLoadComments();
            } catch (error) {
                logger.mustError('Ошибка создания комментария:', error);
            }
        }
    }

    function handleClose() {
        commentsSelectedGroup.set(null);
        commentsShow.set(false);
    }

    function formatDate(dateString: string): string {
        const date = new Date(dateString);
        const now = new Date();
        const isToday = date.toDateString() === now.toDateString();

        if (isToday) {
            return date.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
        }

        return (
            // eslint-disable-next-line prefer-template
            date.toLocaleDateString('ru-RU', {
                day: '2-digit',
                month: '2-digit',
                year: 'numeric',
            }) +
            ' ' +
            date.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
        );
    }

    onMount(async () => {
        await handleLoadComments();
    });
</script>

<div class="comments-overlay"></div>
<div class="comments-container">
    <div class="comments">
        <div class="comments-header">
            <h5>Комментарии</h5>
            <button class="close" onclick={handleClose}>
                <Fa icon={faTimes} />
            </button>
        </div>

        {#each comments as comment (comment.id)}
            <div class="comment" class:own-comment={comment.userName === $currentUser?.username}>
                <div class="comment-header">
                    <span class="user-name">{comment.userName}</span>
                    <span class="created-at">{formatDate(comment.createdAt)}</span>
                </div>
                <div class="comment-text">{comment.text}</div>
            </div>
        {/each}

        <div class="new-comment">
            <textarea bind:value={newCommentText} class="form-control" placeholder="Напишите комментарий..." rows="3"
            ></textarea>
            <button class="btn btn-primary btn-sm" onclick={handleCreateComment} disabled={!newCommentText.trim()}>
                Добавить
            </button>
        </div>
    </div>
</div>

<style>
    .comments-overlay {
        position: fixed;
        top: 0;
        left: 0;
        width: 100vw;
        height: 100vh;
        background: rgba(0, 0, 0, 0.3);
        z-index: 999;
    }

    .comments-container {
        position: absolute;
        top: 74px;
        right: 20px;
        background: rgba(255, 255, 255, 0.95);
        border-radius: 8px;
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
        z-index: 1000;
        width: 70vw;
        max-width: 400px;
        max-height: 70vh;
        overflow-y: auto;
    }

    .comments {
        display: flex;
        flex-direction: column;
        gap: 1rem;
        padding: 1rem;
    }

    .comments-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
    }

    .comments-header h5 {
        margin: 0;
        font-size: 1.2em;
    }

    .close {
        background: none;
        border: none;
        font-size: 1.2em;
        cursor: pointer;
        color: #999;
    }

    .comment {
        border: 1px solid #ddd;
        border-radius: 8px;
        padding: 1rem;
        background-color: #f9f9f9;
    }

    .own-comment {
        background-color: #e3f2fd;
        border-color: #2196f3;
    }

    .comment-header {
        display: flex;
        justify-content: space-between;
        margin-bottom: 0.5rem;
    }

    .user-name {
        font-weight: bold;
    }

    .created-at {
        font-size: 0.9em;
        color: #666;
    }

    .comment-text {
        white-space: pre-wrap;
    }

    .new-comment {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    }

    .new-comment textarea {
        width: 100%;
        border: 1px solid #ddd;
        border-radius: 4px;
        padding: 0.5rem;
        resize: vertical;
    }

    .new-comment button {
        align-self: flex-end;
    }
</style>
