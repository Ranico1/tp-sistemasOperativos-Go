package internal

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
)

type ContextoPedido struct {
	AX    uint32
	BX    uint32
	CX    uint32
	DX    uint32
	EX    uint32
	FX    uint32
	GX    uint32
	HX    uint32
	PC    uint32
	Base  int
	Limit int
}

type ContextoActualizado struct {
	Pid int    `json:"pid"`
	Tid int    `json:"tid"`
	AX  uint32 `json:"ax"`
	BX  uint32 `json:"bx"`
	CX  uint32 `json:"cx"`
	DX  uint32 `json:"dx"`
	EX  uint32 `json:"ex"`
	FX  uint32 `json:"fx"`
	GX  uint32 `json:"gx"`
	HX  uint32 `json:"hx"`
	PC  uint32 `json:"pc"`
}

var mu sync.Mutex

func ConsultarEspacioDisponible(tamanioAOcupar int) (bool, error, bool) {

	var err error

	switch global.MemoryConfig.Scheme {
	case "FIJAS":
		for _, particion := range global.Memoria.Particiones {
			if particion.Libre && particion.Size >= tamanioAOcupar {
				return true, nil, false
			}
		}
		return false, nil, false
	case "DINAMICAS":
		totalLibre := 0
		for _, particion := range global.Memoria.Particiones {
			if particion.Libre {

				totalLibre += particion.Size
				// Tambien va a haber espacio si alguna particion es mayor o igual al tamanio a ocupar
				if particion.Size >= tamanioAOcupar {
					return true, nil, false
				}
			}
		}
		// Si hay espacio suficiente para compactar
		if totalLibre >= tamanioAOcupar {
			return false, nil, true
		}
		return false, nil, false
	}

	return false, err, false

}

// Agregar un archivo de pseudoCodigo a un hilo pasandole el pid, path y tid.

func AgregarArchivoAhilo(pid int, path string, tid int) error {
	var err error
	if proceso, existe1 := global.Procesos[pid]; existe1 {
		//
		if thread, existe2 := proceso.TIDs[tid]; existe2 {

			file, err := os.Open(path) //Abro el archivo que tiene el pseudocodigo (instrucciones assembler)
			if err != nil {
				return err
			}
			defer file.Close()

			var instrucciones []string

			scanner := bufio.NewScanner(file)

			for scanner.Scan() { //Recorro el archivo y guardo las instrucciones
				instrucciones = append(instrucciones, strings.TrimSpace(scanner.Text())) //trim space elimina los espacios en blanco
			}

			if err := scanner.Err(); err != nil {
				return err
			}

			thread.Instrucciones = instrucciones
			proceso.TIDs[tid] = thread
			global.Procesos[pid] = proceso

			return nil

		}
	}
	return err
}

//Se busca la instruccion pedida por CPU y se devuelve la misma en caso de encontrarla
func BuscarInstruccion(pid int, tid int, pc int) (string, error) {
	var err error
	if proceso, existe := global.Procesos[pid]; existe {
		if thread, existe := proceso.TIDs[tid]; existe {
			instruccion := thread.Instrucciones[pc]
			return instruccion, nil
		}
	}
	return "", err

}

//Se busca el contexto pedido por CPU y se devuelve el mismo en caso de encontrarlo
func BuscarContexto(pid int, tid int) (ContextoPedido, error) {
	var contexto ContextoPedido
	var err error
	if proceso, existe := global.Procesos[pid]; existe {
		contexto.Base = proceso.Contexto.Base
		contexto.Limit = proceso.Contexto.Limit
		global.Logger.Log(fmt.Sprintf("PID: %d, BASE DEL CONTEXTO EN FUNC: %d, LIMITE: %d", pid, contexto.Base, contexto.Limit), log.DEBUG)
		if thread, existe := proceso.TIDs[tid]; existe {
			contexto.AX = thread.Contexto.AX
			contexto.BX = thread.Contexto.BX
			contexto.CX = thread.Contexto.CX
			contexto.DX = thread.Contexto.DX
			contexto.EX = thread.Contexto.EX
			contexto.FX = thread.Contexto.FX
			contexto.GX = thread.Contexto.GX
			contexto.HX = thread.Contexto.HX
			contexto.PC = thread.Contexto.PC
			return contexto, nil
		}
	}

	return ContextoPedido{}, err
}

