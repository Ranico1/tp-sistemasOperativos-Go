package global

type Contexto struct {
	PID   int
	TID   int
	AX    uint32
	BX    uint32
	CX    uint32
	DX    uint32
	EX    uint32
	FX    uint32
	GX    uint32
	HX    uint32
	PC    uint32
	Base  int
	Limit int
}

type TCB struct {
	Contexto           Contexto
	Estado             string
	MotivoInterrupcion string
	HilosUnidos        []int
	Instruccion        Instruccion
}

type Instruccion struct {
	Operacion  string
	Parametros []string
}
