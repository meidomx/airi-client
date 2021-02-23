package airi_client

import (
	"fmt"
)

func ExampleSimpleTask() {
	c, err := NewClient("http://localhost:6002")
	if err != nil {
		panic(err)
	}

	task := "demo3"
	err = c.CreateSimpleTask(CreateSimpleTaskReq{
		TaskKey:     task,
		Description: "helloworld",
		EveryType:   EveryMinute,
		At:          1,
	})
	if err != nil {
		panic(err)
	}

	err = c.ListenSimpleTask(task, func(event SimpleTaskEvent) TaskResult {
		fmt.Println(event)
		return TaskResult{}
	})
	if err != nil {
		panic(err)
	}
}
