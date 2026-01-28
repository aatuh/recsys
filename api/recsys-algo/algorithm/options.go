package algorithm

// EngineOption configures optional engine dependencies.
type EngineOption func(*Engine)

// WithClock injects a clock for time-based behavior.
func WithClock(clock Clock) EngineOption {
	return func(e *Engine) {
		if clock != nil {
			e.clock = clock
		}
	}
}

// WithSignalObserver registers an observer for signal availability telemetry.
func WithSignalObserver(observer SignalObserver) EngineOption {
	return func(e *Engine) {
		e.signalObserver = observer
	}
}
