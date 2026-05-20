package model

import (
	"fmt"

	"git.sr.ht/~icikowski/account-center/internal/shared/xerror"
	"git.sr.ht/~icikowski/account-center/internal/shared/xhttp"
)

// Service represents a service that can be accessed by users.
type Service struct {
	Name  string              `yaml:"name"`
	URL   xhttp.URL           `yaml:"url"`
	Icon  *xhttp.URL          `yaml:"icon,omitempty"`
	Roles map[string]UserRole `yaml:"roles,omitempty"`
}

// Validate checks if the [Service] is valid and returns an error if any validation fails.
func (s Service) Validate() error {
	errs := make([]error, 0, 3+len(s.Roles))

	if s.Name == "" {
		errs = append(errs, xerror.NewValidationError("Name is required"))
	}

	if s.URL.String() == "" {
		errs = append(errs, xerror.NewValidationError("URL is required"))
	}

	if s.Icon != nil && s.Icon.String() == "" {
		errs = append(errs, xerror.NewValidationError("icon URL is required if icon is set"))
	}

	for group, role := range s.Roles {
		if !role.IsValid() {
			errs = append(
				errs,
				xerror.NewValidationError(
					fmt.Sprintf("invalid role '%s' for group '%s'", role, group),
				),
			)
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return xerror.NewValidationError("invalid service", errs...)
}

// EffectiveService represents a [Service] with the effective attributes.
type EffectiveService struct {
	Service
	EffectiveRole UserRole `yaml:"-"`
}
