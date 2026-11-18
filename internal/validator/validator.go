package validator

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/thomiceli/opengist/internal/i18n"
)

type OpengistValidator struct {
	v *validator.Validate
}

func NewValidator() *OpengistValidator {
	v := validator.New()
	_ = v.RegisterValidation("notreserved", validateReservedKeywords)
	_ = v.RegisterValidation("alphanumdash", validateAlphaNumDash)
	_ = v.RegisterValidation("alphanumdashorempty", validateAlphaNumDashOrEmpty)
	_ = v.RegisterValidation("alphanumdashunder", validateAlphaNumDashUnder)
	_ = v.RegisterValidation("alphanumdashunderorempty", validateAlphaNumDashUnderOrEmpty)
	_ = v.RegisterValidation("gisttopics", validateGistTopics)
	_ = v.RegisterValidation("expirationdate", validateExpirationDate)
	_ = v.RegisterValidation("themecolor", validateThemeColor)
	return &OpengistValidator{v}
}

var ThemeColors = []string{
	"red", "amber", "emerald", "sky", "indigo", "purple", "neutral",
}

func (cv *OpengistValidator) Validate(i interface{}) error {
	return cv.v.Struct(i)
}

func (cv *OpengistValidator) Var(field interface{}, tag string) error {
	return cv.v.Var(field, tag)
}

func ValidationMessages(err *error, locale *i18n.Locale) string {
	errs := (*err).(validator.ValidationErrors)
	messages := make([]string, len(errs))
	for i, e := range errs {
		switch e.Tag() {
		case "max":
			messages[i] = locale.String("validation.is-too-long", e.Field())
		case "required":
			messages[i] = locale.String("validation.should-not-be-empty", e.Field())
		case "excludes":
			messages[i] = locale.String("validation.should-not-include-sub-directory", e.Field())
		case "alphanum":
			messages[i] = locale.String("validation.should-only-contain-alphanumeric-characters", e.Field())
		case "alphanumdash", "alphanumdashorempty":
			messages[i] = locale.String("validation.should-only-contain-alphanumeric-characters-and-dashes", e.Field())
		case "alphanumdashunder", "alphanumdashunderorempty":
			messages[i] = locale.String("validation.should-only-contain-alphanumeric-characters-and-dashes-and-underscores", e.Field())
		case "min":
			messages[i] = locale.String("validation.not-enough", e.Field())
		case "notreserved":
			messages[i] = locale.String("validation.invalid", e.Field())
		case "gisttopics":
			messages[i] = locale.String("validation.invalid-gist-topics")
		case "expirationdate":
			messages[i] = locale.String("validation.invalid-expiration-date")
		}
	}

	return strings.Join(messages, " ; ")
}

// ReservedUsernames are names that would shadow a top-level route. New
// accounts with these names are rejected, except for "admin": it is such a
// common account name that it stays creatable, and its profile is served from
// /@admin.
var ReservedUsernames = []string{
	"api", "assets", "explore", "init", "healthcheck", "preview", "metrics",
	"mfa", "webauthn", "oauth", "login", "register", "logout", "settings",
	"admin", "search", "topics", "users",
}

// IsReservedUsername reports whether a username collides with a top-level route.
func IsReservedUsername(name string) bool {
	for _, reserved := range ReservedUsernames {
		if name == reserved {
			return true
		}
	}
	return false
}

// IsReservedForSignup is the stricter signup rule. "admin" is deliberately
// allowed so the conventional administrator account keeps working; its links
// use the /@admin profile alias.
func IsReservedForSignup(name string) bool {
	return name != "admin" && IsReservedUsername(name)
}

// ProfilePath returns the URL path fragment for a user's profile. Names that
// are also top-level routes use the /@username alias so /admin keeps opening
// the admin panel while the admin's own profile stays reachable.
func ProfilePath(name string) string {
	if IsReservedUsername(name) {
		return "@" + name
	}
	return name
}

func validateReservedKeywords(fl validator.FieldLevel) bool {
	return !IsReservedForSignup(fl.Field().String())
}

func validateAlphaNumDash(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return regexp.MustCompile(`^[a-zA-Z0-9-]+$`).MatchString(value) &&
		regexp.MustCompile(`[a-zA-Z0-9]`).MatchString(value)
}

func validateAlphaNumDashOrEmpty(fl validator.FieldLevel) bool {
	return regexp.MustCompile(`^$|^[a-zA-Z0-9-]+$`).MatchString(fl.Field().String())
}

func validateAlphaNumDashUnder(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return regexp.MustCompile(`^[a-zA-Z0-9-_]+$`).MatchString(value) &&
		regexp.MustCompile(`[a-zA-Z0-9]`).MatchString(value)
}

func validateAlphaNumDashUnderOrEmpty(fl validator.FieldLevel) bool {
	return regexp.MustCompile(`^$|^[a-zA-Z0-9-_]+$`).MatchString(fl.Field().String())
}

func validateThemeColor(fl validator.FieldLevel) bool {
	color := fl.Field().String()
	if color == "" {
		return true
	}
	for _, c := range ThemeColors {
		if c == color {
			return true
		}
	}
	return false
}

func validateGistTopics(fl validator.FieldLevel) bool {
	topicsInput := fl.Field().String()
	if topicsInput == "" {
		return true
	}

	topics := strings.Fields(topicsInput)

	if len(topics) > 10 {
		return false
	}

	for _, tag := range topics {
		if len(tag) > 50 {
			return false
		}
		if !regexp.MustCompile(`^[\p{L}\p{N}-]+$`).MatchString(tag) {
			return false
		}
	}

	return true
}

var dateTimeLayouts = []string{
	"2006-01-02T15:04",
	"2006-01-02T15:04:05",
	time.RFC3339,
}

func validateExpirationDate(fl validator.FieldLevel) bool {
	expire := fl.Parent().FieldByName("Expire")
	if !expire.IsValid() || expire.String() != "custom" {
		return true
	}

	value := strings.TrimSpace(fl.Field().String())
	if value == "" {
		return false
	}

	t, err := ParseDateTime(value)
	if err != nil {
		return false
	}
	return t.After(time.Now())
}

func ParseDateTime(value string) (time.Time, error) {
	for _, layout := range dateTimeLayouts {
		if t, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid datetime: %q", value)
}
