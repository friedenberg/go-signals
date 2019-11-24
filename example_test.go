package signals

import (
	"os"
	"os/exec"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/friedenberg/go-signals"
)

func Example() {
	shouldRun := true

	wg := &sync.WaitGroup{}
	wg.Add(1)

	handler := &signals.Handler{
		Signals: []os.Signal{syscall.SIGALRM},
		Logger:  &signals.StandardLogger{},
		RunFunc: func() error {
			for shouldRun {
				time.Sleep(time.Second)
			}

			return nil
		},
		ShutdownFunc: func(s os.Signal) error {
			shouldRun = false
			return nil
		},
	}

	go func() {
		handler.Run()
		wg.Done()
	}()

	alarmCmd := exec.Command("kill", "-s", "ALRM", strconv.Itoa(syscall.Getpid()))
	alarmCmd.Run()

	wg.Wait()

	//Output:
	//Calling run handler
	//Received shutdown signal: alarm clock
	//Calling shutdown handler
	//Shutdown handler complete
	//Run handler complete
}
