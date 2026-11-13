package application

import (
	"net/http"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"

	"klubok/internal/application/middleware/inject"
	"klubok/internal/application/middleware/noindex"
)

const (
	// routerTimeoutMargin оставляет запас, чтобы chi успел отдать 504 до того,
	// как http.Server закроет соединение по WriteTimeout.
	routerTimeoutMargin = time.Second

	// defaultRouterTimeout используется, если WriteTimeout выключен.
	defaultRouterTimeout = 60 * time.Second

	defaultLimitByIPWindowLength = 60 * time.Second
	defaultLimitByIPRequestLimit = 100

	defaultThrottleLimit = 15

	defaultCORSMaxAge = 300

	healthLivePath  = "/health/live"
	healthReadyPath = "/health/ready"
)

type MiddlewareManager struct {
	cfg *Configuration

	metricsMiddleware func(http.Handler) http.Handler
	tracerMiddleware  func(http.Handler) http.Handler
}

func NewMiddlewareManager(
	cfg *Configuration,
	metricsMiddleware func(http.Handler) http.Handler,
	tracerMiddleware func(http.Handler) http.Handler,
) *MiddlewareManager {
	return &MiddlewareManager{
		cfg: cfg,

		metricsMiddleware: metricsMiddleware,
		tracerMiddleware:  tracerMiddleware,
	}
}

// SetupAPIMiddlewares собирает полный стек для API.
func (m *MiddlewareManager) SetupAPIMiddlewares(h http.Handler) http.Handler {
	return m.setupDynamic(h, m.apiCORS(), m.tracerMiddleware)
}

// SetupEmbedMiddlewares собирает стек публичного embed API: та же защита, но
// CORS разрешён всем, ответы закрыты от индексации, трейсинг не подключён —
// embed-токен в имени спана даёт неограниченную кардинальность.
func (m *MiddlewareManager) SetupEmbedMiddlewares(h http.Handler) http.Handler {
	return m.setupDynamic(noindex.Handler(h), cors.AllowAll().Handler, nil)
}

// SetupWebMiddlewares закрывает статику и файлы только от паник: лимиты,
// метрики и трейсинг на файлы не распространяются.
func (*MiddlewareManager) SetupWebMiddlewares(h http.Handler) http.Handler {
	return chimiddleware.Recoverer(h)
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation.
func (*MiddlewareManager) SetupMiddlewares(h http.Handler) http.Handler {
	return h
}

// setupDynamic собирает цепочку динамических запросов.
// Порядок навешивания обратен порядку выполнения: последняя строка — внешняя.
func (m *MiddlewareManager) setupDynamic(
	h http.Handler,
	corsHandler func(http.Handler) http.Handler,
	tracerMiddleware func(http.Handler) http.Handler,
) http.Handler {
	h = chimiddleware.StripSlashes(h)
	h = chimiddleware.Timeout(m.routerTimeout())(h)
	h = chimiddleware.Throttle(defaultThrottleLimit)(h)

	h = httprate.Limit(defaultLimitByIPRequestLimit, defaultLimitByIPWindowLength,
		httprate.WithKeyFuncs(keyByClientIP),
	)(h)

	h = chimiddleware.Recoverer(h) // внутри metrics/tracer: паника попадает в 500 и метрики

	if tracerMiddleware != nil {
		h = tracerMiddleware(h)
	}

	h = m.metricsMiddleware(h)

	if corsHandler != nil {
		h = corsHandler(h)
	}

	h = requestLogger(healthLivePath, healthReadyPath)(h)

	h = inject.SetRequestIDToContext(h) // должен выполняться после RequestID
	h = chimiddleware.RequestID(h)

	h = chimiddleware.ClientIPFromHeader("X-Real-IP")(h) // внешняя: IP нужен остальным

	return h
}

// apiCORS возвращает CORS из конфигурации или nil, если CORS выключен.
func (m *MiddlewareManager) apiCORS() func(http.Handler) http.Handler {
	if !m.cfg.HTTP.CORS.Enabled {
		return nil
	}

	return cors.Handler(cors.Options{
		AllowedOrigins:   m.cfg.HTTP.CORS.AllowOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Auth-Token"},
		AllowCredentials: m.cfg.HTTP.CORS.AllowCredentials,
		MaxAge:           defaultCORSMaxAge, // Maximum value not ignored by any of major browsers
		Debug:            m.cfg.HTTP.CORS.Debug,
	})
}

// requestLogger логирует запросы, кроме путей из skipPaths — чтобы пробы куба
// не засоряли лог.
func requestLogger(skipPaths ...string) func(http.Handler) http.Handler {
	skipped := make(map[string]struct{}, len(skipPaths))
	for _, path := range skipPaths {
		skipped[path] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		logged := chimiddleware.Logger(next)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := skipped[r.URL.Path]; ok {
				next.ServeHTTP(w, r)
				return
			}

			logged.ServeHTTP(w, r)
		})
	}
}

// routerTimeout возвращает таймаут хендлера меньше WriteTimeout сервера.
func (m *MiddlewareManager) routerTimeout() time.Duration {
	writeTimeout := m.cfg.HTTP.WriteTimeout
	if writeTimeout <= 0 {
		return defaultRouterTimeout
	}

	timeout := writeTimeout - routerTimeoutMargin
	if timeout <= 0 {
		// WriteTimeout не больше запаса: делим срок, чтобы 504 успел уйти.
		timeout = writeTimeout / 2 //nolint:mnd // 2 — делитель для вырожденного конфига
	}

	return timeout
}

// keyByClientIP берёт IP, который ClientIPFromHeader положил в контекст.
// Заголовки вроде True-Client-IP и X-Forwarded-For клиент может подделать.
func keyByClientIP(r *http.Request) (string, error) {
	if ip := chimiddleware.GetClientIP(r.Context()); ip != "" {
		return ip, nil
	}

	//nolint:wrapcheck // ошибка KeyByIP не используется, ключ всегда строится
	return httprate.KeyByIP(r)
}
