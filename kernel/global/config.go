package global

import (
	"fmt"
	"os"
	"sync"

	config "github.com/sisoputnfrba/tp-golang/utils/config"
	estructuras "github.com/sisoputnfrba/tp-golang/utils/estructuras"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
)

const KernelLog = "./kernel.log"

var InterrumpioCPU chan struct{}
var EsperaRutinaSyscall chan struct{}
var EnRutinaSyscall chan struct{}

var HuboSyscall bool

type Config struct {
	Port               int    `json:"port"`
	IPMemory           string `json:"ip_memory"`
	PortMemory         int    `json:"port_memory"`
	IPCPU              string `json:"ip_cpu"`
	PortCPU            int    `json:"port_cpu"`
	SchedulerAlgorithm string `json:"scheduler_algorithm"`
	Quantum            int    `json:"quantum"`
	LogLevel           string `json:"log_level"`
}

type ProcesoParaNew struct {
	PseudoCodigo string `json:"path"`
	Tamanio      int    `json:"tamanio"`
	PCB          estructuras.PCB
}

type IO struct {
	PID int
	TID int
	MS  int
}

var KernelConfig *Config
var Logger *log.LoggerStruct

var ExisteJoin = true

var ContadorPid = 0

var NinguHiloParaEjecutar = false

var BloqueoPorMutex = false

var InfoEstadoNuevo []ProcesoParaNew

var MutexPID sync.Mutex

var HuboInterrupcionQuantum = false

// Colas segun estado
var EstadoListo map[int][]estructuras.TCB
var EstadoNuevo []estructuras.TCB
var EstadoBloqueado []estructuras.TCB
var EstadoEjecutando []estructuras.TCB
var EstadoSalida []estructuras.TCB

var ProcesosEnSistema []estructuras.PCB // Solo guarda informacion de los PCB

var PrioridadesEnSistema []int

var EsperandoIO []IO

//var TimerIO *time.Timer

var MutexIO sync.Mutex
var MutexColaIO sync.Mutex

// Mutex por colas
var MutexListo sync.Mutex
var MutexNuevo sync.Mutex
var MutexExit sync.Mutex
var MutexBloqueado sync.Mutex
var MutexEjecutando sync.Mutex
var MutexProcesosEnSistema sync.Mutex
var MutexSalida sync.Mutex

var MutexInfoEstadoNuevo sync.Mutex

func InitGlobal() {
	args := os.Args[1:]
	if len(args) <= 2 {
		fmt.Println("Argumentos esperados para iniciar el servidor: ENV=dev | prod CONFIG=config_path")
		os.Exit(1)
	}
	env := args[0]
	archivoConfiguracion := args[1]

	Logger = log.ConfigurarLogger(KernelLog, env)
	KernelConfig = config.CargarConfig[Config](archivoConfiguracion)
	EstadoListo = make(map[int][]estructuras.TCB)
	EstadoListo[0] = []estructuras.TCB{}
	PrioridadesEnSistema = append(PrioridadesEnSistema, 0)
	// var ProcesoEjecutando = 0 // no seria variable global (?)

}
