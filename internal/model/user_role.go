package model

import (
	"slices"
	"sort"
)

// UserRole represents the role assigned to a user for given service.
type UserRole string

// Predefined user role.
const (
	UserRoleSuperuser           UserRole = "superuser"
	UserRoleSystemAdministrator UserRole = "system_administrator"
	UserRoleAdministrator       UserRole = "administrator"
	UserRoleRedactor            UserRole = "redactor"
	UserRoleEditor              UserRole = "editor"
	UserRoleViewer              UserRole = "viewer"
	UserRoleUser                UserRole = "user"
	UserRoleGuest               UserRole = "guest"
	UserRoleGeneralAccess       UserRole = "general_access"
)

// IsValid checks if the [UserRole] is one of the predefined roles.
func (r UserRole) IsValid() bool {
	switch r {
	case UserRoleSuperuser, UserRoleSystemAdministrator, UserRoleAdministrator, UserRoleRedactor,
		UserRoleEditor, UserRoleViewer, UserRoleUser, UserRoleGuest, UserRoleGeneralAccess:
		return true
	default:
		return false
	}
}

var roleOrder = map[UserRole]int{
	UserRoleSuperuser:           0,
	UserRoleSystemAdministrator: 1,
	UserRoleAdministrator:       2,
	UserRoleRedactor:            3,
	UserRoleEditor:              4,
	UserRoleViewer:              5,
	UserRoleUser:                6,
	UserRoleGuest:               7,
	UserRoleGeneralAccess:       100,
}

// OrderRoles takes [UserRole]s and returns unique roles ordered according to their hierarchy.
func OrderRoles(roles []UserRole) []UserRole {
	if len(roles) == 0 {
		return nil
	}

	out := slices.Clone(roles)
	sort.SliceStable(out, func(i, j int) bool {
		return roleOrder[out[i]] < roleOrder[out[j]]
	})
	return slices.Compact(out)
}
