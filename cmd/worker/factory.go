package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/unknownfeature/dcw/cmd/common"
	"github.com/unknownfeature/dcw/cmd/common/config"
	"github.com/unknownfeature/dcw/cmd/common/dto"
	"github.com/unknownfeature/dcw/cmd/worker/client/zmq"
	"github.com/unknownfeature/dcw/cmd/worker/result"
	"github.com/unknownfeature/dcw/cmd/worker/runner"
	"github.com/unknownfeature/dcw/cmd/worker/sfa"
	"github.com/unknownfeature/dcw/cmd/worker/verifier/hr"
	"github.com/unknownfeature/dcw/cmd/worker/verifier/hr/cb"
	"github.com/unknownfeature/dcw/cmd/worker/verifier/hr/cb/vi"
	"golang.org/x/sync/semaphore"
	"log"
	"net/http"
	"runtime"
	"time"
)

func getRunner(commonConfig *config.CommonConfig) (runner.Runner, error) {
	if function, ok := runnerFunctions[commonConfig.JobName]; ok {
		return function(commonConfig)
	}
	return nil, errors.New(fmt.Sprintf("unknown job name %s", commonConfig.JobName))
}

var runnerFunctions = map[string]common.Func[*config.CommonConfig, runner.Runner]{
	config.TestJob: getTestHttpRunner,
	config.CbJob:   getCbRunner,
}

func getCbRunner(commonConfig *config.CommonConfig) (runner.Runner, error) {
	workerConfig, err := config.ReadWorkerConfig[config.HttpRequestVerifier[config.CbCustomConfig]]()
	if err != nil {
		log.Fatal("can't read worker config", err)
	}

	clientConfig := zmq.Config{

		Host:                                 commonConfig.ControllerHost,
		Port:                                 commonConfig.ControllerPort,
		Id:                                   uuid.NewString(),
		MaxConnectAttempts:                   workerConfig.ConnAttempts,
		TimeToSleepBetweenConnectionAttempts: time.Duration(workerConfig.ConnAttemptsTtsSec),
	}

	// todo authorization
	client, err := zmq.NewZMQClient(clientConfig)

	if err != nil {
		return nil, err
	}

	runnerConfig := runner.Config{WorkersCount: runtime.NumCPU() * workerConfig.WorkersFactor}
	sem := semaphore.NewWeighted(int64(runnerConfig.WorkersCount * workerConfig.WorkersSemaphoreWeight))
	ctx := context.Background()

	workReqTrans := dto.NewRequestTransformer[int]()
	workRespTrans := dto.NewResponseTransformer[[]string]()

	resultRequestTrans := dto.NewRequestTransformer[string]()
	headersSupplier := hr.NewSimpleHeadersSupplier(workerConfig.VerifierConfig.Headers)
	method := workerConfig.VerifierConfig.Method
	verifyRequestSupplier := hr.NewRequestSupplier[string](method, headersSupplier, workerConfig.VerifierConfig.Method, cb.NewBodySupplier(workerConfig.VerifierConfig.Body))
	proofTokenRequestSupplier := hr.NewRequestSupplier[map[string]string](method, headersSupplier, workerConfig.VerifierConfig.CustomConfig.VerifyIdentityUrl, vi.NewBodySupplier(workerConfig.VerifierConfig.CustomConfig.VerifyIdentityBody))

	successPredicate := cb.NewResponseHandler(http.DefaultClient, proofTokenRequestSupplier)
	verifier := hr.NewVerifier[string](http.DefaultClient, verifyRequestSupplier, successPredicate)
	worker := sfa.NewWorker(sem, ctx, workRespTrans, verifier)
	workReqSupplier := sfa.NewSupplier(workerConfig.BatchSize, workReqTrans, sem, ctx)
	resHandler := result.NewHandler[string](client, resultRequestTrans)

	return runner.NewDefaultRunner[string](runnerConfig, client, worker, workReqSupplier, resHandler), nil
}

func getTestHttpRunner(commonConfig *config.CommonConfig) (runner.Runner, error) {

	workerConfig, err := config.ReadWorkerConfig[config.HttpRequestVerifier[config.TestHttpCustomConfig]]()
	if err != nil {
		log.Fatal("can't read worker config", err)
	}

	clientConfig := zmq.Config{

		Host:                                 commonConfig.ControllerHost,
		Port:                                 commonConfig.ControllerPort,
		Id:                                   uuid.NewString(),
		MaxConnectAttempts:                   workerConfig.ConnAttempts,
		TimeToSleepBetweenConnectionAttempts: time.Duration(workerConfig.ConnAttemptsTtsSec),
	}

	// todo authorization
	client, err := zmq.NewZMQClient(clientConfig)

	if err != nil {
		return nil, err
	}

	runnerConfig := runner.Config{WorkersCount: runtime.NumCPU() * workerConfig.WorkersFactor}
	sem := semaphore.NewWeighted(int64(runnerConfig.WorkersCount * workerConfig.WorkersSemaphoreWeight))
	ctx := context.Background()

	workReqTrans := dto.NewRequestTransformer[int]()
	workRespTrans := dto.NewResponseTransformer[[]string]()

	resultRequestTrans := dto.NewRequestTransformer[string]()
	headersSupplier := hr.NewSimpleHeadersSupplier(workerConfig.VerifierConfig.Headers)
	method := workerConfig.VerifierConfig.Method
	verifyRequestSupplier := hr.NewRequestSupplier[string](method, headersSupplier, workerConfig.VerifierConfig.Url, hr.NewFormattingBodySupplier[string](workerConfig.VerifierConfig.Body))

	successPredicate := hr.NewResponseHandler(workerConfig.VerifierConfig.CustomConfig.SuccessStatus)
	verifier := hr.NewVerifier[string](http.DefaultClient, verifyRequestSupplier, successPredicate)
	worker := sfa.NewWorker(sem, ctx, workRespTrans, verifier)
	workReqSupplier := sfa.NewSupplier(workerConfig.BatchSize, workReqTrans, sem, ctx)
	resHandler := result.NewHandler[string](client, resultRequestTrans)

	return runner.NewDefaultRunner[string](runnerConfig, client, worker, workReqSupplier, resHandler), nil
}
