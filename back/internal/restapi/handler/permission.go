package handler

import (
	"fmt"

	"klubok/internal/domain"
	"klubok/internal/restapi/oas"
)

func diagramPermissionToDomain(src oas.DiagramCollaboratorPermissionEnum) (domain.DiagramCollaboratorPermission, error) {
	switch src {
	case oas.DiagramCollaboratorPermissionEnumOWNER:
		return domain.DiagramCollaboratorPermissionOwner, nil
	case oas.DiagramCollaboratorPermissionEnumVIEW:
		return domain.DiagramCollaboratorPermissionView, nil
	case oas.DiagramCollaboratorPermissionEnumEDIT:
		return domain.DiagramCollaboratorPermissionEdit, nil
	}

	return domain.DiagramCollaboratorPermission(""), fmt.Errorf("unknown permission: %s", src)
}

func diagramPermissionFromDomain(src domain.DiagramCollaboratorPermission) (oas.DiagramCollaboratorPermissionEnum, error) {
	switch src {
	case domain.DiagramCollaboratorPermissionOwner:
		return oas.DiagramCollaboratorPermissionEnumOWNER, nil
	case domain.DiagramCollaboratorPermissionView:
		return oas.DiagramCollaboratorPermissionEnumVIEW, nil
	case domain.DiagramCollaboratorPermissionEdit:
		return oas.DiagramCollaboratorPermissionEnumEDIT, nil
	}

	return oas.DiagramCollaboratorPermissionEnum(""), fmt.Errorf("unknown permission: %s", src)
}
