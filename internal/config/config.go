package config

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/LiuShen-Fork/QingyuGist/internal/session"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

var QingyuGistVersion = ""

var C *config

var SecretKey []byte

// SSH server modes: the canonical values of ssh.git-enabled. The legacy
// booleans are still accepted (true → builtin, false → disabled).
const (
	SshServerBuiltin  = "builtin"  // QingyuGist's embedded SSH server
	SshServerHost     = "host"     // delegate to the host's OpenSSH
	SshServerDisabled = "disabled" // no SSH git access
)

// Not using nested structs because the library
// doesn't support dot notation in this case sadly
type config struct {
	SecretKey string `yaml:"secret-key" env:"QYG_SECRET_KEY"`

	LogLevel       string `yaml:"log-level" env:"QYG_LOG_LEVEL"`
	LogOutput      string `yaml:"log-output" env:"QYG_LOG_OUTPUT"`
	LogPath        string `yaml:"log-path" env:"QYG_LOG_PATH"`
	ExternalUrl    string `yaml:"external-url" env:"QYG_EXTERNAL_URL"`
	QingyuGistHome string `yaml:"qingyugist-home" env:"QYG_HOME"`
	// LegacyHome accepts the pre-fork YAML key so an existing config.yml keeps
	// pointing at the original data directory during upgrades.
	LegacyHome string `yaml:"opengist-home"`

	DBUri      string `yaml:"db-uri" env:"QYG_DB_URI"`
	DBFilename string `yaml:"db-filename" env:"QYG_DB_FILENAME"` // deprecated

	IndexEnabled  bool   `yaml:"index.enabled" env:"QYG_INDEX_ENABLED"` // deprecated
	Index         string `yaml:"index" env:"QYG_INDEX"`
	BleveDirname  string `yaml:"index.dirname" env:"QYG_INDEX_DIRNAME"` // deprecated
	MeiliHost     string `yaml:"index.meili.host" env:"QYG_MEILI_HOST"`
	MeiliAPIKey   string `yaml:"index.meili.api-key" env:"QYG_MEILI_API_KEY"`
	SearchDefault string `yaml:"search.default" env:"QYG_SEARCH_DEFAULT"`

	GitDefaultBranch string `yaml:"git.default-branch" env:"QYG_GIT_DEFAULT_BRANCH"`

	SqliteJournalMode string `yaml:"sqlite.journal-mode" env:"QYG_SQLITE_JOURNAL_MODE"`

	HttpHost string `yaml:"http.host" env:"QYG_HTTP_HOST"`
	HttpPort string `yaml:"http.port" env:"QYG_HTTP_PORT"`
	HttpGit  bool   `yaml:"http.git-enabled" env:"QYG_HTTP_GIT_ENABLED"`

	ApiEnabled bool `yaml:"api.enabled" env:"QYG_API_ENABLED"`

	DisableFileUpload bool `yaml:"disable-file-upload" env:"QYG_DISABLE_FILE_UPLOAD"`

	UnixSocketPermissions string `yaml:"unix-socket-permissions" env:"QYG_UNIX_SOCKET_PERMISSIONS"`

	SshGit                string `yaml:"ssh.git-enabled" env:"QYG_SSH_GIT_ENABLED"` // builtin | host | disabled (true → builtin, false → disabled)
	SshAuthorizedKeysFile string `yaml:"ssh.authorized-keys-file" env:"QYG_SSH_AUTHORIZED_KEYS_FILE"`
	SshHost               string `yaml:"ssh.host" env:"QYG_SSH_HOST"`
	SshPort               string `yaml:"ssh.port" env:"QYG_SSH_PORT"`
	SshExternalDomain     string `yaml:"ssh.external-domain" env:"QYG_SSH_EXTERNAL_DOMAIN"`
	SshUsername           string `yaml:"ssh.username" env:"QYG_SSH_USERNAME"`

	GithubClientKey string `yaml:"github.client-key" env:"QYG_GITHUB_CLIENT_KEY"`
	GithubSecret    string `yaml:"github.secret" env:"QYG_GITHUB_SECRET"`

	GitlabClientKey string `yaml:"gitlab.client-key" env:"QYG_GITLAB_CLIENT_KEY"`
	GitlabSecret    string `yaml:"gitlab.secret" env:"QYG_GITLAB_SECRET"`
	GitlabUrl       string `yaml:"gitlab.url" env:"QYG_GITLAB_URL"`
	GitlabName      string `yaml:"gitlab.name" env:"QYG_GITLAB_NAME"`

	GiteaClientKey string `yaml:"gitea.client-key" env:"QYG_GITEA_CLIENT_KEY"`
	GiteaSecret    string `yaml:"gitea.secret" env:"QYG_GITEA_SECRET"`
	GiteaUrl       string `yaml:"gitea.url" env:"QYG_GITEA_URL"`
	GiteaName      string `yaml:"gitea.name" env:"QYG_GITEA_NAME"`

	OIDCProviderName   string `yaml:"oidc.provider-name" env:"QYG_OIDC_PROVIDER_NAME"`
	OIDCClientKey      string `yaml:"oidc.client-key" env:"QYG_OIDC_CLIENT_KEY"`
	OIDCSecret         string `yaml:"oidc.secret" env:"QYG_OIDC_SECRET"`
	OIDCDiscoveryUrl   string `yaml:"oidc.discovery-url" env:"QYG_OIDC_DISCOVERY_URL"`
	OIDCGroupClaimName string `yaml:"oidc.group-claim-name" env:"QYG_OIDC_GROUP_CLAIM_NAME"`
	OIDCAdminGroup     string `yaml:"oidc.admin-group" env:"QYG_OIDC_ADMIN_GROUP"`

	MetricsEnabled bool   `yaml:"metrics.enabled" env:"QYG_METRICS_ENABLED"`
	MetricsHost    string `yaml:"metrics.host" env:"QYG_METRICS_HOST"`
	MetricsPort    string `yaml:"metrics.port" env:"QYG_METRICS_PORT"`

	LDAPUrl             string `yaml:"ldap.url" env:"QYG_LDAP_URL"`
	LDAPBindDn          string `yaml:"ldap.bind-dn" env:"QYG_LDAP_BIND_DN"`
	LDAPBindCredentials string `yaml:"ldap.bind-credentials" env:"QYG_LDAP_BIND_CREDENTIALS"`
	LDAPSearchBase      string `yaml:"ldap.search-base" env:"QYG_LDAP_SEARCH_BASE"`
	LDAPSearchFilter    string `yaml:"ldap.search-filter" env:"QYG_LDAP_SEARCH_FILTER"`

	CustomName    string       `yaml:"custom.name" env:"QYG_CUSTOM_NAME"`
	CustomLogo    string       `yaml:"custom.logo" env:"QYG_CUSTOM_LOGO"`
	CustomFavicon string       `yaml:"custom.favicon" env:"QYG_CUSTOM_FAVICON"`
	StaticLinks   []StaticLink `yaml:"custom.static-links" env:"QYG_CUSTOM_STATIC_LINK"`
}

