package model

import (
	"fmt"

	"git.sr.ht/~icikowski/account-center/internal/shared/xerror"
)

// Catalog represents the entire catalog of services and global access roles.
type Catalog struct {
	GlobalAccess map[string]UserRole `yaml:"global_access,omitempty"`
	Services     []Service           `yaml:"services,omitempty"`
}

// Validate checks if the [Catalog] is valid data returns an error if any validation fails.
func (c Catalog) Validate() error {
	errs := make([]error, 0, len(c.GlobalAccess)+max(len(c.Services), 1))

	for group, role := range c.GlobalAccess {
		if !role.IsValid() {
			errs = append(errs, xerror.NewValidationError(fmt.Sprintf(
				"invalid global access role '%s' for group '%s'",
				role,
				group,
			)))
		}
	}

	if len(c.Services) == 0 {
		errs = append(errs, xerror.NewValidationError("at least one service is required"))
	}
	for i, service := range c.Services {
		if err := service.Validate(); err != nil {
			errs = append(errs, xerror.NewItemValidationError(i, err))
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return xerror.NewValidationError("invalid catalog", errs...)
}
