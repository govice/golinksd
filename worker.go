// Copyright 2020 Kevin Gentile
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/govice/golinks/block"
	"github.com/govice/golinks/blockchain"
	"github.com/govice/golinks/blockmap"
	"github.com/rs/xid"
)

type Worker struct {
	daemon           *daemon
	cancelFunc       context.CancelFunc
	RootPath         string `json:"root_path"`
	GenerationPeriod int    `json:"generation_period"`
	running          bool
	id               string
	logger           *log.Logger
}

var ErrBadRootPath = errors.New("bad root_path")

func (w *Worker) Execute(ctx context.Context) error {
	// TODO pull this from its own config file(s)
	fi, err := os.Stat(w.RootPath)
	if err != nil || !fi.IsDir() {
		logln(err)
		return ErrBadRootPath
	}
	absRootPath, err := filepath.Abs(w.RootPath)
	if err != nil {
		return err
	}

	w.logger.Println("starting worker:", w.RootPath)
	w.logger.Println("generating startup blockmap")
	if err := w.generateAndUploadBlockmap(absRootPath); err != nil {
		w.logger.Println("initial blockmap generation failed", err)
	}

	w.logger.Println("scheduling periodic blockmap generation")
	w.logger.Println("generation_period:", w.GenerationPeriod, "ms")
	generationTicker := time.NewTicker(time.Duration(w.GenerationPeriod) * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			w.logger.Println("received termination on worker context")
			generationTicker.Stop()
			return nil //TODO err canceled?
		case <-generationTicker.C:
			w.logger.Println("generating scheduled blockmap for tick")
			if err := w.generateAndUploadBlockmap(absRootPath); err != nil {
				w.logger.Println("scheduled blockmap generation failed. Retrying...")
				generationTicker.Stop()
				if err := w.generateAndUploadBlockmap(absRootPath); err != nil {
					w.logger.Println("blockmap generation and upload failed")
				}
				generationTicker.Reset(time.Duration(w.GenerationPeriod))
			}
		}
	}
}

func (w *Worker) generateAndUploadBlockmap(rootPath string) error {
	blkmap, err := w.generateBlockmap(rootPath)
	if err != nil {
		w.logger.Println("failed to generate blockmap", err)
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
	w.daemon.chainTracker.forceSyncChan <- &wg
	wg.Wait()

	localHeadBlock, err := w.daemon.chainTracker.LocalHead()
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

func (w *Worker) generateBlockmap(rootPath string) (*blockmap.BlockMap, error) {
	w.logger.Println("generating blockmap for", rootPath)
	blkmap := blockmap.New(rootPath)
	if err := blkmap.Generate(); err != nil {
		w.logger.Println("failed to generate blockmap for", rootPath, err)
		return nil, err
	}

	return blkmap, nil
}

func (w *Worker) uploadBlock(blk *block.Block) error {
	return w.daemon.golinksService.UploadBlock(blk)
}

func (w *Worker) logln(v ...interface{}) {
	w.logger.Println(v...)
}

func NewWorker(daemon *daemon, rootPath string, generationPeriod int) (*Worker, error) {
	workerID := xid.NewWithTime(time.Now()).String()
	workerLogsDir := filepath.Join(daemon.configService.HomeDir(), "logs")
	os.Mkdir(workerLogsDir, os.ModePerm)
	logFilePath := filepath.Join(workerLogsDir, workerID+".log")
	f, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	worker := &Worker{
		daemon:           daemon,
		RootPath:         rootPath,
		GenerationPeriod: generationPeriod,
		logger:           log.New(io.MultiWriter(f, os.Stderr), workerID+" ", log.Ltime),
		id:               workerID,
	}

	worker.AddCancelFunc(func() {
		f.Close()
	})

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
