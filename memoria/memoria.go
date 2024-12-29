package main

import (
	"fmt"
	"os"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"github.com/sisoputnfrba/tp-golang/memoria/api"

	
)

func main() {
	global.InitGlobal()

	s := api.CreateServer()

	global.Logger.Log(fmt.Sprintf("Starting Memory server on port: %d", global.MemoryConfig.Port), log.INFO)

	if err := s.Iniciar(); err != nil {
		global.Logger.Log(fmt.Sprintf("Failed to start Memory server: %v", err), log.ERROR)
		os.Exit(1)
	}

	global.Logger.CloseLogger()
}
