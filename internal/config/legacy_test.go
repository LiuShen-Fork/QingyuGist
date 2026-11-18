package config

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnvLookupPrefersNewNamespace(t *testing.T) {
	t.Setenv("QYG_TEST_VALUE", "new")
	t.Setenv("OG_TEST_VALUE", "old")
	if got, ok := envLookup("QYG_TEST_VALUE"); !ok || got != "new" {
		t.Fatalf("envLookup = %q, %v; want new, true", got, ok)
	}
}

func TestEnvLookupFallsBackToLegacy(t *testing.T) {
	t.Setenv("QYG_TEST_ONLY_OLD", "")
	os.Unsetenv("QYG_TEST_ONLY_OLD")
	t.Setenv("OG_TEST_ONLY_OLD", "old")
	if got, ok := envLookup("QYG_TEST_ONLY_OLD"); !ok || got != "old" {
		t.Fatalf("envLookup = %q, %v; want old, true", got, ok)
	}
}

func TestEnvLookupMapsHome(t *testing.T) {
	os.Unsetenv("QYG_HOME")
	t.Setenv(legacyHomeEnv, "/legacy/home")
	if got, ok := envLookup("QYG_HOME"); !ok || got != "/legacy/home" {
		t.Fatalf("envLookup = %q, %v; want legacy home", got, ok)
	}
}

func TestMigrateLegacyDataMovesDatabaseAndSecret(t *testing.T) {
	home := t.TempDir()
	previous := C
	t.Cleanup(func() { C = previous })
	C = &config{QingyuGistHome: home, DBUri: legacyDatabase}

	if err := os.WriteFile(filepath.Join(home, legacyDatabase), []byte("db"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, legacySecretName), []byte("secret"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(home, legacyIndexDir), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, legacyIndexDir, "index_meta.json"), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	migrateLegacyData(&out)

	if C.DBUri != "qingyugist.db" {
		t.Fatalf("DBUri = %q; want qingyugist.db", C.DBUri)
	}
	if _, err := os.Stat(filepath.Join(home, "qingyugist.db")); err != nil {
		t.Fatalf("database was not migrated: %v", err)
	}
	if data, err := os.ReadFile(filepath.Join(home, "qingyugist-secret.key")); err != nil || string(data) != "secret" {
		t.Fatalf("secret was not copied: %q, %v", data, err)
	}
	if _, err := os.Stat(filepath.Join(home, "qingyugist.index", "index_meta.json")); err != nil {
		t.Fatalf("search index was not migrated: %v", err)
	}
	if !strings.Contains(out.String(), "Migrated database") {
		t.Fatalf("migration was not reported: %q", out.String())
	}
}

func TestMigrateLegacyDataKeepsExistingNewDatabase(t *testing.T) {
	home := t.TempDir()
	previous := C
	t.Cleanup(func() { C = previous })
	C = &config{QingyuGistHome: home, DBUri: legacyDatabase}

	for name, data := range map[string]string{
		legacyDatabase:          "old",
		"qingyugist.db":         "new",
		legacySecretName:        "old-secret",
		"qingyugist-secret.key": "new-secret",
	} {
		if err := os.WriteFile(filepath.Join(home, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}

	migrateLegacyData(&bytes.Buffer{})

	got, err := os.ReadFile(filepath.Join(home, "qingyugist.db"))
	if err != nil || string(got) != "new" {
		t.Fatalf("existing database was overwritten: %q, %v", got, err)
	}
	secret, err := os.ReadFile(filepath.Join(home, "qingyugist-secret.key"))
	if err != nil || string(secret) != "new-secret" {
		t.Fatalf("existing secret was overwritten: %q, %v", secret, err)
	}
}
