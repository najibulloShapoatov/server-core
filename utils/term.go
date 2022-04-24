package utils

import (
	"os"
	"os/signal"
	"syscall"
)

func WaitTermSignal() chan os.Signal {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT)
	return signalChan
}
