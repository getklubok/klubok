package application

import (
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"go.uber.org/zap/zapcore"
)

// Configuration представляет конфигурацию приложения.
type Configuration struct {
	// Environment - среда выполнения.
	// Возможные значения: "local", "stage", "prod".
	Environment string `env:"ENVIRONMENT" envDefault:"prod"`

	// BaseURL - базовый URL приложения.
	// Пример: "https://domain.tld".
	AppBaseURL string `env:"APP_BASE_URL,required"`

	Log       Log
	HTTP      HTTPServer
	WebSocket WebSocketServer
	Postgres  Postgres
	Mailer    Mailer
	Metrics   Metrics
	Tracer    Tracer
	Feature   Feature
}

// Log конфигурация логирования.
type Log struct {
	// Уровень логирования.
	// Возможные значения: "debug", "info", "warn", "error".
	Level zapcore.Level `env:"LOG_LEVEL" envDefault:"info"`
}

// HTTPServer конфигурация HTTP сервера.
type HTTPServer struct {
	Port int `env:"HTTP_PORT" envDefault:"3000"`

	ReadTimeout       time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"10s"`
	ReadHeaderTimeout time.Duration `env:"HTTP_READ_HEADER_TIMEOUT" envDefault:"5s"`
	WriteTimeout      time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"30s"`
	IdleTimeout       time.Duration `env:"HTTP_IDLE_TIMEOUT" envDefault:"120s"`

	CORS   HTTPCORS
	Cookie Cookie
}

// HTTPCORS конфигурация CORS для HTTP сервера.
type HTTPCORS struct {
	Enabled bool `env:"HTTP_CORS_ENABLED" envDefault:"true"`

	// AllowOrigins - разрешенные домены.
	// Пример: "http://localhost:5173", "https://domain.tld"
	AllowOrigins []string `env:"HTTP_CORS_ALLOW_ORIGINS,required"`

	// AllowCredentials - разрешены куки.
	AllowCredentials bool `env:"HTTP_CORS_ALLOW_CREDENTIALS" envDefault:"true"`

	// Debug - режим отладки.
	Debug bool `env:"HTTP_CORS_DEBUG" envDefault:"false"`
}

// Cookie конфигурация cookie для HTTP сервера.
type Cookie struct {
	// Name - имя куки.
	Name string `env:"COOKIE_NAME" envDefault:"KLUBOK_SID"`

	// ExpiresDays - количество дней до истечения срока действия куки.
	ExpiresDays int `env:"COOKIE_EXPIRES_DAYS" envDefault:"365"`

	// Secure - куки доступны только через HTTPS.
	Secure bool `env:"COOKIE_SECURE" envDefault:"true"`

	// Domain - домен, на котором куки доступны.
	// Пример: "localhost", "domain.tld"
	Domain string `env:"COOKIE_DOMAIN,required"`
}

// WebSocketServer конфигурация WebSocket сервера.
type WebSocketServer struct {
	Port int `env:"WEBSOCKET_PORT" envDefault:"3030"`
}

// Postgres конфигурация PostgreSQL.
type Postgres struct {
	DSN        PostgresDSN
	PoolConfig PostgresPoolConfig
}

// PostgresDSN параметры соединения PostgreSQL.
type PostgresDSN struct {
	Host     string `env:"POSTGRES_HOST,required"`
	Port     int    `env:"POSTGRES_PORT,required"`
	User     string `env:"POSTGRES_USER,required"`
	Password string `env:"POSTGRES_PASSWORD,required"`
	Database string `env:"POSTGRES_DB,required"`
}

// PostgresPoolConfig параметры пула соединений PostgreSQL.
type PostgresPoolConfig struct {
	// Должно быть меньше лимита в конфигурации БД
	MaxOpenConns int32 `env:"POSTGRES_POOL_MAX_OPEN_CONNS" envDefault:"25"`

	// Всегда меньше или равно MaxOpenConns
	MaxIdleConns int32 `env:"POSTGRES_POOL_MAX_IDLE_CONNS" envDefault:"2"`

	// Не делайте время слишком коротким
	ConnMaxLifetime time.Duration `env:"POSTGRES_POOL_CONN_MAX_LIFETIME" envDefault:"5m"`
}

