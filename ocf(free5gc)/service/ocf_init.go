package service

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gin-contrib/cors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"free5gc/lib/http2_util"
	"free5gc/lib/logger_util"
	"free5gc/lib/openapi/models"
	"free5gc/lib/path_util"
	"free5gc/src/app"
	"free5gc/src/ocf/factory"
	"free5gc/src/ocf/logger"
)

type OCF struct{}

type (
	// Config information.
	Config struct {
		ocfcfg string
	}
)

var config Config

var ocfCLi = []cli.Flag{
	cli.StringFlag{
		Name:  "free5gccfg",
		Usage: "common config file",
	},
	cli.StringFlag{
		Name:  "ocfcfg",
		Usage: "ocf config file",
	},
}

var initLog *logrus.Entry

func init() {
	initLog = logger.InitLog
}

func (*OCF) GetCliCmd() (flags []cli.Flag) {
	return ocfCLi
}

func (*OCF) Initialize(c *cli.Context) {

	config = Config{
		ocfcfg: c.String("ocfcfg"),
	}

	if config.ocfcfg != "" {
		factory.InitConfigFactory(config.ocfcfg)
	} else {
		DefaultOcfConfigPath := path_util.Gofree5gcPath("free5gc/config/ocfcfg.conf")
		factory.InitConfigFactory(DefaultOcfConfigPath)
	}

	if app.ContextSelf().Logger.OCF.DebugLevel != "" {
		level, err := logrus.ParseLevel(app.ContextSelf().Logger.OCF.DebugLevel)
		if err != nil {
			initLog.Warnf("Log level [%s] is not valid, set to [info] level", app.ContextSelf().Logger.OCF.DebugLevel)
			logger.SetLogLevel(logrus.InfoLevel)
		} else {
			logger.SetLogLevel(level)
			initLog.Infof("Log level is set to [%s] level", level)
		}
	} else {
		initLog.Infoln("Log level is default set to [info] level")
		logger.SetLogLevel(logrus.InfoLevel)
	}

	logger.SetReportCaller(app.ContextSelf().Logger.OCF.ReportCaller)

}

func (ocf *OCF) FilterCli(c *cli.Context) (args []string) {
	for _, flag := range ocf.GetCliCmd() {
		name := flag.GetName()
		value := fmt.Sprint(c.Generic(name))
		if value == "" {
			continue
		}

		args = append(args, "--"+name, value)
	}
	return args
}

func (ocf *OCF) Start() {
	initLog.Infoln("Server started")

	router := logger_util.NewGinWithLogrus(logger.GinLog)
	router.Use(cors.New(cors.Config{
		AllowMethods: []string{"GET", "POST", "OPTIONS", "PUT", "PATCH", "DELETE"},
		AllowHeaders: []string{"Origin", "Content-Length", "Content-Type", "User-Agent", "Referrer", "Host",
			"Token", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowAllOrigins:  true,
		MaxAge:           86400,
	}))

	httpcallback.AddService(router)
	oam.AddService(router)
	for _, serviceName := range factory.OcfConfig.Configuration.ServiceNameList {
		switch models.ServiceName(serviceName) {
		case models.ServiceName_NOCF_COMM:
			communication.AddService(router)
		case models.ServiceName_NOCF_EVTS:
			eventexposure.AddService(router)
		case models.ServiceName_NOCF_MT:
			mt.AddService(router)
		case models.ServiceName_NOCF_LOC:
			location.AddService(router)
		}
	}

	self := context.OCF_Self()
	util.InitOcfContext(self)

	addr := fmt.Sprintf("%s:%d", self.BindingIPv4, self.SBIPort)

	ngap_service.Run(self.NgapIpList, 38412, ngap.Dispatch)

	// Register to NRF
	var profile models.NfProfile
	if profileTmp, err := consumer.BuildNFInstance(self); err != nil {
		initLog.Error("Build OCF Profile Error")
	} else {
		profile = profileTmp
	}

	if _, nfId, err := consumer.SendRegisterNFInstance(self.NrfUri, self.NfId, profile); err != nil {
		initLog.Warnf("Send Register NF Instance failed: %+v", err)
	} else {
		self.NfId = nfId
	}

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChannel
		ocf.Terminate()
		os.Exit(0)
	}()

	server, err := http2_util.NewServer(addr, util.OcfLogPath, router)

	if server == nil {
		initLog.Errorf("Initialize HTTP server failed: %+v", err)
		return
	}

	if err != nil {
		initLog.Warnf("Initialize HTTP server: %+v", err)
	}

	serverScheme := factory.OcfConfig.Configuration.Sbi.Scheme
	if serverScheme == "http" {
		err = server.ListenAndServe()
	} else if serverScheme == "https" {
		err = server.ListenAndServeTLS(util.OcfPemPath, util.OcfKeyPath)
	}

	if err != nil {
		initLog.Fatalf("HTTP server setup failed: %+v", err)
	}
}

func (ocf *OCF) Exec(c *cli.Context) error {

	initLog.Traceln("args:", c.String("ocfcfg"))
	args := ocf.FilterCli(c)
	initLog.Traceln("filter: ", args)
	command := exec.Command("./ocf", args...)

	stdout, err := command.StdoutPipe()
	if err != nil {
		initLog.Fatalln(err)
	}
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		in := bufio.NewScanner(stdout)
		for in.Scan() {
			fmt.Println(in.Text())
		}
		wg.Done()
	}()

	stderr, err := command.StderrPipe()
	if err != nil {
		initLog.Fatalln(err)
	}
	go func() {
		in := bufio.NewScanner(stderr)
		for in.Scan() {
			fmt.Println(in.Text())
		}
		wg.Done()
	}()

	go func() {
		if err = command.Start(); err != nil {
			initLog.Errorf("OCF Start error: %+v", err)
		}
		wg.Done()
	}()

	wg.Wait()

	return err
}

func (ocf *OCF) Terminate() {
	logger.InitLog.Infof("Terminating OCF...")
	ocfSelf := context.OCF_Self()

	// deregister with NRF
	problemDetails, err := consumer.SendDeregisterNFInstance()
	if problemDetails != nil {
		logger.InitLog.Errorf("Deregister NF instance Failed Problem[%+v]", problemDetails)
	} else if err != nil {
		logger.InitLog.Errorf("Deregister NF instance Error[%+v]", err)
	} else {
		logger.InitLog.Infof("[OCF] Deregister from NRF successfully")
	}

	logger.InitLog.Infof("Send OCF Status Indication to Notify RANs due to OCF terminating")
	unavailableGuamiList := ngap_message.BuildUnavailableGUAMIList(ocfSelf.ServedGuamiList)
	ocfSelf.OcfRanPool.Range(func(key, value interface{}) bool {
		ran := value.(*context.OcfRan)
		ngap_message.SendOCFStatusIndication(ran, unavailableGuamiList)
		return true
	})

	ngap_service.Stop()

	callback.SendOcfStatusChangeNotify((string)(models.StatusChange_UNAVAILABLE), ocfSelf.ServedGuamiList)
	logger.InitLog.Infof("OCF terminated")
}
