package estructuras

const MutexSinAsignar int = -1

type PCB struct {
	PID                int
	MotivoInterrupcion string
	Registro           RegistrosCPU
	Instruccion        Instruccion
	Threads            []TCB
	ListaMutex         map[string]MapMutex
}

type TCB struct {
	PID         int // Asocio con proceso
	TID         int
	Prioridad   int // Tomamos el 0 como la prioridad máxima
	Estado      string
	HilosUnidos []int
}

type RegistrosCPU struct {
	AX uint32
	BX uint32
	CX uint32
	DX uint32
	EX uint32
	FX uint32
	GX uint32
	HX uint32
	PC uint32
}

type Instruccion struct {
	Operacion  string
	Parametros []string
}

type Proceso struct {
	PID int    `json:"pid"`
	PC  uint32 `json:"pc"`
}

type MapMutex struct {
	TIDAsignado   int
	TIDsEsperando []int
}
