package main

import (
	"github.com/datatogether/task-mgmt/tasks"
)

var TaskRequests = new(Tasks)

type Tasks int

type TasksEnqueueParams struct {
	Type   string
	Params map[string]interface{}
}

func (Tasks) Enqueue(params *TasksEnqueueParams, ok *bool) (err error) {
	if err := EnqueueTask(params.Type, params.Params); err != nil {
		return err
	}
	*ok = true
	return nil
}

type TasksGetParams struct {
	Id string
}

func (Tasks) Get(args *TasksGetParams, res *tasks.Task) (err error) {
	t := &tasks.Task{
		Id: args.Id,
	}
	err = t.Read(store)
	if err != nil {
		return err
	}

	*res = *t
	return nil
}

type TasksListParams struct {
	OrderBy string
	Limit   int
	Offset  int
}

func (Tasks) List(args *TasksListParams, res *[]*tasks.Task) (err error) {
	ts, err := tasks.ReadTasks(store, args.OrderBy, args.Limit, args.Offset)
	if err != nil {
		return err
	}
	*res = ts
	return nil
}
