package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/robfig/cron/v3"

	"nfe-sefaz-sync/configs"
	"nfe-sefaz-sync/internal/handler"
	"nfe-sefaz-sync/internal/repository"
	"nfe-sefaz-sync/internal/service"
	"nfe-sefaz-sync/pkg/certificate"
	"nfe-sefaz-sync/pkg/database"
	"nfe-sefaz-sync/pkg/logger"
)

func main() {
	// Inicializa o logger
	log := logger.New("info")
	log.Info("Iniciando aplicação NFe SEFAZ Sync")

	// Carrega as configurações
	cfg, err := configs.LoadConfig()
	if err != nil {
		log.Fatal("Erro ao carregar configurações", "error", err)
	}

	// Valida as configurações
	if err := cfg.Validate(); err != nil {
		log.Fatal("Configurações inválidas", "error", err)
	}

	log.Info("Configurações carregadas com sucesso",
		"ambiente", cfg.Sefaz.Ambiente,
		"uf", cfg.Sefaz.UF,
	)

	// Conecta ao banco de dados
	db, err := database.NewPostgresConnection(cfg.Database.GetDSN(), cfg.Database.MaxConnections)
	if err != nil {
		log.Fatal("Erro ao conectar ao banco de dados", "error", err)
	}
	defer db.Close()

	log.Info("Conectado ao banco de dados com sucesso")

	// Carrega o certificado digital
	cert, err := certificate.LoadCertificate(cfg.Sefaz.CertPath, cfg.Sefaz.CertPassword)
	if err != nil {
		log.Fatal("Erro ao carregar certificado", "error", err)
	}

	log.Info("Certificado carregado com sucesso")

	// Cria o diretório de armazenamento de XMLs se não existir
	if err := os.MkdirAll(cfg.Storage.XMLPath, 0755); err != nil {
		log.Fatal("Erro ao criar diretório de armazenamento", "error", err)
	}

	// Inicializa as camadas da aplicação
	nfeRepository := repository.NewNFeRepository(db)
	sefazClient := service.NewSefazClient(
		cfg.Sefaz.Ambiente,
		cfg.Sefaz.UF,
		cfg.Sefaz.CNPJ,
		cert,
		cfg.Sefaz.Timeout,
		log,
	)
	nfeService := service.NewNFeService(
		nfeRepository,
		sefazClient,
		cfg.Storage.XMLPath,
		log,
	)

	// Configura o scheduler de sincronização
	if cfg.Sync.Enabled {
		c := cron.New()
		_, err := c.AddFunc(cfg.Sync.CronSchedule, func() {
			log.Info("Iniciando sincronização agendada")
			if _, err := nfeService.SyncNFes(); err != nil {
				log.Error("Erro na sincronização agendada", "error", err)
			}
		})
		if err != nil {
			log.Fatal("Erro ao configurar scheduler", "error", err)
		}
		c.Start()
		defer c.Stop()
		log.Info("Scheduler de sincronização configurado", "schedule", cfg.Sync.CronSchedule)
	}

	// Configura as rotas
	r := chi.NewRouter()

	// Middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","database":"connected","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
	})

	// Registra as rotas da API
	nfeHandler := handler.NewNFeHandler(nfeService, log)
	nfeHandler.RegisterRoutes(r)

	// Configura o servidor HTTP
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Inicia o servidor em uma goroutine
	go func() {
		log.Info("Servidor HTTP iniciado", "address", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Erro ao iniciar servidor", "error", err)
		}
	}()

	// Aguarda sinal de interrupção
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Encerrando aplicação...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Erro ao encerrar servidor", "error", err)
	}

	log.Info("Aplicação encerrada com sucesso")
}