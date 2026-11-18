package i18n

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/LiuShen-Fork/Qingyu-Gist/internal/i18n/locales"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
	"gopkg.in/yaml.v3"
)

var Locales = NewLocaleStore()

type LocaleStore struct {
	Locales map[string]*Locale
}

type Locale struct {
	Code     string
	Name     string
	Messages map[string]string
}

// NewLocaleStore creates a new LocaleStore
func NewLocaleStore() *LocaleStore {
	return &LocaleStore{
		Locales: make(map[string]*Locale),
	}
}

// loadLocaleFromYAML loads a single Locale from a given YAML file
func (store *LocaleStore) loadLocaleFromYAML(localeCode, path string) error {
	a, err := locales.Files.Open(path)
	if err != nil {
		return err
	}
	data, err := io.ReadAll(a)
	if err != nil {
		return err
	}

	tag, err := language.Parse(localeCode)
	if err != nil {
		return err
	}

	name := display.Self.Name(tag)
	switch tag {
	case language.AmericanEnglish:
		name = "English"
	case language.EuropeanSpanish:
		name = "Español"
	}
	switch localeCode {
	case "zh-CN":
		name = "简体中文"
	case "zh-TW":
		name = "繁體中文"
	}

	locale := &Locale{
		Code:     localeCode,
		Name:     cases.Title(language.English).String(name),
		Messages: make(map[string]string),
	}

	err = yaml.Unmarshal(data, &locale.Messages)
	if err != nil {
		return err
	}

	store.Locales[localeCode] = locale
	return nil
}

func (store *LocaleStore) LoadAll() error {
	return fs.WalkDir(locales.Files, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			localeKey := strings.TrimSuffix(path, filepath.Ext(path))
			err := store.loadLocaleFromYAML(localeKey, path)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (store *LocaleStore) GetLocale(lang string) (*Locale, error) {
	_, ok := store.Locales[lang]
	if !ok {
		return nil, fmt.Errorf("locale %s not found", lang)
	}

	return store.Locales[lang], nil
}

func (store *LocaleStore) HasLocale(lang string) bool {
	_, ok := store.Locales[lang]
	return ok
}

func (store *LocaleStore) MatchTag(langs []language.Tag) string {
	for _, lang := range langs {
		if store.HasLocale(lang.String()) {
			return lang.String()
		}
	}

	return "en-US"
}

func (l *Locale) String(key string, args ...any) string {
	message := l.Messages[key]

	if message == "" {
		fallback := Locales.Locales["en-US"]
		if fallback == nil || fallback.Code == l.Code {
			return key
		}
		return fallback.String(key, args...)
	}

	if len(args) == 0 {
		return message
	}

	return fmt.Sprintf(message, args...)
}

func (l *Locale) Tr(key string, args ...any) template.HTML {
	message := l.Messages[key]

	if message == "" {
		fallback := Locales.Locales["en-US"]
		if fallback == nil || fallback.Code == l.Code {
			return template.HTML(key)
		}
		return fallback.Tr(key, args...)
	}

	if len(args) == 0 {
		return template.HTML(message)
	}

	return template.HTML(fmt.Sprintf(message, args...))
}

// TimeAgo renders a Unix timestamp as a localized, approximate relative time
// ("3 minutes ago", "2 个月前"). Using the request locale keeps dates from
// falling back to English halfway through an otherwise translated page.
func (l *Locale) TimeAgo(timestamp int64) template.HTML {
	return l.timeAgo(time.Unix(timestamp, 0), time.Now())
}

// TimeAgoStr is the string-timestamp variant used by git revision metadata.
func (l *Locale) TimeAgoStr(timestamp string) template.HTML {
	seconds, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return l.Tr("time.just-now")
	}
	return l.TimeAgo(seconds)
}

func (l *Locale) timeAgo(t, now time.Time) template.HTML {
	diff := now.Sub(t)
	if diff < 10*time.Second {
		return l.Tr("time.just-now")
	}

	unit := func(key string, value int) template.HTML {
		if value == 1 {
			singular := strings.TrimSuffix(key, "s")
			if message := l.Messages[singular]; message != "" {
				return template.HTML(message)
			}
		}
		return l.Tr(key, value)
	}

	switch {
	case diff < time.Minute:
		return unit("time.seconds", int(diff/time.Second))
	case diff < time.Hour:
		return unit("time.minutes", int(diff/time.Minute))
	case diff < 24*time.Hour:
		return unit("time.hours", int(diff/time.Hour))
	case diff < 30*24*time.Hour:
		return unit("time.days", int(diff/(24*time.Hour)))
	case diff < 365*24*time.Hour:
		return unit("time.months", int(diff/(30*24*time.Hour)))
	default:
		return unit("time.years", int(diff/(365*24*time.Hour)))
	}
}
