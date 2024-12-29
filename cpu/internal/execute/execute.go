package execute

import (
	"encoding/binary"
	"fmt"
	"net/http"
	"strconv"

	//"encoding/json"
	"io"

	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/cpu/internal/mmu"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
	"github.com/sisoputnfrba/tp-golang/utils/metodosHttp"
)

const (
	CONTINUE       = 0
	RETURN_CONTEXT = 1
)

type ContextoEjecucion struct {
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

type PedidoEscritura struct {
	Datos           []byte `json:"datos"`
	DireccionFisica int    `json:"direccionFisica"`
	Tid             int    `json:"tid"`
}

type DumpThread struct {
	HiloInvocador int `json:"TID"`
	PIDAsociado   int `json:"PID"`
}
type NewProcessRequest struct {
	Archivo    string `json:"path"`
	TamMemoria int    `json:"tamanio"`
	Prioridad  int    `json:"prioridad"`
}

type NewThread struct {
	Archivo   string `json:"path"`
	Prioridad int    `json:"prioridad"`
}

type ThreadExitRequest struct {
	HiloInvocador int `json:"TID"`
	PidAsociado   int `json:"PID"`
}

type ThreadToJoin struct {
	TID     int `json:"TID"` //no termino de entender cuál sería el TID de arriba
	TIDjoin int `json:"TIDPadre"`
}

type Mutex struct {
	Nombre string `json:"nombre"`
}

var resultado = CONTINUE

func Execute(tcb *global.TCB, instruccion *global.Instruccion) int {
	// setearRegistro(instruccion.Parametros[1], 0, tcb)

	switch instruccion.Operacion {
	case "SET":
		set(tcb, instruccion)
		resultado = CONTINUE
	case "READ_MEM":
		resultado = readMem(tcb, instruccion)
		//resultado = CONTINUE
	case "WRITE_MEM":
		resultado = writeMem(tcb, instruccion)
		//resultado = CONTINUE
	case "SUM":
		sum(tcb, instruccion)
		resultado = CONTINUE
	case "SUB":
		sub(tcb, instruccion)
		resultado = CONTINUE
	case "JNZ":
		jnz(tcb, instruccion)
		resultado = CONTINUE
	case "LOG":
		Log(instruccion, tcb)
		resultado = CONTINUE

		//Syscalls
	case "DUMP_MEMORY":
		//actualizarContextoEjecucion(tcb)
		dumpMemory()
		resultado = RETURN_CONTEXT
	case "IO":
		//actualizarContextoEjecucion(tcb)
		IO(instruccion.Parametros[0])
		resultado = RETURN_CONTEXT
	case "PROCESS_CREATE":
		//actualizarContextoEjecucion(tcb)
		processCreate(instruccion)
		resultado = RETURN_CONTEXT
	case "THREAD_CREATE":
		//actualizarContextoEjecucion(tcb)
		threadCreate(instruccion)
		resultado = RETURN_CONTEXT
	case "THREAD_JOIN":
		//actualizarContextoEjecucion(tcb)
		threadJoin(tcb, instruccion.Parametros[0])
		resultado = RETURN_CONTEXT
	case "THREAD_CANCEL":
		//actualizarContextoEjecucion(tcb)
		threadCancel(instruccion.Parametros[0])
		resultado = RETURN_CONTEXT
	case "MUTEX_CREATE":
		//actualizarContextoEjecucion(tcb)
		mutexCreate(instruccion.Parametros[0])
		resultado = RETURN_CONTEXT
	case "MUTEX_LOCK":
		//actualizarContextoEjecucion(tcb)
		mutexLock(instruccion.Parametros[0])
		resultado = RETURN_CONTEXT
	case "MUTEX_UNLOCK":
		//actualizarContextoEjecucion(tcb)
		mutexUnlock(instruccion.Parametros[0])
		resultado = RETURN_CONTEXT
	case "THREAD_EXIT":
		//actualizarContextoEjecucion(tcb)
		ThreadExit(tcb.Contexto.PID, tcb.Contexto.TID)
		resultado = RETURN_CONTEXT
	case "PROCESS_EXIT":
		//actualizarContextoEjecucion(tcb)
		processExit()
		resultado = RETURN_CONTEXT
	default:
		global.Logger.Log(fmt.Sprintln("Instruccion invalida"), log.ERROR)
	}

	global.Logger.Log(fmt.Sprintf("## TID: %d - Ejecutando: %s - %s ", tcb.Contexto.TID, instruccion.Operacion, instruccion.Parametros), log.INFO)
	return resultado
}

func Log(instruccion *global.Instruccion, tcb *global.TCB) {
	// Como LOG solo tiene un parametro, utilizo la posicion 0
	global.Logger.Log(fmt.Sprintf("REGISTRO: %s | VALOR: %d", instruccion.Parametros[0], obtenerRegistro(instruccion.Parametros[0], tcb)), log.INFO)
}

func IO(tiempoEnIo string) {
	cliente := http.Client{}

	url := fmt.Sprintf("http://%s:%d/IO/%s", global.CpuConfig.IpKernel, global.CpuConfig.PuertoKernel, tiempoEnIo)
	req, err := http.NewRequest("PUT", url, nil)

	if err != nil {
		global.Logger.Log("ERROR: "+err.Error(), log.ERROR)
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := cliente.Do(req)

	if err != nil {
		global.Logger.Log("ERROR: "+err.Error(), log.ERROR)
	}

	if resp.StatusCode != http.StatusOK {
		global.Logger.Log("Error en el statuscode", log.ERROR)
	}

}

func mutexCreate(nombreMutex string) {
	nuevoMutex := Mutex{
		Nombre: nombreMutex,
	}

	metodosHttp.PutHTTPwithBody[Mutex, string](
		global.CpuConfig.IpKernel,
		global.CpuConfig.PuertoKernel,
		"MutexCreate",
		nuevoMutex,
	)
	// if err != nil {
	// 	global.Logger.Log("Error al crear mutex: "+err.Error(), log.ERROR)
	// }

	global.Logger.Log(fmt.Sprintf("Se creo el mutex: %s ", nombreMutex), log.DEBUG)
}

func mutexLock(nombreMutex string) {
	mutexLock := Mutex{Nombre: nombreMutex}

	metodosHttp.PutHTTPwithBody[Mutex, string](
		global.CpuConfig.IpKernel,
		global.CpuConfig.PuertoKernel,
		"MutexLock",
		mutexLock,
	)

	// if err != nil {
	// 	global.Logger.Log("Error con la request: "+err.Error(), log.ERROR)
	// }

	global.Logger.Log("Se bloqueo el mutex: "+nombreMutex, log.DEBUG)
}

func mutexUnlock(nombreMutex string) {
	mutexLock := Mutex{Nombre: nombreMutex}

	metodosHttp.PutHTTPwithBody[Mutex, string](
		global.CpuConfig.IpKernel,
		global.CpuConfig.PuertoKernel,
		"MutexUnlock",
		mutexLock,
	)

	// if err != nil {
	// 	global.Logger.Log("Error con la request: "+err.Error(), log.ERROR)
	// }

	global.Logger.Log("Se desbloqueo el mutex: "+nombreMutex, log.DEBUG)
}

func processExit() {

	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/proceso/%d", global.CpuConfig.IpKernel, global.CpuConfig.PuertoKernel, global.TidEjecutando)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	respuesta, err := cliente.Do(req)
	if err != nil {
		global.Logger.Log("Error en la request", log.ERROR)
	}

	// Verificar el código de estado de la respuesta
	if respuesta.StatusCode != http.StatusOK {
		//global.Logger.Log("Error en el statuscode", log.ERROR)
	}
}

func set(tcb *global.TCB, instruccion *global.Instruccion) {
	registro := instruccion.Parametros[0]                 //En instruccion.Parametros[0] tengo el registro
	valor, err := strconv.Atoi(instruccion.Parametros[1]) //Porque en instruccion.Parametros[1] tengo el valor de la instruccion pero en string
	if err != nil {
		panic("Error al convertir la instruccion")
	}
	setearRegistro(registro, uint32(valor), tcb)
}

func readMem(tcb *global.TCB, instruccion *global.Instruccion) int { //READ_MEM AX BX, AX es el registro datos y BX es el registro direccion logica

	//global.Logger.Log("PARTE 4", "DEBUG")
	datos := instruccion.Parametros[0]

	direccion := instruccion.Parametros[1]

	// direccionLogica := mmu.TamRegistro(direccion)
	direccionLogica := obtenerRegistro(direccion, tcb)
	direccionFisica := mmu.TraducirDireccion(&tcb.Contexto, direccionLogica)
	//global.Logger.Log("PARTE 1", "DEBUG")
	if direccionFisica == 1 {
		global.Logger.Log("Segmentation Fault", log.DEBUG)
		//tcb.MotivoInterrupcion = "SEGMENTATION_FAULT"
		return RETURN_CONTEXT //Interrumpe por segmentation fault
	}
	valor := leerDeMemoria(tcb.Contexto.TID, int(direccionFisica))

	global.Logger.Log(fmt.Sprintf("## TID: %d - Accion: Leer - Direccion Fisica: %d", tcb.Contexto.TID, direccionFisica), log.INFO)

	setearRegistro(datos, uint32(binary.BigEndian.Uint32(valor)), tcb)

	return CONTINUE //Si esta todo OK sigue
}

func leerDeMemoria(tid int, direccion int) []byte {

	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/%s", global.CpuConfig.IpMemoria, global.CpuConfig.PuertoMemoria, "lecturaMemoria")

	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		global.Logger.Log("ERROR: "+err.Error(), log.ERROR)
	}

	//global.Logger.Log("A", "DEBUG")

	q := req.URL.Query()
	q.Add("tid", fmt.Sprintf("%d", tid))
	q.Add("direccion", strconv.Itoa(direccion))
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/json")

	respuesta, err := cliente.Do(req)

	//global.Logger.Log("B", "DEBUG")

	if err != nil {
		global.Logger.Log("ERROR: "+err.Error(), log.ERROR)
	}

	global.Logger.Log("C", "DEBUG")

	if respuesta.StatusCode != http.StatusOK {
		global.Logger.Log(fmt.Sprintf("ERROR: Estado de respuesta no OK - Código: %d", respuesta.StatusCode), log.ERROR)
		//global.Logger.Log("ERROR: "+err.Error(), log.ERROR)
	}

	global.Logger.Log("PARTE 3", "DEBUG")
	//bodyBytes := json.Unmarshal(req)
	bodyBytes, err := io.ReadAll(respuesta.Body)
	if err != nil {
		global.Logger.Log("ERROR: "+err.Error(), log.ERROR)
	}

	return bodyBytes
}

func writeMem(tcb *global.TCB, instruccion *global.Instruccion) int {
	direccionLogica := obtenerRegistro(instruccion.Parametros[0], tcb)
	datos := obtenerRegistro(instruccion.Parametros[1], tcb) //WRITE_MEM AX BX

	// var datosEnBytes []byte

	// err := binary.Write(buf, binary.BigEndian, number)

	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, datos)

	direccionFisica := mmu.TraducirDireccion(&tcb.Contexto, direccionLogica)
	if direccionFisica == 1 {
		global.Logger.Log("Segmentation Fault", log.DEBUG)
		//tcb.MotivoInterrupcion = "SEGMENTATION_FAULT"
		return RETURN_CONTEXT
	}
	requestEscritura := PedidoEscritura{
		Datos:           bytes,
		DireccionFisica: int(direccionFisica),
		Tid:             tcb.Contexto.TID,
	}

	metodosHttp.PutHTTPwithBody[PedidoEscritura, string](
		global.CpuConfig.IpMemoria,
		global.CpuConfig.PuertoMemoria,
		"escrituraMemoria",
		requestEscritura,
	)
	// if err != nil {
	// 	panic("Error al escribir en memoria")
	// }
	global.Logger.Log(fmt.Sprintf("## TID: %d - Accion: Escribir - Direccion Fisica: %d", tcb.Contexto.TID, direccionFisica), log.INFO)

	return CONTINUE

}

