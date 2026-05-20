package evaluator

import (
	"slices"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"pkg.icikowski.pl/sets"

	"git.sr.ht/~icikowski/account-center/internal/auth"
	"git.sr.ht/~icikowski/account-center/internal/model"
)

// Evaluator represents a component that can evaluate the effective access levels for services based on a catalog and
// user information.
type Evaluator interface {
	// Evaluate determines the effective access level for available services based on the user's data.
	Evaluate(catalogProvider model.Reloader[model.Catalog], user auth.User) []model.EffectiveService
}

type evaluator struct {
	cached *sync.Map
	log    zerolog.Logger
}

// New creates a new [Evaluator] instance with the provided logger.
func New(log zerolog.Logger) Evaluator {
	return &evaluator{
		cached: new(sync.Map),
		log:    log,
	}
}

// Evaluate implements [Evaluator].
func (e *evaluator) Evaluate(catalogProvider model.Reloader[model.Catalog], user auth.User) []model.EffectiveService {
	var (
		subject          = user.Subject
		userGroups       = slices.Clone(user.Groups)
		catalogTimestamp = catalogProvider.LastUpdate()
		l                = e.log.With().Str("subject", subject).Logger()
	)

	if stored, ok := e.cached.Load(subject); ok {
		l.Debug().Msg("cached evaluation found for subject")
		cached, _ := stored.(cachedEvaluation)
		if catalogTimestamp.Equal(cached.catalogTimestamp) && equalGroups(userGroups, cached.groups) {
			return slices.Clone(cached.effectiveServices)
		}
		l.Debug().Msg("cached evaluation is stale, recalculating")
	}

	var (
		catalog      = catalogProvider.Snapshot()
		services     = catalog.Services
		globalAccess = catalog.GlobalAccess
	)

	effective := make([]model.EffectiveService, 0, len(services))
	for _, service := range services {
		effectiveRoles := make([]model.UserRole, 0, len(service.Roles)+len(globalAccess)+1)
		for group, role := range service.Roles {
			if slices.Contains(userGroups, group) {
				effectiveRoles = append(effectiveRoles, role)
			}
		}
		for group, role := range globalAccess {
			if slices.Contains(userGroups, group) {
				effectiveRoles = append(effectiveRoles, role)
			}
		}
		if len(service.Roles) == 0 {
			effectiveRoles = append(effectiveRoles, model.UserRoleGeneralAccess)
		}

		if len(effectiveRoles) == 0 {
			continue
		}

		effectiveRole := model.OrderRoles(effectiveRoles)[0]

		effective = append(effective, model.EffectiveService{
			Service:       service,
			EffectiveRole: effectiveRole,
		})
	}

	e.cached.Store(subject, cachedEvaluation{
		catalogTimestamp:  catalogTimestamp,
		groups:            userGroups,
		effectiveServices: effective,
	})
	l.Debug().Msg("evaluation completed and cached for subject")

	return slices.Clone(effective)
}

type cachedEvaluation struct {
	catalogTimestamp  time.Time
	groups            []string
	effectiveServices []model.EffectiveService
}

func equalGroups(current, stored []string) bool {
	return sets.Equal(sets.NewFromSlice(current), sets.NewFromSlice(stored))
}
