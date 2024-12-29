package execute

import (
	"testing"

	"github.com/sisoputnfrba/tp-golang/cpu/global"
)

func TestSet(t *testing.T) {
	tcb := global.TCB{}
	instruccion := global.Instruccion{
		Operacion:  "SET",
		Parametros: []string{"AX", "100"},
	}

	set(&tcb, &instruccion)

	Expected := 100

	if tcb.Contexto.AX != uint32(Expected) {
		t.Errorf("Expected %d, got %d", Expected, tcb.Contexto.AX)
	}
}

func TestSum(t *testing.T) {
	tcb := global.TCB{
		Contexto: global.Contexto{
			AX: 50,
			BX: 10,
		},
	}
	instruccion := global.Instruccion{
		Operacion:  "SUM",
		Parametros: []string{"AX", "BX"},
	}

	sum(&tcb, &instruccion)

	Expected := 60

	if tcb.Contexto.AX != uint32(Expected) {
		t.Errorf("Expected %d, got %d", Expected, tcb.Contexto.AX)
	}
}

func TestSub(t *testing.T) {
	tcb := global.TCB{
		Contexto: global.Contexto{
			AX: 50,
			BX: 10,
		},
	}
	instruccion := global.Instruccion{
		Operacion:  "SUB",
		Parametros: []string{"AX", "BX"},
	}

	sub(&tcb, &instruccion)

	Expected := 40

	if tcb.Contexto.AX != uint32(Expected) {
		t.Errorf("Expected %d, got %d", Expected, tcb.Contexto.AX)
	}
}

func TestJnz(t *testing.T) {
	tcb := global.TCB{
		Contexto: global.Contexto{
			AX: 2,
			PC: 0,
		},
	}
	instruccion := global.Instruccion{
		Operacion:  "JNZ",
		Parametros: []string{"AX", "10"},
	}

	t.Run("JNZ con registro distinto de 0", func(t *testing.T) {
		jnz(&tcb, &instruccion)

		Expected := 10

		if tcb.Contexto.PC != uint32(Expected) {
			t.Errorf("Expected %d, got %d", Expected, tcb.Contexto.PC)
		}
	})

	tcb.Contexto.AX = 0
	tcb.Contexto.PC = 0
	t.Run("JNZ con registro igual a 0", func(t *testing.T) {
		jnz(&tcb, &instruccion)

		Expected := 0

		if tcb.Contexto.PC != uint32(Expected) {
			t.Errorf("Expected %d, got %d", Expected, tcb.Contexto.PC)
		}
	})
}
