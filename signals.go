package signals

import (
	"os"
	"syscall"
)

// Signals
///////////////////////////////////////////////////////////////////////////////
// REF: runtime/sigtab_linux_generic.go (Go source code)
// These are what is defined below as the shutdown or exit signal group
// SIGHUP: terminal line hangup
// SIGINT: interrupt
// SIGQUIT: quit (core dump expected)
// SIGABRT: abort
// SIGTERM: termination
var ShutdownSignals = []os.Signal{
	syscall.SIGINT,
	syscall.SIGTERM,
	syscall.SIGQUIT,
	syscall.SIGHUP,
	syscall.SIGKILL,
}

var (
	// NOTE: This allows packages using this signal system to have a signal import
	// related to signals and avoid an extra unnecessary syscall import.
	Interrupt os.Signal = syscall.SIGINT
	Terminate os.Signal = syscall.SIGTERM
	Quit      os.Signal = syscall.SIGQUIT
	Hangup    os.Signal = syscall.SIGHUP
	Kill      os.Signal = syscall.SIGKILL
	// Aliasing for a more intuitive API
	SIGINT  = Interrupt
	SIGTERM = Terminate
	SIGQUIT = Quit
	SIGHUP  = Hangup
	SIGKILL = Kill
)
