package main

import (
	"errors"
	"fmt"
	"github.com/unknownfeature/dcw/cmd/common"
	"github.com/unknownfeature/dcw/cmd/common/config"
	"github.com/unknownfeature/dcw/cmd/common/dto"
	"github.com/unknownfeature/dcw/cmd/controller/result"
	"github.com/unknownfeature/dcw/cmd/controller/server"
	"github.com/unknownfeature/dcw/cmd/controller/sfa"
	"log"
)

func getDispatcher(commonConfig *config.CommonConfig, controllerConfig *config.ControllerConfig[config.StringFromAlphabetCustomConfig]) (common.Function[[]byte, []byte], error) {
	if function, ok := dispatcherFunctions[commonConfig.JobName]; ok {
		return function(controllerConfig)
	}
	return nil, errors.New(fmt.Sprintf("unknown job name %s", commonConfig.JobName))
}

var dispatcherFunctions = map[string]common.Func[*config.ControllerConfig[config.StringFromAlphabetCustomConfig], common.Function[[]byte, []byte]]{
	config.TestJob: getSfaBruteForceDispatcher,
	config.CbJob:   getSfaBruteForceDispatcher,
}

func getSfaBruteForceDispatcher(controllerConfig *config.ControllerConfig[config.StringFromAlphabetCustomConfig]) (common.Function[[]byte, []byte], error) {
	workSupplier, err := sfa.ForStandard(controllerConfig.PrecomputeChannelSize, controllerConfig.CustomConfig.Alphabet, controllerConfig.CustomConfig.ResLength, controllerConfig.CustomConfig.Formatter)
	if err != nil {
		log.Fatal("can't create supplier for the server", err)
	}
	workResTrans := dto.NewResponseTransformer[[]string]()
	workHandler := sfa.NewGeneratorHandler(workSupplier, workResTrans)
	resultRespTrans := dto.NewResponseTransformer[string]()

	resultHandler := result.NewHandler[string](resultRespTrans, workHandler)
	handlers := map[dto.Type]common.Function[dto.Request[any], []byte]{dto.Work: workHandler, dto.Result: resultHandler}

	dispatcher := server.NewDispatcher(handlers, dto.NewRequestTransformer[any]())
	return dispatcher, nil
}
