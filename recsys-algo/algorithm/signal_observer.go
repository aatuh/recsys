package algorithm

// SignalOutcome captures whether a signal fetch succeeded.
type SignalOutcome string

const (
	SignalOutcomeSuccess     SignalOutcome = "success"
	SignalOutcomeUnavailable SignalOutcome = "unavailable"
	SignalOutcomeError       SignalOutcome = "error"
)

// SignalObserver records signal availability outcomes for observability.
type SignalObserver interface {
	RecordSignal(signal Signal, outcome SignalOutcome)
}
