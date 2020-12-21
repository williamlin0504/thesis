package main

import (
	"free5gc/src/app"
	"free5gc/src/ocf/logger"
	"free5gc/src/ocf/service"
	"free5gc/src/ocf/version"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var OCF = &service.OCF{}

var appLog *logrus.Entry

var ue = &service.UE{}

func init() {
	appLog = logger.AppLog
}

func main() {
	app := cli.NewApp()
	app.Name = "ocf"
	appLog.Infoln(app.Name)
	appLog.Infoln("OCF version: ", version.GetVersion())
	app.Usage = "-free5gccfg common configuration file -ocfcfg ocf configuration file"
	app.Action = action
	app.Flags = OCF.GetCliCmd()
	if err := app.Run(os.Args); err != nil {
		logger.AppLog.Errorf("OCF Run error: %v", err)
	}
}

func action(c *cli.Context) {
	app.AppInitializeWillInitialize(c.String("free5gccfg"))
	OCF.Initialize(c)
	OCF.Start()
}

func reservationGU(ue ue_ID, db abmf) {
	// 1. Default GU = 100
	int default_GU = 100;



}

func requestGU(ue ue_ID) {
	// 1. Default 要求GU = 10，檢查 ABMF 看此 UE 的餘額
	// 2. 如果餘額足夠，就給此 UE 要求的 GU
	// 3. 如果餘額不足，回傳 “餘額不足的訊息”
	int default_response_GU = 10

	if(abmf.allowance >= default_response_GU){
		return default_response_GU
	}
	else {
		return "Not enough Allowance"
	}
}
