package webmanifest

import (
	"strings"

	"git.sr.ht/~icikowski/account-center/internal/shared/xerror"
)

// Manifest represents a web application manifest, which provides metadata about a web application and is used to
// control how the app appears to users and how it can be launched.
type Manifest struct {
	Name            string      `json:"name"`
	ShortName       string      `json:"short_name"`
	Description     string      `json:"description"`
	Category        Categories  `json:"category,omitempty"`
	Language        string      `json:"lang,omitempty"`
	Display         Display     `json:"display"`
	StartURL        string      `json:"start_url"`
	Scope           string      `json:"scope,omitempty"`
	BackgroundColor string      `json:"background_color,omitempty"`
	ThemeColor      string      `json:"theme_color,omitempty"`
	Icons           Icons       `json:"icons"`
	Shortcuts       Shortcuts   `json:"shortcuts,omitempty"`
	Orientation     Orientation `json:"orientation,omitempty"`
	Direction       Direction   `json:"dir,omitempty"`
}

// Validate checks if the [Manifest] is valid.
func (m Manifest) Validate() error {
	errs := make([]error, 0, 9)

	if m.Name == "" {
		errs = append(errs, xerror.NewValidationError("manifest name cannot be empty"))
	}
	if m.ShortName == "" {
		errs = append(errs, xerror.NewValidationError("manifest short name cannot be empty"))
	}
	if len(m.Category) != 0 {
		if err := m.Category.Validate(); err != nil {
			errs = append(errs, err)
		}
	}
	if !m.Display.IsValid() {
		errs = append(errs, xerror.NewValidationError("invalid display mode"))
	}
	if m.StartURL == "" {
		errs = append(errs, xerror.NewValidationError("manifest start URL cannot be empty"))
	}
	if len(m.Icons) != 0 {
		if err := m.Icons.Validate(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(m.Shortcuts) != 0 {
		if err := m.Shortcuts.Validate(); err != nil {
			errs = append(errs, err)
		}
	}
	if m.Orientation != "" && !m.Orientation.IsValid() {
		errs = append(errs, xerror.NewValidationError("invalid orientation"))
	}
	if m.Direction != "" && !m.Direction.IsValid() {
		errs = append(errs, xerror.NewValidationError("invalid text direction"))
	}

	if len(errs) != 0 {
		return xerror.NewValidationError("invalid manifest", errs...)
	}
	return nil
}

// Category represents the category of a web application.
type Category string

// Valid categories based on https://www.w3.org/TR/appmanifest/#categories-member.
const (
	CategoryBooks           Category = "books"
	CategoryBusiness        Category = "business"
	CategoryEducation       Category = "education"
	CategoryEntertainment   Category = "entertainment"
	CategoryFinance         Category = "finance"
	CategoryFitness         Category = "fitness"
	CategoryFood            Category = "food"
	CategoryGames           Category = "games"
	CategoryGovernment      Category = "government"
	CategoryHealth          Category = "health"
	CategoryKids            Category = "kids"
	CategoryLifestyle       Category = "lifestyle"
	CategoryMagazines       Category = "magazines"
	CategoryMedical         Category = "medical"
	CategoryMusic           Category = "music"
	CategoryNavigation      Category = "navigation"
	CategoryNews            Category = "news"
	CategoryPersonalization Category = "personalization"
	CategoryPhoto           Category = "photo"
	CategoryPolitics        Category = "politics"
	CategoryProductivity    Category = "productivity"
	CategorySecurity        Category = "security"
	CategoryShopping        Category = "shopping"
	CategorySocial          Category = "social"
	CategorySports          Category = "sports"
	CategoryTravel          Category = "travel"
	CategoryUtilities       Category = "utilities"
	CategoryWeather         Category = "weather"
)

// IsValid checks if the [Category] is valid.
func (c Category) IsValid() bool {
	switch c {
	case CategoryBooks,
		CategoryBusiness,
		CategoryEducation,
		CategoryEntertainment,
		CategoryFinance,
		CategoryFitness,
		CategoryFood,
		CategoryGames,
		CategoryGovernment,
		CategoryHealth,
		CategoryKids,
		CategoryLifestyle,
		CategoryMagazines,
		CategoryMedical,
		CategoryMusic,
		CategoryNavigation,
		CategoryNews,
		CategoryPersonalization,
		CategoryPhoto,
		CategoryPolitics,
		CategoryProductivity,
		CategorySecurity,
		CategoryShopping,
		CategorySocial,
		CategorySports,
		CategoryTravel,
		CategoryUtilities,
		CategoryWeather:
		return true
	default:
		return false
	}
}

// Categories represents a list of [Category] values.
type Categories []Category

// Validate checks if the [Categories] are valid.
func (cs Categories) Validate() error {
	errs := make([]error, 0, len(cs))
	for i, c := range cs {
		if !c.IsValid() {
			errs = append(errs, xerror.NewItemValidationError(i, xerror.NewValidationError("invalid category")))
		}
	}

	if len(errs) != 0 {
		return xerror.NewValidationError("invalid categories", errs...)
	}
	return nil
}

// Display represents the display mode of a web application, which determines how the app is presented to users when
// launched.
type Display string

// Valid display modes based on https://www.w3.org/TR/appmanifest/#display-member.
const (
	DisplayStandalone Display = "standalone"
	DisplayFullscreen Display = "fullscreen"
	DisplayMinimalUI  Display = "minimal-ui"
	DisplayBrowser    Display = "browser"
)

// IsValid checks if the [Display] is valid.
func (d Display) IsValid() bool {
	switch d {
	case DisplayStandalone, DisplayFullscreen, DisplayMinimalUI, DisplayBrowser:
		return true
	default:
		return false
	}
}

// Icon represents an icon for a web application, which is used to represent the app in various contexts.
type Icon struct {
	Source  string       `json:"src"`
	Type    string       `json:"type,omitempty"`
	Sizes   IconSizes    `json:"sizes,omitempty"`
	Purpose IconPurposes `json:"purpose,omitempty"`
}

// Validate checks if the [Icon] is valid.
func (i Icon) Validate() error {
	errs := make([]error, 0, 4)

	if i.Source == "" {
		errs = append(errs, xerror.NewValidationError("icon source cannot be empty"))
	}
	if len(i.Sizes) != 0 {
		if err := i.Sizes.Validate(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(i.Purpose) != 0 {
		if err := i.Purpose.Validate(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return xerror.NewValidationError("invalid icon", errs...)
	}
	return nil
}

// Icons represents a list of [Icon] values.
type Icons []Icon

// Validate checks if the [Icons] are valid.
func (is Icons) Validate() error {
	errs := make([]error, 0, len(is))
	for i, icon := range is {
		if err := icon.Validate(); err != nil {
			errs = append(errs, xerror.NewItemValidationError(i, err))
		}
	}

	if len(errs) != 0 {
		return xerror.NewValidationError("invalid icons", errs...)
	}
	return nil
}

// IconSize represents the size of an icon, which is used to specify the dimensions of an icon in a web application
// manifest.
type IconSize string

// Valid icon sizes.
const (
	IconSize64x64   IconSize = "64x64"
	IconSize128x128 IconSize = "128x128"
	IconSize256x256 IconSize = "256x256"
	IconSize512x512 IconSize = "512x512"
)

// IsValid checks if the [IconSize] is valid.
func (s IconSize) IsValid() bool {
	switch s {
	case IconSize64x64, IconSize128x128, IconSize256x256, IconSize512x512:
		return true
	default:
		return false
	}
}

// IconSizes represents a list of [IconSize] values.
type IconSizes []IconSize

// Validate checks if the [IconSizes] are valid.
func (ss IconSizes) Validate() error {
	errs := make([]error, 0, len(ss))
	for i, s := range ss {
		if !s.IsValid() {
			errs = append(errs, xerror.NewItemValidationError(i, xerror.NewValidationError("invalid icon size")))
		}
	}

	if len(errs) != 0 {
		return xerror.NewValidationError("invalid icon sizes", errs...)
	}
	return nil
}

// MarshalJSON converts the [IconSizes] to a JSON string in the format specified by the web application manifest
// standard.
func (ss IconSizes) MarshalJSON() ([]byte, error) {
	if len(ss) == 0 {
		return []byte(`""`), nil
	}

	sizes := make([]string, len(ss))
	for i, s := range ss {
		sizes[i] = string(s)
	}

	return []byte(`"` + strings.Join(sizes, " ") + `"`), nil
}

// UnmarshalJSON parses a JSON string in the format specified by the web application manifest standard and converts it
// to [IconSizes].
func (ss *IconSizes) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`)
	if str == "" {
		*ss = []IconSize{}
		return nil
	}

	sizeStrs := strings.Split(str, " ")
	sizes := make([]IconSize, len(sizeStrs))
	for i, sizeStr := range sizeStrs {
		sizes[i] = IconSize(sizeStr)
	}
	*ss = sizes
	return nil
}

// IconPurpose represents the purpose of an icon, which is used to specify the intended usage of an icon in a web
// application manifest.
type IconPurpose string

// Valid icon purposes based on https://www.w3.org/TR/appmanifest/#purpose-member.
const (
	IconPurposeAny        IconPurpose = "any"
	IconPurposeMaskable   IconPurpose = "maskable"
	IconPurposeMonochrome IconPurpose = "monochrome"
)

// IsValid checks if the [IconPurpose] is valid.
func (p IconPurpose) IsValid() bool {
	switch p {
	case IconPurposeAny, IconPurposeMaskable, IconPurposeMonochrome:
		return true
	default:
		return false
	}
}

// IconPurposes represents a list of [IconPurpose] values.
type IconPurposes []IconPurpose

// Validate checks if the [IconPurposes] are valid.
func (ps IconPurposes) Validate() error {
	errs := make([]error, 0, len(ps))
	for i, p := range ps {
		if !p.IsValid() {
			errs = append(errs, xerror.NewItemValidationError(i, xerror.NewValidationError("invalid icon purpose")))
		}
	}

	if len(errs) != 0 {
		return xerror.NewValidationError("invalid icon purposes", errs...)
	}
	return nil
}

// MarshalJSON converts the [IconPurposes] to a JSON string in the format specified by the web application manifest
// standard.
func (ps IconPurposes) MarshalJSON() ([]byte, error) {
	if len(ps) == 0 {
		return []byte(`""`), nil
	}

	purposes := make([]string, len(ps))
	for i, p := range ps {
		purposes[i] = string(p)
	}

	return []byte(`"` + strings.Join(purposes, " ") + `"`), nil
}

// UnmarshalJSON parses a JSON string in the format specified by the web application manifest standard and converts it
// to [IconPurposes].
func (ps *IconPurposes) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`)
	if str == "" {
		*ps = []IconPurpose{}
		return nil
	}

	purposeStrs := strings.Split(str, " ")
	purposes := make([]IconPurpose, len(purposeStrs))
	for i, purposeStr := range purposeStrs {
		purposes[i] = IconPurpose(purposeStr)
	}
	*ps = purposes
	return nil
}

// Shortcut represents a shortcut for a web application, which provides quick access to specific features or sections of
// the app when launched from a user's device.
type Shortcut struct {
	Name        string `json:"name"`
	ShortName   string `json:"short_name,omitempty"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url"`
	Icons       Icons  `json:"icons,omitempty"`
}

// Validate checks if the [Shortcut] is valid.
func (s Shortcut) Validate() error {
	errs := make([]error, 0, 4)

	if s.Name == "" {
		errs = append(errs, xerror.NewValidationError("shortcut name cannot be empty"))
	}
	if s.URL == "" {
		errs = append(errs, xerror.NewValidationError("shortcut URL cannot be empty"))
	}
	if len(s.Icons) != 0 {
		if err := s.Icons.Validate(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return xerror.NewValidationError("invalid shortcut", errs...)
	}
	return nil
}

// Shortcuts represents a list of [Shortcut] values.
type Shortcuts []Shortcut

// Validate checks if the [Shortcuts] are valid.
func (ss Shortcuts) Validate() error {
	errs := make([]error, 0, len(ss))
	for i, s := range ss {
		if err := s.Validate(); err != nil {
			errs = append(errs, xerror.NewItemValidationError(i, err))
		}
	}

	if len(errs) != 0 {
		return xerror.NewValidationError("invalid shortcuts", errs...)
	}
	return nil
}

// Orientation represents the default orientation for a web application, which determines the orientation of the app.
type Orientation string

// Valid orientations based on https://www.w3.org/TR/appmanifest/#orientation-member.
const (
	OrientationAny                Orientation = "any"
	OrientationNatural            Orientation = "natural"
	OrientationPortrait           Orientation = "portrait"
	OrientationPortraitPrimary    Orientation = "portrait-primary"
	OrientationPortraitSecondary  Orientation = "portrait-secondary"
	OrientationLandscape          Orientation = "landscape"
	OrientationLandscapePrimary   Orientation = "landscape-primary"
	OrientationLandscapeSecondary Orientation = "landscape-secondary"
)

// IsValid checks if the [Orientation] is valid.
func (o Orientation) IsValid() bool {
	switch o {
	case OrientationAny,
		OrientationNatural,
		OrientationPortrait,
		OrientationPortraitPrimary,
		OrientationPortraitSecondary,
		OrientationLandscape,
		OrientationLandscapePrimary,
		OrientationLandscapeSecondary:
		return true
	default:
		return false
	}
}

// Direction represents the text direction for a web application, which determines the direction in which text is
// displayed.
type Direction string

// Valid directions based on https://www.w3.org/TR/appmanifest/#dir-member.
const (
	DirectionAuto Direction = "auto"
	DirectionLTR  Direction = "ltr"
	DirectionRTL  Direction = "rtl"
)

// IsValid checks if the [Direction] is valid.
func (d Direction) IsValid() bool {
	switch d {
	case DirectionAuto, DirectionLTR, DirectionRTL:
		return true
	default:
		return false
	}
}
