package utils

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/kernel/planificadores"
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
	"github.com/sisoputnfrba/tp-golang/utils/metodosHttp"
)

type HiloInterrumpido struct {
	TID                int    `json:"tid"`
	MotivoInterrupcion string `json:"motivo"`
}

type ProcesoParaMemoria struct {
	PseudoCodigo string `json:"path"`
	Tamanio      int    `json:"tamanio"`
	PID          int    `json:"pid"`
}

type HiloParaMemoria struct {
	PseudoCodigo string `json:"path"`
	TID          int    `json:"tid"`
	PID          int    `json:"pid"`
}

type ProcesoAEliminarEnMemoria struct {
	PID int `json:"pid"`
}
type HiloAEliminarEnMemoria struct {
	PID int `json:"pid"`
	TID int `json:"tid"`
}

func CrearProceso(pseudoCodigo string, tamanio int, prioridadMain int) bool {

	pcb := NuevoPCB(prioridadMain)
	global.Logger.Log(fmt.Sprintf("## (%d:0) Se crea el Proceso - Estado: NEW", pcb.PID), log.INFO)
	if len(global.InfoEstadoNuevo) > 0 {
		global.Logger.Log(fmt.Sprintf("PID: %d queda en NEW porque hay otros procesos en la cola", pcb.PID), log.INFO)
		global.MutexInfoEstadoNuevo.Lock()

		global.InfoEstadoNuevo = append(global.InfoEstadoNuevo, global.ProcesoParaNew{PseudoCodigo: pseudoCodigo, Tamanio: tamanio, PCB: pcb})
		global.EstadoNuevo = append(global.EstadoNuevo, pcb.Threads[0])
		global.MutexInfoEstadoNuevo.Unlock()
		return true
	}

	switch SolicitarMemoria(tamanio) {
	case http.StatusOK:

		global.ProcesosEnSistema = append(global.ProcesosEnSistema, pcb)
		nuevoProceso := ProcesoParaMemoria{PseudoCodigo: pseudoCodigo, Tamanio: tamanio, PID: pcb.PID}
		PasarPseudocodigoAMemoria(nuevoProceso)
		planificadores.AsignarColaReady(pcb.Threads[0])
		global.Logger.Log(fmt.Sprintf("PID: %d Pudo pasar a ready", pcb.PID), log.INFO)

	case http.StatusLengthRequired:

		global.MutexListo.Lock() // ? Pausar la planificacion ???????
		global.ProcesosEnSistema = append(global.ProcesosEnSistema, pcb)

		global.Logger.Log("Se va a compactar memoria ", log.DEBUG)
		Compactar()

		global.MutexListo.Unlock()
		nuevoProceso := ProcesoParaMemoria{PseudoCodigo: pseudoCodigo, Tamanio: tamanio, PID: pcb.PID}
		PasarPseudocodigoAMemoria(nuevoProceso)
		planificadores.AsignarColaReady(pcb.Threads[0])
		global.Logger.Log(fmt.Sprintf("PID: %d Pudo pasar a ready", pcb.PID), log.INFO)

	case http.StatusBadRequest:
		global.Logger.Log(fmt.Sprintf("PID: %d No hay particiones disponibles para su tamaño", pcb.PID), log.INFO)
		nuevoProceso := global.ProcesoParaNew{PseudoCodigo: pseudoCodigo, Tamanio: tamanio, PCB: pcb}
		global.MutexNuevo.Lock()
		global.InfoEstadoNuevo = append(global.InfoEstadoNuevo, nuevoProceso)
		global.EstadoNuevo = append(global.EstadoNuevo, pcb.Threads[0])
		global.MutexNuevo.Unlock()

	}

	return true
}

