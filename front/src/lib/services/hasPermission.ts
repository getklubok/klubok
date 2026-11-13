import { type DiagramCollaboratorPermissionEnum } from '../ApiSpec';

const allowedPermissions: Array<DiagramCollaboratorPermissionEnum> = ['OWNER', 'EDIT'];

export default function hasPermission(permission: DiagramCollaboratorPermissionEnum | null | undefined): boolean {
    if (!permission) return false;

    return allowedPermissions.includes(permission);
}
