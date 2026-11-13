package application

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

var allEnvList = []string{ //nolint:gochecknoglobals // it's ok
	"ENVIRONMENT",
	"APP_BASE_URL",
	"LOG_LEVEL",
	"HTTP_PORT",
	"HTTP_READ_TIMEOUT",
	"HTTP_READ_HEADER_TIMEOUT",
	"HTTP_WRITE_TIMEOUT",
	"HTTP_IDLE_TIMEOUT",
	"HTTP_CORS_ENABLED",
	"HTTP_CORS_ALLOW_ORIGINS",
	"HTTP_CORS_ALLOW_CREDENTIALS",
	"HTTP_CORS_DEBUG",
	"COOKIE_NAME",
	"COOKIE_EXPIRES_DAYS",
	"COOKIE_SECURE",
	"COOKIE_DOMAIN",
	"WEBSOCKET_PORT",
	"POSTGRES_HOST",
	"POSTGRES_PORT",
	"POSTGRES_USER",
	"POSTGRES_PASSWORD",
	"POSTGRES_DB",
	"POSTGRES_POOL_MAX_OPEN_CONNS",
	"POSTGRES_POOL_MAX_IDLE_CONNS",
	"POSTGRES_POOL_CONN_MAX_LIFETIME",
	"MAILER_HOST",
	"MAILER_PORT",
	"MAILER_USERNAME",
	"MAILER_PASSWORD",
	"MAILER_FROM",
	"MAILER_SEND_INTERVAL",
	"MAILER_CHANNEL_BUFFER_SIZE",
	"METRICS_PORT",
	"METRICS_READ_TIMEOUT",
	"METRICS_READ_HEADER_TIMEOUT",
	"METRICS_WRITE_TIMEOUT",
	"METRICS_IDLE_TIMEOUT",
	"TRACER_ENABLED",
	"TRACER_ENDPOINT",
	"TRACER_INSECURE",
	"TRACER_SAMPLER_RATIO",
	"TRACER_EXPORT_TIMEOUT",
	"FEATURE_FAKE_AUTO_ACCOUNTS_ENABLED",
	"FEATURE_AUTH_EMAIL_ENABLED",
	"FEATURE_AUTH_SSO_ENABLED",
	"FEATURE_AUTH_SSO_ISSUER_URL",
	"FEATURE_AUTH_SSO_CLIENT_ID",
	"FEATURE_AUTH_SSO_CLIENT_SECRET",
	"FEATURE_AUTH_SSO_REDIRECT_URL",
	"FEATURE_CAPTCHA_ENABLED",
	"FEATURE_CAPTCHA_PROVIDER",
	"FEATURE_CAPTCHA_YANDEX_SERVER_KEY",
	"FEATURE_CAPTCHA_YANDEX_SITE_KEY",
	"FEATURE_STATIC_CLAIMS_ENABLED",
	"FEATURE_STATIC_CLAIMS_SESSION_ID",
	"FEATURE_STATIC_CLAIMS_USER_ID",
	"FEATURE_IP_MASKING_ENABLED",
	"FEATURE_IP_MASKING_V4_OCTETS",
	"FEATURE_IP_MASKING_V6_OCTETS",
	"FEATURE_ALLOWED_EMAIL_DOMAINS_ENABLED",
	"FEATURE_ALLOWED_EMAIL_DOMAINS_LIST",
}

