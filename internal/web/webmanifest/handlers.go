package webmanifest

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"git.sr.ht/~icikowski/account-center/internal/consts"
	"git.sr.ht/~icikowski/account-center/internal/i18n"
	"git.sr.ht/~icikowski/account-center/internal/shared/xhttp"
)

type webManifestHandler struct {
	instanceName string
	assetsPath   string
}

// NewHandler creates a new web manifest handler.
func NewHandler(instanceName, assetsPath string) xhttp.RouteBinder {
	return &webManifestHandler{
		instanceName: instanceName,
		assetsPath:   assetsPath,
	}
}

// Bind implements [xhttp.RouteBinder].
func (h *webManifestHandler) Bind(r chi.Router) {
	r.HandleFunc("/", h.webManifest)
}

func (h *webManifestHandler) webManifest(w http.ResponseWriter, r *http.Request) {
	m := h.generateManifest(r.Context())

	w.Header().Set(consts.HeaderContentType, consts.MIMEApplicationManifestJSON)
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(m); err != nil {
		l := zerolog.Ctx(r.Context())
		l.Error().Err(err).Msg("failed to encode web manifest")
	}
}

func (h *webManifestHandler) generateManifest(ctx context.Context) Manifest {
	title := i18n.T(ctx, i18n.KeyGeneralAccountCenter)
	if h.instanceName != "" {
		title = h.instanceName + " | " + title
	}

	return Manifest{
		Name:        title,
		ShortName:   i18n.T(ctx, i18n.KeyGeneralAccountCenter),
		Description: i18n.T(ctx, i18n.KeyGeneralDescription),
		Category: Categories{
			CategoryLifestyle, CategoryProductivity, CategoryUtilities,
		},
		Language:        i18n.T(ctx, i18n.KeyLanguageCode),
		Display:         DisplayStandalone,
		StartURL:        "../",
		Scope:           "../",
		BackgroundColor: "#7db48f",
		ThemeColor:      "#3c8e96",
		Icons: Icons{
			{
				Source:  h.assetsPath + "/img/icon512_rounded.png",
				Type:    consts.MIMEImagePNG,
				Sizes:   IconSizes{IconSize512x512},
				Purpose: IconPurposes{IconPurposeAny},
			},
			{
				Source:  h.assetsPath + "/img/icon512_maskable.png",
				Type:    consts.MIMEImagePNG,
				Sizes:   IconSizes{IconSize512x512},
				Purpose: IconPurposes{IconPurposeMaskable},
			},
		},
		Shortcuts: Shortcuts{
			{
				Name: i18n.T(ctx, i18n.KeySectionCatalog),
				URL:  consts.RouteCatalog,
				Icons: Icons{
					{
						Source:  h.assetsPath + "/img/shortcut-catalog.png",
						Type:    consts.MIMEImagePNG,
						Sizes:   IconSizes{IconSize128x128},
						Purpose: IconPurposes{IconPurposeAny},
					},
				},
			},
			{
				Name: i18n.T(ctx, i18n.KeySectionKnowledgeBase),
				URL:  consts.RouteKnowledgeBase,
				Icons: Icons{
					{
						Source:  h.assetsPath + "/img/shortcut-kb.png",
						Type:    consts.MIMEImagePNG,
						Sizes:   IconSizes{IconSize128x128},
						Purpose: IconPurposes{IconPurposeAny},
					},
				},
			},
		},
		Orientation: OrientationNatural,
		Direction:   DirectionAuto,
	}
}
