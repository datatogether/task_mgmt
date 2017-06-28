package tasks

// Task is the interface for anything that fits on the task queue
type Task interface {
	// are these task params valid? return error if not
	Valid() error
	// Do the task, returning incremental progress
	Do(progress chan Progress)
}

// NewTaskFunc creates new tasks
type NewTaskFunc func() Task

// Progress represents the current state of a task
type Progress struct {
	Percent float32 `json:"percent"` // percent complete
	Step    int     `json:"step"`    // current Step
	Steps   int     `json:"steps"`   // number of Steps in process
	Status  string  `json:"status"`  // status string representation
	Done    bool    `json:"done"`    // complete flag
	Error   error   `json:"error"`   // error message
}
