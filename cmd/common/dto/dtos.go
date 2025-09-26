package dto

import "encoding/json"

type Type int

type RequestTransformer[In any] interface {
	BytesToRequest([]byte) (Request[In], error)
	RequestToBytes(Request[In]) ([]byte, error)
}

type ResponseTransformer[Out any] interface {
	ResponseToBytes(Response[Out]) ([]byte, error)
	BytesToResponse([]byte) (Response[Out], error)
}

const (
	Work Type = iota
	Result
)

const (
	Done                      = "done"
	PotentialResultsExhausted = "potential results exhausted"
)

type Request[In any] struct {
	Type Type `json:"type"`
	Body In   `json:"body"`
}

type Response[Out any] struct {
	Done bool `json:"done"`
	Body Out  `json:"body"`
}

type Transformer[In any, Out any] struct {
}

func NewRequestTransformer[In any]() RequestTransformer[In] {
	return &Transformer[In, any]{}
}

func NewResponseTransformer[Out any]() ResponseTransformer[Out] {
	return &Transformer[any, Out]{}
}

func (t *Transformer[In, Out]) BytesToRequest(in []byte) (Request[In], error) {
	var req Request[In]
	err := json.Unmarshal(in, &req)
	return req, err
}

func (t *Transformer[In, Out]) RequestToBytes(in Request[In]) ([]byte, error) {
	return json.Marshal(in)
}

func (t *Transformer[In, Out]) BytesToResponse(in []byte) (Response[Out], error) {
	var req Response[Out]
	err := json.Unmarshal(in, &req)
	return req, err
}

func (t *Transformer[In, Out]) ResponseToBytes(in Response[Out]) ([]byte, error) {
	return json.Marshal(in)
}
