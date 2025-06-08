package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"token-transfer-api/internal/db"
	"token-transfer-api/internal/graph"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/vektah/gqlparser/v2/ast"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	dbConnection, err := db.ConnectDb()
	if err != nil {
		log.Fatal(err)
	}

	err = db.CreateDefaultAccount(dbConnection)
	if err != nil {
		err2 := db.CloseDb(dbConnection)
		if err2 != nil {
			log.Print(err2)
		}
		log.Fatal(err)
	}
	defer func() {
		err := db.CloseDb(dbConnection)
		if err != nil {
			log.Fatal(err)
		}
	}()

	srv := handler.New(
		graph.NewExecutableSchema(
			graph.Config{Resolvers: &graph.Resolver{Db: dbConnection}},
		),
	)

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	mux := http.NewServeMux()
	mux.Handle("/", playground.Handler("GraphQL playground", "/query"))
	mux.Handle("/query", srv)

	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
		if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	// Wait for an interrupt signal or a server error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Printf("server error: %v", err)
	case sig := <-quit:
		log.Printf("Shutdown signal %q received, starting shutdown...", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}

	log.Println("Server exiting.")
}
