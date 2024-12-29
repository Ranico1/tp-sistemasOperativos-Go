package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/kernel/global"
	utilsKernel "github.com/sisoputnfrba/tp-golang/kernel/utils"

	logger "github.com/sisoputnfrba/tp-golang/utils/logger"
	serialization "github.com/sisoputnfrba/tp-golang/utils/serializacion"
)

type NewProcessRequest struct {
	Archivo    string `json:"path"`
	TamMemoria int    `json:"tamanio"`
	Prioridad  int    `json:"prioridad"`
}

type NewThread struct {
	Archivo   string `json:"path"`
	Prioridad int    `json:"prioridad"`
}

type ThreadToJoin struct {
	TID     int `json:"TID"` //no termino de entender cuál sería el TID de arriba
	TIDjoin int `json:"TIDPadre"`
}

type DumpThread struct {
	HiloInvocador int `json:"TID"`
	PIDAsociado   int `json:"PID"`
}

type ThreadExitRequest struct {
	HiloInvocador int `json:"TID"`
	PIDAsociado   int `json:"PID"`
}

type Mutex struct {
	Nombre string `json:"nombre"`
}

func ProcessCreate(w http.ResponseWriter, r *http.Request) {
	GestionarCanal()
	var NewProcess NewProcessRequest

	err := serialization.DecodeHTTPBody(r, &NewProcess)

	if err != nil {
		global.Logger.Log("Error al decodear el body: "+err.Error(), logger.ERROR)
		http.Error(w, "Error al decodear el body", http.StatusBadRequest)
		return
	}

	utilsKernel.LogearSyscall("ProcessCreate", global.EstadoEjecutando[0].PID, global.EstadoEjecutando[0].TID)

	if utilsKernel.CrearProceso(NewProcess.Archivo, NewProcess.TamMemoria, NewProcess.Prioridad) {
		TerminarSyscall()
		w.WriteHeader(http.StatusNoContent)
	} else {
		TerminarSyscall()
		w.WriteHeader(http.StatusBadRequest)
	}

}

func ProcessExit(w http.ResponseWriter, r *http.Request) {
	GestionarCanal()
	VerificarCierreCanal()
	tidABuscar, err := strconv.Atoi(r.PathValue("TID"))
	pid := global.EstadoEjecutando[0].PID

	if err != nil {
		global.Logger.Log("Error al decodear el body: "+err.Error(), logger.ERROR)
		http.Error(w, "Error al decodear el body", http.StatusBadRequest)
		return
	}

	procesoAEliminar := utilsKernel.EncontrarTCB(tidABuscar, pid)

	utilsKernel.LogearSyscall("ProcessExit", procesoAEliminar.PID, procesoAEliminar.TID)

	if procesoAEliminar.TID != 0 {
		// no es el hilo principal (no puedo matar el proceso)
		return
	}

	utilsKernel.PCBaExit(procesoAEliminar.PID) // Envia todos los hilos del proceso a EXIT

	w.WriteHeader(http.StatusNoContent)
	TerminarSyscall()
}

func ThreadCreate(w http.ResponseWriter, r *http.Request) {
	GestionarCanal()
	var Thread NewThread

	err := serialization.DecodeHTTPBody(r, &Thread)

	if err != nil {
		global.Logger.Log("Error al decodear el body: "+err.Error(), logger.ERROR)
		http.Error(w, "Error al decodear el body", http.StatusBadRequest)
		return
	}
	utilsKernel.LogearSyscall("ThreadCreate", global.EstadoEjecutando[0].PID, global.EstadoEjecutando[0].TID)

	utilsKernel.CrearHilo(Thread.Archivo, Thread.Prioridad)

	TerminarSyscall()
	w.WriteHeader(http.StatusNoContent)
}

