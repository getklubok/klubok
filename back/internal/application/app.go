package application

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"syscall"

	"github.com/microcosm-cc/bluemonday"
	"github.com/ogen-go/ogen/middleware"
	"github.com/ogen-go/ogen/ogenerrors"
	"github.com/wneessen/go-mail"
	"go.uber.org/zap"

	"klubok/internal/application/metrics"
	"klubok/internal/application/service/cookie"
	"klubok/internal/application/service/diagram"
	"klubok/internal/application/service/embed"
	"klubok/internal/application/service/mailer"
	"klubok/internal/application/service/permission"
	"klubok/internal/application/service/session"
	"klubok/internal/application/service/user"
	"klubok/internal/application/tracer"
	"klubok/internal/application/websocket"
	"klubok/internal/db"
	"klubok/internal/db/sqlc"
	"klubok/internal/domain"
	"klubok/internal/provider/sso"
	"klubok/internal/provider/yandex"
	"klubok/internal/restapi/handler"
	"klubok/internal/restapi/oas"
)

func Run() error { //nolint:gocognit,gocyclo,cyclop,maintidx // It's ok
	ctx := context.Background()

	cfg, err := LoadConfiguration()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	env := domain.EnvironmentFromString(cfg.Environment)
	log, err := NewLogger(cfg.Log.Level, env)
	if err != nil {
		return fmt.Errorf("new logger: %w", err)
	}

	dbpool, err := NewDB(ctx, ConfigDB{
		Host:     cfg.Postgres.DSN.Host,
		Port:     cfg.Postgres.DSN.Port,
		User:     cfg.Postgres.DSN.User,
		Password: cfg.Postgres.DSN.Password,
		Database: cfg.Postgres.DSN.Database,

		MaxOpenConns:    cfg.Postgres.PoolConfig.MaxOpenConns,
		MaxIdleConns:    cfg.Postgres.PoolConfig.MaxIdleConns,
		ConnMaxLifetime: cfg.Postgres.PoolConfig.ConnMaxLifetime,
	})
	if err != nil {
		return fmt.Errorf("new db: %w", err)
	}
	defer dbpool.Close()

	queries := sqlc.New(dbpool)
	dbUseCase := db.NewUseCase(dbpool, queries)

	mMetrics := metrics.New(log)
	metricsServer := metrics.NewServer(
		metrics.ServerConfig{
			Metrics: mMetrics,
			DB: metrics.ServerConfigDB{
				Pool: dbpool,
				Name: cfg.Postgres.DSN.Database,
			},
			HTTP: metrics.ServerConfigHTTP{
				Port: cfg.Metrics.Port,

				ReadTimeout:       cfg.Metrics.ReadTimeout,
				ReadHeaderTimeout: cfg.Metrics.ReadHeaderTimeout,
				WriteTimeout:      cfg.Metrics.WriteTimeout,
				IdleTimeout:       cfg.Metrics.IdleTimeout,
			},
		},
		log,
	)
	defer metricsServer.Stop()

	tracerMiddleware, tracerCleanup, err := initTracer(ctx, cfg, log)
	if err != nil {
		return fmt.Errorf("init tracer: %w", err)
	}
	if tracerCleanup != nil {
		defer tracerCleanup()
	}

	cookieService := cookie.NewService(domain.CookieConfig{
		Name:        cfg.HTTP.Cookie.Name,
		ExpiresDays: cfg.HTTP.Cookie.ExpiresDays,
		Domain:      cfg.HTTP.Cookie.Domain,
		Secure:      cfg.HTTP.Cookie.Secure,
	})

	sessionService := session.NewService(queries, log)
	sessionUpdater := session.NewUpdater(sessionService, cfg.HTTP.Cookie.ExpiresDays, log)
	defer sessionUpdater.Stop()

	var emailCh chan mailer.EmailTask
	if cfg.Feature.Auth.Email.Enabled {
		emailClient, err := mail.NewClient(
			cfg.Mailer.Host,
			mail.WithPort(cfg.Mailer.Port),
			mail.WithSSL(),
			mail.WithTLSPortPolicy(mail.TLSMandatory),
			mail.WithSMTPAuth(mail.SMTPAuthPlain),
			mail.WithUsername(cfg.Mailer.Username),
			mail.WithPassword(cfg.Mailer.Password),
		)
		if err != nil {
			return fmt.Errorf("new mail client: %w", err)
		}

		emailService := mailer.New(cfg.Mailer.From, emailClient, map[string]any{
			"sysAppURL": cfg.AppBaseURL,
		})

		emailCh = make(chan mailer.EmailTask, cfg.Mailer.ChannelBufferSize)
		emailWorker := mailer.NewWorker(emailService, emailCh, cfg.Mailer.SendInterval, log)
		emailWorker.Start(ctx)
		defer emailWorker.Stop()
	}

	ipMasker := domain.IPMasker{
		Enabled:  cfg.Feature.IPMasking.Enabled,
		V4Octets: cfg.Feature.IPMasking.V4Octets,
		V6Octets: cfg.Feature.IPMasking.V6Octets,
	}

	userService := user.NewService(queries, dbUseCase, sessionService, emailCh, ipMasker, log)
	permissionService := permission.NewService(queries)
	wsManager := websocket.NewManager(
		cfg.WebSocket.Port,
		cookieService,
		sessionService,
		userService,
		permissionService,
		ipMasker,
		log,
	)
	defer wsManager.Stop()
	go func() {
		if err := wsManager.Start(); err != nil {
			panic(fmt.Errorf("start websocket: %w", err))
		}
	}()

	sanitizer := bluemonday.UGCPolicy()

	indexHandler := handler.NewIndex()
	healthHandler := handler.NewHealth(dbpool)
	featuresHandler := handler.NewFeatures(
		cfg.Feature.Captcha.Enabled,
		cfg.Feature.Captcha.Yandex.SiteKey,
		cfg.Feature.Auth.Email.Enabled,
		cfg.Feature.Auth.SSO.Enabled,
	)

	diagramService := diagram.NewService(queries, dbUseCase, wsManager, log)
	diagramHandler := handler.NewDiagram(diagramService, permissionService, wsManager, sanitizer, log)

	embedService := embed.NewService(queries, log)
	embedHandler := handler.NewEmbed(embedService, permissionService, log)
	embedPublicHandler := handler.NewEmbedPublic(diagramService, embedService, log)

	shapeHandler := handler.NewShape(diagramService, permissionService, sanitizer, log)
	connectionHandler := handler.NewConnection(diagramService, permissionService, log)
	collaboratorHandler := handler.NewCollaborator(diagramService, userService, permissionService, log)
	commentHandler := handler.NewComment(diagramService, permissionService, sanitizer, log)

	captchaService := yandex.NewCaptchaService(cfg.Feature.Captcha.Yandex.ServerKey, log)
	userHandler := handler.NewUser(
		handler.UserConfig{
			FakeAutoAccountsEnabled:     cfg.Feature.FakeAutoAccounts.Enabled,
			AuthEmailEnabled:            cfg.Feature.Auth.Email.Enabled,
			CaptchaEnabled:              cfg.Feature.Captcha.Enabled,
			AllowedEmailDomainsEnabled:  cfg.Feature.AllowedEmailDomains.Enabled,
			AllowedEmailDomainsSuffixes: cfg.Feature.AllowedEmailDomains.AllowedSuffixes,
		},
		userService,
		cookieService,
		sessionService,
		captchaService,
		sessionUpdater,
		log,
	)

	var ssoProvider handler.SSOProvider
	if cfg.Feature.Auth.SSO.Enabled {
		oidcProvider, err := sso.NewOIDCProvider(ctx, sso.Config{
			IssuerURL:    cfg.Feature.Auth.SSO.IssuerURL,
			ClientID:     cfg.Feature.Auth.SSO.ClientID,
			ClientSecret: cfg.Feature.Auth.SSO.ClientSecret,
			RedirectURL:  cfg.Feature.Auth.SSO.RedirectURL,
		})
		if err != nil {
			return fmt.Errorf("init sso provider: %w", err)
		}
		ssoProvider = oidcProvider
	}
	ssoHandler := handler.NewSSO(
		handler.SSOConfig{
			Enabled:              cfg.Feature.Auth.SSO.Enabled,
			Issuer:               cfg.Feature.Auth.SSO.IssuerURL,
			CookieSecure:         cfg.HTTP.Cookie.Secure,
			PostLoginRedirectURL: cfg.AppBaseURL,
		},
		ssoProvider,
		userService,
		cookieService,
		log,
	)

	apiHandler := handler.NewHandler(
		indexHandler,
		healthHandler,
		featuresHandler,
		diagramHandler,
		shapeHandler,
		connectionHandler,
		collaboratorHandler,
		commentHandler,
		userHandler,
		ssoHandler,
		embedHandler,
		embedPublicHandler,
	)

	environment := domain.EnvironmentFromString(cfg.Environment)
	var staticClaims *domain.Claims
	// ВАЖНО: только в случае локального окружения включаем статичный claims
	if environment == domain.EnvironmentLocal && cfg.Feature.StaticClaims.Enabled {
		staticClaims = &domain.Claims{
			SessionID: domain.SessionID(cfg.Feature.StaticClaims.SessionID),
			UserID:    domain.UserID(cfg.Feature.StaticClaims.UserID),
		}
	}

	securityHandler := handler.NewSecurity(sessionService, sessionUpdater, staticClaims, log)

	middlewareManager := NewMiddlewareManager(cfg, metrics.Middleware(mMetrics), tracerMiddleware)
	srv, err := newOASServer(apiHandler, securityHandler, middlewareManager)
	if err != nil {
		return fmt.Errorf("create server: %w", err)
	}

	mux := newMux(srv, muxMiddlewares{
		api:   middlewareManager.SetupAPIMiddlewares,
		embed: middlewareManager.SetupEmbedMiddlewares,
		web:   middlewareManager.SetupWebMiddlewares,
	}, imageStaticDir, imagePrivacyPath)

	addr := fmt.Sprintf(":%d", cfg.HTTP.Port)

	s := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("HTTP server listen", zap.String("addr", addr))
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("listen: %w", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case <-quit:
		log.Info("Shutting down HTTP server...")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.HTTP.WriteTimeout)
	defer shutdownCancel()

	if err := s.Shutdown(shutdownCtx); err != nil {
		log.Error("HTTP server shutdown failed", zap.Error(err))
	}

	log.Info("HTTP server stopped")

	return nil
}

