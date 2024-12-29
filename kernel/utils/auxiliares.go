package utils

import (
	"fmt"

	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
)

/*===== VARIABLES GLOBALES ======*/

func NuevoPCB(prioridadMain int) estructuras.PCB {
	var nuevoProceso estructuras.PCB

	global.MutexPID.Lock()

	nuevoProceso.PID = global.ContadorPid
	global.ContadorPid++

	global.MutexPID.Unlock()

	var hiloMain estructuras.TCB

	hiloMain.PID = nuevoProceso.PID // El TCB tiene que tener asociado el PID del proceso a crear
	hiloMain.TID = 0
	hiloMain.Prioridad = prioridadMain

	if global.EstadoListo[prioridadMain] == nil {

		global.EstadoListo[prioridadMain] = []estructuras.TCB{}

		i := 0
		for i < len(global.PrioridadesEnSistema) && prioridadMain > global.PrioridadesEnSistema[i] {
			i++
		}
		global.PrioridadesEnSistema = append(global.PrioridadesEnSistema, 0)     // Aumentar el tamaño del slice
		copy(global.PrioridadesEnSistema[i+1:], global.PrioridadesEnSistema[i:]) // Desplazar elementos
		global.PrioridadesEnSistema[i] = prioridadMain
		global.PrioridadesEnSistema = append(global.PrioridadesEnSistema, prioridadMain)
	}

	hiloMain.Estado = "NEW"

	nuevoProceso.Threads = append(nuevoProceso.Threads, hiloMain)

	var Mutexs = make(map[string]estructuras.MapMutex)
	Mutexs = map[string]estructuras.MapMutex{}

	nuevoProceso.ListaMutex = Mutexs

	return nuevoProceso
}

func NuevoTCB(Prioridad int) estructuras.TCB {
	var nuevoHilo estructuras.TCB
	pid := PidProcesoEjecutando()

	nuevoHilo.Prioridad = Prioridad
	nuevoHilo.PID = pid
	nuevoHilo.Estado = "READY" // Por ahora, todos los nuevos hilos iran a READY

	nuevoHilo.TID = TIDIncremental(pid, nuevoHilo)

	return nuevoHilo
}

// Devuelve el PID del proceso que esta ejecutando CPU
func PidProcesoEjecutando() int {
	return global.EstadoEjecutando[0].PID
}

func TIDIncremental(pidABuscar int, nuevoHilo estructuras.TCB) int {
	pid := EncontrarPCB(pidABuscar)
	nuevoHilo.TID = len(global.ProcesosEnSistema[pid].Threads)
	global.ProcesosEnSistema[pid].Threads = append(global.ProcesosEnSistema[pid].Threads, nuevoHilo)
	return nuevoHilo.TID
}

// Busca en la lista de Procesos activos determinado PID y devueve su pocision
func EncontrarPCB(pid int) int {
	for i := 0; i < len(global.ProcesosEnSistema); i++ {
		if pid == global.ProcesosEnSistema[i].PID {
			return i
		}
	}

	return (-1)
}