func ThreadExit(w http.ResponseWriter, r *http.Request) {
	GestionarCanal()
	VerificarCierreCanal()
	var hiloAFinalizar ThreadExitRequest
	//pid := global.EstadoEjecutando[0].PID

	queryParams := r.URL.Query()

	var err error

	hiloAFinalizar.PIDAsociado, _ = strconv.Atoi(queryParams.Get("PID"))
	hiloAFinalizar.HiloInvocador, err = strconv.Atoi(queryParams.Get("TID"))

	if err != nil {
		global.Logger.Log("Error al decodear el body: "+err.Error(), logger.ERROR)
		http.Error(w, "Error al decodear el body", http.StatusBadRequest)
		return
	}

	global.Logger.Log(fmt.Sprintf("Se solicito para eliminar el tid: %d del proceso %d", hiloAFinalizar.HiloInvocador, hiloAFinalizar.PIDAsociado), "DEBUG")

	utilsKernel.LogearSyscall("ThreadEXIT", hiloAFinalizar.PIDAsociado, hiloAFinalizar.HiloInvocador)

	utilsKernel.FinalizarHilo(hiloAFinalizar.HiloInvocador, hiloAFinalizar.PIDAsociado)

	TerminarSyscall()
	w.WriteHeader(http.StatusOK)

}

func ThreadJoin(w http.ResponseWriter, r *http.Request) {
	var join ThreadToJoin
	GestionarCanal()
	pid := global.EstadoEjecutando[0].PID

	err := serialization.DecodeHTTPBody(r, &join)

	if err != nil {
		global.Logger.Log("Error al decodear el body: "+err.Error(), logger.ERROR)
		http.Error(w, "Error al decodear el body", http.StatusBadRequest)
		return
	}

	tcb := utilsKernel.EncontrarTCB(join.TID, pid)
	utilsKernel.LogearSyscall("ThreadJoin", tcb.PID, tcb.TID)

	err1 := utilsKernel.AgregarDependencia(join.TID, join.TIDjoin, pid) // Agrega al TIDjoin la informacion de TID

	if err1 == 0 {
		global.EstadoBloqueado = append(global.EstadoBloqueado, tcb)
	}

	TerminarSyscall()

	global.Logger.Log("HOLA=?", logger.DEBUG)
	w.WriteHeader(http.StatusNoContent)
}

func ThreadCancel(w http.ResponseWriter, r *http.Request) {
	GestionarCanal()
	VerificarCierreCanal()

	tidAFinalizar, err := strconv.Atoi(r.PathValue("TID"))
	pid := global.EstadoEjecutando[0].PID

	if err != nil {
		global.Logger.Log("Error al decodear el body: "+err.Error(), logger.ERROR)
		http.Error(w, "Error al decodear el body", http.StatusBadRequest)
		return
	}

	utilsKernel.LogearSyscall("ThreadCancel", global.EstadoEjecutando[0].PID, global.EstadoEjecutando[0].TID)

	if utilsKernel.ModificarTCB(tidAFinalizar, pid) != nil {
		utilsKernel.FinalizarHilo(tidAFinalizar, pid)
	}

	TerminarSyscall()
	w.WriteHeader(http.StatusOK)
}

func DumpMemory(w http.ResponseWriter, r *http.Request) {
	var HiloDump DumpThread
	GestionarCanal()
	VerificarCierreCanal()
	err := serialization.DecodeHTTPBody(r, &HiloDump)
	if err != nil {
		global.Logger.Log("Error al decodear el body: "+err.Error(), logger.ERROR)
		http.Error(w, "Error al decodear el body", http.StatusBadRequest)
		return
	}

	utilsKernel.LogearSyscall("MemoryDump", HiloDump.PIDAsociado, HiloDump.HiloInvocador)
	utilsKernel.BloquearHilo(HiloDump.HiloInvocador, HiloDump.PIDAsociado)
	TerminarSyscall()
	w.WriteHeader(http.StatusOK)

	if utilsKernel.SolicitarMemoryDump(HiloDump.HiloInvocador, HiloDump.PIDAsociado) == http.StatusOK {
		utilsKernel.BlockToReady(HiloDump.HiloInvocador, HiloDump.PIDAsociado)
		global.Logger.Log("DESBLOQUEADO POR FIN DUMP: ", logger.DEBUG)
	} else {
		utilsKernel.PCBaExit(HiloDump.PIDAsociado)
		global.Logger.Log("FINIQUITEADO POR ERROR DUMP: ", logger.DEBUG)
	}

}

