export default function avatarColor(userName: string): string {
    let hash = 0;
    for (let i = 0; i < userName.length; i++) {
        hash = userName.charCodeAt(i) + ((hash << 5) - hash);
    }

    const color = (hash & 0x00ffffff).toString(16).toUpperCase();

    /*eslint-disable-next-line prefer-template*/
    return `#${('000000' + color).slice(-6)}`;
}
