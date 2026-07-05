package evaluator

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/rs/zerolog"
	"pkg.icikowski.pl/sets"

	"git.sr.ht/~icikowski/account-center/internal/model"
	"git.sr.ht/~icikowski/account-center/internal/store"
)

const evaluationCacheTTL = 24 * time.Hour

// Evaluator represents a component that can evaluate the effective access levels for services based on a catalog and
// user information.
type Evaluator interface {
	// Evaluate determines the effective access level for available services based on the user's data.
	Evaluate(
		ctx context.Context,
		catalogProvider model.Reloader[model.Catalog],
		user model.User,
	) []model.EffectiveService
}

type evaluator struct {
	store store.EvaluationStore
	log   zerolog.Logger
}

// New creates a new [Evaluator] instance with the provided logger and cached evaluations store.
func New(log zerolog.Logger, store store.EvaluationStore) Evaluator {
	return &evaluator{
		store: store,
		log:   log,
	}
}

// Evaluate implements [Evaluator].
func (e *evaluator) Evaluate(
	ctx context.Context,
	catalogProvider model.Reloader[model.Catalog],
	user model.User,
) []model.EffectiveService {
	var (
		subject          = user.Subject
		userGroups       = slices.Clone(user.Groups)
		catalogTimestamp = catalogProvider.LastUpdate()
		l                = e.log.With().Str("subject", subject).Logger()
	)

	if stored, err := e.store.Evaluations().Get(ctx, subject); err == nil {
		l.Debug().Msg("cached evaluation found for subject")
		if catalogTimestamp.Equal(stored.CatalogTimestamp) && equalGroups(userGroups, stored.Groups) {
			return cloneEffectiveServices(stored.EffectiveServices)
		}
		l.Debug().Msg("cached evaluation is stale, recalculating")
	} else if !errors.Is(err, store.ErrNotFound) {
		l.Warn().Err(err).Msg("failed to load cached evaluation")
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

	if err := e.store.Evaluations().Set(ctx, subject, model.Evaluation{
		CatalogTimestamp:  catalogTimestamp,
		Groups:            userGroups,
		EffectiveServices: cloneEffectiveServices(effective),
	}, evaluationCacheTTL); err != nil {
		l.Warn().Err(err).Msg("failed to cache evaluation")
	}
	l.Debug().Msg("evaluation completed and cached for subject")

	return cloneEffectiveServices(effective)
}

func equalGroups(current, stored []string) bool {
	return sets.Equal(sets.NewFromSlice(current), sets.NewFromSlice(stored))
}

func cloneEffectiveServices(in []model.EffectiveService) []model.EffectiveService {
	return slices.Clone(in)
}