func BorrarEstructurasProceso(pid int) {

	delete(global.Procesos, pid)

}

func BorrarEstructurasHilo(pid int, tid int) {
	delete(global.Procesos[pid].TIDs, tid)
}

//Se actualiza el contexto del hilo deseado
func ActualizarRegistros(registrosActualizados ContextoActualizado) {

	var nuevoContexto global.ThreadContext

	proceso := global.Procesos[registrosActualizados.Pid]
	hilo := proceso.TIDs[registrosActualizados.Tid]

	nuevoContexto.AX = registrosActualizados.AX
	nuevoContexto.BX = registrosActualizados.BX
	nuevoContexto.CX = registrosActualizados.CX
	nuevoContexto.DX = registrosActualizados.DX
	nuevoContexto.EX = registrosActualizados.EX
	nuevoContexto.FX = registrosActualizados.FX
	nuevoContexto.GX = registrosActualizados.GX
	nuevoContexto.HX = registrosActualizados.HX
	nuevoContexto.PC = registrosActualizados.PC

	hilo.Contexto = nuevoContexto

	proceso.TIDs[registrosActualizados.Tid] = hilo
	global.Procesos[registrosActualizados.Pid] = proceso

}

func CompactarMemoria() error {
	var err error
	totalHuecosLibres := 0
	for _, particion := range global.Memoria.Particiones {
		if particion.Libre {
			totalHuecosLibres += particion.Size
		}
	}

	// Si no hay huecos libres, no se necesita compactar la memoria
	if totalHuecosLibres == 0 {
		global.Logger.Log("No hay huecos libres", log.ERROR)
		return err
	}

	var particionesCompactadas []global.Particion

	baseActual := 0

	// Crear un nuevo array de bytes para la memoria compactada
	nuevaMemoria := make([]byte, len(global.Memoria.Espacios))

	for _, particion := range global.Memoria.Particiones {
		if !particion.Libre {

			// Copiar el contenido desde la base antigua a la nueva base
			copy(nuevaMemoria[baseActual:baseActual+particion.Size], global.Memoria.Espacios[particion.Base:particion.Base+particion.Size])

			// Actualizo la base de la partición ocupada y el proceso que esta en la misma
			particion.Base = baseActual
			proceso := global.Procesos[particion.Pid]
			proceso.Contexto.Base = baseActual
			proceso.Contexto.Limit = baseActual + particion.Size - 1
			global.Procesos[particion.Pid] = proceso

			particionesCompactadas = append(particionesCompactadas, particion)
			baseActual += particion.Size
		}
	}

	// Agrego una partición libre con el tamaño de totalHuecosLibres
	particionesCompactadas = append(particionesCompactadas, global.Particion{
		Base:  baseActual,
		Size:  totalHuecosLibres,
		Libre: true,
		Pid: -1,
	})

	// Actualizar la memoria global con la memoria compactada
	global.Memoria.Espacios = nuevaMemoria

	// Actualizar la lista de particiones en la memoria global
	global.Memoria.Particiones = particionesCompactadas

	return nil
}

func VerSiPidEstaRepetido(pid int) bool {

	for _, particion := range global.Memoria.Particiones {
		if particion.Pid == pid {
			return true
		}
	}
	return false
}

func AgregarProceso(pid int, tamanioProceso int) {
	estrategia := global.MemoryConfig.Search_algorithm
	var indice int

	switch estrategia {
	case "FIRST":
		indice = FirstFit(tamanioProceso)
	case "BEST":
		indice = BestFit(tamanioProceso)
	case "WORST":
		indice = WorstFit(tamanioProceso)
	}

	particion := global.Memoria.Particiones[indice]

	//Marco la particion como ocupada
	global.Memoria.Particiones[indice].Libre = false
	global.Memoria.Particiones[indice].Pid = pid

	if particion.Size > tamanioProceso && global.MemoryConfig.Scheme == "DINAMICAS" { //Si la particion es mas grande que el tamaño requerido, la divido

		nuevaParticion := global.Particion{
			Base:  particion.Base + tamanioProceso,
			Size:  particion.Size - tamanioProceso,
			Libre: true,
			Pid:   -1,
		}

		global.Memoria.Particiones[indice].Size = tamanioProceso
		global.Memoria.Particiones = append(global.Memoria.Particiones[:indice+1], append([]global.Particion{nuevaParticion}, global.Memoria.Particiones[indice+1:]...)...)
	}

	
	//Actualizo la base y el limite del proceso correspondiente
	var contexto global.ProcessContext
	// proceso := global.Procesos[pid]

	contexto.Base = particion.Base
	
	if global.MemoryConfig.Scheme == "DINAMICAS" {
		contexto.Limit = particion.Base + tamanioProceso - 1
	} else {
		contexto.Limit = particion.Base + particion.Size - 1
	}

	var proceso global.Process
	
	

	global.Logger.Log(fmt.Sprintf("Base: %d, Limite: %d: ", contexto.Base, contexto.Limit), log.DEBUG)
	proceso.Contexto = contexto

	// Inicializo el mapa de hilos, con el hilo 0
	var Hilos = make(map[int]global.Thread)
	Hilos = map[int]global.Thread{}

	threadCero := global.Thread{Contexto: global.ThreadContext{AX: 0, BX: 0, CX: 0, DX: 0, EX: 0, FX: 0, GX: 0, HX: 0, PC: 0}, Archivo: "", Instrucciones: []string{}}

	Hilos[0] = threadCero

	proceso.TIDs = Hilos

	global.Procesos[pid] = proceso

}

