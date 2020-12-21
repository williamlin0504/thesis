package factory

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"

	"free5gc/src/ocf/logger"
)

var OcfConfig Config

func checkErr(err error) {
	if err != nil {
		err = fmt.Errorf("[Configuration] %s", err.Error())
		logger.AppLog.Fatal(err)
	}
}

// TODO: Support configuration update from REST api
func InitConfigFactory(f string) {
	content, err := ioutil.ReadFile(f)
	checkErr(err)

	OcfConfig = Config{}

	err = yaml.Unmarshal([]byte(content), &OcfConfig)
	checkErr(err)

	logger.InitLog.Infof("Successfully initialize configuration %s", f)
}
