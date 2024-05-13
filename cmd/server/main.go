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

	"github.com/funayman/ebook-uploader/cmd/server/handler"
	"github.com/funayman/ebook-uploader/upload"
	"github.com/funayman/ebook-uploader/upload/stores/uploadfs"
	"github.com/funayman/ebook-uploader/upload/stores/uploadgcs"
	"github.com/funayman/ebook-uploader/upload/stores/uploadmulti"
	"github.com/funayman/ebook-uploader/web"
	"github.com/funayman/ebook-uploader/web/debug"
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
			MaxFileSize        string        `conf:"default:50MB"`
		}
		Upload struct {
			FS struct {
				Dirs []string `conf:"default:./uploads"`
			}
			GCP struct {
				Buckets []string
			}
			S3 struct {
				Buckets []string
			}
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

	stores := []upload.Storer{}

	if len(config.Upload.FS.Dirs) > 0 {
		for _, dir := range config.Upload.FS.Dirs {
			uploadStoreFS, err := uploadfs.NewStore(log, dir)
			if err != nil {
				return err
			}
			stores = append(stores, uploadStoreFS)
		}
	}

	if len(config.Upload.GCP.Buckets) > 0 {
		for _, bucket := range config.Upload.GCP.Buckets {
			uploadStoreGCS, err := uploadgcs.NewStore(log, bucket)
			if err != nil {
				return err
			}
			stores = append(stores, uploadStoreGCS)
		}
	}

	// TODO implement
	// if len(config.Upload.S3.Buckets) > 0 {
	// 	for _, bucket := range config.Upload.S3.Buckets {
	// 		uploadStoreS3, err := uploads3.NewStore(log, bucket)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		stores = append(stores, uploadStoreS3)
	// 	}
	// }

	store, err := uploadmulti.NewStore(log, stores...)
	if err != nil {
		return err
	}

	uploadCore := upload.NewCore(log, store)

	// -------------------------------------------------------------------------
	// main web service

	maxUploadSize, err := bytesize.Parse(config.Web.MaxFileSize)
	if err != nil {
		return err
	}

	mux := handler.Mux(handler.Config{
		ShutdownCh:    shutdownCh,
		CORSOrigins:   config.Web.CORSAllowedOrigins,
		Log:           log,
		UploadCore:    uploadCore,
		MaxUploadSize: int64(maxUploadSize),
	})

	svr := http.Server{
		Addr:         config.Web.HostPort,
		Handler:      mux,
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
