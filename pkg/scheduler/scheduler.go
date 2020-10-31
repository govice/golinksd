package scheduler

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/govice/golinks-daemon/pkg/log"
	"github.com/rs/xid"
)

type Task interface {
	ID() string
	Work() func() error
}

type Scheduler struct {
	id    string
	queue []Task
	sem   chan struct{}
}

func New(concurrencyCeiling int) (*Scheduler, error) {
	return &Scheduler{
		id:  xid.NewWithTime(time.Now()).String(),
		sem: make(chan struct{}, concurrencyCeiling),
	}, nil
}

var ErrTaskScheduled = errors.New("ErrTaskScheduled: task already scheduled")

func (s *Scheduler) Schedule(task Task) error {
	for _, t := range s.queue {
		if t.ID() == task.ID() {
			return ErrTaskScheduled
		}
	}

	s.queue = append(s.queue, task)

	return nil
}

func (s *Scheduler) Run(c context.Context) {
	var wg sync.WaitGroup

	defer func() {
		log.Logln("waiting for running work to finish...")
		wg.Wait()
	}()

	for {
		var t Task
		if len(s.queue) == 0 {
			time.Sleep(time.Second * 1)
			continue
		} else if len(s.queue) > 1 {
			t, s.queue = s.queue[0], s.queue[1:]
		} else {
			t, s.queue = s.queue[0], nil
		}

		select {
		case s.sem <- struct{}{}:
			wg.Add(1)
			go func() {
				log.Logln(t.ID(), "executing...")
				if err := t.Work()(); err != nil {
					log.Errln(t.ID(), "failed")
				}
				<-s.sem
				wg.Done()
			}()
		case <-c.Done():
			log.Logln(s.id, "stopping scheduler limiter")
			return
		}
	}
}
