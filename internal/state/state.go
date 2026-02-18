package state

type State string

const (
	Queued            State = "queued"
	Processing        State = "processing"
	Done              State = "done"
	Retrying          State = "retrying"
	DLQ               State = "dlq"
	SagaRunning       State = "saga_running"
	SagaStepFailed    State = "saga_step_failed"
	SagaCompensating  State = "saga_compensating"
	SagaCompensated   State = "saga_compensated"
)

var allStates = []State{
	Queued,
	Processing,
	Done,
	Retrying,
	DLQ,
	SagaRunning,
	SagaStepFailed,
	SagaCompensating,
	SagaCompensated,
}

var transitions = map[State]map[State]bool{
	Queued: {
		Processing: true,
	},
	Processing: {
		Done:     true,
		Retrying: true,
		DLQ:      true,
	},
	Retrying: {
		Queued: true,
	},
	SagaRunning: {
		SagaStepFailed: true,
	},
	SagaStepFailed: {
		Retrying: true,
	},
	SagaCompensating: {
		SagaCompensated: true,
	},
	SagaCompensated: {
		DLQ: true,
	},
}

func AllStates() []State {
	out := make([]State, len(allStates))
	copy(out, allStates)
	return out
}

func CanTransition(from, to State) bool {
	next, ok := transitions[from]
	if !ok {
		return false
	}
	return next[to]
}

func IsTerminal(s State) bool {
	switch s {
	case Done, DLQ:
		return true
	default:
		return false
	}
}
