package utils

import (
	"fmt"

	"github.com/sisoputnfrba/tp-golang/kernel/global"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
)

func LogearSyscall(syscall string, pid int, tid int) {
	body := fmt.Sprintf("## (%d:%d) - Solicito syscall: %s", pid, tid, syscall)
	global.Logger.Log(body, log.INFO)
}
