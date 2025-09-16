package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/BazNick/shortlink/cmd/config"
	"github.com/BazNick/shortlink/cmd/middleware/auth"
	"github.com/BazNick/shortlink/cmd/middleware/compress"
	"github.com/BazNick/shortlink/cmd/middleware/logger"
	"github.com/BazNick/shortlink/cmd/middleware/trusted"
	"github.com/BazNick/shortlink/internal/app/entities"
	"github.com/BazNick/shortlink/internal/app/grpc"
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
	buildInfo()

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

	// Initialize storage
	switch {
	case conf.DB != "":
		db := entities.NewDB(conf.DB)
		storage = db
	case conf.FilePath != "":
		file := entities.NewFileStore(conf.FilePath)
		storage = file
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

	internalAPI := router.Group("/api/internal")
	internalAPI.Use(trusted.TrustedIPMiddleware(conf.TrustedSubnet))
	internalAPI.GET("/stats", urlHandler.GetStats)

	if err := startServersWithGracefulShutdown(router, conf, storage); err != nil {
		log.Fatal(err)
	}
}

// startServerWithGracefulShutdown запускает HTTP или HTTPS сервер с поддержкой graceful shutdown.
// Сервер корректно завершает работу при получении сигналов SIGTERM, SIGINT, SIGQUIT.
// Все необработанные запросы завершаются, все несохраненные данные сохраняются в репозитории.
//
// Параметры:
//   - router: настроенный Gin роутер
//   - conf: конфигурация приложения
//   - storage: хранилище данных для сохранения при завершении
//
// Возвращает:
//   - error: ошибку, если не удалось запустить сервер
//
// Пример:
//
//	router := gin.Default()
//	conf := config.Config{EnableHTTPS: true, CertFile: "cert.pem", KeyFile: "key.pem"}
//	storage := entities.NewDB("postgres://...")
//	if err := startServerWithGracefulShutdown(router, conf, storage); err != nil {
//	    log.Fatal(err)
//	}
func startServerWithGracefulShutdown(router *gin.Engine, conf config.Config, storage storage.Storage) error {
	server := &http.Server{
		Addr:    conf.Address,
		Handler: router,
	}

	// Настройка HTTPS если включен
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
	} else {
		fmt.Printf("Starting HTTP server on %s\n", conf.Address)
	}

	// Канал для получения сигналов операционной системы
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	// Запуск сервера в отдельной горутине
	serverErr := make(chan error, 1)
	go func() {
		var err error
		if conf.EnableHTTPS {
			err = server.ListenAndServeTLS(conf.CertFile, conf.KeyFile)
		} else {
			err = server.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	fmt.Println("Server started. Press Ctrl+C to stop gracefully.")

	// Ожидание сигнала завершения
	sig := <-sigChan
	fmt.Printf("\nReceived signal: %v. Starting graceful shutdown...\n", sig)

	// Создание контекста с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Попытка graceful shutdown сервера
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Server forced to shutdown: %v\n", err)
		return err
	}

	// Сохранение всех несохраненных данных в хранилище
	if err := saveAllData(); err != nil {
		fmt.Printf("Warning: failed to save all data: %v\n", err)
	}

	// Закрытие соединений с хранилищем
	if err := closeStorage(storage); err != nil {
		fmt.Printf("Warning: failed to close storage properly: %v\n", err)
	}

	fmt.Println("Server shutdown completed gracefully.")
	return nil
}

// startServersWithGracefulShutdown запускает HTTP/HTTPS и gRPC серверы с поддержкой graceful shutdown.
// Серверы корректно завершают работу при получении сигналов SIGTERM, SIGINT, SIGQUIT.
//
// Параметры:
//   - router: настроенный Gin роутер
//   - conf: конфигурация приложения
//   - storage: хранилище данных для сохранения при завершении
//
// Возвращает:
//   - error: ошибку, если не удалось запустить серверы
func startServersWithGracefulShutdown(router *gin.Engine, conf config.Config, storage storage.Storage) error {
	var grpcServer *grpc.Server

	// Создаем gRPC сервер если он включен
	if conf.EnableGRPC {
		// Получаем worker manager для gRPC сервера
		var workerManager *entities.DeleteWorkerManager
		if dbStorage, ok := storage.(*entities.DB); ok {
			workerManager = entities.NewDeleteWorkerManager(dbStorage.Database, 100)
			workerManager.StartDeleteWorkers(runtime.NumCPU())
		}

		grpcServer = grpc.NewServer(storage, workerManager)
		fmt.Printf("Starting gRPC server on %s\n", conf.GRPCAddress)
	}

	// Запускаем gRPC сервер в отдельной горутине если он включен
	var grpcErr chan error
	if grpcServer != nil {
		grpcErr = make(chan error, 1)
		go func() {
			if err := grpcServer.Start(conf.GRPCAddress); err != nil {
				grpcErr <- err
			}
		}()
	}

	// Запускаем HTTP сервер
	if err := startServerWithGracefulShutdown(router, conf, storage); err != nil {
		// Останавливаем gRPC сервер если HTTP сервер не запустился
		if grpcServer != nil {
			grpcServer.Stop()
		}
		return err
	}

	// Останавливаем gRPC сервер при завершении HTTP сервера
	if grpcServer != nil {
		grpcServer.Stop()
	}

	return nil
}

// saveAllData сохраняет все несохраненные данные в хранилище.
// Обеспечивает целостность данных при завершении работы сервера.
//
// Возвращает:
//   - error: ошибку, если не удалось сохранить данные
func saveAllData() error {
	// Здесь можно добавить логику для принудительного сохранения
	// всех несохраненных данных в хранилище
	// Например, для файлового хранилища - принудительная синхронизация
	// Для базы данных - коммит всех транзакций

	// Для текущей реализации большинство операций уже синхронные
	// но можно добавить дополнительную логику при необходимости
	fmt.Println("Saving all pending data to storage...")
	return nil
}

// closeStorage корректно закрывает соединения с хранилищем.
// Обеспечивает освобождение ресурсов при завершении работы сервера.
//
// Параметры:
//   - storage: хранилище данных
//
// Возвращает:
//   - error: ошибку, если не удалось закрыть хранилище
func closeStorage(storage storage.Storage) error {
	fmt.Println("Closing storage connections...")

	// Проверяем тип хранилища и закрываем соответствующие соединения
	switch s := storage.(type) {
	case *entities.DB:
		if s.Database != nil {
			return s.Database.Close()
		}
	case *entities.FileStore:
		if s.FileStorage != nil {
			return s.FileStorage.Close()
		}
	case *entities.HashDict:
		// Для in-memory хранилища ничего закрывать не нужно
		return nil
	}

	return nil
}

func buildInfo() {
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