func TestLoadConfiguration(t *testing.T) {
	tests := []struct {
		name       string
		setEnvFunc func()
		want       *Configuration
		wantErr    error
	}{
		{
			name: "Should load configuration from env",
			setEnvFunc: func() {
				t.Setenv("ENVIRONMENT", "local")

				t.Setenv("APP_BASE_URL", "https://domain.tld")

				t.Setenv("LOG_LEVEL", "debug")

				t.Setenv("HTTP_PORT", "5172")

				t.Setenv("HTTP_READ_TIMEOUT", "15s")
				t.Setenv("HTTP_READ_HEADER_TIMEOUT", "3s")
				t.Setenv("HTTP_WRITE_TIMEOUT", "20s")
				t.Setenv("HTTP_IDLE_TIMEOUT", "60s")

				t.Setenv("HTTP_CORS_ENABLED", "false")
				t.Setenv("HTTP_CORS_ALLOW_ORIGINS", "http://localhost:5173")
				t.Setenv("HTTP_CORS_ALLOW_CREDENTIALS", "false")
				t.Setenv("HTTP_CORS_DEBUG", "true")

				t.Setenv("WEBSOCKET_PORT", "5171")

				t.Setenv("COOKIE_NAME", "KLUBOK_SID_TEST")
				t.Setenv("COOKIE_EXPIRES_DAYS", "100")
				t.Setenv("COOKIE_SECURE", "false")
				t.Setenv("COOKIE_DOMAIN", "localhost")

				t.Setenv("POSTGRES_HOST", "localhost")
				t.Setenv("POSTGRES_PORT", "50432")
				t.Setenv("POSTGRES_USER", "postgres")
				t.Setenv("POSTGRES_PASSWORD", "password")
				t.Setenv("POSTGRES_DB", "klubok_db")

				t.Setenv("POSTGRES_POOL_MAX_OPEN_CONNS", "10")
				t.Setenv("POSTGRES_POOL_MAX_IDLE_CONNS", "5")
				t.Setenv("POSTGRES_POOL_CONN_MAX_LIFETIME", "1m")

				t.Setenv("MAILER_HOST", "localhost")
				t.Setenv("MAILER_PORT", "25")
				t.Setenv("MAILER_USERNAME", "username")
				t.Setenv("MAILER_PASSWORD", "password")
				t.Setenv("MAILER_FROM", "from@localhost.tld")
				t.Setenv("MAILER_SEND_INTERVAL", "500ms")
				t.Setenv("MAILER_CHANNEL_BUFFER_SIZE", "100")

				t.Setenv("METRICS_PORT", "8080")

				t.Setenv("METRICS_READ_TIMEOUT", "150s")
				t.Setenv("METRICS_READ_HEADER_TIMEOUT", "30s")
				t.Setenv("METRICS_WRITE_TIMEOUT", "200s")
				t.Setenv("METRICS_IDLE_TIMEOUT", "600s")

				t.Setenv("TRACER_ENABLED", "true")
				t.Setenv("TRACER_ENDPOINT", "domain.tld:4318")
				t.Setenv("TRACER_INSECURE", "false")
				t.Setenv("TRACER_SAMPLER_RATIO", "0.01")
				t.Setenv("TRACER_EXPORT_TIMEOUT", "600s")

				t.Setenv("FEATURE_FAKE_AUTO_ACCOUNTS_ENABLED", "true")

				t.Setenv("FEATURE_AUTH_EMAIL_ENABLED", "true")

				t.Setenv("FEATURE_AUTH_SSO_ENABLED", "false")
				t.Setenv("FEATURE_AUTH_SSO_ISSUER_URL", "https://sso.domain.tld")
				t.Setenv("FEATURE_AUTH_SSO_CLIENT_ID", "client-id")
				t.Setenv("FEATURE_AUTH_SSO_CLIENT_SECRET", "client-secret")
				t.Setenv("FEATURE_AUTH_SSO_REDIRECT_URL", "https://domain.tld/auth/sso/callback")

				t.Setenv("FEATURE_CAPTCHA_ENABLED", "true")
				t.Setenv("FEATURE_CAPTCHA_PROVIDER", "yandex")
				t.Setenv("FEATURE_CAPTCHA_YANDEX_SERVER_KEY", "key")
				t.Setenv("FEATURE_CAPTCHA_YANDEX_SITE_KEY", "site-key")

				t.Setenv("FEATURE_STATIC_CLAIMS_ENABLED", "true")
				t.Setenv("FEATURE_STATIC_CLAIMS_SESSION_ID", "session_id")
				t.Setenv("FEATURE_STATIC_CLAIMS_USER_ID", "user_id")

				t.Setenv("FEATURE_IP_MASKING_ENABLED", "true")
				t.Setenv("FEATURE_IP_MASKING_V4_OCTETS", "2")
				t.Setenv("FEATURE_IP_MASKING_V6_OCTETS", "6")

				t.Setenv("FEATURE_ALLOWED_EMAIL_DOMAINS_ENABLED", "true")
				t.Setenv("FEATURE_ALLOWED_EMAIL_DOMAINS_LIST", " example.com , Corp.Example.com , ")
			},
			want: &Configuration{
				Environment: "local",
				AppBaseURL:  "https://domain.tld",
				Log: Log{
					Level: zapcore.DebugLevel,
				},
				HTTP: HTTPServer{
					Port:              5172,
					ReadTimeout:       15 * time.Second,
					ReadHeaderTimeout: 3 * time.Second,
					WriteTimeout:      20 * time.Second,
					IdleTimeout:       60 * time.Second,
					CORS: HTTPCORS{
						Enabled:          false,
						AllowOrigins:     []string{"http://localhost:5173"},
						AllowCredentials: false,
						Debug:            true,
					},
					Cookie: Cookie{
						Name:        "KLUBOK_SID_TEST",
						ExpiresDays: 100,
						Secure:      false,
						Domain:      "localhost",
					},
				},
				WebSocket: WebSocketServer{
					Port: 5171,
				},
				Postgres: Postgres{
					DSN: PostgresDSN{
						Host:     "localhost",
						Port:     50432,
						User:     "postgres",
						Password: "password",
						Database: "klubok_db",
					},
					PoolConfig: PostgresPoolConfig{
						MaxOpenConns:    10,
						MaxIdleConns:    5,
						ConnMaxLifetime: 1 * time.Minute,
					},
				},
				Mailer: Mailer{
					Host:              "localhost",
					Port:              25,
					Username:          "username",
					Password:          "password",
					From:              "from@localhost.tld",
					SendInterval:      500 * time.Millisecond,
					ChannelBufferSize: 100,
				},
				Metrics: Metrics{
					Port: 8080,

					ReadTimeout:       150 * time.Second,
					ReadHeaderTimeout: 30 * time.Second,
					WriteTimeout:      200 * time.Second,
					IdleTimeout:       600 * time.Second,
				},
				Tracer: Tracer{
					Enabled:       true,
					Endpoint:      "domain.tld:4318",
					Insecure:      false,
					SamplerRatio:  0.01,
					ExportTimeout: 600 * time.Second,
				},
				Feature: Feature{
					FakeAutoAccounts: FakeAutoAccounts{
						Enabled: true,
					},
					Auth: Auth{
						Email: AuthEmail{
							Enabled: true,
						},
						SSO: AuthSSO{
							Enabled:      false,
							IssuerURL:    "https://sso.domain.tld",
							ClientID:     "client-id",
							ClientSecret: "client-secret",
							RedirectURL:  "https://domain.tld/auth/sso/callback",
						},
					},
					Captcha: Captcha{
						Enabled:  true,
						Provider: "yandex",
						Yandex: CaptchaYandex{
							ServerKey: "key",
							SiteKey:   "site-key",
						},
					},
					StaticClaims: StaticClaims{
						Enabled:   true,
						SessionID: "session_id",
						UserID:    "user_id",
					},
					IPMasking: IPMasking{
						Enabled:  true,
						V4Octets: 2,
						V6Octets: 6,
					},
					AllowedEmailDomains: AllowedEmailDomains{
						Enabled:         true,
						AllowedSuffixes: []string{"example.com", "corp.example.com"},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "Should load default configuration",
			setEnvFunc: func() {
				t.Setenv("MAILER_HOST", "localhost")
				t.Setenv("MAILER_PORT", "25")
				t.Setenv("MAILER_USERNAME", "username")
				t.Setenv("MAILER_PASSWORD", "password")
				t.Setenv("MAILER_FROM", "from@localhost.tld")
			},
			want: &Configuration{
				Environment: "prod",
				Log: Log{
					Level: zapcore.InfoLevel,
				},
				HTTP: HTTPServer{
					Port:              3000,
					ReadTimeout:       10 * time.Second,
					ReadHeaderTimeout: 5 * time.Second,
					WriteTimeout:      30 * time.Second,
					IdleTimeout:       120 * time.Second,
					CORS: HTTPCORS{
						Enabled:          true,
						AllowCredentials: true,
					},
					Cookie: Cookie{
						Name:        "KLUBOK_SID",
						ExpiresDays: 365,
						Secure:      true,
					},
				},
				WebSocket: WebSocketServer{
					Port: 3030,
				},
				Postgres: Postgres{
					PoolConfig: PostgresPoolConfig{
						MaxOpenConns:    25,
						MaxIdleConns:    2,
						ConnMaxLifetime: 5 * time.Minute,
					},
				},
				Mailer: Mailer{
					Host:              "localhost",
					Port:              25,
					Username:          "username",
					Password:          "password",
					From:              "from@localhost.tld",
					SendInterval:      1 * time.Second,
					ChannelBufferSize: 100,
				},
				Metrics: Metrics{
					Port: 6060,

					ReadTimeout:       10 * time.Second,
					ReadHeaderTimeout: 5 * time.Second,
					WriteTimeout:      30 * time.Second,
					IdleTimeout:       120 * time.Second,
				},
				Tracer: Tracer{
					Endpoint:      "localhost:4318",
					Insecure:      true,
					SamplerRatio:  1,
					ExportTimeout: 10 * time.Second,
				},
				Feature: Feature{
					Auth: Auth{
						Email: AuthEmail{
							Enabled: true,
						},
					},
					Captcha: Captcha{
						Provider: "yandex",
					},
					IPMasking: IPMasking{
						V4Octets: 3,
						V6Octets: 7,
					},
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetConfigEnv(t)
			tt.setEnvFunc()
			got, err := LoadConfiguration()
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAllowedEmailDomains_LoadConfiguration(t *testing.T) {
	tests := []struct {
		name       string
		setEnvFunc func(t *testing.T)
		want       AllowedEmailDomains
		wantErr    error
	}{
		{
			name: "Should load configuration with allowed email domains enabled and non-empty list",
			setEnvFunc: func(t *testing.T) {
				t.Helper()
				t.Setenv("APP_BASE_URL", "https://domain.tld")
				t.Setenv("HTTP_CORS_ALLOW_ORIGINS", "http://localhost:5173")
				t.Setenv("COOKIE_DOMAIN", "localhost")
				t.Setenv("POSTGRES_HOST", "localhost")
				t.Setenv("POSTGRES_PORT", "5432")
				t.Setenv("POSTGRES_USER", "postgres")
				t.Setenv("POSTGRES_PASSWORD", "password")
				t.Setenv("POSTGRES_DB", "klubok_db")
				t.Setenv("MAILER_HOST", "localhost")
				t.Setenv("MAILER_PORT", "25")
				t.Setenv("MAILER_USERNAME", "username")
				t.Setenv("MAILER_PASSWORD", "password")
				t.Setenv("MAILER_FROM", "from@localhost.tld")
				t.Setenv("FEATURE_ALLOWED_EMAIL_DOMAINS_ENABLED", "true")
				t.Setenv("FEATURE_ALLOWED_EMAIL_DOMAINS_LIST", " example.com , Corp.Example.com , ")
			},
			want: AllowedEmailDomains{
				Enabled:         true,
				AllowedSuffixes: []string{"example.com", "corp.example.com"},
			},
			wantErr: nil,
		},
		{
			name: "Should fail to load configuration with allowed email domains enabled and empty list",
			setEnvFunc: func(t *testing.T) {
				t.Helper()
				t.Setenv("APP_BASE_URL", "https://domain.tld")
				t.Setenv("HTTP_CORS_ALLOW_ORIGINS", "http://localhost:5173")
				t.Setenv("COOKIE_DOMAIN", "localhost")
				t.Setenv("POSTGRES_HOST", "localhost")
				t.Setenv("POSTGRES_PORT", "5432")
				t.Setenv("POSTGRES_USER", "postgres")
				t.Setenv("POSTGRES_PASSWORD", "password")
				t.Setenv("POSTGRES_DB", "klubok_db")
				t.Setenv("MAILER_HOST", "localhost")
				t.Setenv("MAILER_PORT", "25")
				t.Setenv("MAILER_USERNAME", "username")
				t.Setenv("MAILER_PASSWORD", "password")
				t.Setenv("MAILER_FROM", "from@localhost.tld")
				t.Setenv("FEATURE_ALLOWED_EMAIL_DOMAINS_ENABLED", "true")
			},
			want:    AllowedEmailDomains{},
			wantErr: errors.New("validate allowed email domains: FEATURE_ALLOWED_EMAIL_DOMAINS_LIST is empty while FEATURE_ALLOWED_EMAIL_DOMAINS_ENABLED=true"),
		},
		{
			name: "Should load configuration with allowed email domains disabled without list",
			setEnvFunc: func(t *testing.T) {
				t.Helper()
				t.Setenv("APP_BASE_URL", "https://domain.tld")
				t.Setenv("HTTP_CORS_ALLOW_ORIGINS", "http://localhost:5173")
				t.Setenv("COOKIE_DOMAIN", "localhost")
				t.Setenv("POSTGRES_HOST", "localhost")
				t.Setenv("POSTGRES_PORT", "5432")
				t.Setenv("POSTGRES_USER", "postgres")
				t.Setenv("POSTGRES_PASSWORD", "password")
				t.Setenv("POSTGRES_DB", "klubok_db")
				t.Setenv("MAILER_HOST", "localhost")
				t.Setenv("MAILER_PORT", "25")
				t.Setenv("MAILER_USERNAME", "username")
				t.Setenv("MAILER_PASSWORD", "password")
				t.Setenv("MAILER_FROM", "from@localhost.tld")
				t.Setenv("FEATURE_ALLOWED_EMAIL_DOMAINS_ENABLED", "false")
			},
			want: AllowedEmailDomains{
				Enabled: false,
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetConfigEnv(t)
			tt.setEnvFunc(t)
			got, err := LoadConfiguration()
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got.Feature.AllowedEmailDomains)
			}
		})
	}
}

func TestAuth_LoadConfiguration(t *testing.T) {
	setBaseEnv := func(t *testing.T) {
		t.Helper()
		t.Setenv("APP_BASE_URL", "https://domain.tld")
		t.Setenv("HTTP_CORS_ALLOW_ORIGINS", "http://localhost:5173")
		t.Setenv("COOKIE_DOMAIN", "localhost")
		t.Setenv("POSTGRES_HOST", "localhost")
		t.Setenv("POSTGRES_PORT", "5432")
		t.Setenv("POSTGRES_USER", "postgres")
		t.Setenv("POSTGRES_PASSWORD", "password")
		t.Setenv("POSTGRES_DB", "klubok_db")
		t.Setenv("MAILER_HOST", "localhost")
		t.Setenv("MAILER_PORT", "25")
		t.Setenv("MAILER_USERNAME", "username")
		t.Setenv("MAILER_PASSWORD", "password")
		t.Setenv("MAILER_FROM", "from@localhost.tld")
	}

	setValidSSO := func(t *testing.T) {
		t.Helper()
		t.Setenv("FEATURE_AUTH_SSO_ENABLED", "true")
		t.Setenv("FEATURE_AUTH_SSO_ISSUER_URL", "https://sso.domain.tld")
		t.Setenv("FEATURE_AUTH_SSO_CLIENT_ID", "client-id")
		t.Setenv("FEATURE_AUTH_SSO_CLIENT_SECRET", "client-secret")
		t.Setenv("FEATURE_AUTH_SSO_REDIRECT_URL", "https://domain.tld/auth/sso/callback")
	}

	tests := []struct {
		name           string
		setEnvFunc     func(t *testing.T)
		want           Auth
		wantErr        error
		wantErrContain string
	}{
		{
			name: "Should load defaults with only email enabled",
			setEnvFunc: func(t *testing.T) {
				t.Helper()
				setBaseEnv(t)
			},
			want: Auth{
				Email: AuthEmail{Enabled: true},
			},
			wantErr: nil,
		},
		{
			name: "Should load with only SSO enabled",
			setEnvFunc: func(t *testing.T) {
				t.Helper()
				setBaseEnv(t)
				t.Setenv("FEATURE_AUTH_EMAIL_ENABLED", "false")
				setValidSSO(t)
			},
			want: Auth{
				SSO: AuthSSO{
					Enabled:      true,
					IssuerURL:    "https://sso.domain.tld",
					ClientID:     "client-id",
					ClientSecret: "client-secret",
					RedirectURL:  "https://domain.tld/auth/sso/callback",
				},
			},
			wantErr: nil,
		},
		{
			name: "Should load with email and SSO enabled",
			setEnvFunc: func(t *testing.T) {
				t.Helper()
				setBaseEnv(t)
				setValidSSO(t)
			},
			want: Auth{
				Email: AuthEmail{Enabled: true},
				SSO: AuthSSO{
					Enabled:      true,
					IssuerURL:    "https://sso.domain.tld",
					ClientID:     "client-id",
					ClientSecret: "client-secret",
					RedirectURL:  "https://domain.tld/auth/sso/callback",
				},
			},
			wantErr: nil,
		},
		{
			name: "Should load with fake accounts and disabled SSO",
			setEnvFunc: func(t *testing.T) {
				t.Helper()
				setBaseEnv(t)
				t.Setenv("FEATURE_FAKE_AUTO_ACCOUNTS_ENABLED", "true")
			},
			want: Auth{
				Email: AuthEmail{Enabled: true},
			},
			wantErr: nil,
		},
		{
			name: "Should fail with SSO enabled without params",
			setEnvFunc: func(t *testing.T) {
				t.Helper()
				setBaseEnv(t)
				t.Setenv("FEATURE_AUTH_SSO_ENABLED", "true")
			},
			wantErr: errors.New("validate auth sso: FEATURE_AUTH_SSO_ISSUER_URL is empty while FEATURE_AUTH_SSO_ENABLED=true"),
		},
		{
			name: "Should fail with SSO enabled and invalid issuer URL",
			setEnvFunc: func(t *testing.T) {
				t.Helper()
				setBaseEnv(t)
				setValidSSO(t)
				t.Setenv("FEATURE_AUTH_SSO_ISSUER_URL", "://bad-url")
			},
			wantErrContain: "validate auth sso: invalid FEATURE_AUTH_SSO_ISSUER_URL:",
		},
		{
			name: "Should fail with SSO enabled and non-http issuer scheme",
			setEnvFunc: func(t *testing.T) {
				t.Helper()
				setBaseEnv(t)
				setValidSSO(t)
				t.Setenv("FEATURE_AUTH_SSO_ISSUER_URL", "ftp://sso.domain.tld")
			},
			wantErr: errors.New(`validate auth sso: invalid FEATURE_AUTH_SSO_ISSUER_URL: scheme must be http or https, got "ftp"`),
		},
		{
			name: "Should fail with SSO enabled without client id",
			setEnvFunc: func(t *testing.T) {
				t.Helper()
				setBaseEnv(t)
				setValidSSO(t)
				t.Setenv("FEATURE_AUTH_SSO_CLIENT_ID", "")
			},
			wantErr: errors.New("validate auth sso: FEATURE_AUTH_SSO_CLIENT_ID is empty while FEATURE_AUTH_SSO_ENABLED=true"),
		},
		{
			name: "Should fail with SSO enabled without client secret",
			setEnvFunc: func(t *testing.T) {
				t.Helper()
				setBaseEnv(t)
				setValidSSO(t)
				t.Setenv("FEATURE_AUTH_SSO_CLIENT_SECRET", "")
			},
			wantErr: errors.New("validate auth sso: FEATURE_AUTH_SSO_CLIENT_SECRET is empty while FEATURE_AUTH_SSO_ENABLED=true"),
		},
		{
			name: "Should fail with SSO enabled without redirect URL",
			setEnvFunc: func(t *testing.T) {
				t.Helper()
				setBaseEnv(t)
				setValidSSO(t)
				t.Setenv("FEATURE_AUTH_SSO_REDIRECT_URL", "")
			},
			wantErr: errors.New("validate auth sso: FEATURE_AUTH_SSO_REDIRECT_URL is empty while FEATURE_AUTH_SSO_ENABLED=true"),
		},
		{
			name: "Should load with only fake accounts enabled",
			setEnvFunc: func(t *testing.T) {
				t.Helper()
				setBaseEnv(t)
				t.Setenv("FEATURE_AUTH_EMAIL_ENABLED", "false")
				t.Setenv("FEATURE_FAKE_AUTO_ACCOUNTS_ENABLED", "true")
			},
			want: Auth{},
		},
		{
			name: "Should fail with fake accounts and SSO enabled",
			setEnvFunc: func(t *testing.T) {
				t.Helper()
				setBaseEnv(t)
				t.Setenv("FEATURE_FAKE_AUTO_ACCOUNTS_ENABLED", "true")
				setValidSSO(t)
			},
			wantErr: errors.New("validate auth: FEATURE_FAKE_AUTO_ACCOUNTS_ENABLED and FEATURE_AUTH_SSO_ENABLED must not be enabled at the same time"),
		},
		{
			name: "Should fail with email, SSO and fake accounts disabled",
			setEnvFunc: func(t *testing.T) {
				t.Helper()
				setBaseEnv(t)
				t.Setenv("FEATURE_AUTH_EMAIL_ENABLED", "false")
			},
			wantErr: errors.New(
				"validate auth: at least one of FEATURE_AUTH_EMAIL_ENABLED, FEATURE_AUTH_SSO_ENABLED, FEATURE_FAKE_AUTO_ACCOUNTS_ENABLED must be enabled",
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetConfigEnv(t)
			tt.setEnvFunc(t)
			got, err := LoadConfiguration()
			if tt.wantErrContain != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContain)
				return
			}
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got.Feature.Auth)
			}
		})
	}
}

func TestMailer_LoadConfiguration(t *testing.T) {
	setBaseEnv := func(t *testing.T) {
		t.Helper()
		t.Setenv("HTTP_CORS_ALLOW_ORIGINS", "http://localhost:5173")
		t.Setenv("COOKIE_DOMAIN", "localhost")
		t.Setenv("POSTGRES_HOST", "localhost")
		t.Setenv("POSTGRES_PORT", "5432")
		t.Setenv("POSTGRES_USER", "postgres")
		t.Setenv("POSTGRES_PASSWORD", "password")
		t.Setenv("POSTGRES_DB", "klubok_db")
	}

	setValidMailerEnv := func(t *testing.T) {
		t.Helper()
		t.Setenv("MAILER_HOST", "localhost")
		t.Setenv("MAILER_PORT", "25")
		t.Setenv("MAILER_USERNAME", "username")
		t.Setenv("MAILER_PASSWORD", "password")
		t.Setenv("MAILER_FROM", "from@localhost.tld")
	}

	tests := []struct {
		name         string
		emailEnabled string
		setEnvFunc   func(t *testing.T)
		wantErr      error
	}{
		{
			name:         "Should require mailer settings when email auth enabled",
			emailEnabled: "true",
			setEnvFunc: func(t *testing.T) {
				t.Helper()
			},
			wantErr: errors.New("validate mailer: MAILER_HOST is empty while FEATURE_AUTH_EMAIL_ENABLED=true"),
		},
		{
			name:         "Should require mailer port when email auth enabled",
			emailEnabled: "true",
			setEnvFunc: func(t *testing.T) {
				t.Helper()
				t.Setenv("MAILER_HOST", "localhost")
			},
			wantErr: errors.New("validate mailer: MAILER_PORT is empty while FEATURE_AUTH_EMAIL_ENABLED=true"),
		},
		{
			name:         "Should load with mailer settings when email auth enabled",
			emailEnabled: "true",
			setEnvFunc:   setValidMailerEnv,
		},
		{
			name:         "Should not require mailer settings when email auth disabled",
			emailEnabled: "false",
			setEnvFunc: func(t *testing.T) {
				t.Helper()
				t.Setenv("FEATURE_FAKE_AUTO_ACCOUNTS_ENABLED", "true")
			},
		},
		{
			name:         "Should ignore partial mailer settings when email auth disabled",
			emailEnabled: "false",
			setEnvFunc: func(t *testing.T) {
				t.Helper()
				t.Setenv("FEATURE_FAKE_AUTO_ACCOUNTS_ENABLED", "true")
				t.Setenv("MAILER_HOST", "localhost")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetConfigEnv(t)
			setBaseEnv(t)
			t.Setenv("FEATURE_AUTH_EMAIL_ENABLED", tt.emailEnabled)
			tt.setEnvFunc(t)

			_, err := LoadConfiguration()
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCaptcha_Validate(t *testing.T) {
	tests := []struct {
		name    string
		srvFunc func(t *testing.T) Captcha
		wantErr error
	}{
		{
			name: "Should validate captcha",
			srvFunc: func(t *testing.T) Captcha {
				t.Helper()

				return Captcha{
					Enabled:  true,
					Provider: "yandex",
					Yandex: CaptchaYandex{
						ServerKey: "key",
						SiteKey:   "site-key",
					},
				}
			},
			wantErr: nil,
		},
		{
			name: "Should validate disabled captcha with empty key",
			srvFunc: func(t *testing.T) Captcha {
				t.Helper()

				return Captcha{
					Enabled:  false,
					Provider: "yandex",
					Yandex: CaptchaYandex{
						ServerKey: "",
					},
				}
			},
			wantErr: nil,
		},
		{
			name: "Should unknown captcha provider",
			srvFunc: func(t *testing.T) Captcha {
				t.Helper()

				return Captcha{
					Enabled:  true,
					Provider: "yohohondex",
				}
			},
			wantErr: errors.New("unknown captcha provider"),
		},
		{
			name: "Should not validate captcha",
			srvFunc: func(t *testing.T) Captcha {
				t.Helper()

				return Captcha{
					Enabled:  true,
					Provider: "yandex",
					Yandex: CaptchaYandex{
						ServerKey: "",
					},
				}
			},
			wantErr: errors.New("yandex server key is empty"),
		},
		{
			name: "Should not validate captcha without site key",
			srvFunc: func(t *testing.T) Captcha {
				t.Helper()

				return Captcha{
					Enabled:  true,
					Provider: "yandex",
					Yandex: CaptchaYandex{
						ServerKey: "key",
					},
				}
			},
			wantErr: errors.New("yandex site key is empty"),
		},
		{
			name: "Should validate disabled captcha with unknown provider",
			srvFunc: func(t *testing.T) Captcha {
				t.Helper()

				return Captcha{
					Enabled:  false,
					Provider: "yohohondex",
				}
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := tt.srvFunc(t)
			err := srv.validate()
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestStaticClaims_Validate(t *testing.T) {
	tests := []struct {
		name    string
		srvFunc func(t *testing.T) StaticClaims
		wantErr error
	}{
		{
			name: "Should validate static claims - enabled",
			srvFunc: func(t *testing.T) StaticClaims {
				t.Helper()

				return StaticClaims{
					Enabled:   true,
					SessionID: "01JHHXTVAMTYVE7Q3H6VRP7206",
					UserID:    "01KPDB4VR5B34RB62YXDVMYK4S",
				}
			},
			wantErr: nil,
		},
		{
			name: "Should validate static claims - disabled",
			srvFunc: func(t *testing.T) StaticClaims {
				t.Helper()

				return StaticClaims{}
			},
			wantErr: nil,
		},
		{
			name: "Should not validate static claims - session id is empty",
			srvFunc: func(t *testing.T) StaticClaims {
				t.Helper()

				return StaticClaims{
					Enabled:   true,
					SessionID: "",
					UserID:    "01KPDB4VR5B34RB62YXDVMYK4S",
				}
			},
			wantErr: errors.New("session id is empty"),
		},
		{
			name: "Should not validate static claims - user id is empty",
			srvFunc: func(t *testing.T) StaticClaims {
				t.Helper()

				return StaticClaims{
					Enabled:   true,
					SessionID: "01JHHXTVAMTYVE7Q3H6VRP7206",
					UserID:    "",
				}
			},
			wantErr: errors.New("user id is empty"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := tt.srvFunc(t)
			err := srv.validate()
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestAllowedEmailDomains_Validate(t *testing.T) {
	tests := []struct {
		name    string
		srvFunc func(t *testing.T) AllowedEmailDomains
		wantErr error
	}{
		{
			name: "Should validate allowed email domains - enabled with non-empty list",
			srvFunc: func(t *testing.T) AllowedEmailDomains {
				t.Helper()

				return AllowedEmailDomains{
					Enabled:         true,
					AllowedSuffixes: []string{"example.com"},
				}
			},
			wantErr: nil,
		},
		{
			name: "Should not validate allowed email domains - enabled with empty list",
			srvFunc: func(t *testing.T) AllowedEmailDomains {
				t.Helper()

				return AllowedEmailDomains{
					Enabled: true,
				}
			},
			wantErr: errors.New("FEATURE_ALLOWED_EMAIL_DOMAINS_LIST is empty while FEATURE_ALLOWED_EMAIL_DOMAINS_ENABLED=true"),
		},
		{
			name: "Should validate allowed email domains - disabled without list",
			srvFunc: func(t *testing.T) AllowedEmailDomains {
				t.Helper()

				return AllowedEmailDomains{}
			},
			wantErr: nil,
		},
		{
			name: "Should validate allowed email domains - disabled with non-empty list",
			srvFunc: func(t *testing.T) AllowedEmailDomains {
				t.Helper()

				return AllowedEmailDomains{
					Enabled:         false,
					AllowedSuffixes: []string{"example.com"},
				}
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := tt.srvFunc(t)
			err := srv.validate()
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestIPMasking_Validate(t *testing.T) {
	tests := []struct {
		name    string
		srvFunc func(t *testing.T) IPMasking
		wantErr error
	}{
		{
			name: "Should validate ip masking - enabled with defaults",
			srvFunc: func(t *testing.T) IPMasking {
				t.Helper()

				return IPMasking{
					Enabled:  true,
					V4Octets: 3,
					V6Octets: 7,
				}
			},
			wantErr: nil,
		},
		{
			name: "Should validate ip masking - enabled with custom octets",
			srvFunc: func(t *testing.T) IPMasking {
				t.Helper()

				return IPMasking{
					Enabled:  true,
					V4Octets: 1,
					V6Octets: 8,
				}
			},
			wantErr: nil,
		},
		{
			name: "Should validate ip masking - disabled with invalid values",
			srvFunc: func(t *testing.T) IPMasking {
				t.Helper()

				return IPMasking{
					Enabled:  false,
					V4Octets: 0,
					V6Octets: 9,
				}
			},
			wantErr: nil,
		},
		{
			name: "Should not validate ip masking - v4 octets below range",
			srvFunc: func(t *testing.T) IPMasking {
				t.Helper()

				return IPMasking{
					Enabled:  true,
					V4Octets: 0,
					V6Octets: 7,
				}
			},
			wantErr: errors.New("v4 octets must be in range 1..4, got 0"),
		},
		{
			name: "Should not validate ip masking - v4 octets above range",
			srvFunc: func(t *testing.T) IPMasking {
				t.Helper()

				return IPMasking{
					Enabled:  true,
					V4Octets: 5,
					V6Octets: 7,
				}
			},
			wantErr: errors.New("v4 octets must be in range 1..4, got 5"),
		},
		{
			name: "Should not validate ip masking - v6 octets below range",
			srvFunc: func(t *testing.T) IPMasking {
				t.Helper()

				return IPMasking{
					Enabled:  true,
					V4Octets: 3,
					V6Octets: 0,
				}
			},
			wantErr: errors.New("v6 octets must be in range 1..8, got 0"),
		},
		{
			name: "Should not validate ip masking - v6 octets above range",
			srvFunc: func(t *testing.T) IPMasking {
				t.Helper()

				return IPMasking{
					Enabled:  true,
					V4Octets: 3,
					V6Octets: 9,
				}
			},
			wantErr: errors.New("v6 octets must be in range 1..8, got 9"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := tt.srvFunc(t)
			err := srv.validate()
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestLoadConfiguration_SSOLoginRedirectURL(t *testing.T) {
	setBaseEnv := func(t *testing.T) {
		t.Helper()
		t.Setenv("HTTP_CORS_ALLOW_ORIGINS", "http://localhost:5173")
		t.Setenv("COOKIE_DOMAIN", "localhost")
		t.Setenv("POSTGRES_HOST", "localhost")
		t.Setenv("POSTGRES_PORT", "5432")
		t.Setenv("POSTGRES_USER", "postgres")
		t.Setenv("POSTGRES_PASSWORD", "password")
		t.Setenv("POSTGRES_DB", "klubok_db")
		t.Setenv("MAILER_HOST", "localhost")
		t.Setenv("MAILER_PORT", "25")
		t.Setenv("MAILER_USERNAME", "username")
		t.Setenv("MAILER_PASSWORD", "password")
		t.Setenv("MAILER_FROM", "from@localhost.tld")
	}

	setValidSSO := func(t *testing.T) {
		t.Helper()
		t.Setenv("FEATURE_AUTH_SSO_ENABLED", "true")
		t.Setenv("FEATURE_AUTH_SSO_ISSUER_URL", "https://sso.domain.tld")
		t.Setenv("FEATURE_AUTH_SSO_CLIENT_ID", "client-id")
		t.Setenv("FEATURE_AUTH_SSO_CLIENT_SECRET", "client-secret")
		t.Setenv("FEATURE_AUTH_SSO_REDIRECT_URL", "https://domain.tld/auth/sso/callback")
	}

	tests := []struct {
		name        string
		appBaseURL  string
		enableSSO   bool
		wantURL     string
		wantErrPart string
	}{
		{
			name:       "SSO enabled with valid front URL",
			appBaseURL: "http://localhost:5173",
			enableSSO:  true,
			wantURL:    "http://localhost:5173",
		},
		{
			name:       "SSO enabled strips trailing slash",
			appBaseURL: "http://localhost:5173/",
			enableSSO:  true,
			wantURL:    "http://localhost:5173",
		},
		{
			name:        "SSO enabled rejects non-http scheme",
			appBaseURL:  "ftp://localhost:5173",
			enableSSO:   true,
			wantErrPart: `invalid APP_BASE_URL: scheme must be http or https, got "ftp"`,
		},
		{
			name:        "SSO enabled rejects unparsable URL",
			appBaseURL:  "://bad-url",
			enableSSO:   true,
			wantErrPart: "invalid APP_BASE_URL:",
		},
		{
			name:       "SSO disabled does not validate front URL",
			appBaseURL: "ftp://localhost:5173",
			enableSSO:  false,
			wantURL:    "ftp://localhost:5173",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetConfigEnv(t)
			setBaseEnv(t)
			t.Setenv("APP_BASE_URL", tt.appBaseURL)
			if tt.enableSSO {
				setValidSSO(t)
			}
			got, err := LoadConfiguration()
			if tt.wantErrPart != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrPart)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantURL, got.AppBaseURL)
		})
	}
}

// resetConfigEnv сбрасывает переменные окружения конфига, которые могут быть унаследованы
// из внешней среды (например, директивы из .envrc), чтобы тесты LoadConfiguration не зависели
// от состояния окружения. Пустая строка для caarlos0/env v11 означает «переменная не задана»,
// поэтому применяются значения envDefault из config.go.
func resetConfigEnv(t *testing.T) {
	t.Helper()

	for _, name := range allEnvList {
		t.Setenv(name, "")
	}
}
