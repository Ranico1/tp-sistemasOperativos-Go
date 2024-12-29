package global

import (
	"fmt"
	"io"
	"math"
	"os"

	config "github.com/sisoputnfrba/tp-golang/utils/config"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
)

type Config struct {
	Port             int    `json:"port"`
	IPMemory         string `json:"ip_memory"`
	PortMemory       int    `json:"port_memory"`
	MountDir         string `json:"mount_dir"`
	BlockSize        int    `json:"block_size"`
	BlockCount       int    `json:"block_count"`
	BlockAccessDelay int    `json:"block_access_delay"`
	LogLevel         string `json:"log_level"`
}

type Archivo struct {
	IndexBlock int `json:"index_block"`
	Size       int `json:"size"`
}

var FSConfig *Config

var Logger *log.LoggerStruct

const FSLOG = "./filesystem.log"

// var Bloques []byte

var ArchivoBitmap *os.File

var Bitmap []byte // Esta formado por ceros y unos donde:  0 = Vacio y 1 = Ocupado

// Nombre  Contenido(IndexBlock, Size)
//var ArchivosMetadata map[string]Archivo

func InitGlobal() {
	args := os.Args[1:]
	if len(args) != 2 {
		fmt.Println("Argumentos esperados para iniciar el servidor: ENV=dev | prod CONFIG=config_path")
		os.Exit(1)
	}
	env := args[0]
	archivoConfiguracion := args[1]

	Logger = log.ConfigurarLogger(FSLOG, env)
	FSConfig = config.CargarConfig[Config](archivoConfiguracion)

	CrearCarpetaMountDir()

	//ArchivosMetadata = map[string]Archivo{}
	LevantarFS(FSConfig)

}

func LevantarFS(fsconfig *Config) {

	AbrirBitmapDat(fsconfig)

	//AbrirBloquesDat(fsconfig)

	//CargarArchivosMetadata(fsconfig)

}

func AbrirBitmapDat(fsconfig *Config) {

	ruta := fsconfig.MountDir + "/bitmap.dat"
	var err error
	// Como el tamanio se tiene que redondear al superior, hago la cuenta en float y despues lo redondeo
	tamanio := float64(fsconfig.BlockCount) / 8.0
	tamanioRedondeado := int64(math.Ceil(tamanio))

	Bitmap = make([]byte, tamanioRedondeado) // Le asigno espacio a la variable global y se inicializa en 0

	ArchivoBitmap, err = os.OpenFile(ruta, os.O_RDWR|os.O_CREATE, 0666) // 0666 son permisos de lectura y escritura para todos
	if err != nil {
		Logger.Log(fmt.Sprint("Error al crear/abrir archivo: ", err), log.ERROR)
		return
	}

	// Ajusto el archivo que se creo/abrio

	err = ArchivoBitmap.Truncate(tamanioRedondeado)
	if err != nil {
		Logger.Log(fmt.Sprint("Error al ajustar el archivo: ", err), log.ERROR)
	}

	// Cargo el slice de Bitmap
	_, err = ArchivoBitmap.ReadAt(Bitmap, 0)
	if err != nil && err != io.EOF {
		Logger.Log(fmt.Sprint("Error al leer el archivo para cargarlo: ", err), log.ERROR)
	}

	bitsAdicionales := int(tamanioRedondeado*8) - fsconfig.BlockCount
	if bitsAdicionales > 0 {
		for i := 0; i < bitsAdicionales; i++ {
			Bitmap[len(Bitmap)-1] |= (1 << i)
		}
	}
}

func CrearCarpetaMountDir() {
	rutaMountDir := FSConfig.MountDir

	//Verificar si existe
	_, err := os.Stat(rutaMountDir)
	if os.IsNotExist(err) {
		//Crear directorio
		err = os.MkdirAll(rutaMountDir, 0777)
		if err != nil {
			Logger.Log(fmt.Sprintf("Error al crear directorio %s: %s", rutaMountDir, err.Error()), log.ERROR)
			return
		}
	}

	rutaFiles := FSConfig.MountDir + "/files"
	_, err1 := os.Stat(rutaFiles)
	if os.IsNotExist(err1) {
		//Crear directorio
		err1 = os.MkdirAll(rutaFiles, 0777)
		if err1 != nil {
			Logger.Log(fmt.Sprintf("Error al crear carpeta files %s: %s", rutaFiles, err1.Error()), log.ERROR)
			return
		}
	}
}

// func AbrirBloquesDat(fsconfig *Config) {

// 	ruta := fsconfig.MountDir + "/bloques.dat"

// 	tamanio := fsconfig.BlockCount * fsconfig.BlockSize

// 	Bloques = make([]byte, tamanio)

// 	archivo, err := os.OpenFile(ruta, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
// 	if err != nil {
// 		Logger.Log(fmt.Sprint("Error al crear/abrir archivo: ", err), log.ERROR)
// 		return
// 	}

// 	defer archivo.Close()

// 	err = archivo.Truncate(int64(tamanio))
// 	if err != nil {
// 		Logger.Log(fmt.Sprint("Error al ajustar el archivo: ", err), log.ERROR)
// 	}

// 	// Cargo el slice de Bloques
// 	_, err = archivo.Read(Bitmap)
// 	if err != nil && err != io.EOF {
// 		Logger.Log(fmt.Sprint("Error al leer el archivo para cargarlo: ", err), log.ERROR)
// 	}

// }

// func CargarArchivosMetadata(fsconfig *Config) {
// 	ruta := fsconfig.MountDir + "/files"

// 	archivos, err := os.ReadDir(ruta)
// 	if err != nil {
// 		Logger.Log(fmt.Sprintf("Error al leer el directorio %s ", err.Error()), log.ERROR)
// 		return
// 	}
// 	// Me fijo que el archivo no sea un directorio y que tenga en el nombre dmp
// 	for _, archivo := range archivos {
// 		if !archivo.IsDir() && strings.Contains(archivo.Name(), "dmp") {
// 			agregarArchivoMetadata(archivo.Name())
// 		}
// 	}
// }

// func agregarArchivoMetadata(nombreArchivo string) {
// 	var archivoMetadata Archivo

// 	ruta := FSConfig.MountDir + "/files/" + nombreArchivo

// 	archivo, err := os.Open(ruta)

// 	if err != nil {
// 		Logger.Log(fmt.Sprintf("Error al abrir el archivo %s :  %s", ruta, err.Error()), log.ERROR)
// 		return
// 	}

// 	defer archivo.Close()

// 	decoder := json.NewDecoder(archivo)
// 	err = decoder.Decode(&archivoMetadata)
// 	if err != nil {
// 		Logger.Log(fmt.Sprintf("Error al decodificar el archivo %s :  %s", ruta, err.Error()), log.ERROR)
// 		return
// 	}

// 	ArchivosMetadata[nombreArchivo] = archivoMetadata
// 	Logger.Log(fmt.Sprintf("Archivo metadata decodeado: %+v", archivoMetadata), log.DEBUG)

// }
