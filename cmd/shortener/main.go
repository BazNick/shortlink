package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	"github.com/BazNick/shortlink/cmd/config"
	"github.com/BazNick/shortlink/cmd/middleware/auth"
	"github.com/BazNick/shortlink/cmd/middleware/compress"
	"github.com/BazNick/shortlink/cmd/middleware/logger"
	"github.com/BazNick/shortlink/internal/app/entities"
	"github.com/BazNick/shortlink/internal/app/handlers"
	"github.com/BazNick/shortlink/internal/app/storage"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	printBuildInfo()

	var (
		conf    config.Config
		router  = gin.Default()
		storage storage.Storage
		err     error
	)

	conf, err = config.GetCLParams()
	if err != nil {
		log.Fatal(err)
	}

	pprof.Register(router)

	switch {
	case conf.DB != "":
		db := entities.NewDB(conf.DB)
		storage = db

		defer db.Database.Close()
	case conf.FilePath != "":
		file := entities.NewFileStore(conf.FilePath)
		storage = file

		defer file.FileStorage.Close()
	default:
		hashDict := entities.NewHashDict()
		storage = hashDict
	}

	urlHandler := handlers.NewURLHandler(
		storage,
		conf.FilePath,
		conf.DB,
	)

	router.Use(
		logger.WithLogging(),
		logger.WithLogging(),
		compress.GzipHandle(),
		auth.Auth(conf.SecretKey),
	)

	router.GET("/:id", urlHandler.GetLink)
	router.POST("/", urlHandler.AddLink)
	router.POST("/api/shorten", urlHandler.PostJSONLink)
	router.GET("/ping", urlHandler.DBPingConn)
	router.POST("/api/shorten/batch", urlHandler.BatchLinks)
	router.GET("/api/user/urls", urlHandler.GetUserLinks)
	router.DELETE("/api/user/urls", urlHandler.DeleteUserLinks)

	// Start server with HTTP or HTTPS based on configuration
	if err := startServer(router, conf); err != nil {
		log.Fatal(err)
	}
}

// startServer запускает HTTP или HTTPS сервер в зависимости от конфигурации.
// Если включен HTTPS, использует http.ListenAndServeTLS с указанными сертификатами.
// В противном случае запускает обычный HTTP сервер.
//
// Параметры:
//   - router: настроенный Gin роутер
//   - conf: конфигурация приложения
//
// Возвращает:
//   - error: ошибку, если не удалось запустить сервер
//
// Пример:
//
//	router := gin.Default()
//	conf := config.Config{EnableHTTPS: true, CertFile: "cert.pem", KeyFile: "key.pem"}
//	if err := startServer(router, conf); err != nil {
//	    log.Fatal(err)
//	}
func startServer(router *gin.Engine, conf config.Config) error {
	server := &http.Server{
		Addr:    conf.Address,
		Handler: router,
	}

	if conf.EnableHTTPS {
		if conf.CertFile == "" || conf.KeyFile == "" {
			return fmt.Errorf("HTTPS enabled but certificate files not provided. Use -cert and -key flags or CERT_FILE and KEY_FILE environment variables")
		}

		server.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}

		fmt.Printf("Starting HTTPS server on %s\n", conf.Address)
		fmt.Printf("Certificate file: %s\n", conf.CertFile)
		fmt.Printf("Private key file: %s\n", conf.KeyFile)

		return server.ListenAndServeTLS(conf.CertFile, conf.KeyFile)
	}

	fmt.Printf("Starting HTTP server on %s\n", conf.Address)
	return server.ListenAndServe()
}

func printBuildInfo() {
	version := buildVersion
	if version == "" {
		version = "N/A"
	}

	date := buildDate
	if date == "" {
		date = "N/A"
	}

	commit := buildCommit
	if commit == "" {
		commit = "N/A"
	}

	fmt.Printf("Build version: %s\n", version)
	fmt.Printf("Build date: %s\n", date)
	fmt.Printf("Build commit: %s\n", commit)
}
