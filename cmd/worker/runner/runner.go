package runner

import (
	"github.com/unknownfeature/dcw/cmd/common"
	"github.com/unknownfeature/dcw/cmd/common/dto"
	"github.com/unknownfeature/dcw/cmd/worker/client"
	"log"
	"sync"
	"sync/atomic"
)

type Runner interface {
	Start() *sync.WaitGroup
	Stop()
}

type Config struct {
	WorkersCount int
}
type DefaultRunner[Result any] struct {
	config          Config
	client          client.Client
	worker          common.Function[[]byte, *dto.Request[Result]]
	requestSupplier common.Supplier[[]byte]
	resultHandler   common.Consumer[dto.Request[Result]]
	stop            *atomic.Bool
}

func NewDefaultRunner[Result any](config Config, client client.Client, worker common.Function[[]byte, *dto.Request[Result]], requestSupplier common.Supplier[[]byte], resultHandler common.Consumer[dto.Request[Result]]) Runner {

	return &DefaultRunner[Result]{config, client, worker, requestSupplier, resultHandler, &atomic.Bool{}}
}

func (r *DefaultRunner[Result]) Start() *sync.WaitGroup {

	wg := sync.WaitGroup{}
	wg.Add(r.config.WorkersCount)

	for i := r.config.WorkersCount; i > 0; i-- {
		go r.runWorker(&wg)
	}
	return &wg
}

func (r *DefaultRunner[Result]) Stop() {
	r.stop.Store(true)
}

func (r *DefaultRunner[Result]) runWorker(wg *sync.WaitGroup) {
	defer wg.Done()

	for !r.stop.Load() {

		if supply, err := r.requestSupplier.Supply(); err != nil {
			log.Printf("error calling supply, will exit %s", err.Error())
		} else {
			r.doWork(supply)
		}

	}
}

func (r *DefaultRunner[Result]) doWork(req []byte) {

	// call server  to get the next chunk of work
	resp, err := r.client.Call(req)

	// if all options were tried and to success then the server returns this error
	if err != nil && err.Error() == dto.PotentialResultsExhausted {
		log.Printf("error calling the server %s", err.Error())

		// client just stops
		r.Stop()
		return
	}

	// process next batch from the server
	if res, err := r.worker.Apply(resp); err == nil || err.Error() == dto.PotentialResultsExhausted {
		if err != nil {
			log.Printf("result has been found by someone else")
		} else if err = r.resultHandler.Consume(*res); err != nil {
			// todo should let the server know? But also if it errors then most likely there is a problem with connection to the server
			log.Printf("error handling result %s", err.Error())
		}
		// stop regardless of the error
		r.Stop()
	} else {
		log.Printf("error processing work from server %s", err.Error())
		// todo this about this case
	}

}