// Busca en todas las colas el tid que corresponde con ese pid. Una vez que encuentra el tcb, lo saca y lo devuelve (actucalizando la cola correspondiente).
func EncontrarTCB(tid int, pid int) estructuras.TCB {

	var tcb estructuras.TCB

	global.MutexListo.Lock()
	for i := 0; i < len(global.PrioridadesEnSistema); i++ {
		for j := 0; j < len(global.EstadoListo[global.PrioridadesEnSistema[i]]); j++ {
			if tid == global.EstadoListo[global.PrioridadesEnSistema[i]][j].TID && pid == global.EstadoListo[global.PrioridadesEnSistema[i]][j].PID {
				tcb = global.EstadoListo[global.PrioridadesEnSistema[i]][j]

				global.EstadoListo[global.PrioridadesEnSistema[i]] = append(global.EstadoListo[global.PrioridadesEnSistema[i]][:j], global.EstadoListo[global.PrioridadesEnSistema[i]][j+1:]...)
				global.MutexListo.Unlock()
				return tcb
			}
		}
	}
	global.MutexListo.Unlock()

	global.MutexBloqueado.Lock()
	for i := 0; i < len(global.EstadoBloqueado); i++ {
		if tid == global.EstadoBloqueado[i].TID && pid == global.EstadoBloqueado[i].PID {
			tcb = global.EstadoBloqueado[i]
			global.EstadoBloqueado = append(global.EstadoBloqueado[:i], global.EstadoBloqueado[i+1:]...)
			global.MutexBloqueado.Unlock()
			return tcb
		}
	}
	global.MutexBloqueado.Unlock()

	global.MutexNuevo.Lock()
	for i := 0; i < len(global.EstadoNuevo); i++ {
		if tid == global.EstadoNuevo[i].TID && pid == global.EstadoNuevo[i].PID {
			tcb = global.EstadoNuevo[i]
			global.EstadoNuevo = append(global.EstadoNuevo[:i], global.EstadoNuevo[i+1:]...)
			global.MutexNuevo.Unlock()
			return tcb
		}
	}
	global.MutexNuevo.Unlock()

	global.MutexSalida.Lock()
	for i := 0; i < len(global.EstadoSalida); i++ {
		if tid == global.EstadoSalida[i].TID && pid == global.EstadoSalida[i].PID {
			tcb = global.EstadoSalida[i]

			global.EstadoSalida = append(global.EstadoSalida[:i], global.EstadoSalida[i+1:]...)
			global.MutexSalida.Unlock()
			return tcb
		}
	}
	global.MutexSalida.Unlock()

	if tid == global.EstadoEjecutando[0].TID && pid == global.EstadoEjecutando[0].PID {
		return global.EstadoEjecutando[0]
	}

	global.Logger.Log((fmt.Sprintf("El proceso %d quiso buscar el hilo %d pero no existe", pid, tid)), log.ERROR)
	return tcb
}

// Busca en todas las colas el tid que corresponde con ese pid. Una vez que encuentra el tcb, devuelve el puntero a él.
func ModificarTCB(tid int, pid int) *estructuras.TCB {

	for i := 0; i < len(global.PrioridadesEnSistema); i++ {
		for j := 0; j < len(global.EstadoListo[global.PrioridadesEnSistema[i]]); j++ {
			if tid == global.EstadoListo[global.PrioridadesEnSistema[i]][j].TID && pid == global.EstadoListo[global.PrioridadesEnSistema[i]][j].PID {
				return &global.EstadoListo[global.PrioridadesEnSistema[i]][j]
			}
		}
	}

	for i := 0; i < len(global.EstadoBloqueado); i++ {
		if tid == global.EstadoBloqueado[i].TID && pid == global.EstadoBloqueado[i].PID {
			return &global.EstadoBloqueado[i]
		}
	}

	for i := 0; i < len(global.EstadoNuevo); i++ {
		if tid == global.EstadoNuevo[i].TID && pid == global.EstadoNuevo[i].PID {
			return &global.EstadoNuevo[i]
		}
	}

	if tid == global.EstadoEjecutando[0].TID && pid == global.EstadoEjecutando[0].PID {
		return &global.EstadoEjecutando[0]
	}

	return nil
}

func AgregarDependencia(tidAUnir int, tidABuscar int, pid int) int {
	tcb := ModificarTCB(tidABuscar, pid)
	if tcb != nil {
		tcb.HilosUnidos = append(tcb.HilosUnidos, tidAUnir)
		global.Logger.Log(fmt.Sprintf("## (%d:%d) - Bloqueado por THREAD_JOIN", tcb.PID, tidAUnir), log.INFO)
		global.ExisteJoin = true
		return 0
	}
	global.ExisteJoin = false
	return -1
}

func HayHiloMayorPrioridad() bool {
	global.MutexListo.Lock()
	defer global.MutexListo.Unlock() // Asegura siempre liberar el mutex

	for i := 0; i < len(global.PrioridadesEnSistema); i++ {
		prioridad := global.PrioridadesEnSistema[i]
		if prioridad < global.EstadoEjecutando[0].Prioridad {
			if len(global.EstadoListo[prioridad]) != 0 {
				//fmt.Print("\nSE ENCONTRO UNO CON MAS PRIORIDAD\n")
				return true
			}
		}
	}
	return false
}