func AgregarNuevoHilo(pid int, tid int, archivo string) error {

	nuevoHilo := global.Thread{Contexto: global.ThreadContext{AX: 0, BX: 0, CX: 0, DX: 0, EX: 0, FX: 0, GX: 0, HX: 0, PC: 0}, Archivo: archivo, Instrucciones: []string{}}

	// fmt.Printf("\n%d\n", pid) // ! borrar
	// fmt.Printf("\n%d\n", tid)

	ruta := "/home/utnso/tp-2024-2c-Futbol-y-Negocios/Pruebas"
	rutaPseudocodigo := filepath.Join(ruta, archivo)

	proceso := global.Procesos[pid]

	proceso.TIDs[tid] = nuevoHilo

	global.Procesos[pid] = proceso

	err := AgregarArchivoAhilo(pid, rutaPseudocodigo, tid)
	if err != nil {
		global.Logger.Log(fmt.Sprintf("Error al cargar instrucciones al hilo - (PID:TID) - (%d,%d)", pid, tid), log.ERROR)
		return err
	}

	return nil
}

func FirstFit(tamanioProceso int) int {
	for i, particion := range global.Memoria.Particiones {
		if particion.Libre && particion.Size >= tamanioProceso {
			return i
		}
	}
	return -1
}

func BestFit(tamanioProceso int) int {
	indice := -1
	tamanioMinimo := int(^uint(0) >> 1)
	for i, particion := range global.Memoria.Particiones {
		if particion.Libre && particion.Size >= tamanioProceso && particion.Size < tamanioMinimo {
			indice = i
			tamanioMinimo = particion.Size
		}
	}
	return indice
}

func WorstFit(tamanioProceso int) int {
	indice := -1
	tamanioMaximo := 0
	for i, particion := range global.Memoria.Particiones {
		if particion.Libre && particion.Size >= tamanioProceso && particion.Size > tamanioMaximo {
			indice = i
			tamanioMaximo = particion.Size
		}
	}
	return indice
}

func LiberarParticion(pid int) int {
	tamanioParticion := -1
	for i, particion := range global.Memoria.Particiones {
		if particion.Pid == pid {
			tamanioParticion = global.Memoria.Particiones[i].Size
			global.Memoria.Particiones[i].Libre = true
			global.Memoria.Particiones[i].Pid = -1

			if global.MemoryConfig.Scheme == "DINAMICAS" {
				if ((i+1) < len(global.Memoria.Particiones) && global.Memoria.Particiones[i+1].Libre) || ((i-1) != -1 && global.Memoria.Particiones[i-1].Libre) {
					ConsolidarParticiones(i)
				}
			}
		}
	}
	return tamanioParticion
}

