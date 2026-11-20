package i18n

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
)

// newTestStore builds a store with the given messages per locale code, without
// touching the embedded files.
func newTestStore(messages map[string]map[string]string) *LocaleStore {
	store := NewLocaleStore()
	for code, entries := range messages {
		store.Locales[code] = &Locale{Code: code, Name: code, Messages: entries}
	}
	store.markCoverage()
	return store
}

func entries(n int) map[string]string {
	out := make(map[string]string, n)
	for i := range n {
		out["key."+string(rune('a'+i))] = "value"
	}
	return out
}

func TestMarkCoverageDropsPartialLocales(t *testing.T) {
	store := newTestStore(map[string]map[string]string{
		"en-US": entries(20),
		"zh-CN": entries(20),
		// 18 of 20 is 90%: finished enough to be useful, but still short of the
		// bar, so it must not appear in the switcher.
		"de-DE": limit(entries(20), 18),
		// 19 of 20 is 95%: served.
		"fr-FR": limit(entries(20), 19),
		// An empty file is a stub, not a translation.
		"ja-JP": {},
	})

	require.True(t, store.Locales["en-US"].Complete)
	require.True(t, store.Locales["zh-CN"].Complete)
	require.True(t, store.Locales["fr-FR"].Complete)
	require.False(t, store.Locales["de-DE"].Complete)
	require.False(t, store.Locales["ja-JP"].Complete)

	offered := store.Offered()
	require.Len(t, offered, 3)
	for _, code := range []string{"en-US", "zh-CN", "fr-FR"} {
		require.Contains(t, offered, code)
	}
	require.Equal(t, []string{"en-US", "fr-FR", "zh-CN"}, store.offeredCodes)
}

func limit(m map[string]string, n int) map[string]string {
	i := 0
	for key := range m {
		if i >= n {
			delete(m, key)
		}
		i++
	}
	return m
}

func TestMarkCoverageAlwaysKeepsTheDefault(t *testing.T) {
	// A locale is scored against en-US, so en-US has to be scored against its
	// own keys. Nine blanks out of ten leaves it below the bar.
	stub := entries(10)
	for key := range stub {
		stub[key] = ""
	}
	stub["key.a"] = "value"

	store := newTestStore(map[string]map[string]string{
		"en-US": stub,
		"zh-CN": entries(10),
	})

	require.False(t, store.Locales["en-US"].Complete)
	require.Contains(t, store.Offered(), "en-US")
}

func TestMatchTagIgnoresIncompleteLocales(t *testing.T) {
	store := newTestStore(map[string]map[string]string{
		"en-US": entries(10),
		"zh-CN": entries(10),
		"de-DE": limit(entries(10), 5),
	})

	require.Equal(t, "en-US", store.MatchTag(ParseAcceptLanguage("de-DE,de;q=0.9")))
	require.Equal(t, "en-US", store.MatchTag(ParseAcceptLanguage("ja-JP")))
	require.Equal(t, "en-US", store.MatchTag(nil))
}

func TestMatchTagMatchesBaseLanguage(t *testing.T) {
	store := newTestStore(map[string]map[string]string{
		"en-US": entries(10),
		"zh-CN": entries(10),
	})

	require.Equal(t, "zh-CN", store.MatchTag(ParseAcceptLanguage("zh")))
	require.Equal(t, "zh-CN", store.MatchTag(ParseAcceptLanguage("zh-Hans-CN")))
	require.Equal(t, "en-US", store.MatchTag(ParseAcceptLanguage("en-GB")))
	require.Equal(t, "en-US", store.MatchTag([]language.Tag{language.English}))
}

func TestUsableRejectsIncompleteLocales(t *testing.T) {
	store := newTestStore(map[string]map[string]string{
		"en-US": entries(10),
		"de-DE": limit(entries(10), 5),
	})

	require.True(t, store.Usable("en-US"))
	require.False(t, store.Usable("de-DE"))
	require.False(t, store.Usable("xx-YY"))
}
