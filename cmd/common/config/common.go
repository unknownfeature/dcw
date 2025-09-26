package config

import (
	"github.com/unknownfeature/dcw/cmd/util"
)

type Type int

const (
	Common Type = iota
	Controller
	Worker
)

const (
	CbJob   = "CB_OTP_HTTP"
	TestJob = "Test_HTTP"
)

var configNames = map[Type]string{
	Common:     util.GetEnvString("COMMON_CONFIG_LOC", "common.json"),
	Controller: util.GetEnvString("CONTROLLER_CONFIG_LOC", "controller.json"),
	Worker:     util.GetEnvString("WORKER_CONFIG_LOC", "worker.json"),
}

type CommonConfig struct {
	ControllerHost string `json:"controllerHost"`
	ControllerPort int    `json:"controllerPort"`
	JobName        string `json:"jobName"`
}

func ReadCommonConfig() (*CommonConfig, error) {

	return util.ReadToStruct[CommonConfig](configNames[Common], func() *CommonConfig { return &CommonConfig{} })

}