func processCreate(instruccion *global.Instruccion) {
	convTamMemoria, _ := strconv.Atoi(instruccion.Parametros[1])
	convPrioridad, _ := strconv.Atoi(instruccion.Parametros[2])
	newProcess := NewProcessRequest{
		Archivo:    instruccion.Parametros[0],
		TamMemoria: convTamMemoria,
		Prioridad:  convPrioridad,
	}
	metodosHttp.PutHTTPwithBody[NewProcessRequest, string](
		global.CpuConfig.IpKernel,
		global.CpuConfig.PuertoKernel,
		"crearProceso",
		newProcess,
	)
}

func threadCancel(TID string) {
	cliente := http.Client{}

	url := fmt.Sprintf("http://%s:%d/hilo/%s", global.CpuConfig.IpKernel, global.CpuConfig.PuertoKernel, TID)
	req, err := http.NewRequest("DELETE", url, nil)

	if err != nil {
		global.Logger.Log("ERROR: "+err.Error(), log.ERROR)
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := cliente.Do(req)

	if err != nil {
		global.Logger.Log("ERROR: "+err.Error(), log.ERROR)
	}

	if resp.StatusCode != http.StatusOK {
		global.Logger.Log("Error en el statuscode", log.ERROR)
	}
}

func threadCreate(instruccion *global.Instruccion) {
	convPrioridad, _ := strconv.Atoi(instruccion.Parametros[1])
	newThread := NewThread{
		Archivo:   instruccion.Parametros[0],
		Prioridad: convPrioridad,
	}
	metodosHttp.PutHTTPwithBody[NewThread, string](
		global.CpuConfig.IpKernel,
		global.CpuConfig.PuertoKernel,
		"crearHilo",
		newThread,
	)
}

func threadJoin(tcb *global.TCB, TID string) {
	TIDconv, _ := strconv.Atoi(TID)
	threadToJoin := ThreadToJoin{
		TID:     tcb.Contexto.TID,
		TIDjoin: TIDconv,
	}
	metodosHttp.PutHTTPwithBody[ThreadToJoin, string](
		global.CpuConfig.IpKernel,
		global.CpuConfig.PuertoKernel,
		"unirHilo",
		threadToJoin,
	)
	global.Logger.Log("hasta acva llega", log.DEBUG)
}

func ThreadExit(pid int, tid int) {
	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/hilo", global.CpuConfig.IpKernel, global.CpuConfig.PuertoKernel)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return
	}

	q := req.URL.Query()
	q.Add("PID", strconv.Itoa(pid)) // ! ARREGLAR
	q.Add("TID", strconv.Itoa(tid))
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/json")
	respuesta, err := cliente.Do(req)
	if err != nil {
		return
	}

	// Verificar el código de estado de la respuesta
	if respuesta.StatusCode != http.StatusOK {
		return
	}
}

