package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/ardanlabs/conf/v3"
	"github.com/funayman/logger"
	"github.com/inhies/go-bytesize"
	"go.uber.org/zap"

	"github.com/funayman/ebook-uploader/cmd/server/build/all"
	"github.com/funayman/ebook-uploader/upload"
	"github.com/funayman/ebook-uploader/upload/stores/uploadfs"
	"github.com/funayman/ebook-uploader/web"
	"github.com/funayman/ebook-uploader/web/debug"
	"github.com/funayman/ebook-uploader/web/mux"
)

const (
	service = "EBOOK-UPLOADER"
)

var (
	build = "devel"
)

func main() {
	log, err := logger.New(service)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	if err := run(ctx, log); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(ctx context.Context, log *zap.SugaredLogger) error {
	config := struct {
		Web struct {
			ReadTimeout        time.Duration `conf:"default:60s"`
			WriteTimeout       time.Duration `conf:"default:10s"`
			IdleTimeout        time.Duration `conf:"default:120s"`
			ShutdownTimeout    time.Duration `conf:"default:20s"`
			HostPort           string        `conf:"default:0.0.0.0:8000"`
			DebugHostPort      string        `conf:"default:0.0.0.0:4000"`
			CORSAllowedOrigins []string      `conf:"default:*"`
		}
		Upload struct {
			Dir         string `conf:"default:./uploads"`
			MaxFileSize string `conf:"default:50MB"`
		}
		conf.Version
	}{
		Version: conf.Version{
			Build: build,
			Desc:  "Simple File Uploader",
		},
	}

	help, err := conf.Parse("", &config)
	switch {
	case errors.Is(err, conf.ErrHelpWanted):
		fmt.Println(help)
		return nil
	case err != nil:
		return fmt.Errorf("conf.Parse: %w", err)
	}

	log.Infow("starting", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM, web.SIGWEB)

	// -------------------------------------------------------------------------
	// debug service

	go func() {
		log.Infow("startup", "status", "debug v1 router started", "host", config.Web.DebugHostPort)
		log.Infow("statsviz available", "uri", config.Web.DebugHostPort+"/debug/statsviz/")

		if err := http.ListenAndServe(config.Web.DebugHostPort, debug.Mux()); err != nil {
			log.Errorw("shutdown", "status", "debug v1 router closed", "host", config.Web.DebugHostPort, "msg", err)
		}
	}()

	// -------------------------------------------------------------------------
	// storage for uploads

	maxUploadSize, err := bytesize.Parse(config.Upload.MaxFileSize)
	if err != nil {
		return err
	}
	uploadStoreFS, err := uploadfs.NewStore(log, config.Upload.Dir)
	if err != nil {
		return err
	}
	uploadCore := upload.NewCore(log, uploadStoreFS)

	// -------------------------------------------------------------------------
	// main web service

	muxConfig := mux.Config{
		Build:         build,
		ShutdownCh:    shutdownCh,
		Log:           log,
		UploadCore:    uploadCore,
		MaxUploadSize: int64(maxUploadSize),
	}
	webAPIMux := mux.Web(muxConfig, config.Web.CORSAllowedOrigins, all.Routes())

	svr := http.Server{
		Addr:         config.Web.HostPort,
		Handler:      webAPIMux,
		ReadTimeout:  config.Web.ReadTimeout,
		WriteTimeout: config.Web.WriteTimeout,
		IdleTimeout:  config.Web.IdleTimeout,
		ErrorLog:     logger.NewStdLogger(log),
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Infow("startup", "status", "starting api server", "host", svr.Addr)
		serverErrors <- svr.ListenAndServe()
	}()

	// -------------------------------------------------------------------------
	// shutdown

	select {
	case sig := <-shutdownCh:
		log.Infow("shutdown", "shutdown started", "signal", sig.String())
		defer log.Infow("shutdown", "status", "shutdown complete", "signal", sig.String())

		ctx, cancel := context.WithTimeout(ctx, config.Web.ShutdownTimeout)
		defer cancel()

		if err := svr.Shutdown(ctx); err != nil {
			svr.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	case err := <-serverErrors:
		return fmt.Errorf("cannot start server: %w", err)
	}

	return nil
}
