package main

import (
	"fmt"
	"os"

	"github.com/sisoputnfrba/tp-golang/cpu/api"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
)

func main() {
	global.InitGlobal()
	server := api.CreateServer()

	global.Logger.Log(fmt.Sprintf("Iniciando servidor CPU en el puerto: %d", global.CpuConfig.Puerto), log.INFO)

	err := server.Iniciar()
	if err != nil {
		global.Logger.Log(fmt.Sprintf("Error al iniciar el servidor CPU | Error: %v", err), log.ERROR)
		os.Exit(1)
	}
	global.Logger.CloseLogger()
}
