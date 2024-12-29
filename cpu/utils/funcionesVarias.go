package utils

import (
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	//log "github.com/sisoputnfrba/tp-golang/utils/logger"
	"github.com/sisoputnfrba/tp-golang/utils/metodosHttp"
)

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

type InterrupcionRequest struct {
	TID                int    `json:"tid"`
	MotivoInterrupcion string `json:"motivo"`
}

func EnviarContextoAMemoria(contextoEjecucion *global.Contexto) {

	contextoActualizado := &ContextoActualizado{
		Pid: contextoEjecucion.PID,
		Tid: contextoEjecucion.TID,
		AX:  contextoEjecucion.AX,
		BX:  contextoEjecucion.BX,
		CX:  contextoEjecucion.CX,
		DX:  contextoEjecucion.DX,
		EX:  contextoEjecucion.EX,
		FX:  contextoEjecucion.FX,
		GX:  contextoEjecucion.GX,
		HX:  contextoEjecucion.HX,
		PC:  contextoEjecucion.PC,
	}

	metodosHttp.PutHTTPwithBody[*ContextoActualizado, string](
		global.CpuConfig.IpMemoria,
		global.CpuConfig.PuertoMemoria,
		"actualizacionContexto",
		contextoActualizado,
	)

	// Saco lo de "error al enviar .... porque se comparaba con nil"
}

func EnviarInterrupcionAKernel(proceso *global.TCB) {
	global.EsPrimeraEjecucion = true
	requestInterrupcion := InterrupcionRequest{
		TID:                proceso.Contexto.TID,
		MotivoInterrupcion: proceso.MotivoInterrupcion,
	}

	metodosHttp.PutHTTPwithBody[InterrupcionRequest, string](
		global.CpuConfig.IpKernel,
		global.CpuConfig.PuertoKernel,
		"interrupcion",
		requestInterrupcion,
	)

	// fmt.Print(err)

	// if err != nil {
	// 	global.Logger.Log("Error al enviar la interrupcion a kernel", log.ERROR)
	// }

	// Saco lo de "error al enviar .... porque se comparaba con nil"
}
