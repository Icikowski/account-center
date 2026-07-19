// Package main provides the account-center command.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"git.sr.ht/~icikowski/account-center/internal/auth"
	"git.sr.ht/~icikowski/account-center/internal/buildinfo"
	"git.sr.ht/~icikowski/account-center/internal/catalog"
	"git.sr.ht/~icikowski/account-center/internal/config"
	"git.sr.ht/~icikowski/account-center/internal/evaluator"
	"git.sr.ht/~icikowski/account-center/internal/knowledgebase"
	"git.sr.ht/~icikowski/account-center/internal/model"
	"git.sr.ht/~icikowski/account-center/internal/shared/xlog"
	"git.sr.ht/~icikowski/account-center/internal/store"
	"git.sr.ht/~icikowski/account-center/internal/web"
)

func main() {
	log := xlog.InitialLogger()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load configuration")
	}

	log = xlog.New(cfg.Log.Level, cfg.Log.Pretty)

	ver := buildinfo.Get()
	log.Info().
		Str(xlog.FieldVersion, ver.Version).
		Str(xlog.FieldCommit, ver.GitReference).
		Time(xlog.FieldBuildTime, ver.BuildTime).
		Str(xlog.FieldGoVersion, ver.GoVersion).
		Msg("initializing application")

	trustedProxies, err := auth.NewTrustedProxies(cfg.Server.TrustedProxyCIDRs)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse trusted proxies")
	}

	catalogProvider, err := catalog.NewWatcher(ctx, cfg.Catalog.Path, cfg.Catalog.ReloadDebounce, log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create catalog watcher")
	}

	var knowledgeBaseProvider model.Reloader[model.KnowledgeBase]
	if cfg.KnowledgeBase.Enabled {
		knowledgeBaseProvider, err = knowledgebase.NewWatcher(
			ctx,
			cfg.KnowledgeBase.Path,
			cfg.KnowledgeBase.ReloadDebounce,
			log,
		)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create knowledge base watcher")
		}
	}

	var storageBackend store.StorageBackend
	if cfg.Redis.Enabled {
		redis.SetLogger(xlog.NewRedisLogger(log))
		redisClient := cfg.Redis.Client()
		if err := redisClient.Ping(ctx).Err(); err != nil {
			log.Fatal().Err(err).Msg("failed to connect to Redis")
		}
		storageBackend = store.NewRedis(redisClient, cfg.Redis.KeyPrefix)
	} else {
		storageBackend = store.NewMemory(ctx)
	}

	authService, err := auth.NewService(
		ctx,
		cfg.OIDC.ProviderURL, cfg.OIDC.ClientID, cfg.OIDC.ClientSecret,
		cfg.Instance.BaseURL,
		cfg.OIDC.RefreshBefore, cfg.Auth.SessionTTL, cfg.Auth.LoginStateTTL,
		storageBackend,
		auth.WithTrustedProxies(trustedProxies),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create auth service")
	}

	evaluator := evaluator.New(log, storageBackend)

	webHandler := web.NewHandler(
		cfg.Instance.Name,
		catalogProvider,
		knowledgeBaseProvider,
		storageBackend,
		authService,
		cfg.Auth.SessionCookieName,
		trustedProxies,
		evaluator,
		log,
	)

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Address, cfg.Server.Port),
		Handler: webHandler,
	}

	go func() {
		sctx := context.WithoutCancel(ctx)

		<-ctx.Done()
		log.Info().Str(xlog.FieldCause, context.Cause(ctx).Error()).Msg("shutting down")

		sctx, cancel := context.WithTimeout(sctx, 5*time.Second)
		defer cancel()

		if err := server.Shutdown(sctx); err != nil {
			log.Error().Err(err).Msg("failed to shutdown server")
		}
	}()

	log.Info().Msg("started application server")
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal().Err(err).Msg("server error")
	}
	log.Info().Msg("stopped application server")
}
