package main

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/rs/xid"
)

type Task struct {
	ID   string
	Work func() error
}

type Scheduler struct {
	id    string
	queue []*Task
	sem   chan struct{}
}

func NewScheduler(concurrencyCeiling int) (*Scheduler, error) {
	return &Scheduler{
		id:  xid.NewWithTime(time.Now()).String(),
		sem: make(chan struct{}, concurrencyCeiling),
	}, nil
}

var ErrTaskScheduled = errors.New("ErrTaskScheduled: task already scheduled")

func (s *Scheduler) Schedule(id string, work func() error) error {
	for _, task := range s.queue {
		if task.ID == id {
			return ErrTaskScheduled
		}
	}

	s.queue = append(s.queue, &Task{
		ID:   id,
		Work: work,
	})

	return nil
}

func (s *Scheduler) Run(c context.Context) {
	var t *Task
	var wg sync.WaitGroup

	defer func() {
		logln("waiting for running work to finish...")
		wg.Wait()
	}()

	for {
		if len(s.queue) == 0 {
			logln("scheduler queue empty, sleeping...")
		}
		t, s.queue = s.queue[0], s.queue[1:]
		select {
		case s.sem <- struct{}{}:
			wg.Add(1)
			go func() {
				logln(t.ID, "executing...")
				if err := t.Work(); err != nil {
					errln(t.ID, "failed")
				}
				<-s.sem
				wg.Done()
			}()
		case <-c.Done():
			logln(s.id, "stopping scheduler limiter")
			return

		default:
			wg.Add(1)
			logln(t.ID, "executing...")
			if err := t.Work(); err != nil {
				errln(t.ID, "failed")
			}
			wg.Done()
		}
	}

}
