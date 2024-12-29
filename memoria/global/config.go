package global

import (
	"fmt"
	"os"


	config "github.com/sisoputnfrba/tp-golang/utils/config"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
)

const MEMORYLOG = "./memoria.log"

type Config struct {
	Port             int      `json:"port"`
	Memory_size      int      `json:"memory_size"`
	Instruction_path string   `json:"instruction_path"`
	Response_delay   int      `json:"response_delay"`
	Ip_kernel        string   `json:"ip_kernel"`
	Port_kernel      int      `json:"port_kernel"`
	Ip_cpu           string   `json:"ip_cpu"`
	Port_cpu         int      `json:"port_cpu"`
	Ip_filesystem    string   `json:"ip_filesystem"`
	Port_filesystem  int      `json:"port_filesystem"`
	Scheme           string   `json:"scheme"`
	Search_algorithm string   `json:"search_algorithm"`
	Partitions       []int    `json:"partitions"`
	Log_level		 string   `json:"log_level"`
}

// ProcessContext almacena la base y el límite para la traducción de direcciones
type ProcessContext struct {
	Base  int
	Limit int
}


// ThreadContext almacena los registros de la CPU para cada hilo
type ThreadContext struct {
	AX uint32
	BX uint32
	CX uint32
	DX uint32
	EX uint32
	FX uint32
	GX uint32
	HX uint32
	PC uint32
}

type Process struct {
	Contexto ProcessContext
	TIDs     map[int]Thread
}

type Thread struct {
	Contexto ThreadContext
	Archivo string
	Instrucciones []string
}

// Particion representa una partición de memoria
type Particion struct {
	Base  int
	Size  int
	Libre bool
	Pid	  int
}

type EstructuraMemoria struct {
	Espacios    []byte			//memoria usuario (solo voy a leer y escribir)
	Particiones []Particion 	
}

// Mapas para almacenar los contextos de ejecución y sus
var Procesos = make(map[int]Process)
var Hilos = make(map[int]Thread)



var Memoria *EstructuraMemoria


var MemoryConfig *Config
var Logger *log.LoggerStruct

func InitGlobal() {
	args := os.Args[1:]
	if len(args) != 2 {
		fmt.Println("Argumentos esperados para iniciar el servidor: ENV=dev | prod CONFIG=config_path")
		os.Exit(1)
	}
	env := args[0]
	archivoConfiguracion := args[1]

	Logger = log.ConfigurarLogger(MEMORYLOG, env)
	MemoryConfig = config.CargarConfig[Config](archivoConfiguracion)
	Memoria = InicializarMemoria()


	
	Procesos = map[int]Process{}
	Hilos = map[int]Thread{}

	


}

func InicializarMemoria() *EstructuraMemoria {
	byteArray := make([]byte, MemoryConfig.Memory_size)
	memoria := EstructuraMemoria{Espacios: byteArray}
	switch MemoryConfig.Scheme {
	case "FIJAS":
		base := 0
		//agregue lo del tamaño de la particion, ver si esta bien 
		for _, size := range MemoryConfig.Partitions { //Voy tomando cada elemento del array de particiones: _ es el indice, size es lo que hay en el array
			memoria.Particiones = append(memoria.Particiones, Particion{Base: base, Size: size, Libre: true, Pid : -1})
			base += size
		}
	case "DINAMICAS":
		// Inicializamos una partición dinámica que cubre toda la memoria
		// A la particion inicial unica, se la va ocupando desde el principio y queda la parte libre al final,
		//
		memoria.Particiones = append(memoria.Particiones, Particion{Base: 0, Size: MemoryConfig.Memory_size, Libre: true, Pid : -1})
	}

	return &memoria
}
