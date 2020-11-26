package daemon

import (
	"context"
	"path/filepath"
	"sync"

	"github.com/govice/golinksd/internal/webserver"
	"github.com/govice/golinksd/pkg/authentication"
	"github.com/govice/golinksd/pkg/blockchain"
	"github.com/govice/golinksd/pkg/chaintracker"
	"github.com/govice/golinksd/pkg/config"
	"github.com/govice/golinksd/pkg/golinks"
	"github.com/govice/golinksd/pkg/log"
	"github.com/govice/golinksd/pkg/worker"
	"github.com/kardianos/service"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
)

type Daemon struct {
	primaryCancel         context.CancelFunc
	cancelFuncs           []context.CancelFunc
	service               service.Service
	logger                service.Logger
	errorGroup            errgroup.Group
	blockchainService     *blockchain.Service
	configService         *config.Service
	golinksService        *golinks.Service
	webserver             *webserver.Webserver
	workerService         *worker.Service
	chainTrackerService   *chaintracker.Service
	authenticationService *authentication.Service

	// chainMutex sync.Mutex
}

func New() (*Daemon, error) {
	// SERVICES
	d := &Daemon{}
	if err := d.initializeServices(); err != nil {
		return nil, err
	}

	if err := d.initializeBackgroundTasks(); err != nil {
		return nil, err
	}

	// DAEMON CONFIG
	serviceConfig := &service.Config{
		Name:        "golinksd",
		DisplayName: "golinksd",
		Description: "golinks daemon",
	}

	s, err := service.New(d, serviceConfig)
	if err != nil {
		return nil, err
	}
	d.service = s

	d.logger, err = s.Logger(nil)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Daemon) Execute() error {
	if err := d.service.Run(); err != nil {
		return err
	}
	return nil
}

func (d *Daemon) Stop(s service.Service) error {
	d.primaryCancel()
	d.errorGroup.Wait()
	return nil
}

func (d *Daemon) Start(s service.Service) error {
	go d.run()
	return nil
}

func (d *Daemon) StopDaemon() error {
	return d.Stop(d.service)
}

func (d *Daemon) run() error {
	primaryContext, primaryCancel := context.WithCancel(context.Background())
	d.primaryCancel = primaryCancel

	if viper.GetBool("development") {
		d.errorGroup.Go(func() error {
			return d.ExecuteFrontend(primaryContext)
		})
	}

	chainTrackerCtx, cancelChainTracker := context.WithCancel(primaryContext)
	d.errorGroup.Go(func() error {
		return d.ExecuteChainTracker(chainTrackerCtx)
	})
	d.cancelFuncs = append(d.cancelFuncs, cancelChainTracker)

	var initialSyncWg sync.WaitGroup
	initialSyncWg.Add(1)

	log.Logln("performing initial chain sync...")
	d.ChainTrackerService().ForceSync(&initialSyncWg)
	initialSyncWg.Wait()

	workerCtx, cancelDaemon := context.WithCancel(primaryContext)

	d.errorGroup.Go(func() error {
		return d.ExecuteWorkerManager(workerCtx)
	})
	d.cancelFuncs = append(d.cancelFuncs, cancelDaemon)

	<-primaryContext.Done()
	return nil
}

func (d *Daemon) initializeServices() error {
	cs, err := config.New()
	if err != nil {
		log.Errln("failed to initialize configuration service")
		return err
	}
	d.configService = cs

	gs, err := golinks.New(d.configService)
	if err != nil {
		log.Errln("failed to iniitalize golinks service")
		return err
	}
	d.golinksService = gs

	bs, err := blockchain.New()
	if err != nil {
		log.Errln("failed to initialize blockchain service")
		return err
	}
	d.blockchainService = bs

	cts, err := chaintracker.New(d)
	if err != nil {
		log.Errln("failed to initialize chain tracker service")
		return err
	}
	d.chainTrackerService = cts

	as, err := authentication.New()
	if err != nil {
		log.Errln("failed to initialize authenticaiton service")
		return err
	}
	d.authenticationService = as

	workerConfigPath := filepath.Join(d.ConfigService().HomeDir(), "workers.json")
	ws, err := worker.New(d, &WorkerConfigManager{Path: workerConfigPath})
	if err != nil {
		log.Errln("failed to initialize worker service")
		return err
	}
	d.workerService = ws

	return nil
}

func (d *Daemon) ExecuteFrontend(ctx context.Context) error {
	return d.webserver.Execute(ctx)
}

func (d *Daemon) ExecuteWorkerManager(ctx context.Context) error {
	return d.workerService.Execute(ctx)
}

func (d *Daemon) ExecuteChainTracker(ctx context.Context) error {
	return d.chainTrackerService.Execute(ctx)
}

func (d *Daemon) initializeBackgroundTasks() error {
	// WORKERS
	webserver, err := webserver.New(d)
	if err != nil {
		log.Errln("failed to initialize webserver")
		return err
	}
	d.webserver = webserver

	return nil
}

func (d *Daemon) AuthenticationService() *authentication.Service {
	return d.authenticationService
}

func (d *Daemon) BlockchainService() *blockchain.Service {
	return d.blockchainService
}

func (d *Daemon) ConfigService() *config.Service {
	return d.configService
}

func (d *Daemon) WorkerService() *worker.Service {
	return d.workerService
}

func (d *Daemon) GolinksService() *golinks.Service {
	return d.golinksService
}

func (d *Daemon) ChainTrackerService() *chaintracker.Service {
	return d.chainTrackerService
}