func CrearProcesoDeNew(pseudoCodigo string, tamanio int, pcb estructuras.PCB) bool {
	switch SolicitarMemoria(tamanio) {

	case http.StatusOK:
		global.Logger.Log("SE CREA Y NO COMPACTA. DESDE NEW", log.DEBUG)

		global.ProcesosEnSistema = append(global.ProcesosEnSistema, pcb)
		nuevoProceso := ProcesoParaMemoria{PseudoCodigo: pseudoCodigo, Tamanio: tamanio, PID: pcb.PID}
		PasarPseudocodigoAMemoria(nuevoProceso)
		global.InfoEstadoNuevo = global.InfoEstadoNuevo[1:]
		EncontrarTCB(pcb.Threads[0].TID, pcb.PID)
		planificadores.AsignarColaReady(pcb.Threads[0])
		global.Logger.Log(fmt.Sprintf("PID: %d Pudo pasar a ready", pcb.PID), log.INFO)

		return true

	case http.StatusLengthRequired:

		global.Logger.Log("NECESITO COMPACTAR. DESDE NEW", log.DEBUG)

		global.MutexProcesosEnSistema.Lock() // ? Pausar la planificacion ???????
		global.ProcesosEnSistema = append(global.ProcesosEnSistema, pcb)
		Compactar()
		nuevoProceso := ProcesoParaMemoria{PseudoCodigo: pseudoCodigo, Tamanio: tamanio, PID: pcb.PID}
		PasarPseudocodigoAMemoria(nuevoProceso)
		global.InfoEstadoNuevo = global.InfoEstadoNuevo[1:]
		global.Logger.Log("SE HIZO EL REQUEST", log.DEBUG)

		planificadores.AsignarColaReady(pcb.Threads[0])

		global.MutexProcesosEnSistema.Unlock()
		global.Logger.Log(fmt.Sprintf("PID: %d Pudo pasar a ready", pcb.PID), log.INFO)

		return true
	case http.StatusBadRequest:

		global.Logger.Log("El proceso se queda en NEW", log.DEBUG)

		return false
	}

	return false
}

func CrearHilo(pseudoCodigo string, prioridad int) bool {

	tcb := NuevoTCB(prioridad)

	nuevoHilo := HiloParaMemoria{PseudoCodigo: pseudoCodigo, TID: tcb.TID, PID: tcb.PID}

	PasarPseudocodigoAMemoriaHilo(nuevoHilo)
	planificadores.AsignarColaReady(tcb)

	global.Logger.Log(fmt.Sprintf("## (%d:%d) Se crea el Hilo - Estado: READY", tcb.PID, tcb.TID), log.INFO)

	return true
}

func PCBaExit(pid int) {
	SolicitarEliminacionProceso(pid)
	TCBsAExit(pid)
	pos := EncontrarPCB(pid)

	global.ProcesosEnSistema = append(global.ProcesosEnSistema[:pos], global.ProcesosEnSistema[pos+1:]...)

	global.Logger.Log(fmt.Sprintf("## Finaliza el proceso %d", pid), log.INFO)

	VerficarEstadoNew()
}

func VerficarEstadoNew() {
	if len(global.InfoEstadoNuevo) != 0 {

		InfoNuevoProceso := global.InfoEstadoNuevo[0]

		if CrearProcesoDeNew(InfoNuevoProceso.PseudoCodigo, InfoNuevoProceso.Tamanio, InfoNuevoProceso.PCB) {
			VerficarEstadoNew()
		}

	}
}

func TCBsAExit(pid int) {

	pos := EncontrarPCB(pid)

	pcb := global.ProcesosEnSistema[pos]

	for i := len(pcb.Threads) - 1; i >= 0; i-- { // Libero primero el tcb de tid mas grande, ultimo el tcb de tid 0
		tcb := pcb.Threads[i]
		tcb = EncontrarTCB(tcb.TID, pid)
		tcb.Estado = "EXIT"

		global.MutexExit.Lock()
		global.EstadoSalida = append(global.EstadoSalida, tcb)
		global.MutexExit.Unlock()
	}
}

func BloquearHilo(tid int, pid int) {
	tcb := EncontrarTCB(tid, pid)
	tcb.Estado = "BLOCK"

	global.MutexBloqueado.Lock()
	global.EstadoBloqueado = append(global.EstadoBloqueado, tcb)
	global.MutexBloqueado.Unlock()
	global.Logger.Log(fmt.Sprintf("Se bloqueo el hilo %d del proceso %d", tcb.TID, tcb.PID), log.DEBUG)
}


func BlockToReady(tid int, pid int) {
	tcb := EncontrarTCB(tid, pid)
	tcb.Estado = "READY"

	planificadores.AsignarColaReady(tcb)

}

func SolicitarMemoria(tamanio int) int {
	cliente := &http.Client{}

	endpoint := "tamanioProceso/" + strconv.Itoa(tamanio)

	url := fmt.Sprintf("http://%s:%d/%s", global.KernelConfig.IPMemory, global.KernelConfig.PortMemory, endpoint)
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return -1
	}

	req.Header.Set("Content-Type", "application/json")

	respuesta, _ := cliente.Do(req)

	return respuesta.StatusCode
}

