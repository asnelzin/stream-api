package rest

import (
	"log"

	"net/http"

	"context"
	"fmt"
	"sync"
	"time"

	"github.com/asnelzin/stream-api/pkg/store"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type Server struct {
	Version   string
	DataStore store.Engine

	lock       sync.Mutex
	httpServer *http.Server
}

func (s *Server) Run(port int) {
	log.Printf("[INFO] activate rest server on port %d", port)

	router := s.routes()

	s.lock.Lock()
	s.httpServer = &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	s.lock.Unlock()

	err := s.httpServer.ListenAndServe()
	log.Printf("[WARN] http server terminated, %s", err)
}

// Shutdown rest http server
func (s *Server) Shutdown() {
	log.Print("[WARN] shutdown rest server")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	s.lock.Lock()
	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("[DEBUG] rest shutdown error, %s", err)
	}
	log.Print("[DEBUG] shutdown rest server completed")
	s.lock.Unlock()
}

func (s *Server) routes() chi.Router {
	router := chi.NewRouter()
	router.Use(middleware.RequestID, middleware.Recoverer)
	router.Use(AppInfo("stream-api", s.Version), Ping)

	router.Route("/v1", func(r chi.Router) {
		r.Use(middleware.Logger)

		r.Mount("/streams", MakeHandler(s.DataStore))
	})

	return router
}
