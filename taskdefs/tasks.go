package taskdefs

// Progress represents the current state of a task
type Progress struct {
	Percent float32 `json:"percent"` // percent complete
	Status  string  `json:"status"`  // status string representation
	Done    bool    `json:"done"`    // complete flag
	Error   error   `json:"error"`   // error message
}

func RunTask(task TaskType, opts map[string]interface{}) (progress chan Progress) {

}

type Task interface {
	Run() (progress chan Progress)
}