func SolicitarMemoryDump(tid int, pid int) int {
	cliente := &http.Client{}

	endpoint := "memoryDump/" + strconv.Itoa(pid) + "/" + strconv.Itoa(tid)

	url := fmt.Sprintf("http://%s:%d/%s", global.KernelConfig.IPMemory, global.KernelConfig.PortMemory, endpoint)
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return -1
	}
	req.Header.Set("Content-Type", "application/json")

	respuesta, _ := cliente.Do(req)

	return respuesta.StatusCode
}

func FinalizarHilo(tid int, pid int) {
	tcb := *ModificarTCB(tid, pid)

	tcb = EncontrarTCB(tcb.TID, pid)
	tcb.Estado = "EXIT"

	global.MutexExit.Lock()
	global.EstadoSalida = append(global.EstadoSalida, tcb)
	global.MutexExit.Unlock()

	LiberarHilosJoin(tcb)
	global.Logger.Log(fmt.Sprintf("## (%d:%d) Finaliza el Hilo", tcb.PID, tcb.TID), log.INFO)
	SolicitarEliminacionHilo(tcb.TID, pid)
}

// Toma el nombre del mutex y lo agrega (sin asignar) al map de mutex del pid pasado por parametro.
func AgregarMutexAProceso(nombre string, pid int) {
	posicionPCB := EncontrarPCB(pid)

	mutexParaAgregar := estructuras.MapMutex{TIDAsignado: estructuras.MutexSinAsignar, TIDsEsperando: nil}

	global.MutexProcesosEnSistema.Lock()
	global.ProcesosEnSistema[posicionPCB].ListaMutex[nombre] = mutexParaAgregar
	global.MutexProcesosEnSistema.Unlock()
}

// Funciona como mutex.Lock(). Si no esta asignado, se lo da al hilo que lo solicito, sino, lo bloquea.
func BloquearMutex(nombre string, pid int, tid int) {
	posicionPCB := EncontrarPCB(pid)

	switch global.ProcesosEnSistema[posicionPCB].ListaMutex[nombre].TIDAsignado {

	// Caso que el recurso no exita (0 es el valor generico) // TODO cambiar numero
	case 0:
		FinalizarHilo(tid, pid)

	// Caso que no este asignado (uso algo como "macro" para no poner -1)
	case estructuras.MutexSinAsignar:
		cambio := global.ProcesosEnSistema[posicionPCB].ListaMutex[nombre]
		cambio.TIDAsignado = tid
		global.MutexProcesosEnSistema.Lock()
		global.ProcesosEnSistema[posicionPCB].ListaMutex[nombre] = cambio
		global.MutexProcesosEnSistema.Unlock()
		global.Logger.Log(fmt.Sprintf("Se bloqueo mutex: %s por el hilo %d por sin asignar", nombre, tid), "DEBUG")

	// El mutex ya esta bloqueado por otro hilo
	default:
		BloquearHilo(tid, pid)
		global.BloqueoPorMutex = true
		cambio := global.ProcesosEnSistema[posicionPCB].ListaMutex[nombre]
		cambio.TIDsEsperando = append(cambio.TIDsEsperando, tid)
		global.MutexProcesosEnSistema.Lock()
		global.ProcesosEnSistema[posicionPCB].ListaMutex[nombre] = cambio
		global.MutexProcesosEnSistema.Unlock()
		global.Logger.Log(fmt.Sprintf("## (%d:%d) Bloqueado por: MUTEX", pid, tid), log.INFO)
	}
}

// Funciona como mutex.Unlock(). Si esta asignado al hilo que corresponde, se libera y asigna al siguiente esperando. Sino, no hace nada.
func DesbloquearMutex(nombre string, pid int, tid int) {
	posicionPCB := EncontrarPCB(pid)

	switch global.ProcesosEnSistema[posicionPCB].ListaMutex[nombre].TIDAsignado {

	// Caso que el recurso no exita (0 es el valor generico)
	case 0:
		if tid != 0 {
			FinalizarHilo(tid, pid)
		}

	// Caso que este asignado a ese tid (lo libera y asigna al siguiente)
	case tid:

		if len(global.ProcesosEnSistema[posicionPCB].ListaMutex[nombre].TIDsEsperando) != 0 {
			cambio := global.ProcesosEnSistema[posicionPCB].ListaMutex[nombre]
			cambio.TIDAsignado = cambio.TIDsEsperando[0]
			cambio.TIDsEsperando = cambio.TIDsEsperando[1:]
			global.ProcesosEnSistema[posicionPCB].ListaMutex[nombre] = cambio
			BlockToReady(cambio.TIDAsignado, pid)
			global.Logger.Log(fmt.Sprintf("Se bloqueo mutex: %s por el hilo %d por desbloqueo", nombre, cambio.TIDAsignado), "DEBUG")
		} else {
			cambio := global.ProcesosEnSistema[posicionPCB].ListaMutex[nombre]
			cambio.TIDAsignado = -1
			global.ProcesosEnSistema[posicionPCB].ListaMutex[nombre] = cambio
		}

	// El mutex existe pero esta bloqueado por otro hilo
	default:
		return
	}
}