type StaticLink struct {
	Name string `yaml:"name" env:"QYG_CUSTOM_STATIC_LINK_#_NAME"`
	Path string `yaml:"path" env:"QYG_CUSTOM_STATIC_LINK_#_PATH"`
}

// SshEnabled reports whether SSH git access is offered in any form.
func (c *config) SshEnabled() bool {
	return c.SshGit == SshServerBuiltin || c.SshGit == SshServerHost
}

// SshBuiltin reports whether QingyuGist runs its own embedded SSH server.
func (c *config) SshBuiltin() bool {
	return c.SshGit == SshServerBuiltin
}

// SshManagesAuthorizedKeys reports whether QingyuGist maintains an authorized_keys
// file: host mode with a configured file path.
func (c *config) SshManagesAuthorizedKeys() bool {
	return c.SshGit == SshServerHost && c.SshAuthorizedKeysFile != ""
}

func configWithDefaults() (*config, error) {
	c := &config{}

	c.SecretKey = ""

	c.LogLevel = "warn"
	c.LogOutput = "stdout,file"
	c.QingyuGistHome = ""
	c.DBUri = "qingyugist.db"
	c.Index = "bleve"
	c.SearchDefault = "content"

	c.SqliteJournalMode = "WAL"

	c.HttpHost = "0.0.0.0"
	c.HttpPort = "6157"
	c.HttpGit = true

	c.ApiEnabled = true

	c.UnixSocketPermissions = "0666"

	c.SshGit = SshServerBuiltin
	c.SshHost = "0.0.0.0"
	c.SshPort = "2222"

	c.GitlabName = "GitLab"

	c.GiteaUrl = "https://gitea.com"
	c.GiteaName = "Gitea"

	c.MetricsEnabled = false
	c.MetricsHost = "0.0.0.0"
	c.MetricsPort = "6158"

	return c, nil
}

