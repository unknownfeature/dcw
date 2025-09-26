package hr

import (
	"github.com/unknownfeature/dcw/cmd/common"
	"log"
	"net/http"
)

// this is sort of response post processing
type Verifier[In any] struct {
	client          *http.Client
	requestSupplier common.Function[In, *http.Request]
	onResponse      common.Predicate[*http.Response]
}

func NewVerifier[In any](client *http.Client, requestSupplier common.Function[In, *http.Request], onResponse common.Predicate[*http.Response]) common.Predicate[In] {
	return &Verifier[In]{client, requestSupplier, onResponse}
}

func (v *Verifier[In]) Test(in In) (bool, error) {

	if req, err := v.requestSupplier.Apply(in); err != nil {
		log.Printf("can't verify because request can't be created %s", err.Error())
		return false, err
	} else if resp, err := v.client.Do(req); err != nil {
		log.Printf("error calling http request %s", err.Error())
		return false, err
	} else if success, err := v.onResponse.Test(resp); err != nil {
		log.Printf("error verifying http request %s", err.Error())
		return false, err
	} else {
		return success, nil
	}

}
