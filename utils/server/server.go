package server
import (
	"fmt"
	"net/http"
)

type Server struct {
	router *http.ServeMux
	port   int
}

type Config struct {
	Port     int
	Handlers map[string]http.HandlerFunc
}

func NuevoServer(config Config) *Server {
	s := &Server{
		router: http.NewServeMux(),
		port:   config.Port,
	}

	for path, handler := range config.Handlers {
		s.router.HandleFunc(path, handler)
	}

	return s
}

func (s *Server) Iniciar() error {
	addr := fmt.Sprintf(":%d", s.port)
	return http.ListenAndServe(addr, s.router)
}