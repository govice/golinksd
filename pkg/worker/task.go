package worker

type WorkerTask struct {
	id   string
	work func() error
}

func (wt *WorkerTask) ID() string {
	return wt.id
}

func (wt *WorkerTask) Work() func() error {
	return wt.work
}
