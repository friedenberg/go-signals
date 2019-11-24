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
	Signals []os.Signal
	Logger
	RunFunc      RunFunc
	ShutdownFunc ShutdownFunc
}

func (h *Handler) Run() error {
	if len(h.Signals) == 0 {
		h.Info("No signals set, using defaults of SIGHUP and SIGTERM")
		h.Signals = []os.Signal{syscall.SIGHUP, syscall.SIGTERM}
	}

	wg := sync.WaitGroup{}

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel)

	runComplete := makeBoolMutex()
	shutdownComplete := makeBoolMutex()

	var runErr error
	shutdownErr := makeErrorMutex()

	go func() {
		sigHupOrTerm := <-signalChannel

		//intentionally only increment the group if a signal has been received
		//Otherwise, a race condition exists where the shutdown handler may be
		//terminated prematurely
		wg.Add(1)
		defer wg.Done()

		defer shutdownComplete.setTrue()

		defer func() {
			if r := recover(); r != nil {
				if err, ok := r.(error); ok {
					h.Error("shutdown handler panicked with error:", err)
					shutdownErr.set(err)
				} else {
					h.Error("shutdown handler panicked:", r)
					shutdownErr.set(fmt.Errorf("shutdown handler panicked: %v", r))
				}
			}
		}()

		if runComplete.read() {
			h.Info("run completed, not running shutdown handler")
			return
		}

		h.Info(fmt.Sprintf("Received shutdown signal: %s", sigHupOrTerm))
		h.Info("Calling shutdown handler")
		shutdownErr.set(h.ShutdownFunc(sigHupOrTerm))
		h.Info("Shutdown handler complete")

		if shutdownErr.read() != nil {
			h.Error(errors.Wrap(shutdownErr.read(), "Shutdown handler failed"))
		}
	}()

	h.Info("Calling run handler")
	runErr = h.RunFunc()
	runComplete.setTrue()
	h.Info("Run handler complete")

	if runErr != nil {
		h.Error(errors.Wrap(runErr, "Run handler failed"))
	}

	wg.Wait()

	//only log the run handler failure if shutdown wasn't executed
	//This is because structs like net/http.Server _always_ return errors when
	//methods like ListenAndServe() are called
	if shutdownComplete.read() {
		return shutdownErr.read()
	}

	return runErr
}
