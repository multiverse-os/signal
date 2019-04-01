package signals

import (
	"os"
	"os/signal"
	"sync"
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

type Handler struct {
	hooks     map[os.Signal][]func(os.Signal)
	ignored   []os.Signal
	mutex     *sync.Mutex
	channel   chan os.Signal
	listening bool
}

// TODO: Should it be optional to declare the channel handling here so the
// design where one uses signal handling to hold open the application?
func NewHandler() Handler {
	handler := Handler{
		hooks:   map[os.Signal][]func(os.Signal){},
		ignored: []os.Signal{},
		mutex:   &sync.Mutex{},
		channel: make(chan os.Signal, 1),
	}
	go func() {
		for {
			incomingSignal := <-handler.channel
			handler.handle(incomingSignal)
		}
	}()
	return handler
}

func ShutdownHandler(function func(os.Signal)) Handler {
	handler := NewHandler()
	return handler.OnShutdown(function)
}

func (self Handler) handle(incomingSignal os.Signal) {
	functions := self.hooks[incomingSignal]
	for _, function := range functions {
		function(incomingSignal)
	}
}

///////////////////////////////////////////////////////////////////////////////
func (self Handler) Add(function func(os.Signal), signals ...os.Signal) Handler {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	for _, s := range signals {
		self.hooks[s] = append(self.hooks[s], function)
		signal.Notify(self.channel, s)
	}
	return self
}

// NOTE: Shutdown adds the defined hook to all shutdown/exit signals
func (self Handler) OnShutdown(function func(os.Signal)) Handler {
	return self.Add(function, ShutdownSignals...)
}

func (self Handler) OnInterrupt(function func(os.Signal)) Handler {
	return self.Add(function, Interrupt)
}
func (self Handler) OnTerminate(function func(os.Signal)) Handler {
	return self.Add(function, Terminate)
}
func (self Handler) OnQuit(function func(os.Signal)) Handler   { return self.Add(function, Quit) }
func (self Handler) OnHangup(function func(os.Signal)) Handler { return self.Add(function, Hangup) }
func (self Handler) OnKill(function func(os.Signal)) Handler   { return self.Add(function, Kill) }

///////////////////////////////////////////////////////////////////////////////
// TODO: Would be nice to eventually build in functionality to turly ignore
// signals even force kill signals; possibly via uninterruptable sleep or some
// similar technique to make this a truly general use signal handling library
func (self Handler) Ignore(signals ...os.Signal) Handler {
	self.ignored = append(self.ignored, signals...)
	for _, ignoredSignal := range self.ignored {
		self.Remove(ignoredSignal)
		signal.Ignore(self.Signals()...)
	}
	return self
}

// NOTE: Stop ignoring will stop listening too since it requires use of reset
// which resets both Notify and Ignore calls
func (self Handler) StopIgnoring() Handler {
	signal.Reset()
	return self
}

func (self Handler) StopListening() Handler {
	signal.Stop(self.channel)
	return self
}

///////////////////////////////////////////////////////////////////////////////
func (self Handler) Remove(s os.Signal) Handler {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	delete(self.hooks, s)
	return self
}

func (self Handler) Clear() Handler {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	for s := range self.hooks {
		delete(self.hooks, s)
	}
	return self
}

func (self Handler) Reset() Handler {
	self.Clear()
	self.StopListening()
	return self
}

///////////////////////////////////////////////////////////////////////////////
func (self Handler) Signals() []os.Signal {
	signals := make([]os.Signal, len(self.hooks))
	for s, _ := range self.hooks {
		signals = append(signals, s)
	}
	return signals
}
