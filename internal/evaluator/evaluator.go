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
			return slices.Clone(stored.EffectiveServices)
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
		effectiveRole, ok := effectiveRoleForService(userGroups, service, globalAccess)
		if !ok {
			continue
		}

		effective = append(effective, model.EffectiveService{
			Service:       service,
			EffectiveRole: effectiveRole,
		})
	}

	if err := e.store.Evaluations().Set(ctx, subject, model.Evaluation{
		CatalogTimestamp:  catalogTimestamp,
		Groups:            userGroups,
		EffectiveServices: slices.Clone(effective),
	}, evaluationCacheTTL); err != nil {
		l.Warn().Err(err).Msg("failed to cache evaluation")
	}
	l.Debug().Msg("evaluation completed and cached for subject")

	return slices.Clone(effective)
}

func equalGroups(current, stored []string) bool {
	return sets.Equal(sets.NewFromSlice(current), sets.NewFromSlice(stored))
}

func effectiveRoleForService(
	userGroups []string,
	service model.Service,
	globalAccess map[string]model.UserRole,
) (model.UserRole, bool) {
	effectiveRoles := matchingRoles(userGroups, service.Roles)
	effectiveRoles = append(effectiveRoles, matchingRoles(userGroups, globalAccess)...)
	if len(service.Roles) == 0 {
		effectiveRoles = append(effectiveRoles, model.UserRoleGeneralAccess)
	}
	if len(effectiveRoles) == 0 {
		return "", false
	}

	return model.OrderRoles(effectiveRoles)[0], true
}

func matchingRoles(
	userGroups []string,
	roles map[string]model.UserRole,
) []model.UserRole {
	if len(roles) == 0 {
		return nil
	}

	matched := make([]model.UserRole, 0, len(roles))
	for group, role := range roles {
		if slices.Contains(userGroups, group) {
			matched = append(matched, role)
		}
	}
	return matched
}
