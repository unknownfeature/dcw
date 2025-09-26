package hr

import (
	"bytes"
	"github.com/unknownfeature/dcw/cmd/common"
	"log"
	"net/http"
)

// this type's responsibility is to build HTTP request based on the body and headers templates it received st the moment of construction
// and in parameter it will be receiving at runtime
type RequestSupplier[In any] struct {
	method          string
	headersSupplier common.Function[any, map[string]string]
	url             string
	bodySupplier    common.Function[In, []byte]
}

func NewRequestSupplier[In any](method string, headersSupplier common.Function[any, map[string]string], url string, bodySupplier common.Function[In, []byte]) common.Function[In, *http.Request] {
	return &RequestSupplier[In]{
		method, headersSupplier, url, bodySupplier,
	}
}

func (r *RequestSupplier[In]) Apply(in In) (*http.Request, error) {

	supply, err := r.bodySupplier.Apply(in)

	req, err := http.NewRequest(r.method, r.url, bytes.NewReader(supply))

	if err != nil {
		log.Printf("can't create http request %s", err.Error())
		return nil, err
	}
	headers, err := r.headersSupplier.Apply(in)

	if err != nil {
		log.Printf("can't create http request %s", err.Error())
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return req, nil

}
