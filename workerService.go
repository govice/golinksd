package main

import "errors"

type WorkerService struct {
	daemon *daemon
}

func NewWorkerService(daemon *daemon) (*WorkerService, error) {
	ws := &WorkerService{
		daemon: daemon,
	}

	return ws, nil
}

var ErrWorkerIndexOutOfBonds = errors.New("worker index out of bounds")

func (ws *WorkerService) getWorkerByIndex(index int) (*Worker, error) {
	if index < 0 || index > ws.daemon.workerManager.WorkerConfig.Length()-1 {
		return nil, ErrWorkerIndexOutOfBonds
	}

	return ws.daemon.workerManager.WorkerConfig.Workers[index], nil
}

func (ws *WorkerService) deleteWorkerByIndex(index int) error {
	if index < 0 || index > ws.daemon.workerManager.WorkerConfig.Length()-1 {
		return ErrWorkerIndexOutOfBonds
	}
	return ws.daemon.workerManager.removeWorker(index)
}