// Mailer конфигурация почтовика.
// Поддерживается только TLS.
// Обязательные поля проверяются только при FEATURE_AUTH_EMAIL_ENABLED=true.
type Mailer struct {
	// Host - хост почтового сервера.
	// Пример: "smtp.timeweb.ru"
	Host string `env:"MAILER_HOST"`

	// Port - порт почтового сервера.
	// Пример: 465
	Port int `env:"MAILER_PORT"`

	// Username - имя пользователя почтового сервера.
	// Пример: "no-reply@domain.tld"
	Username string `env:"MAILER_USERNAME"`

	// Password - пароль пользователя почтового сервера.
	Password string `env:"MAILER_PASSWORD"`

	// From - адрес отправителя.
	// Пример: "Klubok <no-reply@domain.tld>"
	From string `env:"MAILER_FROM"`

	// SendInterval - интервал между отправками писем воркером.
	SendInterval time.Duration `env:"MAILER_SEND_INTERVAL" envDefault:"1s"`

	// ChannelBufferSize - размер буфера канала для асинхронной отправки писем.
	ChannelBufferSize int `env:"MAILER_CHANNEL_BUFFER_SIZE" envDefault:"100"`
}

func (m Mailer) validate() error {
	if m.Host == "" {
		return errors.New("MAILER_HOST is empty while FEATURE_AUTH_EMAIL_ENABLED=true")
	}
	if m.Port == 0 {
		return errors.New("MAILER_PORT is empty while FEATURE_AUTH_EMAIL_ENABLED=true")
	}
	if m.Username == "" {
		return errors.New("MAILER_USERNAME is empty while FEATURE_AUTH_EMAIL_ENABLED=true")
	}
	if m.Password == "" {
		return errors.New("MAILER_PASSWORD is empty while FEATURE_AUTH_EMAIL_ENABLED=true")
	}
	if m.From == "" {
		return errors.New("MAILER_FROM is empty while FEATURE_AUTH_EMAIL_ENABLED=true")
	}

	return nil
}

// Metrics конфигурация метрик.
type Metrics struct {
	Port int `env:"METRICS_PORT" envDefault:"6060"`

	ReadTimeout       time.Duration `env:"METRICS_READ_TIMEOUT" envDefault:"10s"`
	ReadHeaderTimeout time.Duration `env:"METRICS_READ_HEADER_TIMEOUT" envDefault:"5s"`
	WriteTimeout      time.Duration `env:"METRICS_WRITE_TIMEOUT" envDefault:"30s"`
	IdleTimeout       time.Duration `env:"METRICS_IDLE_TIMEOUT" envDefault:"120s"`
}

// Tracer конфигурация OpenTelemetry трейсера.
type Tracer struct {
	// Enabled - включить трейсинг.
	Enabled bool `env:"TRACER_ENABLED" envDefault:"false"`

	// Endpoint - адрес OTLP HTTP коллектора.
	// Пример: "localhost:4318"
	Endpoint string `env:"TRACER_ENDPOINT" envDefault:"localhost:4318"`

	// Insecure - использовать HTTP вместо HTTPS.
	Insecure bool `env:"TRACER_INSECURE" envDefault:"true"`

	// SamplerRatio - доля запросов для трейсинга (0.0-1.0).
	// 1.0 = все запросы, 0.01 = 1%.
	SamplerRatio float64 `env:"TRACER_SAMPLER_RATIO" envDefault:"1"`

	// ExportTimeout - таймаут экспорта.
	ExportTimeout time.Duration `env:"TRACER_EXPORT_TIMEOUT" envDefault:"10s"`
}

type Feature struct {
	// Автоматические обезличенные аккаунты
	FakeAutoAccounts FakeAutoAccounts

	// Способы входа
	Auth Auth

	// Капча
	Captcha Captcha

	// Статичный claims
	// Срабатывает только при environment=local
	// Необходим для отладки визуальной части с помощью ИИ, чтобы он смог сам увидеть что происходит
	StaticClaims StaticClaims

	// Маскировка IP-адресов при записи в сохраняемые места
	// (privacy_log, логи WebSocket, атрибуты трейсера).
	IPMasking IPMasking

	// Ограничение регистрации по суффиксам доменов почты.
	AllowedEmailDomains AllowedEmailDomains
}

type FakeAutoAccounts struct {
	// По умолчанию: false.
	Enabled bool `env:"FEATURE_FAKE_AUTO_ACCOUNTS_ENABLED" envDefault:"false"`
}

// Auth способы входа (хотя бы один MUST быть включён).
type Auth struct {
	Email AuthEmail
	SSO   AuthSSO
}

// AuthEmail вход и регистрация по email-коду.
type AuthEmail struct {
	// По умолчанию: true (перенос существующего поведения).
	Enabled bool `env:"FEATURE_AUTH_EMAIL_ENABLED" envDefault:"true"`
}

