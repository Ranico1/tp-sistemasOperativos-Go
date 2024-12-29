package handlers

import (
	//"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	global "github.com/sisoputnfrba/tp-golang/memoria/global"
	internal "github.com/sisoputnfrba/tp-golang/memoria/internal"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
	metodosHttp "github.com/sisoputnfrba/tp-golang/utils/metodosHttp"
)

type EspacioMemDisponible struct {
	TamMemoria int `json:"tamanio"`
}

// Struct con lo que pasa Kernel para crear un proceso
type ProcesoAcrear struct {
	Archivo    string `json:"path"`
	TamMemoria int    `json:"tamanio"`
	Pid        int    `json:"pid"`
}

// Struct con lo que pasa CPU para pedir una instruccion
type PedidoInstruccion struct {
	Pid int `json:"pid"`
	Tid int `json:"tid"`
	Pc  int `json:"programCounter"`
}

// Struct con lo que pasa Kernel para crear un hilo
type HiloAcrear struct {
	Archivo string `json:"path"`
	Pid     int    `json:"pid"`
	Tid     int    `json:"tid"`
}

// Struct con lo que pasa kernel para eliminar un hilo 
type HiloAeliminar struct {
	Pid int `json:"pid"`
	Tid int `json:"tid"`
}

// Struct con lo que pasa CPU para pedir el contexto de un hilo 
type PedidoContexto struct {
	Pid int `json:"pid"`
	Tid int `json:"tid"`
}

// Struct con lo que pasa Kernel para finalizar un proceso
type ProcesoAfinalizar struct {
	Pid int `json:"pid"`
}

// Struct con lo que pasa CPU para leer memoria
type PedidoLectura struct {
	DireccionFisica int `json:"direccionFisica"`
	Tid             int `json:"tid"`
}

// Struct con lo que le devolvemos a CPU cuando nos pide lectura 
type RespuestaLectura struct {
	Datos []byte `json:"datos"`
}

// Struct con lo que pasa CPU para escribir en memoria
type PedidoEscritura struct {
	Datos           []byte `json:"datos"`
	DireccionFisica int    `json:"direccionFisica"`
	Tid             int    `json:"tid"`
}

// Struct con lo que nos pasa Kernel para hacer el memory dump
type MemoriaDump struct {
	NombreArchivo string `json:"nombre_archivo"`
	Tamanio       int    `json:"tamanio"`
	Contenido     []byte `json:"contenido"`
}


//Recibir tamanio del proceso para ver si se puede crear o no.
func RecibirTamanioProceso(w http.ResponseWriter, r *http.Request) {

	// queryParams := r.URL.Query()
	// tamanioEnInt, _ := strconv.Atoi(queryParams.Get("tamanio"))

	//if err != nil {
	//	global.Logger.Log("Error al decodificar body: "+err.Error(), log.ERROR)
	//	http.Error(w, "Error al decodificar body", http.StatusBadRequest)
	//	return
	//}

	tamanioEnInt, _ := strconv.Atoi(r.PathValue("tamanio"))

	hayParticionLibre, err, hacerCompactacion := internal.ConsultarEspacioDisponible(tamanioEnInt)

	if err != nil {
		global.Logger.Log("Fallo la busqueda de espacio para el proceso", log.ERROR)
	}

	// Se responde que hay particion libre
	if hayParticionLibre {
		w.WriteHeader(http.StatusOK)
		global.Logger.Log("Hay espacio para almacenar al proceso", log.INFO)
		return
	}

	//Se responde que  se puede hacer la compactacion
	if hacerCompactacion {
		w.WriteHeader(http.StatusLengthRequired)
		return
	}

	// En caso de no haber espacio (de ningun tipo, ya sea particion libre o para compactar) se responde como Error
	if !hayParticionLibre && !hacerCompactacion {
		global.Logger.Log("No hay espacio suficiente para crear el proceso", log.ERROR)
		http.Error(w, "No hay espacio suficiente para crear el proceso", http.StatusBadRequest)

		return
	}

}

