package planificadores

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
	metodos "github.com/sisoputnfrba/tp-golang/utils/metodosHttp"
)

type MotivoInterrupcion struct {
	Motivo string `json:"motivo"`
	Tid    int    `json:"tid"`
}

type HiloAEjecutar struct {
	PID int `json:"pid"`
	TID int `json:"tid"`
}

var mutexFin sync.Mutex

func AsignarColaReady(tcb estructuras.TCB) {
	tcb.Estado = "READY"
	switch global.KernelConfig.SchedulerAlgorithm {
	case "FIFO":
		fifo(tcb)
	case "PRIORIDADES":
		prioridades(tcb)
	case "CMN":
		cmn(tcb)
	}

}

func fifo(tcb estructuras.TCB) {

	global.MutexListo.Lock()

	global.EstadoListo[0] = append(global.EstadoListo[0], tcb)

	global.Logger.Log(fmt.Sprintf("Se encolo por fifo el tid: %d del proceso %d", tcb.TID, tcb.PID), "DEBUG")

	global.MutexListo.Unlock()
}

func prioridades(tcb estructuras.TCB) {
	var i int

	global.MutexListo.Lock()
	for i < len(global.EstadoListo[0]) && global.EstadoListo[0][i].Prioridad <= tcb.Prioridad {
		i++
	}

	global.EstadoListo[0] = append(global.EstadoListo[0], estructuras.TCB{}) // Aumentar el tamaño del slice
	copy(global.EstadoListo[0][i+1:], global.EstadoListo[0][i:])             // Desplazar elementos
	global.EstadoListo[0][i] = tcb

	//global.EstadoListo[0][i] = tcb

	global.MutexListo.Unlock()

}

// Unico que usa mas de una "fila" del map
func cmn(tcb estructuras.TCB) {

	prioridad := tcb.Prioridad

	global.MutexListo.Lock()
	if global.EstadoListo[prioridad] != nil {
		global.EstadoListo[tcb.Prioridad] = append(global.EstadoListo[tcb.Prioridad], tcb)
	} else {
		global.EstadoListo[tcb.Prioridad] = []estructuras.TCB{}
		global.EstadoListo[tcb.Prioridad] = append(global.EstadoListo[tcb.Prioridad], tcb)

		i := 0
		for i < len(global.PrioridadesEnSistema) && prioridad > global.PrioridadesEnSistema[i] {
			i++
		}
		global.PrioridadesEnSistema = append(global.PrioridadesEnSistema, 0)     // Aumentar el tamaño del slice
		copy(global.PrioridadesEnSistema[i+1:], global.PrioridadesEnSistema[i:]) // Desplazar elementos
		global.PrioridadesEnSistema[i] = prioridad
		global.PrioridadesEnSistema = append(global.PrioridadesEnSistema, tcb.Prioridad)
	}
	global.MutexListo.Unlock()
	global.Logger.Log(fmt.Sprintf("Se quiere encolar (cmn) el tid: %d del proceso %d con prioridad %d", tcb.TID, tcb.PID, tcb.Prioridad), "DEBUG")
}