// TODO: VER QUE FUNCIONE XD
func dumpMemory() {
	dumpMemoryReq := DumpThread{
		PIDAsociado:   global.PIDEjecutando,
		HiloInvocador: global.TidEjecutando,
	}

	metodosHttp.PutHTTPwithBody[DumpThread, string](
		global.CpuConfig.IpKernel,
		global.CpuConfig.PuertoKernel,
		"dumpProceso",
		dumpMemoryReq,
	)
}

func sum(tcb *global.TCB, instruccion *global.Instruccion) { //Por ejemplo SUM AX BX, BX es Origen y AX es Destino
	destino := obtenerRegistro(instruccion.Parametros[0], tcb)
	origen := obtenerRegistro(instruccion.Parametros[1], tcb)
	resultado := origen + destino
	setearRegistro(instruccion.Parametros[0], resultado, tcb)
}

func sub(tcb *global.TCB, instruccion *global.Instruccion) {
	destino := obtenerRegistro(instruccion.Parametros[0], tcb)
	origen := obtenerRegistro(instruccion.Parametros[1], tcb)
	resultado := destino - origen
	setearRegistro(instruccion.Parametros[0], resultado, tcb)
}

func jnz(tcb *global.TCB, instruccion *global.Instruccion) { //Salta a la linea que le llega por parametro si el registro es distinto de 0
	valor := obtenerRegistro(instruccion.Parametros[0], tcb)
	if valor != 0 {
		nuevoPC, _ := strconv.Atoi(instruccion.Parametros[1])
		tcb.Contexto.PC = uint32(nuevoPC)
	}
}

