package worker

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	glog "log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/govice/golinks/block"
	"github.com/govice/golinks/blockchain"
	"github.com/govice/golinks/blockmap"
	"github.com/govice/golinksd/pkg/log"
	"github.com/govice/golinksd/pkg/scheduler"
	"github.com/rs/xid"
)

type Worker struct {
	cancelFunc       context.CancelFunc
	RootPath         string   `json:"root_path"`
	GenerationPeriod int      `json:"generation_period"`
	IgnorePaths      []string `json:"ignore_paths"`
	running          bool
	id               string
	logger           *glog.Logger
	servicer         Servicer
}

type NewWorkerConfig struct {
	RootPath         string
	GenerationPeriod int
	IgnorePaths      []string
	WorkerID         string
}

// DefaultLogger satisfies io.WriteCloser interface to supply io.Multiwriter with a
// closer func for cleanup.
type DefaultLogger struct {
	logDirPath string
	id         string
}

func NewDefaultLogger(id, logDir string) *DefaultLogger {
	return &DefaultLogger{
		logDirPath: logDir,
		id:         id,
	}
}

func (wlm *DefaultLogger) Write(p []byte) (n int, err error) {
	logFilePath := filepath.Join(wlm.logDirPath, wlm.id+".log")
	f, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return -1, err
	}

	//TODO do we worry about the OS file descriptor limit? io.WriteCloser?
	// wrap io.Writer to use type referenced file, implement io.Closer to close file
	defer f.Close()
	return f.Write(p)
}

var ErrBadRootPath = errors.New("bad root_path")

func (w *Worker) Execute(ctx context.Context) error {
	// TODO pull this from its own config file(s)
	fi, err := os.Stat(w.RootPath)
	if err != nil || !fi.IsDir() {
		log.Logln(err)
		return ErrBadRootPath
	}

	w.logger.Println("starting worker:", w.RootPath)
	genDuration := time.Duration(w.GenerationPeriod) * time.Millisecond
	generationTicker := time.NewTicker(genDuration)
	schedulerFunc := func() {
		generationTicker.Stop()
		if err := w.servicer.WorkerService().ScheduleWork(w.id, func() error {
			berr := w.generateAndUploadBlockmap()
			log.Logln(w.id, "resetting generation ticker...")
			generationTicker.Reset(genDuration)
			return berr
		}); errors.Is(err, scheduler.ErrTaskScheduled) {
			log.Logln(w.id, "task already scheduled. waiting until next epoch...")
			generationTicker.Reset(genDuration)
		} else if err != nil {
			log.Errln(w.id, err)
			generationTicker.Reset(genDuration)
		}
		w.logger.Println(w.id, "finished scheduled generation epoch")
	}

	w.logger.Println("scheduling startup blockmap")
	schedulerFunc()

	w.logger.Println("scheduling periodic blockmap generation")
	w.logger.Println("generation_period:", w.GenerationPeriod, "ms")

	for {
		select {
		case <-ctx.Done():
			w.logger.Println("received termination on worker context")
			generationTicker.Stop()
			return nil //TODO err canceled?
		case <-generationTicker.C:
			w.logger.Println("generating scheduled blockmap for tick")
			schedulerFunc()
		}
	}
}

func (w *Worker) generateAndUploadBlockmap() error {
	blkmap := blockmap.New(w.RootPath)
	blkmap.AutoIgnore = true
	blkmap.FailOnError = false
	blkmap.IOThrottleSize = 1024 * 100 //100 MB
	blkmap.SetIgnorePaths(w.IgnorePaths)
	var generationErr *blockmap.GenerationError
	if err := blkmap.Generate(); errors.As(err, &generationErr) {
		w.logger.Println(generationErr)
	} else if err != nil {
		w.logger.Println("failed to generate blockmap for", w.RootPath, err)
		return err
	}

	blockmapBytes, err := json.Marshal(blkmap)
	if err != nil {
		w.logger.Println("failed to marshal blockmap")
		return err
	}

	w.logger.Println("sending force sync to chain tracker")
	var wg sync.WaitGroup
	wg.Add(1)
	w.servicer.ChainTrackerService().ForceSync(&wg)
	wg.Wait()

	localHeadBlock, err := w.servicer.ChainTrackerService().LocalHead()
	if err != nil {
		w.logger.Println("failed to get local head block", err)
		return err
	}

	stagedBlock := block.NewSHA512(localHeadBlock.Index+1, blockmapBytes, localHeadBlock.BlockHash)

	subchain := &blockchain.Blockchain{
		Blocks: []block.Block{*localHeadBlock, *stagedBlock},
	}

	if err := subchain.Validate(); err != nil {
		w.logger.Println("failed to validate subchain")
		return err
	}

	if err := w.uploadBlock(stagedBlock); err != nil {
		w.logger.Println("failed to upload staged block")
		return err
	}

	return nil
}

func (w *Worker) uploadBlock(blk *block.Block) error {
	return w.servicer.GolinksService().UploadBlock(blk)
}

func (w *Worker) logln(v ...interface{}) {
	w.logger.Println(v...)
}

func NewWorker(servicer Servicer, config *NewWorkerConfig, logWriterProucer func(id string) io.Writer) (*Worker, error) {
	if config.WorkerID == "" {
		config.WorkerID = xid.NewWithTime(time.Now()).String()
	}
	logWriter := logWriterProucer(config.WorkerID)
	worker := &Worker{
		cancelFunc:       func() {},
		RootPath:         config.RootPath,
		GenerationPeriod: config.GenerationPeriod,
		logger:           glog.New(io.MultiWriter(logWriter, os.Stderr), config.WorkerID+" ", glog.Ltime),
		id:               config.WorkerID,
		IgnorePaths:      config.IgnorePaths,
		servicer:         servicer,
	}

	return worker, nil
}

func (w *Worker) AddCancelFunc(cancelFunc func()) {
	if w.cancelFunc == nil {
		w.cancelFunc = cancelFunc
		return
	}
	oldCancelFunc := w.cancelFunc
	w.cancelFunc = func() {
		oldCancelFunc()
		cancelFunc()
	}
}
