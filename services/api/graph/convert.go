package graph

import (
	"fmt"

	"github.com/icerde/api/internal/store"
)

func mapUser(user store.User) (*User, error) {
	kind, err := mapWorkspaceKind(user.WorkspaceKind)
	if err != nil {
		return nil, err
	}
	return &User{
		ID:            user.ID,
		Email:         user.Email,
		WorkspaceKind: kind,
	}, nil
}

func mapWorkspaceKind(kind store.WorkspaceKind) (WorkspaceKind, error) {
	switch kind {
	case store.WorkspacePersonal:
		return WorkspaceKindPersonal, nil
	case store.WorkspaceOrganization:
		return WorkspaceKindOrganization, nil
	default:
		var unknown store.WorkspaceKind = kind
		return "", fmt.Errorf("unhandled workspace kind: %s", unknown)
	}
}

func mapProjects(rows []store.Project) []*Project {
	out := make([]*Project, 0, len(rows))
	for _, row := range rows {
		item := row
		out = append(out, &Project{
			ID:     item.ID,
			Name:   item.Name,
			Stack:  item.Stack,
			Status: item.Status,
		})
	}
	return out
}