func SeleccionarProximaEjecucion() {
	var i int

	global.HuboInterrupcionQuantum = false

	global.MutexListo.Lock()
	for i < len(global.PrioridadesEnSistema) && len(global.EstadoListo[global.PrioridadesEnSistema[i]]) == 0 {
		i++
	}

	if i >= len(global.PrioridadesEnSistema) {
		global.Logger.Log("No quedan hilos por ejecutar", log.DEBUG)
		global.NinguHiloParaEjecutar = true
		global.MutexListo.Unlock()
		return
	}

	hiloSeleccionado := global.EstadoListo[global.PrioridadesEnSistema[i]][0]

	global.Logger.Log(fmt.Sprintf("Se selecciono para ejecutar el tid: %d del proceso %d, en estado: %s", hiloSeleccionado.TID, hiloSeleccionado.PID, hiloSeleccionado.Estado), "DEBUG")

	global.EstadoListo[global.PrioridadesEnSistema[i]] = global.EstadoListo[global.PrioridadesEnSistema[i]][1:]

	global.MutexListo.Unlock()

	hiloSeleccionado.Estado = "EXEC"
	global.EstadoEjecutando = nil
	global.EstadoEjecutando = append(global.EstadoEjecutando, hiloSeleccionado)

	global.Logger.Log(fmt.Sprintf("El hilo en EstadoEjecutando es: %d del proceso %d", global.EstadoEjecutando[0].TID, global.EstadoEjecutando[0].PID), "DEBUG")

	global.EsperaRutinaSyscall = make(chan struct{})
	if global.KernelConfig.SchedulerAlgorithm != "CMN" {
		Ejecutar(hiloSeleccionado)
	} else {
		EjecutarCMN(hiloSeleccionado)
	}

}

func Ejecutar(unHilo estructuras.TCB) {

	hiloParaCPU := HiloAEjecutar{PID: unHilo.PID, TID: unHilo.TID}

	metodos.PutHTTPwithBody[HiloAEjecutar, HiloAEjecutar](global.KernelConfig.IPCPU, global.KernelConfig.PortCPU, "recibirTIDYPID", hiloParaCPU)
}

func EjecutarCMN(unHilo estructuras.TCB) {

	mutexFin = sync.Mutex{}
	mutexFin.Lock()

	hiloParaCPU := HiloAEjecutar{PID: unHilo.PID, TID: unHilo.TID}

	url := fmt.Sprintf("http://%s:%d/%s", global.KernelConfig.IPCPU, global.KernelConfig.PortCPU, "recibirTIDYPID")
	body, _ := json.Marshal(hiloParaCPU)

	req, _ := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 20 * time.Millisecond}
	go ManejarRR()
	client.Do(req)

	global.Logger.Log("Se solicito a CPU ejecutar", log.DEBUG)
	

	mutexFin.Lock()
}

func ManejarRR() {

	global.HuboSyscall = false
	global.InterrumpioCPU = make(chan struct{})
	global.EnRutinaSyscall = make(chan struct{})
	timer := time.NewTimer(time.Duration(global.KernelConfig.Quantum) * time.Millisecond)
	global.Logger.Log("Se inicio el timer", "DEBUG")
	select {
	case <-timer.C:
		global.HuboInterrupcionQuantum = true

		if global.HuboSyscall {
			<-global.EnRutinaSyscall

		}
		global.Logger.Log(fmt.Sprintf("## (%d:%d) - Desalojado por fin de Quantum", global.EstadoEjecutando[0].PID, global.EstadoEjecutando[0].TID), log.INFO)

		hiloParaCPU := MotivoInterrupcion{Motivo: "QUANTUM", Tid: global.EstadoEjecutando[0].TID}
		mutexFin.Unlock()
		metodos.PutHTTPwithBody[MotivoInterrupcion, MotivoInterrupcion](global.KernelConfig.IPCPU, global.KernelConfig.PortCPU, "interrupcion", hiloParaCPU)
	case <-global.InterrumpioCPU:
		global.Logger.Log("Interrupcion por CPU", "DEBUG")
		if !timer.Stop() {
			<-timer.C
		}
		mutexFin.Unlock()
		return

		// default:
		// 	time.Sleep(time.Millisecond)
	}
}

func SeguirEjecucion() {
	hiloParaCPU := HiloAEjecutar{PID: global.EstadoEjecutando[0].PID, TID: global.EstadoEjecutando[0].TID}
	global.Logger.Log("Sigue porque queda quantum", "DEBUG")

	url := fmt.Sprintf("http://%s:%d/%s", global.KernelConfig.IPCPU, global.KernelConfig.PortCPU, "recibirTIDYPID")
	body, _ := json.Marshal(hiloParaCPU)

	req, _ := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 20 * time.Millisecond}
	client.Do(req)

}
