package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/cko-recruitment/payment-gateway-challenge-go/docs"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"

	"golang.org/x/sync/errgroup"
)

type Api struct {
	router          *chi.Mux
	paymentsHandler *PaymentsHandler
}

func New(paymentsHandler *PaymentsHandler) *Api {
	a := &Api{
		paymentsHandler: paymentsHandler,
	}

	a.setupRouter()

	return a
}

func (a *Api) Run(ctx context.Context, addr string) error {
	httpServer := &http.Server{
		Addr:        addr,
		Handler:     a.router,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		<-ctx.Done()
		fmt.Printf("shutting down HTTP server\n")
		return httpServer.Shutdown(ctx)
	})

	g.Go(func() error {
		fmt.Printf("starting HTTP server on %s\n", addr)
		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			return err
		}

		return nil
	})

	return g.Wait()
}

func (a *Api) setupRouter() {
	a.router = chi.NewRouter()

	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
	)

	a.router.Use(middleware.RequestID)
	a.router.Use(RequestLogger(logger))
	a.router.Use(middleware.Recoverer)
	a.router.Use(middleware.Timeout(60 * time.Second))

	a.router.Get("/swagger/*", a.SwaggerHandler())

	a.router.Route("/api/v1", func(r chi.Router) {
		r.Get("/ping", a.PingHandler())

		r.Get("/payments/{id}", a.paymentsHandler.GetHandler())
		r.Post("/payments", a.paymentsHandler.PostHandler())
	})
}

type pong struct {
	Message string `json:"message"`
}

// PingHandler godoc
//
// @Summary     Health check
// @Description Simple health check endpoint used to verify service availability
// @Tags        health
// @Success     204 "Service is healthy"
// @Failure     500 "Internal server error"
// @Router      /api/v1/ping [get]
func (a *Api) PingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		if err := json.NewEncoder(w).Encode(pong{Message: "pong"}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

// SwaggerHandler returns an http.HandlerFunc that handles HTTP Swagger related requests.
func (a *Api) SwaggerHandler() http.HandlerFunc {
	return httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://%s/swagger/doc.json", docs.SwaggerInfo.Host)),
	)
}