// initTracer initializes the OpenTelemetry tracer if enabled.
// Returns nil middleware and nil cleanup if tracer is disabled.
func initTracer(ctx context.Context, cfg *Configuration, log *zap.Logger) (func(http.Handler) http.Handler, func(), error) {
	if !cfg.Tracer.Enabled {
		return nil, nil, nil
	}

	tp, err := tracer.New(ctx, tracer.Config{
		ServiceName:   "klubok",
		Endpoint:      cfg.Tracer.Endpoint,
		Insecure:      cfg.Tracer.Insecure,
		SamplerRatio:  cfg.Tracer.SamplerRatio,
		ExportTimeout: cfg.Tracer.ExportTimeout,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("init tracer: %w", err)
	}

	cleanup := func() { //nolint:contextcheck // cleanup creates its own timeout context for shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.HTTP.WriteTimeout)
		defer shutdownCancel()

		if err := tp.Shutdown(shutdownCtx); err != nil {
			log.Error("tracer shutdown failed", zap.Error(err))
		}
	}

	ipMasker := domain.IPMasker{
		Enabled:  cfg.Feature.IPMasking.Enabled,
		V4Octets: cfg.Feature.IPMasking.V4Octets,
		V6Octets: cfg.Feature.IPMasking.V6Octets,
	}

	return tracer.Middleware("klubok/http", ipMasker, healthLivePath, healthReadyPath), cleanup, nil
}

// newOASServer creates and configures the ogen HTTP server.
func newOASServer(
	apiHandler *handler.Handler,
	securityHandler *handler.Security,
	middlewareManager *MiddlewareManager,
) (*oas.Server, error) {
	notFoundHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message": "page not found"}`))
	}
	methodNotAllowedHandler := func(w http.ResponseWriter, r *http.Request, allowed string) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Allow", allowed)
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = fmt.Fprintf(w, `{"message": "use one of the following methods: %s"}`, allowed)
	}
	errorHandler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
		code := ogenerrors.ErrorCode(err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_, _ = fmt.Fprintf(w, `{"message": "%v"}`, err)
	}

	srv, err := oas.NewServer(
		apiHandler,
		securityHandler,
		oas.WithNotFound(notFoundHandler),
		oas.WithMethodNotAllowed(methodNotAllowedHandler),
		oas.WithErrorHandler(errorHandler),
		oas.WithMiddleware(httpToOgenMiddleware(middlewareManager.SetupMiddlewares)),
	)
	if err != nil {
		return nil, fmt.Errorf("create server: %w", err)
	}

	return srv, nil
}

// httpToOgenMiddleware adapts a standard net/http middleware (func(http.Handler) http.Handler)
// into an ogen middleware.Middleware.
func httpToOgenMiddleware(httpMiddleware func(http.Handler) http.Handler) middleware.Middleware {
	return func(req middleware.Request, next middleware.Next) (middleware.Response, error) {
		var resp middleware.Response

		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			req.Raw = r
			var err error
			resp, err = next(req)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})

		wrapped := httpMiddleware(inner)
		wrapped.ServeHTTP(httptest.NewRecorder(), req.Raw)

		return resp, nil
	}
}
