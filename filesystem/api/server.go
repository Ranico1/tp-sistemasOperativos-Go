package api

import (
	"net/http"

	"github.com/sisoputnfrba/tp-golang/filesystem/api/handlers"
	global "github.com/sisoputnfrba/tp-golang/filesystem/global"
	"github.com/sisoputnfrba/tp-golang/utils/server"
)

func CreateServer() *server.Server {
	configServer := server.Config{
		Port:     global.FSConfig.Port,
		Handlers: map[string]http.HandlerFunc{
			"PUT /memoryDump": 					handlers.CrearArchivoDump,
		},
	}
	return server.NuevoServer(configServer)
}
