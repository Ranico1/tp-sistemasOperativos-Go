package global

import (
	"fmt"
	"os"
	"sync"

	config "github.com/sisoputnfrba/tp-golang/utils/config"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
)

type Config struct {
	IpMemoria     string `json:"ip_memory"`
	PuertoMemoria int    `json:"port_memory"`
	IpKernel      string `json:"ip_kernel"`
	PuertoKernel  int    `json:"port_kernel"`
	Puerto        int    `json:"port"`
	LogLevel      string `json:"log_level"`
}

var HuboInterrupcion = false
var EsPrimeraEjecucion = true

var ArrancaEjecucion chan struct{}
var Ejecutando bool
var CpuConfig *Config
var MotivoInterrupcion string
var Logger *log.LoggerStruct

var PCBMutex sync.Mutex
var MutexEjecucion sync.Mutex

var TidEjecutando int
var PIDEjecutando int

const CPULog = "./cpu.log"

func InitGlobal() {
	args := os.Args[1:]
	if len(args) != 2 {
		fmt.Println("Argumentos esperados para iniciar el servidor: ENV=dev | prod CONFIG=config_path")
		os.Exit(1)
	}
	env := args[0]
	archivoConfiguracion := args[1]

	Logger = log.ConfigurarLogger(CPULog, env)
	CpuConfig = config.CargarConfig[Config](archivoConfiguracion)
}
