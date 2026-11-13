import { get } from 'svelte/store';
import { onlineUsers } from '../stores';
import { type OnlineUser } from '../ApiSpec';

class OnlineUserStoreManager {
    public load(users: OnlineUser[]) {
        onlineUsers.set(users);
    }

    public add(user: OnlineUser) {
        const found = get(onlineUsers).find((el) => el.id === user.id);
        if (found) return;

        onlineUsers.update((u) => [...u, user]);
    }

    public remove(id: string) {
        onlineUsers.update((u) => u.filter((el) => el.id !== id));
    }

    public clear() {
        onlineUsers.set([]);
    }
}

export const onlineUserStoreManager = new OnlineUserStoreManager();
