package main

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"gophKeeper/internal/config"
	"gophKeeper/internal/cookie"
	"gophKeeper/internal/gophkeeper"
	"gophKeeper/internal/storage"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var gophKeeperInstance *gophkeeper.GophKeeper

func main() {
	configuration := config.Configuration{}
	cfg, err := configuration.InitConfig()

	if err != nil {
		log.Fatalf("Error while initializing configuration: %v", err)
	}

	var dataUsersStorage storage.UserStorageInterface
	dataUsersStorage = &storage.UsersStorage{}
	dataUsersStorage.Init(cfg.DatabaseDsn)
	migrations(cfg)

	cookieManager := cookie.CookieManager{
		Storage: dataUsersStorage,
	}

	server := &http.Server{}

	idleConnsClosed := make(chan struct{})
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sigint
		// получили сигнал os.Interrupt, запускаем процедуру graceful shutdown

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			// ошибки закрытия Listener
			log.Printf("HTTP server Shutdown: %v", err)
		}

		close(idleConnsClosed)
	}()

	gophKeeperInstance = &gophkeeper.GophKeeper{
		Storage:       dataUsersStorage,
		ServerAddress: cfg.ServerAddress,
		BaseURL:       cfg.BaseURL,
		CookieManager: &cookieManager,
	}

	initHTTPServer(&cookieManager, cfg, server)

	<-idleConnsClosed

	fmt.Println("Server Shutdown gracefully")
}

func migrations(cfg *config.ConfigData) {
	dsn := cfg.DatabaseDsn
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	err = db.AutoMigrate(storage.Users{})

	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	err = db.AutoMigrate(storage.UserCookie{})

	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	err = db.AutoMigrate(storage.UserData{})

	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
}

func initHTTPServer(
	cookieManager *cookie.CookieManager,
	cfg *config.ConfigData,
	server *http.Server,
) {
	r := chi.NewRouter()
	r.With(cookieManager.CookieHandler).Route("/", func(r chi.Router) {
		r.Post("/add_data", gophKeeperInstance.AddDataHandler)
		r.Get("/get_data", gophKeeperInstance.GetDataHandler)
		r.Get("/delete_data", gophKeeperInstance.DeleteDataHandler)
		r.Post("/edit_data", gophKeeperInstance.EditDataHandler)
		//r.Post("/", gophKeeperInstance.PostHandler)
		//r.Get("/{id}", gophKeeperInstance.GetHandler)
		//r.Post("/api/shorten", gophKeeperInstance.JSONPostHandler)
		//r.Post("/api/shorten/batch", gophKeeperInstance.JSONBatchHandler)
		//r.Get("/api/internal/stats", gophKeeperInstance.GetInternalStats)
		//
		//r.With(cookieManager.AuthMiddleware).Get("/api/user/urls", gophKeeperInstance.GetUserUrls)
		//r.With(cookieManager.AuthMiddleware).Delete("/api/user/urls", gophKeeperInstance.DeleteUserUrls)
	})

	r.Post("/registration", gophKeeperInstance.RegistrationHandler)
	r.Post("/authentication", gophKeeperInstance.AuthenticationHandler)
	r.Get("/ping", gophKeeperInstance.PingHandler)

	server.Addr = cfg.ServerAddress
	server.Handler = r
	err := server.ListenAndServe()

	if err != nil {
		log.Printf("Error starting server: %s", err)
	}
}
