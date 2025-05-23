package main

import (
	"net/http"
	"os"
	"os/signal"
	"subscriptions/internal/config"
	"subscriptions/internal/graphql/graph"
	"subscriptions/internal/graphql/graph/model"
	"subscriptions/internal/lib/logger/handlers/slogpretty"
	"subscriptions/internal/storage/postgres"
	"syscall"

	"golang.org/x/exp/slog"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/vektah/gqlparser/v2/ast"
	"golang.org/x/time/rate"
)

const (
	envLocal    = "local"
	envDev      = "dev"
	envProd     = "prod"
	limit       = rate.Limit(1000)
	burst       = 1000
	defaultPort = "8080"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)

	log.Info("starting subscriptions", slog.String("env", cfg.Env), slog.String("version", "123"))
	log.Debug("debug messages are enabled")

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	storage, err := postgres.InitDB(cfg.StoragePath, cfg.DBSaver)
	if err != nil {
		log.Error("failed to initialize database", err)
	}
	defer storage.CloseDB()

	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{Logger: log, CommentAddedNotification: make(chan *model.Comment, 1), Storage: storage}}))
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	// srv.AddTransport()
	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{Cache: lru.New[string](100)})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Debug("connect to http://localhost:%s/ for GraphQL playground", port)

	srvChan := make(chan os.Signal, 1)
	signal.Notify(srvChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Error("failed to start server", err)
		}
	}()

	<-srvChan
	log.Info("shutting down server...")
	log.Info("server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)
	return slog.New(handler)
}
