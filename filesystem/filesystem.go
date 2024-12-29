package main

import (
	"fmt"
	"os"

	"github.com/sisoputnfrba/tp-golang/filesystem/api"
	"github.com/sisoputnfrba/tp-golang/filesystem/global"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
)

func main() {

	global.InitGlobal()

	s := api.CreateServer()

	global.Logger.Log(fmt.Sprintf("Iniciando FS server on port: %d", global.FSConfig.Port), log.INFO)

	err := s.Iniciar()
	if err != nil {
		global.Logger.Log(fmt.Sprintf("Fallo para iniciar FS server: %v", err), log.ERROR)
		os.Exit(1)
	}

	global.Logger.CloseLogger()

}