func obtenerRegistro(registro string, tcb *global.TCB) uint32 {
	switch registro {
	case "AX":
		return tcb.Contexto.AX
	case "BX":
		return tcb.Contexto.BX
	case "CX":
		return tcb.Contexto.CX
	case "DX":
		return tcb.Contexto.DX
	case "EX":
		return tcb.Contexto.EX
	case "FX":
		return tcb.Contexto.FX
	case "GX":
		return tcb.Contexto.GX
	case "HX":
		return tcb.Contexto.HX
	default:
		panic("No se pudo obtener el registro")
	}
}

func setearRegistro(registro string, valor uint32, tcb *global.TCB) {
	switch registro {
	case "AX":
		tcb.Contexto.AX = valor
	case "BX":
		tcb.Contexto.BX = valor
	case "CX":
		tcb.Contexto.CX = valor
	case "DX":
		tcb.Contexto.DX = valor
	case "EX":
		tcb.Contexto.EX = valor
	case "FX":
		tcb.Contexto.FX = valor
	case "GX":
		tcb.Contexto.GX = valor
	case "HX":
		tcb.Contexto.HX = valor
	case "PC":
		tcb.Contexto.PC = valor
	}
}

// func actualizarContextoEjecucion(tcb *global.TCB) {
// 	contextoAActualizar := &ContextoEjecucion{
// 		Pid: tcb.Contexto.PID,
// 		Tid: tcb.Contexto.TID,
// 		AX:  tcb.Contexto.AX,
// 		BX:  tcb.Contexto.BX,
// 		CX:  tcb.Contexto.CX,
// 		DX:  tcb.Contexto.DX,
// 		EX:  tcb.Contexto.EX,
// 		FX:  tcb.Contexto.FX,
// 		GX:  tcb.Contexto.GX,
// 		HX:  tcb.Contexto.HX,
// 		PC:  tcb.Contexto.PC,
// 	}
// 	metodosHttp.PutHTTPwithBody[*ContextoEjecucion, string](
// 		global.CpuConfig.IpMemoria,
// 		global.CpuConfig.PuertoMemoria,
// 		"/actualizacionContexto", //Puede cambiar
// 		contextoAActualizar,
// 	)
// 	global.Logger.Log(fmt.Sprintf("## TID: %d - Actualizo Contexto Ejecucion", tcb.Contexto.TID), log.INFO)
// }
