package cicloInstruccion

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/cpu/internal/execute"
	"github.com/sisoputnfrba/tp-golang/cpu/utils"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
)

func CicloInstruccion(proceso *global.TCB) {
	close(global.ArrancaEjecucion)
	for global.Ejecutando {
		instruccion, _ := Fetch(proceso)
		proceso.Contexto.PC++ //Incremento el PC
		instruccionDecodeada := Decode(instruccion)
		proceso.Instruccion = *instruccionDecodeada
		proceso.Estado = "EJECUTANDO" //No se si es necesario esto
		exec_result := execute.Execute(proceso, &proceso.Instruccion)
		//global.Logger.Log("LLEGA HASTA ACA? 2", "DEBUG")
		if exec_result == execute.RETURN_CONTEXT || global.HuboInterrupcion {
			//global.Logger.Log("LLEGA HASTA ACA?", "DEBUG")
			global.Ejecutando = false
			proceso.Estado = "INTERRUMPIDO"
		}
	}
	InterrumpoPor(proceso)

	if proceso.MotivoInterrupcion != "PROCESS_EXIT" { // TODO: Poner en el switch de kernel la eliminacion de datos en memoria ?
		utils.EnviarContextoAMemoria(&proceso.Contexto)
	}

	global.Logger.Log(fmt.Sprintf("## TID: %d - Actualizo Contexto Ejecucion", global.TidEjecutando), log.INFO)

	global.Logger.Log(fmt.Sprintf("Se interrumpio el tid: %d, por %s", proceso.Contexto.TID, proceso.MotivoInterrupcion), log.DEBUG)

	utils.EnviarInterrupcionAKernel(proceso)
}

func Fetch(proceso *global.TCB) (string, error) {

	global.PCBMutex.Lock()
	pid := proceso.Contexto.PID
	tid := proceso.Contexto.TID
	pc := proceso.Contexto.PC
	global.PCBMutex.Unlock()

	instruccion, err := obtenerInstruccion(pid, tid, pc)
	if err != nil {
		global.Logger.Log("No se pudo obtener la instruccion | ERROR: "+err.Error(), log.ERROR)
	}
	global.Logger.Log(fmt.Sprintf("TID: %d - FETCH - Program Counter: %d", proceso.Contexto.TID, proceso.Contexto.PC), log.INFO)

	global.Logger.Log(fmt.Sprintf("PID: %d TID: %d, Instruccion: %s | PC: %d", proceso.Contexto.PID, proceso.Contexto.TID, instruccion, proceso.Contexto.PC), log.DEBUG)

	return instruccion, err
}

func obtenerInstruccion(Pid int, tid int, pc uint32) (string, error) {
	var RespData string
	url := fmt.Sprintf("http://%s:%d/%s", global.CpuConfig.IpMemoria, global.CpuConfig.PuertoMemoria, "instruccion")

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return RespData, err
	}
	q := req.URL.Query()
	q.Add("pid", fmt.Sprintf("%d", Pid))
	q.Add("tid", fmt.Sprintf("%d", tid))
	q.Add("pc", fmt.Sprintf("%d", int(pc)))
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/json")

	respuesta, err := http.DefaultClient.Do(req)
	if err != nil {
		return RespData, err
	}

	if respuesta.StatusCode != http.StatusOK {
		return RespData, err
	}

	bodyBytes, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return RespData, err
	}

	instruccion := strings.Trim(string(bodyBytes), `"`)

	return instruccion, nil

}

func Decode(instruccion string) *global.Instruccion {
	sliceInstruccion := strings.Fields(instruccion) //Fields separa una cadena por sus espacios en blanco (similar a strtok con un " ")

	instruccionDecodeada := &global.Instruccion{ //Por ejemplo, en [0] tengo un MOV y en [1] y [2] un AX 100
		Operacion:  sliceInstruccion[0],
		Parametros: sliceInstruccion[1:],
	}

	if instruccionDecodeada.Operacion == "" {
		global.Logger.Log("No hay instruccion", log.ERROR)
		return nil
	}

	return instruccionDecodeada
}

func InterrumpoPor(proceso *global.TCB) {

	if proceso.Instruccion.Operacion == "IO" {
		proceso.MotivoInterrupcion = "IO"
	} else if proceso.Instruccion.Operacion == "DUMP_MEMORY" {
		proceso.MotivoInterrupcion = "DUMP_MEMORY"
	} else if proceso.Instruccion.Operacion == "PROCESS_CREATE" {
		proceso.MotivoInterrupcion = "PROCESS_CREATE"
	} else if proceso.Instruccion.Operacion == "THREAD_CREATE" {
		proceso.MotivoInterrupcion = "THREAD_CREATE"
	} else if proceso.Instruccion.Operacion == "THREAD_JOIN" {
		proceso.MotivoInterrupcion = "THREAD_JOIN"
	} else if proceso.Instruccion.Operacion == "THREAD_CANCEL" {
		proceso.MotivoInterrupcion = "THREAD_CANCEL"
	} else if proceso.Instruccion.Operacion == "MUTEX_CREATE" {
		proceso.MotivoInterrupcion = "MUTEX_CREATE"
	} else if proceso.Instruccion.Operacion == "MUTEX_LOCK" {
		proceso.MotivoInterrupcion = "MUTEX_LOCK"
	} else if proceso.Instruccion.Operacion == "MUTEX_UNLOCK" {
		proceso.MotivoInterrupcion = "MUTEX_UNLOCK"
	} else if proceso.Instruccion.Operacion == "THREAD_EXIT" {
		proceso.MotivoInterrupcion = "THREAD_EXIT"
	} else if proceso.Instruccion.Operacion == "PROCESS_EXIT" {
		proceso.MotivoInterrupcion = "PROCESS_EXIT"
	} else if global.MotivoInterrupcion == "SEGMENTATION_FAULT" {
		proceso.MotivoInterrupcion = "SEGMENTATION_FAULT"
	} else if global.MotivoInterrupcion == "FIN_IO" {
		proceso.MotivoInterrupcion = "FIN_IO"
	} else if global.MotivoInterrupcion == "QUANTUM" {
		proceso.MotivoInterrupcion = "QUANTUM"
	}
}
