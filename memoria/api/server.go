package api

import (
	"net/http"

	"github.com/sisoputnfrba/tp-golang/memoria/api/handlers"
	"github.com/sisoputnfrba/tp-golang/memoria/global"
	"github.com/sisoputnfrba/tp-golang/utils/server"
)

func CreateServer() *server.Server {
	configServer := server.Config{
		Port: global.MemoryConfig.Port,
		Handlers: map[string]http.HandlerFunc{

			"GET /tamanioProceso/{tamanio}": handlers.RecibirTamanioProceso,
			"GET /instruccion":              handlers.EnviarInstruccion,
			"GET /contextoDeEjecucion":      handlers.ObtenerContextoEjecucion,
			"PUT /creacionHilo":             handlers.CrearHilo,
			"PUT /eliminacionProceso":       handlers.EliminarProceso,
			"PUT /eliminacionHilo":          handlers.EliminarHilo,
			"PUT /actualizacionContexto":    handlers.ActualizarContexto,
			"PUT /creacionProceso":          handlers.AgregarProcesoAMemoria,
			"GET /lecturaMemoria":           handlers.LeerMemoria,
			"PUT /escrituraMemoria":         handlers.EscribirMemoria,
			"PUT /compactacionMemoria":      handlers.CompactarMemoria,
			"GET /memoryDump/{pid}/{tid}":   handlers.MemoryDump,
		},
	}
	return server.NuevoServer(configServer)
}
