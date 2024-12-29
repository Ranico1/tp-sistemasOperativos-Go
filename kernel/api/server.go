package api

import (
	"net/http"

	"github.com/sisoputnfrba/tp-golang/kernel/api/handlers"
	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/utils/server"
)

func CrearServer() *server.Server {

	configServer := server.Config{
		Port: global.KernelConfig.Port,
		Handlers: map[string]http.HandlerFunc{
			"PUT /interrupcion":      handlers.Interrupcion,
			"PUT /dumpProceso":       handlers.DumpMemory,
			"PUT /crearProceso":      handlers.ProcessCreate,
			"PUT /crearHilo":         handlers.ThreadCreate,
			"PUT /unirHilo":          handlers.ThreadJoin,
			"DELETE /hilo":           handlers.ThreadExit,
			"DELETE /hilo/{TID}":     handlers.ThreadCancel,
			"DELETE /proceso/{TID}":  handlers.ProcessExit,
			"PUT /IO/{milisegundos}": handlers.IO,
			"PUT /MutexCreate":       handlers.MutexCreate,
			"PUT /MutexLock":         handlers.MutexLock,
			"PUT /MutexUnlock":       handlers.MutexUnlock,
		},
	}
	return server.NuevoServer(configServer)
}
