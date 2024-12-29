package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	api "github.com/sisoputnfrba/tp-golang/kernel/api"
	global "github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/kernel/planificadores"
	"github.com/sisoputnfrba/tp-golang/kernel/utils"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
	metodos "github.com/sisoputnfrba/tp-golang/utils/metodosHttp"
)

type MainProceso struct {
	PID int `json:"pid"`
	TID int `json:"tid"`
}

type MotivoInterrupcion struct {
	Motivo string `json:"motivo"`
	Tid    int    `json:"tid"`
}

func main() {

	global.InitGlobal()

	s := api.CrearServer()

	global.Logger.Log(fmt.Sprintf("Iniciando servidor del kernel en puerto: %d", global.KernelConfig.Port), log.INFO)

	go LanzarPrimerProceso()
	go InterfaceDeIO()

	if err := s.Iniciar(); err != nil {
		global.Logger.Log(fmt.Sprintf("Error al iniciar el servidor: %v", err), log.ERROR)
		os.Exit(1)
	}

}

func LanzarPrimerProceso() {
	pseudoCodigo := os.Args[3]

	tamanio, _ := strconv.Atoi(os.Args[4])

	utils.CrearProceso(pseudoCodigo, tamanio, 0)

	planificadores.SeleccionarProximaEjecucion()
}

func InterfaceDeIO() {

	global.MutexIO.Lock() // Se bloquea la primera vez para que entre al for y se quede esperando.
	for {

		global.MutexIO.Lock()

		hiloEjecutando := global.EsperandoIO[0]

		// global.Logger.Log(fmt.Sprintf("UN BLOQUEO POR %d", time.Duration(hiloEjecutando.MS)*time.Millisecond), log.DEBUG)

		time.Sleep(time.Duration(hiloEjecutando.MS) * time.Millisecond)

		utils.BlockToReady(hiloEjecutando.TID, hiloEjecutando.PID)

		body := fmt.Sprintf("## (%d:%d) finalizo IO y pasa a READY", hiloEjecutando.PID, hiloEjecutando.TID)
		global.Logger.Log(body, log.INFO)

		global.MutexColaIO.Lock()
		global.EsperandoIO = global.EsperandoIO[1:]
		global.MutexColaIO.Unlock()

		if utils.ModificarTCB(hiloEjecutando.TID, hiloEjecutando.PID).Prioridad < global.EstadoEjecutando[0].Prioridad {
			hiloParaCPU := MotivoInterrupcion{Motivo: "FIN_IO", Tid: global.EstadoEjecutando[0].TID}
			metodos.PutHTTPwithBody[MotivoInterrupcion, MotivoInterrupcion](global.KernelConfig.IPCPU, global.KernelConfig.PortCPU, "interrupcion", hiloParaCPU)
		}

		if len(global.EsperandoIO) != 0 {
			global.MutexIO.Unlock()
		}

		if global.NinguHiloParaEjecutar {
			global.NinguHiloParaEjecutar = false
			planificadores.SeleccionarProximaEjecucion()
		}
	}

}
