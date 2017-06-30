package tasks

import (
	"database/sql"
	"github.com/ipfs/go-datastore"
)

// Taskable anything that fits on a task queue
type Taskable interface {
	// are these task params valid? return error if not
	Valid() error
	// Do the task, returning incremental progress updates
	// it's expected that the func will send either
	// p.Done == true or p.Error != nil at least once
	// to signal that the task is either done or errored
	Do(updates chan Progress)
}

// NewTaskFunc creates new tasks
type NewTaskFunc func() Taskable

// Progress represents the current state of a task
// tasks will be given a Progress channel to send updates
type Progress struct {
	Percent float32 `json:"percent"` // percent complete between 0.0 & 1.0
	Step    int     `json:"step"`    // current Step
	Steps   int     `json:"steps"`   // number of Steps in the task
	Status  string  `json:"status"`  // status string that describes what is currently happening
	Done    bool    `json:"done"`    // complete flag
	Error   error   `json:"error"`   // error message
}

// SqlDbTaskable is a task that has a method for assigning
// a datastore to the task
type DatastoreTaskable interface {
	Taskable
	SetDatastore(ds datastore.Datastore)
}

// SqlDbTaskable is a task that has a method for assigning a
// database connection to the task
type SqlDbTaskable interface {
	Taskable
	SetSqlDB(db *sql.DB)
}
