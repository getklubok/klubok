import { describe, it, expect, beforeEach } from 'vitest';
import { get } from 'svelte/store';
import { onlineUsers } from '../stores';
import { onlineUserStoreManager } from './OnlineUserStoreManager';
import type { OnlineUser } from '../ApiSpec';

const user1: OnlineUser = { id: '01AAA', userName: 'alice' };
const user2: OnlineUser = { id: '02BBB', userName: 'bob' };
const user3: OnlineUser = { id: '03CCC', userName: 'charlie' };

describe('OnlineUserStoreManager', () => {
    beforeEach(() => {
        onlineUsers.set([]);
    });

    describe('load', () => {
        it('должен устанавливать список пользователей', () => {
            onlineUserStoreManager.load([user1, user2]);
            expect(get(onlineUsers)).toEqual([user1, user2]);
        });

        it('должен заменять предыдущий список', () => {
            onlineUserStoreManager.load([user1]);
            onlineUserStoreManager.load([user2, user3]);
            expect(get(onlineUsers)).toEqual([user2, user3]);
        });

        it('должен устанавливать пустой массив', () => {
            onlineUserStoreManager.load([user1]);
            onlineUserStoreManager.load([]);
            expect(get(onlineUsers)).toEqual([]);
        });
    });

    describe('add', () => {
        it('должен добавлять пользователя в пустой store', () => {
            onlineUserStoreManager.add(user1);
            expect(get(onlineUsers)).toEqual([user1]);
        });

        it('должен добавлять пользователя к существующим', () => {
            onlineUserStoreManager.load([user1]);
            onlineUserStoreManager.add(user2);
            expect(get(onlineUsers)).toEqual([user1, user2]);
        });

        it('не должен добавлять дубликат по id', () => {
            onlineUserStoreManager.load([user1]);
            const duplicate: OnlineUser = { id: '01AAA', userName: 'alice-updated' };
            onlineUserStoreManager.add(duplicate);
            expect(get(onlineUsers)).toEqual([user1]);
        });
    });

    describe('remove', () => {
        it('должен удалять пользователя по id', () => {
            onlineUserStoreManager.load([user1, user2, user3]);
            onlineUserStoreManager.remove('02BBB');
            expect(get(onlineUsers)).toEqual([user1, user3]);
        });

        it('не должен изменять store если id не найден', () => {
            onlineUserStoreManager.load([user1, user2]);
            onlineUserStoreManager.remove('nonexistent');
            expect(get(onlineUsers)).toEqual([user1, user2]);
        });

        it('должен работать с пустым store', () => {
            onlineUserStoreManager.remove('01AAA');
            expect(get(onlineUsers)).toEqual([]);
        });
    });

    describe('clear', () => {
        it('должен очищать всех пользователей', () => {
            onlineUserStoreManager.load([user1, user2, user3]);
            onlineUserStoreManager.clear();
            expect(get(onlineUsers)).toEqual([]);
        });

        it('должен работать с уже пустым store', () => {
            onlineUserStoreManager.clear();
            expect(get(onlineUsers)).toEqual([]);
        });
    });
});
