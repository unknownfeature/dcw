package config

import "github.com/unknownfeature/dcw/cmd/util"

type WorkerConfig[T HttpRequestVerifier[C], C CbCustomConfig | TestHttpCustomConfig] struct {
	ConnAttempts           int `json:"connAttempts"`
	ConnAttemptsTtsSec     int `json:"connAttemptsTtsSec"`
	BatchSize              int `json:"batchSize"`
	WorkersFactor          int `json:"workersFactor"`
	WorkersSemaphoreWeight int `json:"workersSemaphoreWeight"`
	VerifierConfig         T   `json:"verifierConfig"`
}

type HttpRequestVerifier[T CbCustomConfig | TestHttpCustomConfig] struct {
	Method       string            `json:"method"`
	Url          string            `json:"url"`
	Headers      map[string]string `json:"headers"`
	Body         string            `json:"body"`
	CustomConfig T                 `json:"customConfig"`
}

type CbCustomConfig struct {
	VerifyIdentityUrl  string `json:"verifyIdentityUrl"`
	VerifyIdentityBody string `json:"verifyIdentityBody"`
}

type TestHttpCustomConfig struct {
	SuccessStatus int `json:"successStatus"`
}

func ReadWorkerConfig[T HttpRequestVerifier[C], C CbCustomConfig | TestHttpCustomConfig]() (*WorkerConfig[T, C], error) {

	return util.ReadToStruct[WorkerConfig[T, C]](configNames[Worker], func() *WorkerConfig[T, C] { return &WorkerConfig[T, C]{} })

}