func InitConfig(configPath string, out io.Writer) error {
	// Default values
	c, err := configWithDefaults()
	if err != nil {
		return err
	}

	if err = loadConfigFromYaml(c, configPath, out); err != nil {
		return err
	}

	if err = loadConfigFromEnv(c, out); err != nil {
		return err
	}

	// ssh.git-enabled accepts the explicit modes (builtin, host, disabled) as well
	// as the legacy booleans (true → builtin, false → disabled). Collapse whatever
	// was provided into a canonical mode.
	c.SshGit = normalizeSshGitMode(c.SshGit)

	if c.QingyuGistHome == "" && c.LegacyHome != "" {
		c.QingyuGistHome = c.LegacyHome
	}

	if c.QingyuGistHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("qingyugist home directory is not set and current user home directory could not be determined; please specify the qingyugist home directory manually via the configuration")
		}

		c.QingyuGistHome = filepath.Join(homeDir, ".qingyugist")
		if !pathExists(c.QingyuGistHome) {
			legacyHome := filepath.Join(homeDir, "."+legacyName)
			if pathExists(legacyHome) {
				c.QingyuGistHome = legacyHome
				_, _ = fmt.Fprintf(out, "Using existing data directory %s; move it to %s when convenient.\n", legacyHome, filepath.Join(homeDir, ".qingyugist"))
			}
		}
	}

	if err = checks(c); err != nil {
		return err
	}

	C = c

	if c.LogPath == "" {
		c.LogPath = filepath.Join(GetHomeDir(), "log")
	}

	migrateLegacyData(out)

	if err = migrateConfig(); err != nil {
		return err
	}

	if err = os.Setenv("QYG_HOME_INTERNAL", GetHomeDir()); err != nil {
		return err
	}
	ExportLegacyEnv()
	return nil
}

func InitLog() {
	if err := os.MkdirAll(C.LogPath, 0755); err != nil {
		panic(err)
	}

	var level zerolog.Level
	level, err := zerolog.ParseLevel(C.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}

	var logWriters []io.Writer
	logOutputTypes := strings.Split(strings.ToLower(C.LogOutput), ",")
	slices.Sort(logOutputTypes)
	logOutputTypes = slices.Compact(logOutputTypes)

	consoleWriter := zerolog.NewConsoleWriter(
		func(w *zerolog.ConsoleWriter) {
			w.TimeFormat = time.TimeOnly
			w.FormatCaller = func(i interface{}) string {
				file := i.(string)
				index := strings.Index(file, "internal")
				if index == -1 {
					return file
				}
				return file[index:]
			}
		},
	)

	for _, logOutputType := range logOutputTypes {
		logOutputType = strings.TrimSpace(logOutputType)
		if !slices.Contains([]string{"stdout", "file"}, logOutputType) {
			defer func() { log.Warn().Msg("Invalid log output type: " + logOutputType) }()
			continue
		}

		switch logOutputType {
		case "stdout":
			logWriters = append(logWriters, consoleWriter)
			defer func() { log.Debug().Msg("Logging to stdout") }()
		case "file":
			file, err := os.OpenFile(filepath.Join(C.LogPath, "qingyugist.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				panic(err)
			}
			logWriters = append(logWriters, file)
			defer func() { log.Debug().Msg("Logging to file: " + file.Name()) }()
		}
	}
	if len(logWriters) == 0 {
		logWriters = append(logWriters, consoleWriter)
		defer func() { log.Warn().Msg("No valid log outputs, defaulting to stdout") }()
	}

	multi := zerolog.MultiLevelWriter(logWriters...)
	log.Logger = zerolog.New(multi).Level(level).With().Caller().Timestamp().Logger()

	if !slices.Contains([]string{"debug", "info", "warn", "error", "fatal"}, strings.ToLower(C.LogLevel)) {
		log.Warn().Msg("Invalid log level: " + C.LogLevel)
	}
}

// gitVersionRegex extracts the major and minor numbers from a git version string.
// It tolerates a leading "git version " prefix as well as suffixes such as the
// ".windows.1" appended by Git for Windows (e.g. "git version 2.50.1.windows.1").
var gitVersionRegex = regexp.MustCompile(`(\d+)\.(\d+)`)

func CheckGitVersion(version string) (bool, error) {
	matches := gitVersionRegex.FindStringSubmatch(version)
	if matches == nil {
		return false, fmt.Errorf("invalid version string: %q", version)
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return false, fmt.Errorf("invalid major version number")
	}
	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return false, fmt.Errorf("invalid minor version number")
	}

	// Check if version is prior to 2.28
	if major < 2 || (major == 2 && minor < 28) {
		return false, nil
	}
	return true, nil
}

func GetHomeDir() string {
	absolutePath, _ := filepath.Abs(C.QingyuGistHome)
	return filepath.Clean(absolutePath)
}

func SetupSecretKey() {
	if C.SecretKey == "" {
		path := filepath.Join(GetHomeDir(), "qingyugist-secret.key")
		SecretKey, _ = session.GenerateSecretKey(path)
	} else {
		SecretKey = []byte(C.SecretKey)
	}
}