/* REQUEST PARA MEMORIA */

func PasarPseudocodigoAMemoria(proceso ProcesoParaMemoria) {
	metodosHttp.PutHTTPwithBody[ProcesoParaMemoria, interface{}](global.KernelConfig.IPMemory, global.KernelConfig.PortMemory, "creacionProceso", proceso)
}

func PasarPseudocodigoAMemoriaHilo(hilo HiloParaMemoria) {
	metodosHttp.PutHTTPwithBody[HiloParaMemoria, interface{}](global.KernelConfig.IPMemory, global.KernelConfig.PortMemory, "creacionHilo", hilo)
}

func Compactar() {
	metodosHttp.PutHTTPwithBody[int, error](global.KernelConfig.IPMemory, global.KernelConfig.PortMemory, "compactacionMemoria", 0)
}

func SolicitarEliminacionProceso(pid int) {
	proceso := ProcesoAEliminarEnMemoria{PID: pid}
	metodosHttp.PutHTTPwithBody[ProcesoAEliminarEnMemoria, error](global.KernelConfig.IPMemory, global.KernelConfig.PortMemory, "eliminacionProceso", proceso)
}

func SolicitarEliminacionHilo(tid int, pid int) {
	hilo := HiloAEliminarEnMemoria{PID: pid, TID: tid}
	metodosHttp.PutHTTPwithBody[HiloAEliminarEnMemoria, error](global.KernelConfig.IPMemory, global.KernelConfig.PortMemory, "eliminacionHilo", hilo)
}

// Gestiona el hilo segun el tipo de interrupcion (de CPU) que haya tenido.
func GestionarInterrupcion(infoHiloInterrumpido HiloInterrumpido, hilo estructuras.TCB) {
	global.Logger.Log(fmt.Sprintf("Se interrumpio el tid: %d por %s", hilo.TID, infoHiloInterrumpido.MotivoInterrupcion), log.DEBUG)

	switch infoHiloInterrumpido.MotivoInterrupcion { // TODO agregar todos los casos.
	// case "QUANTUM":

	// 	planificadores.AsignarColaReady(global.EstadoEjecutando[0])
	// 	planificadores.SeleccionarProximaEjecucion()

	case "THREAD_EXIT":

		// FinalizarHilo(hilo.TID, hilo.PID) ya se hizo con la syscall

		planificadores.SeleccionarProximaEjecucion()

	case "THREAD_JOIN":
		global.Logger.Log("A", log.DEBUG)
		if global.ExisteJoin {
			planificadores.SeleccionarProximaEjecucion()
		} else {
			planificadores.SeguirEjecucion()
		}

	case "PROCESS_EXIT":

		planificadores.SeleccionarProximaEjecucion()

	case "PROCESS_CREATE":

		if HayHiloMayorPrioridad() || global.HuboInterrupcionQuantum {
			planificadores.AsignarColaReady(global.EstadoEjecutando[0])
			planificadores.SeleccionarProximaEjecucion()
		} else {
			// La interrupcion no es bloqueante.
			global.EstadoEjecutando[0] = hilo
			planificadores.SeguirEjecucion()
			// if global.KernelConfig.SchedulerAlgorithm != "CMN" {
			// 	planificadores.Ejecutar(global.EstadoEjecutando[0])
			// } else {
			// 	planificadores.EjecutarCMN(global.EstadoEjecutando[0])

			// }
		}

	case "IO":

		planificadores.SeleccionarProximaEjecucion()

	case "DUMP_MEMORY": //Esta syscall bloqueará al hilo que la invocó hasta que el módulo memoria confirme la finalización de la operación, en caso de error, el proceso se enviará a EXIT. Caso contrario, el hilo se desbloquea normalmente pasando a READY.
		planificadores.SeleccionarProximaEjecucion()

	case "FIN_IO":
		planificadores.AsignarColaReady(global.EstadoEjecutando[0])
		planificadores.SeleccionarProximaEjecucion()

	case "MUTEX_LOCK":
		if global.BloqueoPorMutex {
			planificadores.SeleccionarProximaEjecucion()
		} else if HayHiloMayorPrioridad() || global.HuboInterrupcionQuantum {
			planificadores.AsignarColaReady(global.EstadoEjecutando[0])
			planificadores.SeleccionarProximaEjecucion()
		} else {
			global.EstadoEjecutando[0] = hilo
			planificadores.SeguirEjecucion()
		}
	case "MUTEX_UNLOCK":

		if HayHiloMayorPrioridad() || global.HuboInterrupcionQuantum {
			planificadores.AsignarColaReady(global.EstadoEjecutando[0])
			planificadores.SeleccionarProximaEjecucion()
		} else {
			// La interrupcion no es bloqueante.
			global.EstadoEjecutando[0] = hilo
			planificadores.SeguirEjecucion()
		}

	case "SEGMENTATION_FAULT":
		//FinalizarHilo(global.EstadoEjecutando[0].TID, global.EstadoEjecutando[0].PID)
		global.Logger.Log(fmt.Sprintf("## Proceso: %d eliminado por SEGMENTATION FAULT", global.EstadoEjecutando[0].PID), log.INFO)
		PCBaExit(global.EstadoEjecutando[0].PID)
		planificadores.SeleccionarProximaEjecucion()

	default:
		if global.HuboInterrupcionQuantum {
			planificadores.AsignarColaReady(global.EstadoEjecutando[0])
			planificadores.SeleccionarProximaEjecucion()
		} else {
			// La interrupcion no es bloqueante.
			global.EstadoEjecutando[0] = hilo
			planificadores.SeguirEjecucion()
			
		}

	}

}

