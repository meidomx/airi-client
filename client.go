package airi_client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type EveryType int

const (
	EveryDay    EveryType = 1
	EveryHour   EveryType = 2
	EveryMinute EveryType = 3
	EverySecond EveryType = 4
)

const (
	RespStatusSuccess       = 0
	RespStatusErrParam      = 1
	RespStatusErrGeneral    = 2
	RespStatusListenTimeout = 3
)

type TaskResult struct {
}

type SimpleTaskEvent struct {
	TaskKey     string
	Parameter   string
	TriggerTime int
}

type SimpleTaskCallback func(event SimpleTaskEvent) TaskResult

type CreateSimpleTaskReq struct {
	TaskKey     string
	Description string
	EveryType   EveryType
	At          int
}

type Client interface {
	CreateSimpleTask(req CreateSimpleTaskReq) error
	ListenSimpleTask(taskKey string, callback SimpleTaskCallback) error
}

type clientImpl struct {
	Addr string
}

func (c *clientImpl) CreateSimpleTask(req CreateSimpleTaskReq) error {
	var every string
	switch req.EveryType {
	case EveryDay:
		every = "day"
	case EveryHour:
		every = "hour"
	case EveryMinute:
		every = "minute"
	case EverySecond:
		every = "second"
	default:
		return errors.New("unknown EveryType")
	}

	s := struct {
		TaskKey     string `json:"task_key"`
		Description string `json:"description"`
		EveryType   string `json:"every"`
		At          int    `json:"at"`
	}{
		TaskKey:     req.TaskKey,
		Description: req.Description,
		EveryType:   every,
		At:          req.At,
	}
	d, err := json.Marshal(s)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.Addr+"/api/v1/simple_task", "application/json", bytes.NewReader(d))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println()

	ss := struct {
		Status       int    `json:"status"`
		ErrorMessage string `json:"errormessage"`
	}{}
	err = json.Unmarshal(body, &ss)
	if err != nil {
		return err
	}

	if ss.Status == 0 {
		return nil
	} else {
		return errors.New("error response:" + ss.ErrorMessage)
	}
}

func (c *clientImpl) ListenSimpleTask(taskKey string, callback SimpleTaskCallback) error {
	for {
		ev, err := listenSimpleTaskLoop(c, taskKey)
		if err != nil {
			log.Println("[ERROR] ListenSimpleTask error.", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}
		if ev != nil {
			callback(*ev)
		}
	}
}

func listenSimpleTaskLoop(c *clientImpl, taskKey string) (*SimpleTaskEvent, error) {
	resp, err := http.Get(c.Addr + "/api/v1/simple_task/" + taskKey)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	s := struct {
		Status       int    `json:"status"`
		ErrorMessage string `json:"errormessage"`
		TaskKey      string `json:"task_key"`
		Parameter    string `json:"parameter"`
		TriggerTime  int    `json:"trigger_time"`
	}{}
	err = json.Unmarshal(all, &s)
	if err != nil {
		return nil, err
	}

	if s.Status == RespStatusListenTimeout {
		return nil, nil
	}

	if s.Status != 0 {
		return nil, errors.New("error response:" + s.ErrorMessage)
	} else {
		return &SimpleTaskEvent{
			TaskKey:     s.TaskKey,
			Parameter:   s.Parameter,
			TriggerTime: s.TriggerTime,
		}, nil
	}
}

var _ Client = &clientImpl{}

func NewClient(endpoint string) (*clientImpl, error) {
	return &clientImpl{
		Addr: endpoint,
	}, nil
}