func CompactarMemoria(w http.ResponseWriter, r *http.Request) {

	//global.Logger.Log("Se hizo la request ", log.INFO)
	
	//Se compacta la memoria
	err := internal.CompactarMemoria()

	// En caso de que haya habido un error al compactar, se responde con un error

	// LOG NO OBLIGATORIO PARA LA COMPACTACION
	global.Logger.Log(("Se ha compactado la memoria"), log.INFO)

	// RECORRIDO DE LAS PARTICIONES PARA QUE ME LAS MUESTRE
	for i, particion := range global.Memoria.Particiones {
		//fmt.Printf("Partición %d: Base=%d, Size=%d, Libre=%t, Pid=%d\n", i, particion.Base, particion.Size, particion.Libre, particion.Pid)
		global.Logger.Log(fmt.Sprintf("Partición %d: Base=%d, Size=%d, Libre=%t, Pid=%d", i, particion.Base, particion.Size, particion.Libre, particion.Pid), log.LogLevel("INFO"))

	}
	if err == nil {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

// Funcion para agregar el archivo de pseudocodigo al hilo cero en la memoria del sistema y reservar la memoria de usuario
func AgregarProcesoAMemoria(w http.ResponseWriter, r *http.Request) {
	var proceso ProcesoAcrear
	err := metodosHttp.DecodeHTTPBody(r, &proceso)
	if err != nil {
		global.Logger.Log("Error al decodificar body: "+err.Error(), log.ERROR)
		http.Error(w, "Error al decodificar body", http.StatusBadRequest)
		return
	}

	ruta := "/home/utnso/tp-2024-2c-Futbol-y-Negocios/Pruebas"
	rutaPseudocodigo := filepath.Join(ruta, proceso.Archivo)

	// Inicializar el map de los TIDs del proceso (?

	//Revisar que no sea un PID repetido
	estaRepetido := internal.VerSiPidEstaRepetido(proceso.Pid)
	if estaRepetido {
		global.Logger.Log("El pid ingresado ya existe: ", log.ERROR)
		http.Error(w, "Error al decodificar body", http.StatusBadRequest)
		return
	}

	//Funcion para agregar el proceso a la particion fija o dinamica
	internal.AgregarProceso(proceso.Pid, proceso.TamMemoria)

	//Se hace el log
	global.Logger.Log(fmt.Sprintf("## Proceso creado - PID: %d - Tamaño: %d", proceso.Pid, proceso.TamMemoria), log.INFO)

	//Se agrega el archivo de pseudocodigo al hilo cero
	err2 := internal.AgregarArchivoAhilo(proceso.Pid, rutaPseudocodigo, 0)
	if err2 != nil {
		global.Logger.Log("No se cargo el archivo porque no existe el hilo", log.ERROR)
	}

	for i, particion := range global.Memoria.Particiones {
		//fmt.Printf("Partición %d: Base=%d, Size=%d, Libre=%t, Pid=%d\n", i, particion.Base, particion.Size, particion.Libre, particion.Pid)
		global.Logger.Log(fmt.Sprintf("Partición %d: Base=%d, Size=%d, Libre=%t, Pid=%d", i, particion.Base, particion.Size, particion.Libre, particion.Pid), log.LogLevel("INFO"))
	}
}

func pasarAint(cadena string) int {
	num, err := strconv.Atoi(cadena)
	if err != nil {
		global.Logger.Log("Error al convertir a INT", log.ERROR)
	}
	return num
}

//Enviar instruccion a CPU
func EnviarInstruccion(w http.ResponseWriter, r *http.Request) {
	//var Pedido PedidoInstruccion
	// err := metodosHttp.DecodeHTTPBody(r, &Pedido)
	// if err != nil {
	// 	global.Logger.Log("Error al decodificar body: "+err.Error(), log.ERROR)
	// 	http.Error(w, "Error al decodificar body", http.StatusBadRequest)
	// 	return
	// }

	queryParams := r.URL.Query()
	Pid := pasarAint(queryParams.Get("pid"))
	Tid := pasarAint(queryParams.Get("tid"))
	Pc := pasarAint(queryParams.Get("pc"))

	//Delay que hay que ponerle a las respuestas a los requests
	time.Sleep(time.Duration(global.MemoryConfig.Response_delay) * time.Millisecond)

	instruccion, err := internal.BuscarInstruccion(Pid, Tid, Pc)
	if err != nil {
		fmt.Printf("La instruccion buscada para el PID: %d no fue encontrada", Pid)
		http.Error(w, "Instruccion no encontrada", http.StatusBadRequest)
		return
	}

	instruccionRespuesta, err1 := json.Marshal(instruccion)
	if err1 != nil {
		http.Error(w, "Error al codificar instruccion", http.StatusBadRequest)
		return
	}

	//Se hace el log
	global.Logger.Log(fmt.Sprintf(("## Obtener instrucción - (PID:TID) - (%d:%d) - Instrucción: %s"), Pid, Tid, instruccion), log.INFO)

	//Se envia la instruccion a CPU
	w.WriteHeader(http.StatusOK)
	w.Write(instruccionRespuesta)
}

func ObtenerContextoEjecucion(w http.ResponseWriter, r *http.Request) {
	//var PedidoContexto PedidoContexto
	//err := metodosHttp.DecodeHTTPBody(r, &PedidoContexto)
	//	if err != nil {
	//		global.Logger.Log("Error al decodificar body: "+err.Error(), log.ERROR)
	//		http.Error(w, "Error al decodificar body", http.StatusBadRequest)
	//		return
	//	}

	queryParams := r.URL.Query()
	Pid := pasarAint(queryParams.Get("pid"))
	Tid := pasarAint(queryParams.Get("tid"))

	// Delay de respuesta para las requests
	time.Sleep(time.Duration(global.MemoryConfig.Response_delay) * time.Millisecond)

	// Buscar contexto pedido por CPU mediante un PID-TID y marshalizarlo para poder enviarlo
	contextoConseguido, err := internal.BuscarContexto(Pid, Tid)
	if err != nil {
		http.Error(w, "Error al buscar contexto", http.StatusBadRequest)
		return
	}

	//fmt.Println("CONTEXTO BASE:", contextoConseguido.Base)

	contextoRespuesta, err := json.Marshal(contextoConseguido)
	if err != nil {
		http.Error(w, "Error al codificar contexto", http.StatusBadRequest)
		return
	}

	//Se hace el log
	global.Logger.Log(fmt.Sprintf("## Contexto Solicitado - (PID:TID) - (%d,%d)", Pid, Tid), log.INFO)

	// Enviar el contexto
	w.WriteHeader(http.StatusOK)
	w.Write(contextoRespuesta)
}

// Crear hilo a partir de un Pid, Tid y su archivo de pseudocodigo
func CrearHilo(w http.ResponseWriter, r *http.Request) {
	var PedidoHilo HiloAcrear
	err := metodosHttp.DecodeHTTPBody(r, &PedidoHilo)
	if err != nil {
		global.Logger.Log("Error al decodificar body: "+err.Error(), log.ERROR)
		http.Error(w, "Error al decodificar body", http.StatusBadRequest)
		return
	}

	err = internal.AgregarNuevoHilo(PedidoHilo.Pid, PedidoHilo.Tid, PedidoHilo.Archivo)
	if err != nil {
		global.Logger.Log("Error al agregar hilo", log.ERROR)
		http.Error(w, "Error al agregar hilo", http.StatusBadRequest)
		return
	}

	//TODO:Se hace el log
	global.Logger.Log(fmt.Sprintf("## Hilo creado - (PID:TID) - (%d,%d)", PedidoHilo.Pid, PedidoHilo.Tid), log.INFO)

	// Se responde como "ok" a (Kernel) a la creacion del hilo
	w.WriteHeader(http.StatusOK)

	// proceso := global.Procesos[PedidoHilo.Pid]
	// for i := range proceso.TIDs {
	// 	fmt.Printf("numero hilo: %d", i)
	// }
}

func EliminarProceso(w http.ResponseWriter, r *http.Request) {
	var pedidoAeliminar ProcesoAfinalizar
	err := metodosHttp.DecodeHTTPBody(r, &pedidoAeliminar)
	if err != nil {
		global.Logger.Log("Error al decodificar body: "+err.Error(), log.ERROR)
		http.Error(w, "Error al decodificar body", http.StatusBadRequest)
		return
	}

	//Eliminar estructuras de memoria de sistema
	internal.BorrarEstructurasProceso(pedidoAeliminar.Pid)

	//Funcion para liberar particion ocupada y hacer la compactacion de particiones libres aledañas.
	tamanioProceso := internal.LiberarParticion(pedidoAeliminar.Pid)

	//Significa que encontro el proceso
	if tamanioProceso != -1 {
		global.Logger.Log(fmt.Sprintf("## Proceso destruido - PID: %d - Tamaño: %d", pedidoAeliminar.Pid, tamanioProceso), log.INFO)
	} else {
		global.Logger.Log(fmt.Sprintf("Proceso no encontrado - PID: %d", pedidoAeliminar.Pid), log.DEBUG)
	}

	for i, particion := range global.Memoria.Particiones {
		//fmt.Printf("Partición %d: Base=%d, Size=%d, Libre=%t, Pid=%d\n", i, particion.Base, particion.Size, particion.Libre, particion.Pid)
		global.Logger.Log(fmt.Sprintf("Partición %d: Base=%d, Size=%d, Libre=%t, Pid=%d", i, particion.Base, particion.Size, particion.Libre, particion.Pid), log.LogLevel("INFO"))
	}
}

// Funcion para eliminar el hilo pedido por Kernel
func EliminarHilo(w http.ResponseWriter, r *http.Request) {
	var PedidoHiloAeliminar HiloAeliminar
	err := metodosHttp.DecodeHTTPBody(r, &PedidoHiloAeliminar)
	if err != nil {
		global.Logger.Log("Error al decodificar body: "+err.Error(), log.ERROR)
		http.Error(w, "Error al decodificar body", http.StatusBadRequest)
		return
	}

	// Funcion para eliminar las estructuras del hilo en memoria del sistema
	internal.BorrarEstructurasHilo(PedidoHiloAeliminar.Pid, PedidoHiloAeliminar.Tid)

	//Se hace el log
	global.Logger.Log(fmt.Sprintf("## Hilo destruido - (PID:TID) - (%d,%d)", PedidoHiloAeliminar.Pid, PedidoHiloAeliminar.Tid), log.INFO)

	// Responde a Kernel con un "ok"
	w.WriteHeader(http.StatusOK)

	//proceso := global.Procesos[PedidoHiloAeliminar.Pid]
	// for i := range proceso.TIDs {
	// 	fmt.Printf("numero hilo: %d", i)
	// }
}

func ActualizarContexto(w http.ResponseWriter, r *http.Request) {

	var Contexto internal.ContextoActualizado
	err := metodosHttp.DecodeHTTPBody(r, &Contexto)
	if err != nil {
		global.Logger.Log("Error al decodificar body: "+err.Error(), log.ERROR)
		http.Error(w, "Error al decodificar body", http.StatusBadRequest)
		return
	}
	if _, existe := global.Procesos[Contexto.Pid]; existe {
		internal.ActualizarRegistros(Contexto)

		//Se hace el log
		global.Logger.Log(fmt.Sprintf("## Contexto Actualizado - (PID-TID) - (%d,%d)", Contexto.Pid, Contexto.Tid), log.INFO)

		global.Logger.Log(fmt.Sprintf("Contexto (HILO - %d - PID - %d) , es: %d, %d, %d, %d, %d, %d, %d, %d, %d \n", Contexto.Tid, Contexto.Pid, Contexto.AX, Contexto.BX, Contexto.CX, Contexto.DX, Contexto.EX, Contexto.FX, Contexto.GX, Contexto.HX, Contexto.PC), log.DEBUG)

	}

}

func LeerMemoria(w http.ResponseWriter, r *http.Request) {

	queryParams := r.URL.Query()
	Tid := pasarAint(queryParams.Get("tid"))
	DireccionFisica := pasarAint(queryParams.Get("direccion"))

	//Delay de respuesta para las requests
	time.Sleep(time.Duration(global.MemoryConfig.Response_delay) * time.Millisecond)

	valorParaEnviar := internal.LeerDireccion(DireccionFisica)

	var datos RespuestaLectura = RespuestaLectura{
		Datos: valorParaEnviar,
	}

	//Printeamos lo del array para ver que onda

	//fmt.Println(datos.Datos)

	// datosAenviar, err := json.Marshal(datos)
	// if err != nil {
	// 	http.Error(w, "Error al codificar datos", http.StatusBadRequest)
	// 	return
	// }

	pidDelProcesoAleer := internal.ObtenerPidPorDireccionFisica(DireccionFisica)

	global.Logger.Log("A", "DEBUG")

	if pidDelProcesoAleer == -1 {
		global.Logger.Log("No se encontro el PID correspondiente al TID solicitado", log.ERROR)
		http.Error(w, "No se encontro el PID correspondiente al TID solicitado", http.StatusBadRequest)
		return
	}

	global.Logger.Log("B", "DEBUG")
	//Se hace el log
	global.Logger.Log(fmt.Sprintf("## Lectura - (PID:TID) - (%d,%d) - Dir.Fisica: %d - Tamaño: 4 bytes", pidDelProcesoAleer, Tid, DireccionFisica), log.INFO)

	//Enviar datos a CPU
	//fmt.Println("DATOS A ENVIAR PARA CPU:", datos.Datos)
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	w.Write(datos.Datos)
}

func EscribirMemoria(w http.ResponseWriter, r *http.Request) {
	var Pedido PedidoEscritura
	err := metodosHttp.DecodeHTTPBody(r, &Pedido)
	if err != nil {
		global.Logger.Log("Error al decodificar body: "+err.Error(), log.ERROR)
		http.Error(w, "Error al decodificar body", http.StatusBadRequest)
		return
	}

	// Delay de respuesta para las requests
	time.Sleep(time.Duration(global.MemoryConfig.Response_delay) * time.Millisecond)

	internal.EscribirEspacioMemoria(Pedido.Datos, Pedido.DireccionFisica)

	pidDelProcesoAescribir := internal.ObtenerPidPorDireccionFisica(Pedido.DireccionFisica)

	if pidDelProcesoAescribir == -1 {
		global.Logger.Log("Error no se encontro el PID del TID solicitado", log.DEBUG)
		http.Error(w, "No se encontro el PID correspondiente al TID solicitado", http.StatusBadRequest)
		return
	}

	//Se hace el log
	global.Logger.Log(fmt.Sprintf("## Escritura - (PID:TID) - (%d,%d) - Dir.Fisica: %d - Tamaño: 4 bytes", pidDelProcesoAescribir, Pedido.Tid, Pedido.DireccionFisica), log.INFO)

	//Responder ok a que se pudo escribir en memoria
	w.WriteHeader(http.StatusOK)

	// array := global.Memoria.Espacios
	// for i := range array {
	// 	fmt.Print(array[i])
	// }
}

func MemoryDump(w http.ResponseWriter, r *http.Request) {
	// queryParams := r.URL.Query()
	// Pid := pasarAint(queryParams.Get("pid"))
	// Tid := pasarAint(queryParams.Get("tid"))

	Pid, _ := strconv.Atoi(r.PathValue("pid"))
	Tid, _ := strconv.Atoi(r.PathValue("tid"))

	contenidoProceso, tamanioEnviar := internal.ObtenerContenidoYtamanioProceso(Pid)
	tiempoActual := time.Now()

	tiempoFormateado := tiempoActual.Format("15:04:05:000")

	nombreDelArchivo := fmt.Sprintf("%d-%d-%s.dmp", Pid, Tid, tiempoFormateado)

	contenidoAenviar := MemoriaDump{
		NombreArchivo: nombreDelArchivo,
		Tamanio:       tamanioEnviar,
		Contenido:     contenidoProceso,
	}

	global.Logger.Log(fmt.Sprintf("## Memory Dump solicitado - (PID:TID) - (%d - %d)", Pid, Tid), log.INFO)
	_, err := metodosHttp.PutHTTPwithBody[MemoriaDump, int](global.MemoryConfig.Ip_filesystem, global.MemoryConfig.Port_filesystem, "memoryDump", contenidoAenviar)

	if err == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		w.WriteHeader(http.StatusOK)
	}

	// respuesta, err := json.Marshal(resp)
	// if err != nil {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }

}

// //recibir pid y path de las instrucciones
// func RecibirProceso(w http.ResponseWriter, r *http.Request) {
// 	// tiempo de delay que hay ante una request
// 	DelayResponse := time.Duration(global.MemoryConfig.Response_delay)
// 	time.Sleep(DelayResponse * time.Millisecond)

// 	var nuevoProceso NewProcessRequest
// 	err := metodosHttp.DecodeHTTPBody(r, &nuevoProceso)
// 	if err != nil {
// 		global.Logger.Log("Error al decodificar body: "+err.Error(), log.ERROR)
// 		http.Error(w, "Error al decodificar body", http.StatusBadRequest)
// 		return
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte("ok"))
// }

// obtener contexto de ejecucion segun pid-tid
