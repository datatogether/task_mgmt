package tasks

import (
	"github.com/ipfs/go-datastore"
)

// TaskRequests encapsulates all types of requests that can be made
// in relation to tasks
// TODO - should this internal state be moved into the package level
// via package-level setter funcs?
type TaskRequests struct {
	// url to amqp server for enqueuing tasks, only required
	// to fullfill requests, not submit them
	AmqpUrl string
	// Store to read / write tasks to only required
	// to fulfill requests, not submit them
	Store datastore.Datastore
}

type TasksEnqueueParams struct {
	Type   string
	Params map[string]interface{}
}

func (t TaskRequests) Enqueue(params *TasksEnqueueParams, ok *bool) (err error) {
	if err := EnqueueTask(t.AmqpUrl, params.Type, params.Params); err != nil {
		return err
	}
	*ok = true
	return nil
}

type TasksGetParams struct {
	Id string
}

func (t TaskRequests) Get(args *TasksGetParams, res *Task) (err error) {
	tsk := &Task{
		Id: args.Id,
	}
	err = tsk.Read(t.Store)
	if err != nil {
		return err
	}

	*res = *tsk
	return nil
}

type TasksListParams struct {
	OrderBy string
	Limit   int
	Offset  int
}

func (t TaskRequests) List(args *TasksListParams, res *[]*Task) (err error) {
	ts, err := ReadTasks(t.Store, args.OrderBy, args.Limit, args.Offset)
	if err != nil {
		return err
	}
	*res = ts
	return nil
}