func loadConfigFromYaml(c *config, configPath string, out io.Writer) error {
	if configPath != "" {
		absolutePath, _ := filepath.Abs(configPath)
		absolutePath = filepath.Clean(absolutePath)
		file, err := os.Open(absolutePath)
		if err != nil {
			if !os.IsNotExist(err) {
				return err
			}
			_, _ = fmt.Fprintln(out, "No YAML config file found at "+absolutePath)
		} else {
			_, _ = fmt.Fprintln(out, "Using YAML config file: "+absolutePath)

			// Override default values with values from config.yml
			d := yaml.NewDecoder(file)
			if err = d.Decode(&c); err != nil {
				return err
			}
			defer file.Close()
		}
	} else {
		_, _ = fmt.Fprintln(out, "No YAML config file specified.")
	}

	return nil
}

func loadConfigFromEnv(c *config, out io.Writer) error {
	v := reflect.ValueOf(c).Elem()
	var envVars []string

	for i := 0; i < v.NumField(); i++ {
		tag := v.Type().Field(i).Tag.Get("env")

		if tag == "" {
			continue
		}

		envValue, present := envLookup(strings.ToUpper(tag))
		if !present {
			envValue = ""
		}
		if envValue == "" && v.Field(i).Kind() != reflect.Slice {
			continue
		}

		switch v.Field(i).Kind() {
		case reflect.String:
			v.Field(i).SetString(envValue)
			envVars = append(envVars, tag)
		case reflect.Bool:
			boolVal, err := strconv.ParseBool(envValue)
			if err != nil {
				return err
			}
			v.Field(i).SetBool(boolVal)
			envVars = append(envVars, tag)
		case reflect.Slice:
			if v.Type().Field(i).Type.Elem().Kind() == reflect.Struct {
				prefix := strings.ToUpper(tag) + "_"
				var sliceValue reflect.Value
				elemType := v.Type().Field(i).Type.Elem()

				for index := 0; ; index++ {
					allFieldsPresent := true
					elemValue := reflect.New(elemType).Elem()

					for j := 0; j < elemValue.NumField() && allFieldsPresent; j++ {
						elemField := elemValue.Type().Field(j)
						envName := fmt.Sprintf("%s%d_%s", prefix, index, strings.ToUpper(elemField.Name))
						envValue, present := envLookup(envName)

						if !present {
							allFieldsPresent = false
							break
						}

						envVars = append(envVars, envName)
						elemValue.Field(j).SetString(envValue)
					}

					if !allFieldsPresent {
						break
					}

					if sliceValue.Kind() != reflect.Slice {
						sliceValue = reflect.MakeSlice(v.Type().Field(i).Type, 0, index+1)
					}
					sliceValue = reflect.Append(sliceValue, elemValue)
				}

				if sliceValue.IsValid() {
					v.Field(i).Set(sliceValue)
				}
			}
		default:
			return fmt.Errorf("unsupported type: %s", v.Field(i).Kind())
		}

	}

	if len(envVars) > 0 {
		_, _ = fmt.Fprintln(out, "Using environment variables config: "+strings.Join(envVars, ", "))
	} else {
		_, _ = fmt.Fprintln(out, "No environment variables config specified.")
	}

	return nil
}

// normalizeSshGitMode maps every accepted ssh.git-enabled value to a canonical
// mode. The legacy booleans are preserved for backward compatibility (true →
// builtin, false → disabled); unknown values are returned untouched so checks()
// can reject them.
func normalizeSshGitMode(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "host":
		return SshServerHost
	case "false", "f", "0", "no", "off", "disabled":
		return SshServerDisabled
	case "true", "t", "1", "yes", "on", "builtin", "":
		return SshServerBuiltin
	default:
		return strings.TrimSpace(s)
	}
}

func checks(c *config) error {
	switch c.SshGit {
	case SshServerBuiltin, SshServerHost, SshServerDisabled:
	default:
		return fmt.Errorf("invalid ssh.git-enabled %q (must be %q, %q or %q, or a boolean)", c.SshGit, SshServerBuiltin, SshServerHost, SshServerDisabled)
	}

	if _, err := url.Parse(c.ExternalUrl); err != nil {
		return err
	}

	if _, err := url.Parse(c.GiteaUrl); err != nil {
		return err
	}

	if _, err := url.Parse(c.OIDCDiscoveryUrl); err != nil {
		return err
	}

	if c.LDAPUrl != "" && c.LDAPBindDn != "" && c.LDAPBindCredentials == "" {
		return fmt.Errorf("ldap.bind-dn is set but ldap.bind-credentials is empty; " +
			"set the password, or leave ldap.bind-dn empty to query the LDAP server anonymously")
	}

	return nil
}
