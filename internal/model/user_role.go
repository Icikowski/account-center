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
	UserRoleOwner               UserRole = "owner"
	UserRoleRedactor            UserRole = "redactor"
	UserRoleEditor              UserRole = "editor"
	UserRoleContributor         UserRole = "contributor"
	UserRoleMember              UserRole = "member"
	UserRoleViewer              UserRole = "viewer"
	UserRoleUser                UserRole = "user"
	UserRoleGuest               UserRole = "guest"
	UserRoleGeneralAccess       UserRole = "general_access"
)

// IsValid checks if the [UserRole] is one of the predefined roles.
func (r UserRole) IsValid() bool {
	switch r {
	case UserRoleSuperuser, UserRoleSystemAdministrator, UserRoleAdministrator, UserRoleOwner, UserRoleRedactor,
		UserRoleEditor, UserRoleContributor, UserRoleMember, UserRoleViewer, UserRoleUser, UserRoleGuest,
		UserRoleGeneralAccess:
		return true
	default:
		return false
	}
}

var roleOrder = map[UserRole]int{
	UserRoleSuperuser:           0,
	UserRoleSystemAdministrator: 1,
	UserRoleAdministrator:       2,
	UserRoleOwner:               3,
	UserRoleRedactor:            4,
	UserRoleEditor:              5,
	UserRoleContributor:         6,
	UserRoleMember:              7,
	UserRoleViewer:              8,
	UserRoleUser:                9,
	UserRoleGuest:               10,
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