// AuthSSO вход через SSO по протоколу OIDC.
type AuthSSO struct {
	// По умолчанию: false.
	Enabled bool `env:"FEATURE_AUTH_SSO_ENABLED" envDefault:"false"`

	// IssuerURL - issuer URL OIDC-провайдера для discovery.
	// Пример: "https://accounts.google.com"
	IssuerURL string `env:"FEATURE_AUTH_SSO_ISSUER_URL" envDefault:""`

	// ClientID - идентификатор клиента у провайдера.
	// Пример: "799...734.apps.googleusercontent.com"
	ClientID string `env:"FEATURE_AUTH_SSO_CLIENT_ID" envDefault:""`

	// ClientSecret - секрет клиента у провайдера.
	// Пример: "GOCSPX-..."
	ClientSecret string `env:"FEATURE_AUTH_SSO_CLIENT_SECRET" envDefault:""`

	// RedirectURL - URL callback для возврата от провайдера.
	// Пример: "http://localhost:8080/auth/sso/callback"
	RedirectURL string `env:"FEATURE_AUTH_SSO_REDIRECT_URL" envDefault:""`
}

func (s AuthSSO) validate() error {
	if !s.Enabled {
		return nil
	}

	if s.IssuerURL == "" {
		return errors.New("FEATURE_AUTH_SSO_ISSUER_URL is empty while FEATURE_AUTH_SSO_ENABLED=true")
	}
	if err := validateHTTPURL(s.IssuerURL); err != nil {
		return fmt.Errorf("invalid FEATURE_AUTH_SSO_ISSUER_URL: %w", err)
	}
	if s.ClientID == "" {
		return errors.New("FEATURE_AUTH_SSO_CLIENT_ID is empty while FEATURE_AUTH_SSO_ENABLED=true")
	}
	if s.ClientSecret == "" {
		return errors.New("FEATURE_AUTH_SSO_CLIENT_SECRET is empty while FEATURE_AUTH_SSO_ENABLED=true")
	}
	if s.RedirectURL == "" {
		return errors.New("FEATURE_AUTH_SSO_REDIRECT_URL is empty while FEATURE_AUTH_SSO_ENABLED=true")
	}
	if err := validateHTTPURL(s.RedirectURL); err != nil {
		return fmt.Errorf("invalid FEATURE_AUTH_SSO_REDIRECT_URL: %w", err)
	}

	return nil
}

func validateHTTPURL(raw string) error {
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return fmt.Errorf("parse URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("scheme must be http or https, got %q", u.Scheme)
	}
	if u.Host == "" {
		return errors.New("host is empty")
	}

	return nil
}

// Captcha конфигурация капчи.
type Captcha struct {
	// Enabled - включить проверку капчи.
	// По умолчанию: false.
	Enabled bool `env:"FEATURE_CAPTCHA_ENABLED" envDefault:"false"`

	// Provider - провайдер капчи.
	// Возможные значения: "yandex".
	Provider string `env:"FEATURE_CAPTCHA_PROVIDER" envDefault:"yandex"`

	// Yandex - конфигурация капчи Yandex.
	Yandex CaptchaYandex
}

// CaptchaYandex конфигурация капчи Yandex.
type CaptchaYandex struct {
	// ServerKey - ключ сервера.
	ServerKey string `env:"FEATURE_CAPTCHA_YANDEX_SERVER_KEY" envDefault:""`

	// SiteKey - публичный ключ сайта, отдаётся фронтенду через API.
	SiteKey string `env:"FEATURE_CAPTCHA_YANDEX_SITE_KEY" envDefault:""`
}

func (c Captcha) validate() error {
	if !c.Enabled {
		return nil
	}

	providers := []string{"yandex"}
	if !slices.Contains(providers, c.Provider) {
		return errors.New("unknown captcha provider")
	}

	if c.Provider == "yandex" && c.Yandex.ServerKey == "" {
		return errors.New("yandex server key is empty")
	}

	if c.Provider == "yandex" && c.Yandex.SiteKey == "" {
		return errors.New("yandex site key is empty")
	}

	return nil
}

type StaticClaims struct {
	// По умолчанию: false.
	Enabled bool `env:"FEATURE_STATIC_CLAIMS_ENABLED" envDefault:"false"`

	SessionID string `env:"FEATURE_STATIC_CLAIMS_SESSION_ID" envDefault:""`
	UserID    string `env:"FEATURE_STATIC_CLAIMS_USER_ID" envDefault:""`
}

func (f StaticClaims) validate() error {
	if f.Enabled {
		if f.SessionID == "" {
			return errors.New("session id is empty")
		}
		if f.UserID == "" {
			return errors.New("user id is empty")
		}
	}

	return nil
}

type IPMasking struct {
	// По умолчанию: false.
	Enabled bool `env:"FEATURE_IP_MASKING_ENABLED" envDefault:"false"`

	// Количество сохраняемых октетов IPv4 (1..4).
	V4Octets int `env:"FEATURE_IP_MASKING_V4_OCTETS" envDefault:"3"`

	// Количество сохраняемых групп IPv6 (1..8).
	V6Octets int `env:"FEATURE_IP_MASKING_V6_OCTETS" envDefault:"7"`
}

