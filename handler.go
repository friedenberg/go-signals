package signals

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/pkg/errors"
)

type RunFunc func() error
type ShutdownFunc func(os.Signal) error

type Handler struct {
	Logger
	RunFunc      RunFunc
	ShutdownFunc ShutdownFunc
}

func (h *Handler) Run() error {
	wg := sync.WaitGroup{}

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGHUP, syscall.SIGTERM)

	shutdownExecuted := false

	var runErr, shutdownErr error

	go func() {
		sigHupOrTerm := <-signalChannel

		//intentionally only increment the group if a signal has been received
		//Otherwise, a race condition exists where the shutdown handler may be
		//terminated prematurely
		wg.Add(1)

		h.Info(fmt.Sprintf("Received shutdown signal: %s", sigHupOrTerm))
		h.Info("Calling shutdown handler")
		shutdownErr = h.ShutdownFunc(sigHupOrTerm)
		shutdownExecuted = true
		h.Info("Shutdown handler complete")

		if shutdownErr != nil {
			h.Error(errors.Wrap(shutdownErr, "Shutdown handler failed"))
		}

		wg.Done()
	}()

	h.Info("Calling run handler")
	runErr = h.RunFunc()
	h.Info("Run handler complete")

	if runErr != nil {
		h.Error(errors.Wrap(runErr, "Run handler failed"))
	}

	wg.Wait()

	//only log the run handler failure if shutdown wasn't executed
	//This is because structs like net/http.Server _always_ return errors when
	//methods like ListenAndServe() are called
	if shutdownExecuted && shutdownErr != nil {
		return shutdownErr
	}

	return runErr
}