// En caso de que la particion liberada tenga particiones aledañas libres, se consolidan
func ConsolidarParticiones(i int) {

	// Con el append lo que logramos es que se elimine la particion que se consolidan y se sumen sus tamaños a la particion que esta mas a la izquierda
	// En el caso que sean 3 particiones juntas libres, se eliminan las 2 que le siguen a la primera y se suman sus tamaños a la primera, y en el append
	// se agrega las particiones hasta la primera y la segunda y tercera se eliminan
	if i < len(global.Memoria.Particiones)-1 && global.Memoria.Particiones[i+1].Libre && i > 0 && global.Memoria.Particiones[i-1].Libre {
		global.Memoria.Particiones[i-1].Size += global.Memoria.Particiones[i].Size + global.Memoria.Particiones[i+1].Size
		global.Memoria.Particiones = append(global.Memoria.Particiones[:i], global.Memoria.Particiones[i+2:]...)
		return
	}

	// Cuando la particion de la derecha esta libre se la elimina y se suma su tamaño a la particion de la derecha
	if i < len(global.Memoria.Particiones)-1 && global.Memoria.Particiones[i+1].Libre {
		global.Memoria.Particiones[i].Size += global.Memoria.Particiones[i+1].Size
		global.Memoria.Particiones = append(global.Memoria.Particiones[:i+1], global.Memoria.Particiones[i+2:]...)
		return
	}
	if i > 0 && global.Memoria.Particiones[i-1].Libre {
		global.Memoria.Particiones[i-1].Size += global.Memoria.Particiones[i].Size
		global.Memoria.Particiones = append(global.Memoria.Particiones[:i], global.Memoria.Particiones[i+1:]...)
		return
	}
}

// Funcion para la lectura de la memoria en el array de bytes contiguo

func LeerDireccion(direccionFisica int) []byte {
	mu.Lock()
	defer mu.Unlock()

	arrayMemoria := global.Memoria.Espacios
    arrayAenviar := make([]byte, 4)

    copy(arrayAenviar, arrayMemoria[direccionFisica:direccionFisica+4])
	global.Logger.Log(fmt.Sprintf("Leyendo de memoria - Base: %d, Contenido: %v", direccionFisica, arrayAenviar), log.DEBUG)

	//fmt.Println("LECTURA: ", arrayAenviar)
	return arrayAenviar
}

// Funcion para la escritura de la memoria

func EscribirEspacioMemoria(dato []byte, direccionFisica int) {
	mu.Lock()
	defer mu.Unlock()

	//fmt.Println("DATO A ESCRIBIR: ", dato)
	copy(global.Memoria.Espacios[direccionFisica:direccionFisica+4], dato)

	global.Logger.Log(fmt.Sprintf("Escribiendo en memoria - Base: %d, Contenido: %v", direccionFisica, dato), log.DEBUG)

}

func ObtenerPidPorDireccionFisica(direccionFisica int) int {
	for _, particion := range global.Memoria.Particiones {
		if particion.Base <= direccionFisica && direccionFisica < particion.Base+particion.Size {
			return particion.Pid
		}
	}
	return -1 // Indica que no se encontró ninguna partición
}

func ObtenerContenidoYtamanioProceso(pid int) ([]byte, int) {
	var contenido []byte
	var tamanio int
	for _, particion := range global.Memoria.Particiones {
		if particion.Pid == pid {
			// Leer el contenido de la partición del array de bytes
			global.Logger.Log(fmt.Sprintf("Leyendo partición - PID: %d, Base: %d, Size: %d", pid, particion.Base, particion.Size), log.DEBUG)
			contenido = append(contenido, global.Memoria.Espacios[particion.Base:particion.Base+particion.Size]...)
			tamanio = particion.Size
		}
	}
	return contenido, tamanio
}

// AsignarMemoria asigna memoria a un nuevo proceso
// func AsignarMemoria(pid int, size int) error {
// 	for i, particion := range global.Memoria.Particiones {
// 		if particion.Libre && particion.Size >= size {
// 			// Asignar la partición
// 			global.Memoria.Particiones[i].Libre = false
// 			// Si la partición es más grande que el tamaño requerido, dividirla
// 			if particion.Size > size {
// 				nuevaParticion := global.Particion{
// 					Base:  particion.Base + size,
// 					Size:  particion.Size - size,
// 					Libre: true,
// 				}
// 				global.Memoria.Particiones[i].Size = size
// 				global.Memoria.Particiones = append(global.Memoria.Particiones[:i+1], append([]global.Particion{nuevaParticion}, global.Memoria.Particiones[i+1:]...)...)
// 			}
// 			// Actualizar el contexto del proceso
// 			processContexts[pid] = global.ProcessContext{Base: particion.Base, Limit: particion.Base + size - 1}

// 			return nil
// 		}
// 	}
// 	return fmt.Errorf("no hay suficiente memoria disponible")
// }
