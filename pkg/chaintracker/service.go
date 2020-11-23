package chaintracker

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/govice/golinks/block"
	"github.com/govice/golinksd/pkg/config"
	"github.com/govice/golinksd/pkg/golinks"
	"github.com/govice/golinksd/pkg/log"
	"github.com/spf13/viper"
)

type Service struct {
	servicer      Servicer
	forceSyncChan chan *sync.WaitGroup
}

type ConfigServicer interface {
	ConfigService() *config.Service
}

type GolinksServicer interface {
	GolinksService() *golinks.Service
}

type Servicer interface {
	ConfigServicer
	GolinksServicer
}

func New(servicer Servicer) (*Service, error) {
	return &Service{
		servicer:      servicer,
		forceSyncChan: make(chan *sync.WaitGroup),
	}, nil
}

func (ct *Service) Execute(ctx context.Context) error {
	log.Logln("starting chain tracker")
	if err := ct.initialize(); err != nil {
		return err
	}
	trackingPeriod := viper.GetInt("tracking_period")
	log.Logln("tracking period:", trackingPeriod)
	syncTicker := time.NewTicker(time.Millisecond * time.Duration(trackingPeriod))
	for {
		select {
		case <-syncTicker.C:
			log.Logln("running periodic sync...")
			if err := ct.checkAndSync(); err != nil {
				log.Errln("check and sync failed", err)
			}
		case wg := <-ct.forceSyncChan:
			log.Logln("received force sync...")
			if err := ct.checkAndSync(); err != nil {
				log.Errln("force sync failed", err)
			}
			wg.Done()
		case <-ctx.Done():
			log.Logln("received termination on chain tracker context")
			return nil
		}
	}
}

func (ct *Service) initialize() error {
	os.Mkdir(ct.chainDir(), os.ModePerm)
	return nil
}

func (ct *Service) chainDir() string {
	return filepath.Join(ct.servicer.ConfigService().HomeDir(), "chain")
}

func (ct *Service) checkAndSync() error {
	syncInfo, err := ct.getSyncInfo()
	if errors.Is(err, ErrChainDesync) {
		if err := ct.clearLocalChain(); err != nil {
			log.Errln("failed to clear local chain", err)
			return err
		}
		syncInfo, err = ct.getSyncInfo()
		if err != nil {
			log.Errln("failed to get sync info:", err)
			return err
		}
	} else if err != nil {
		log.Errln("failed to get sync info:", err)
		return err
	}

	if syncInfo.NeedsSync {
		log.Logf("synchronizing local chain (%d) with remote (%d)\n", syncInfo.LocalLength, syncInfo.RemoteLength)
		if err := ct.synchronize(syncInfo); err != nil {
			log.Errln("failed to synchronize chain", err)
			return err
		}
	}

	return nil
}

func (ct *Service) clearLocalChain() error {
	fis, err := ioutil.ReadDir(ct.chainDir())
	if err != nil {
		return err
	}

	for _, f := range fis {
		fn := path.Join(ct.chainDir(), f.Name())
		if err := os.RemoveAll(fn); err != nil {
			log.Errln("failed to remove file:", fn)
			return nil
		}
	}

	return nil
}

func (ct *Service) synchronize(syncInfo *SyncInfo) error {
	blocks, err := ct.requestBlockRange(syncInfo.LocalLength, syncInfo.RemoteLength-1)
	if err != nil {
		log.Errln("failed to get block range:", syncInfo.LocalLength, syncInfo.RemoteLength-1)
		return err
	}

	for _, b := range blocks {
		blockBytes, err := json.Marshal(b)
		if err != nil {
			log.Errln("failed to marshal block", b.Index)
			return err
		}

		fileName := filepath.Join(ct.chainDir(), strconv.Itoa(b.Index)+".json")
		if err := ioutil.WriteFile(fileName, blockBytes, os.ModePerm); err != nil {
			log.Errln("failed to write block file", fileName)
			return err
		}
	}

	return nil
}

func (ct *Service) localChainFileLength() (int, error) {
	files, err := ct.readChainDir()
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
			log.Errln("found non-json file in chainDir")
			continue
		}

		if strings.HasPrefix(file.Name(), strconv.Itoa(index)) {
			length++
		} else {
			log.Errln("file name", file.Name(), "does not have expected prefix", strconv.Itoa(index))
			return -1, ErrChainDesync
		}
	}

	return length, nil
}

// ErrChainDesync indicates the local chain head is out of sync with a remote block of the same index.
var ErrChainDesync = errors.New("chaintracker: desync in local chain with remote")

func (ct *Service) getSyncInfo() (*SyncInfo, error) {
	remoteLength, err := ct.servicer.GolinksService().GetLength()
	if err != nil {
		log.Errln("failed to get remote length")
		return nil, err
	}

	localLength, err := ct.localChainFileLength()
	if err != nil {
		log.Errln("failed to get local chain length")
		return nil, err
	}

	localHead, err := ct.LocalHead()
	if !errors.Is(err, ErrMissingLocalHead) {
		// enforce the daemon population by a single remote chain
		remoteLocalHead, err := ct.servicer.GolinksService().GetBlock(localHead.Index)
		if err != nil {
			log.Errln("failed to get local head from remote")
			return nil, err
		}

		if !block.Equal(localHead, remoteLocalHead) || (localLength > remoteLength) {
			log.Errln(ErrChainDesync)
			return nil, ErrChainDesync
		}
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

func (ct *Service) requestBlockRange(startIndex, endIndex int) ([]*block.Block, error) {
	var blocks []*block.Block
	for index := startIndex; index <= endIndex; index++ {
		block, err := ct.servicer.GolinksService().GetBlock(index)
		if err != nil {
			log.Errln("failed to get block:", index, err)
			return nil, err
		}
		blocks = append(blocks, block)
	}
	return blocks, nil
}

var ErrMissingLocalHead = errors.New("chaintracker: missing local head")

func (ct *Service) LocalHead() (*block.Block, error) {

	files, err := ct.readChainDir()
	if err != nil {
		log.Errln("failed to read chain directory")
		return nil, err
	}

	if files == nil {
		return nil, ErrMissingLocalHead
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

func (ct *Service) readChainDir() ([]os.FileInfo, error) {
	files, err := ioutil.ReadDir(ct.chainDir())
	if err != nil {
		return nil, err
	}

	var filesOut []os.FileInfo
	// omit any OS generated hidden files/folders
	for _, fi := range files {
		if strings.HasSuffix(fi.Name(), ".json") {
			filesOut = append(filesOut, fi)
		}
	}

	sort.Sort(NumericalFileInfos(filesOut))

	return filesOut, nil
}

type SyncInfo struct {
	NeedsSync    bool
	LocalLength  int
	RemoteLength int
}

type NumericalFileInfos []os.FileInfo

func (nfi NumericalFileInfos) Len() int {
	return len(nfi)
}

func (nfi NumericalFileInfos) Swap(i, j int) {
	nfi[i], nfi[j] = nfi[j], nfi[i]
}

func (nfi NumericalFileInfos) Less(i, j int) bool {
	pathA := nfi[i].Name()
	pathB := nfi[j].Name()

	a, err := strconv.Atoi(pathA[0:strings.LastIndex(pathA, ".")])
	if err != nil {
		return pathA < pathB
	}
	b, err := strconv.Atoi(pathB[0:strings.LastIndex(pathB, ".")])
	if err != nil {
		return pathA < pathB
	}

	return a < b
}

func (ct *Service) ForceSync(wg *sync.WaitGroup) {
	ct.forceSyncChan <- wg
}