func IO(w http.ResponseWriter, r *http.Request) {
	GestionarCanal()
	VerificarCierreCanal()
	hiloEjecutando := global.EstadoEjecutando[0]

	utilsKernel.LogearSyscall("IO", hiloEjecutando.PID, hiloEjecutando.TID)

	milisegundos := r.PathValue("milisegundos") // milisegundos lo lee como String

	ms, _ := strconv.Atoi(milisegundos) // ms son los milisegundos convertidos a Int

	utilsKernel.GestionarIO(hiloEjecutando.TID, hiloEjecutando.PID, ms)

	TerminarSyscall()
	w.WriteHeader(http.StatusOK)

}

func MutexCreate(w http.ResponseWriter, r *http.Request) {
	var NuevoMutex Mutex
	GestionarCanal()
	hiloEjecutando := global.EstadoEjecutando[0]

	err := serialization.DecodeHTTPBody(r, &NuevoMutex)

	if err != nil {
		global.Logger.Log("Error al decodear el body: "+err.Error(), logger.ERROR)
		http.Error(w, "Error al decodear el body", http.StatusBadRequest)
		return
	}
	utilsKernel.LogearSyscall("Mutex_Create", hiloEjecutando.PID, hiloEjecutando.TID)

	utilsKernel.AgregarMutexAProceso(NuevoMutex.Nombre, hiloEjecutando.PID)

	TerminarSyscall()
	w.WriteHeader(http.StatusOK)

}

func MutexLock(w http.ResponseWriter, r *http.Request) {
	var MutexABloquear Mutex
	GestionarCanal()
	hiloEjecutando := global.EstadoEjecutando[0]

	err := serialization.DecodeHTTPBody(r, &MutexABloquear)

	if err != nil {
		global.Logger.Log("Error al decodear el body: "+err.Error(), logger.ERROR)
		http.Error(w, "Error al decodear el body", http.StatusBadRequest)
		return
	}
	utilsKernel.LogearSyscall("Mutex_Lock", hiloEjecutando.PID, hiloEjecutando.TID)

	utilsKernel.BloquearMutex(MutexABloquear.Nombre, hiloEjecutando.PID, hiloEjecutando.TID)

	TerminarSyscall()
	w.WriteHeader(http.StatusOK)
}

func MutexUnlock(w http.ResponseWriter, r *http.Request) {
	var MutexADesbloquear Mutex
	GestionarCanal()
	hiloEjecutando := global.EstadoEjecutando[0]

	err := serialization.DecodeHTTPBody(r, &MutexADesbloquear)

	if err != nil {
		global.Logger.Log("Error al decodear el body: "+err.Error(), logger.ERROR)
		http.Error(w, "Error al decodear el body", http.StatusBadRequest)
		return
	}
	utilsKernel.LogearSyscall("Mutex_Unlock", hiloEjecutando.PID, hiloEjecutando.TID)

	utilsKernel.DesbloquearMutex(MutexADesbloquear.Nombre, hiloEjecutando.PID, hiloEjecutando.TID)

	TerminarSyscall()
	w.WriteHeader(http.StatusOK)
}

func Interrupcion(w http.ResponseWriter, r *http.Request) {
	var hilo utilsKernel.HiloInterrumpido
	err := serialization.DecodeHTTPBody(r, &hilo)

	if err != nil {
		global.Logger.Log("Error al decodear el body: "+err.Error(), logger.ERROR)
		http.Error(w, "Error al decodear el body", http.StatusBadRequest)
		return
	}

	hiloEjecutando := global.EstadoEjecutando[0]

	if global.KernelConfig.SchedulerAlgorithm == "FIFO" {
		utilsKernel.GestionarInterrupcionFIFO(hilo, hiloEjecutando)
	} else {
		utilsKernel.GestionarInterrupcion(hilo, hiloEjecutando)
	}
	w.WriteHeader(http.StatusNoContent)
}

func GestionarCanal() {
	global.HuboSyscall = true
}

func TerminarSyscall() {
	if global.HuboInterrupcionQuantum {
		close(global.EnRutinaSyscall)
	}
	global.HuboSyscall = false
}

func VerificarCierreCanal() {
	if global.KernelConfig.SchedulerAlgorithm == "CMN" {
		close(global.InterrumpioCPU)
	}
}
