package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/cpu/internal/cicloInstruccion"
	"github.com/sisoputnfrba/tp-golang/cpu/utils"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
	serialization "github.com/sisoputnfrba/tp-golang/utils/serializacion"
)

type KernelRequest struct {
	PID int `json:"pid"`
	TID int `json:"tid"`
}

type RequestContextoEjecucion struct {
	AX    uint32 `json:"ax"`
	BX    uint32 `json:"bx"`
	CX    uint32 `json:"cx"`
	DX    uint32 `json:"dx"`
	EX    uint32 `json:"ex"`
	FX    uint32 `json:"fx"`
	GX    uint32 `json:"gx"`
	HX    uint32 `json:"hx"`
	PC    uint32 `json:"pc"`
	Base  int    `json:"base"`
	Limit int    `json:"limit"`
}

type MotivoInterrupcion struct {
	Motivo string `json:"motivo"`
	Tid    int    `json:"tid"`
}

func Interrupcion(w http.ResponseWriter, r *http.Request) {
	motivo := MotivoInterrupcion{}
	err := serialization.DecodeHTTPBody(r, &motivo)
	if err != nil {
		http.Error(w, "Error al decodificar el body", http.StatusBadRequest)
		global.Logger.Log(fmt.Sprintf("Error al decodificar la interrupcion | Error: %v", err), log.ERROR)
		return
	}
	global.Logger.Log("## Llega interrupcion al puerto Interrupt", log.INFO)

	//global.Logger.Log("VIENE ACA", "DEBUG")

	if global.Ejecutando {
		global.Logger.Log("caso 1", "DEBUG")
		<-global.ArrancaEjecucion
		global.MutexEjecucion.Lock()
		global.Ejecutando = false
		global.MutexEjecucion.Unlock()
	} else {
		global.Logger.Log("caso 2", "DEBUG")
		global.HuboInterrupcion = true
	}

	global.MotivoInterrupcion = motivo.Motivo
	w.WriteHeader(http.StatusNoContent)
	global.Logger.Log(fmt.Sprintf("Interrupcion recibida: %v", motivo.Motivo), log.DEBUG)
}

func PedirContextoAMemoria(kernelRequest *KernelRequest) {
	url := fmt.Sprintf("http://%s:%d/%s", global.CpuConfig.IpMemoria, global.CpuConfig.PuertoMemoria, "contextoDeEjecucion")

	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		global.Logger.Log("ERROR: "+err.Error(), log.ERROR)
	}

	q := req.URL.Query()
	q.Add("pid", fmt.Sprintf("%d", kernelRequest.PID))
	q.Add("tid", fmt.Sprintf("%d", kernelRequest.TID))
	// q.Add("pid", string(kernelRequest.PID))
	// q.Add("tid", string(kernelRequest.TID))
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/json")

	respuesta, err := http.DefaultClient.Do(req)

	if err != nil {
		global.Logger.Log("ERROR: "+err.Error(), log.ERROR)
	}

	if respuesta.StatusCode != http.StatusOK {
		global.Logger.Log("ERROR: "+err.Error(), log.ERROR)
	}

	contextoEjecucion := &RequestContextoEjecucion{}

	error := json.NewDecoder(respuesta.Body).Decode(contextoEjecucion)

	if error != nil {
		global.Logger.Log("ERROR: "+error.Error(), log.ERROR)
	}
	global.PCBMutex.Lock()

	contexto := &global.Contexto{
		PID:   kernelRequest.PID,
		TID:   kernelRequest.TID,
		AX:    contextoEjecucion.AX,
		BX:    contextoEjecucion.BX,
		CX:    contextoEjecucion.CX,
		DX:    contextoEjecucion.DX,
		EX:    contextoEjecucion.EX,
		FX:    contextoEjecucion.FX,
		GX:    contextoEjecucion.GX,
		HX:    contextoEjecucion.HX,
		PC:    contextoEjecucion.PC,
		Base:  contextoEjecucion.Base,
		Limit: contextoEjecucion.Limit,
	}
	global.PCBMutex.Unlock()

	tcb := &global.TCB{
		Contexto: *contexto,
		Estado:   "",
	}
	if global.HuboInterrupcion && !global.EsPrimeraEjecucion {
		// "Devolver contexto a memoria y esperar otro thread de kernel"
		global.Logger.Log(fmt.Sprintf("operacion: %s", tcb.Instruccion.Operacion), log.DEBUG)
		cicloInstruccion.InterrumpoPor(tcb)
		utils.EnviarInterrupcionAKernel(tcb)

	} else {
		global.Ejecutando = true
		global.TidEjecutando = kernelRequest.TID
		global.PIDEjecutando = kernelRequest.PID
		global.Logger.Log(fmt.Sprintf("Ejecutando TID: %d", kernelRequest.TID), log.DEBUG)
		global.EsPrimeraEjecucion = false
		cicloInstruccion.CicloInstruccion(tcb)
	}
}

func ObtenerDeKernel(w http.ResponseWriter, r *http.Request) {
	global.HuboInterrupcion = false
	var kernelRequest *KernelRequest
	err := serialization.DecodeHTTPBody(r, &kernelRequest)
	if err != nil {
		http.Error(w, "Error al decodificar", http.StatusBadRequest)
		global.Logger.Log(fmt.Sprintf("Error al decodificar | Error: %v", err), log.ERROR)
		return
	}
	//serialization.EncodeHTTPResponse(w, kernelRequest, http.StatusOK)
	global.Logger.Log("Llega un TID y PID desde Kernel", log.DEBUG)
	global.Logger.Log(fmt.Sprintf("## TID: %d - Solicito Contexto Ejecucion", kernelRequest.TID), log.INFO)
	global.ArrancaEjecucion = make(chan struct{})
	w.WriteHeader(http.StatusNoContent)
	PedirContextoAMemoria(kernelRequest)
}

// TODO Definir si sirve esto o no
// func ObtenerContextoDeEjecucion(w http.ResponseWriter, r *http.Request) {
// 	var contexto *RequestContextoEjecucion
// 	err := serialization.DecodeHTTPBody(r, &contexto)
// 	if err != nil {
// 		http.Error(w, "Error al decodificar", http.StatusBadRequest)
// 		global.Logger.Log(fmt.Sprintf("Error al decodificar | Error: %v", err), log.ERROR)
// 		return
// 	}
// 	global.Logger.Log(fmt.Sprintf("Llega un contexto de ejecucion %v", contexto), log.DEBUG)
// 	serialization.EncodeHTTPResponse(w, contexto, http.StatusOK)
// }