func (f IPMasking) validate() error {
	if f.Enabled {
		if f.V4Octets < 1 || f.V4Octets > 4 {
			return fmt.Errorf("v4 octets must be in range 1..4, got %d", f.V4Octets)
		}
		if f.V6Octets < 1 || f.V6Octets > 8 {
			return fmt.Errorf("v6 octets must be in range 1..8, got %d", f.V6Octets)
		}
	}

	return nil
}

// AllowedEmailDomains конфигурация ограничения регистрации по доменам почты.
type AllowedEmailDomains struct {
	// Enabled - включить ограничение регистрации по суффиксам доменов.
	// По умолчанию: false.
	Enabled bool `env:"FEATURE_ALLOWED_EMAIL_DOMAINS_ENABLED" envDefault:"false"`

	// AllowedSuffixes - разрешённые суффиксы доменов (через запятую).
	// Пример: "example.com,corp.example.com".
	AllowedSuffixes []string `env:"FEATURE_ALLOWED_EMAIL_DOMAINS_LIST"`
}

func (d AllowedEmailDomains) validate() error {
	if !d.Enabled {
		return nil
	}

	if len(d.AllowedSuffixes) == 0 {
		return errors.New("FEATURE_ALLOWED_EMAIL_DOMAINS_LIST is empty while FEATURE_ALLOWED_EMAIL_DOMAINS_ENABLED=true")
	}

	return nil
}

// normalize суффиксы доменов при загрузке: нижний регистр, обрезка пробелов, отбрасывание пустых элементов.
func (d AllowedEmailDomains) normalize() []string {
	if len(d.AllowedSuffixes) == 0 {
		return d.AllowedSuffixes
	}

	normalized := make([]string, 0, len(d.AllowedSuffixes))
	for _, suffix := range d.AllowedSuffixes {
		suffix = strings.ToLower(strings.TrimSpace(suffix))
		if suffix == "" {
			continue
		}
		normalized = append(normalized, suffix)
	}

	return normalized
}

// LoadConfiguration возвращает новую конфигурацию приложения на основе переменных среды.
func LoadConfiguration() (*Configuration, error) {
	var config Configuration
	if err := env.Parse(&config); err != nil {
		return nil, fmt.Errorf("parse configuration: %w", err)
	}

	config.AppBaseURL = strings.TrimSuffix(config.AppBaseURL, "/")

	if err := config.Feature.Captcha.validate(); err != nil {
		return nil, fmt.Errorf("validate captcha: %w", err)
	}

	if err := config.Feature.StaticClaims.validate(); err != nil {
		return nil, fmt.Errorf("validate static claims: %w", err)
	}

	if err := config.Feature.IPMasking.validate(); err != nil {
		return nil, fmt.Errorf("validate ip masking: %w", err)
	}

	config.Feature.AllowedEmailDomains.AllowedSuffixes = config.Feature.AllowedEmailDomains.normalize()
	if err := config.Feature.AllowedEmailDomains.validate(); err != nil {
		return nil, fmt.Errorf("validate allowed email domains: %w", err)
	}

	if err := config.Feature.Auth.SSO.validate(); err != nil {
		return nil, fmt.Errorf("validate auth sso: %w", err)
	}

	if config.Feature.Auth.SSO.Enabled {
		if err := validateHTTPURL(config.AppBaseURL); err != nil {
			return nil, fmt.Errorf("invalid APP_BASE_URL: %w", err)
		}
	}

	if config.Feature.Auth.Email.Enabled {
		if err := config.Mailer.validate(); err != nil {
			return nil, fmt.Errorf("validate mailer: %w", err)
		}
	}

	if err := config.Feature.validate(); err != nil {
		return nil, fmt.Errorf("validate auth: %w", err)
	}

	return &config, nil
}

func (f Feature) validate() error {
	if f.FakeAutoAccounts.Enabled && f.Auth.SSO.Enabled {
		return errors.New("FEATURE_FAKE_AUTO_ACCOUNTS_ENABLED and FEATURE_AUTH_SSO_ENABLED must not be enabled at the same time")
	}

	if !f.Auth.Email.Enabled && !f.Auth.SSO.Enabled && !f.FakeAutoAccounts.Enabled {
		return errors.New(
			"at least one of FEATURE_AUTH_EMAIL_ENABLED, FEATURE_AUTH_SSO_ENABLED, FEATURE_FAKE_AUTO_ACCOUNTS_ENABLED must be enabled",
		)
	}

	return nil
}
