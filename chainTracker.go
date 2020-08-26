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
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/govice/golinks/block"
	"github.com/spf13/viper"
)

type ChainTracker struct {
	daemon *daemon
}

func NewChainTracker(daemon *daemon) (*ChainTracker, error) {
	return &ChainTracker{
		daemon: daemon,
	}, nil
}

func (ct *ChainTracker) Execute(ctx context.Context) error {
	logln("starting chain tracker")
	if err := ct.initialize(); err != nil {
		return err
	}
	trackingPeriod := viper.GetInt("tracking_period")
	logln("tracking period:", trackingPeriod)
	syncTicker := time.NewTicker(time.Millisecond * time.Duration(trackingPeriod))
	if err := ct.checkAndSync(); err != nil {
		errln("initial check and sync failed", err)
	}
	for {
		select {
		case <-syncTicker.C:
			if err := ct.checkAndSync(); err != nil {
				errln("check and sync failed", err)
			}
		case <-ctx.Done():
			logln("received termination on chain tracker context")
			return nil
		}
	}
}

func (ct *ChainTracker) initialize() error {
	os.Mkdir(ct.chainDir(), os.ModePerm)
	return nil
}

func (ct *ChainTracker) chainDir() string {
	return filepath.Join(ct.daemon.HomeDir(), "chain")
}

func (ct *ChainTracker) checkAndSync() error {
	ct.daemon.chainMutex.Lock()
	defer ct.daemon.chainMutex.Unlock()
	syncInfo, err := ct.getSyncInfo()
	if err != nil {
		errln("failed to get sync info:", err)
		return err
	}

	logln("Local chain length", syncInfo.LocalLength)
	logln("Remote chain length", syncInfo.RemoteLength)
	if syncInfo.NeedsSync {
		logln("synchronizing local chain with remote")
		if err := ct.synchronize(syncInfo); err != nil {
			errln("failed to synchronize chain", err)
			return err
		}
	} else {
		logln("local chain update to date with remote")
	}

	return nil
}

func (ct *ChainTracker) synchronize(syncInfo *SyncInfo) error {
	blocks, err := ct.requestBlockRange(syncInfo.LocalLength, syncInfo.RemoteLength-1)
	if err != nil {
		errln("failed to get block range:", syncInfo.LocalLength, syncInfo.RemoteLength-1)
		return err
	}

	for _, b := range blocks {
		blockBytes, err := json.Marshal(b)
		if err != nil {
			errln("failed to marshal block", b.Index)
			return err
		}

		fileName := filepath.Join(ct.chainDir(), strconv.Itoa(b.Index)+".json")
		if err := ioutil.WriteFile(fileName, blockBytes, os.ModePerm); err != nil {
			errln("failed to write block file", fileName)
			return err
		}
	}

	return nil
}

func (ct *ChainTracker) localChainFileLength() (int, error) {
	files, err := ioutil.ReadDir(ct.chainDir())
	if err != nil {
		return -1, err
	}

	if len(files) == 0 {
		return 0, nil
	}

	//files should already be sorted alphanumerically
	length := 0
	for index, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			errln("found non-json file in chainDir")
			continue
		}

		if strings.HasPrefix(file.Name(), strconv.Itoa(index)) {
			length++
		} else {
			errln("file name", file.Name(), "does not match expected prefix", strconv.Itoa(index))
		}
	}

	return length, nil
}

func (ct *ChainTracker) getSyncInfo() (*SyncInfo, error) {
	remoteLength, err := ct.daemon.golinksService.GetLength()
	if err != nil {
		errln("failed to get remote length")
		return nil, err
	}

	localLength, err := ct.localChainFileLength()
	if err != nil {
		errln("failed to get local chain length")
		return nil, err
	}

	syncInfo := &SyncInfo{
		RemoteLength: remoteLength,
		LocalLength:  localLength,
		NeedsSync:    false,
	}

	if remoteLength > localLength {
		syncInfo.NeedsSync = true
	}

	return syncInfo, nil
}

func (ct *ChainTracker) requestBlockRange(startIndex, endIndex int) ([]*block.Block, error) {
	var blocks []*block.Block
	for index := startIndex; index <= endIndex; index++ {
		block, err := ct.daemon.golinksService.GetBlock(index)
		if err != nil {
			errln("failed to get block:", index, err)
			return nil, err
		}
		blocks = append(blocks, block)
	}
	return blocks, nil
}

func (ct *ChainTracker) LocalHead() (*block.Block, error) {
	files, err := ioutil.ReadDir(ct.chainDir())
	if err != nil {
		return nil, err
	}

	fileAbs := filepath.Join(ct.chainDir(), files[len(files)-1].Name())

	blockBytes, err := ioutil.ReadFile(fileAbs)
	if err != nil {
		return nil, err
	}

	b := &block.Block{}
	if err := json.Unmarshal(blockBytes, b); err != nil {
		return nil, err
	}

	return b, nil
}

type SyncInfo struct {
	NeedsSync    bool
	LocalLength  int
	RemoteLength int
}
