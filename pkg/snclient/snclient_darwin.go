package snclient

import (
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"

	"pkg/utils"
)

func (snc *Agent) daemonize() {
	snc.RunBackground()
}

func isInteractive() bool {
	o, _ := os.Stdout.Stat()
	// check if attached to terminal.
	return (o.Mode() & os.ModeCharDevice) == os.ModeCharDevice
}

func setupUsrSignalChannel(osSignalUsrChannel chan os.Signal) {
	signal.Notify(osSignalUsrChannel, syscall.SIGUSR1)
	signal.Notify(osSignalUsrChannel, syscall.SIGUSR2)
}

func mainSignalHandler(sig os.Signal, snc *Agent) MainStateType {
	switch sig {
	case syscall.SIGTERM:
		log.Infof("got sigterm, quiting gracefully")

		return ShutdownGraceFully
	case os.Interrupt, syscall.SIGINT:
		log.Infof("got sigint, quitting")

		return Shutdown
	case syscall.SIGHUP:
		log.Infof("got sighup, reloading configuration...")

		return Reload
	case syscall.SIGUSR1:
		log.Errorf("requested thread dump via signal %s", sig)
		utils.LogThreadDump(log)

		return Resume
	case syscall.SIGUSR2:
		if snc.flags.flagMemProfile == "" {
			log.Errorf("requested memory profile, but flag -memprofile missing")

			return (Resume)
		}

		memFile, err := os.Create(snc.flags.flagMemProfile)
		if err != nil {
			log.Errorf("could not create memory profile: %s", err.Error())
		}
		defer memFile.Close()

		runtime.GC()

		if err := pprof.WriteHeapProfile(memFile); err != nil {
			log.Errorf("could not write memory profile: %s", err.Error())
		}

		log.Warnf("memory profile written to: %s", snc.flags.flagMemProfile)

		return (Resume)
	default:
		log.Warnf("Signal not handled: %v", sig)
	}

	return Resume
}

func (snc *Agent) finishUpdate(binPath string) {
	log.Tracef("[update] reexec into new file %s %v", binPath, os.Args[1:])
	err := syscall.Exec(binPath, os.Args, os.Environ()) //nolint:gosec // false positive? There should be no tainted input here
	if err != nil {
		log.Errorf("restart failed: %s", err.Error())
	}
	os.Exit(ExitCodeError)
}