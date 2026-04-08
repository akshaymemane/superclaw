package agent

import "sync"

// Status represents the lifecycle phase of an agent run.
type Status int

const (
	StatusIdle    Status = iota
	StatusRunning        // actively calling LLM or executing tools
	StatusDone           // completed successfully
	StatusError          // halted due to error or step limit
)

func (s Status) String() string {
	switch s {
	case StatusIdle:
		return "idle"
	case StatusRunning:
		return "running"
	case StatusDone:
		return "done"
	case StatusError:
		return "error"
	default:
		return "unknown"
	}
}

// State captures observable progress for a single agent run.
// All fields are safe to read after Run returns.
type State struct {
	mu     sync.Mutex
	Steps  int
	Status Status
	Result string // final text response from the model
	Err    error
}

func newState() *State {
	return &State{Status: StatusIdle}
}

func (s *State) incrementStep() {
	s.mu.Lock()
	s.Steps++
	s.mu.Unlock()
}

func (s *State) finish(result string) {
	s.mu.Lock()
	s.Result = result
	s.Status = StatusDone
	s.mu.Unlock()
}

func (s *State) fail(err error) {
	s.mu.Lock()
	s.Err = err
	s.Status = StatusError
	s.mu.Unlock()
}