func GestionarInterrupcionFIFO(infoHiloInterrumpido HiloInterrumpido, hilo estructuras.TCB) {
	global.Logger.Log(fmt.Sprintf("Se interrumpio el tid: %d por %s", hilo.TID, infoHiloInterrumpido.MotivoInterrupcion), log.DEBUG)

	switch infoHiloInterrumpido.MotivoInterrupcion { // TODO agregar todos los casos.

	case "THREAD_EXIT":

		planificadores.SeleccionarProximaEjecucion()

	case "THREAD_JOIN":

		planificadores.SeleccionarProximaEjecucion()

	case "PROCESS_EXIT":

		planificadores.SeleccionarProximaEjecucion()

	case "IO":

		planificadores.SeleccionarProximaEjecucion()

	case "FIN_IO":

		planificadores.SeleccionarProximaEjecucion()

	case "MUTEX_LOCK":

		planificadores.SeleccionarProximaEjecucion()

	case "DUMP_MEMORY":

		planificadores.SeleccionarProximaEjecucion()

	case "SEGMENTATION_FAULT":
		//FinalizarHilo(global.EstadoEjecutando[0].TID, global.EstadoEjecutando[0].PID)
		global.Logger.Log(fmt.Sprintf("## Proceso: %d eliminado por SEGMENTATION FAULT", global.EstadoEjecutando[0].PID), log.INFO)
		PCBaExit(global.EstadoEjecutando[0].PID)
		planificadores.SeleccionarProximaEjecucion()
	default:

		planificadores.SeguirEjecucion()
	}

}

func LiberarHilosJoin(tcb estructuras.TCB) {
	for i := 0; i < len(tcb.HilosUnidos); i++ {
		BlockToReady(tcb.HilosUnidos[i], tcb.PID)
	}
}

func GestionarIO(unTid int, unPid int, ms int) {

	NuevaReqIO := global.IO{PID: unPid, TID: unTid, MS: ms}
	BloquearHilo(unTid, unPid)
	global.Logger.Log(fmt.Sprintf("## (%d:%d) - Bloqueado por IO", unPid, unTid), log.INFO)
	if len(global.EsperandoIO) > 0 {
		global.MutexColaIO.Lock()
		global.EsperandoIO = append(global.EsperandoIO, NuevaReqIO)
		global.MutexColaIO.Unlock()
	} else {
		global.Logger.Log("Se inicia un nuevo ciclo de IO", log.DEBUG)
		//	global.TimerIO = time.NewTimer(time.Duration(ms) * time.Millisecond)
		global.MutexColaIO.Lock()
		global.EsperandoIO = append(global.EsperandoIO, NuevaReqIO)
		global.MutexColaIO.Unlock()
		global.MutexIO.Unlock()

	}

}
