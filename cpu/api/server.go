package api

import (
	"net/http"

	"github.com/sisoputnfrba/tp-golang/cpu/api/handlers"
	global "github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/utils/server"
)

func CreateServer() *server.Server {
	configServer := server.Config{
		Port: global.CpuConfig.Puerto,
		Handlers: map[string]http.HandlerFunc{
			"PUT /interrupcion":   handlers.Interrupcion,
			"PUT /recibirTIDYPID": handlers.ObtenerDeKernel,
		},
	}
	return server.NuevoServer(configServer)
}
