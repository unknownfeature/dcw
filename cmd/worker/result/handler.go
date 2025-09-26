package result

import (
	"github.com/unknownfeature/dcw/cmd/common"
	"github.com/unknownfeature/dcw/cmd/common/dto"
	"github.com/unknownfeature/dcw/cmd/worker/client"
)

type Handler[In any] struct {
	client             client.Client
	requestTransformer dto.RequestTransformer[In]
}

func NewHandler[In any](client client.Client, requestTransformer dto.RequestTransformer[In]) common.Consumer[dto.Request[In]] {
	return &Handler[In]{client, requestTransformer}
}

func (h *Handler[In]) Consume(res dto.Request[In]) error {

	if req, err := h.requestTransformer.RequestToBytes(res); err != nil {
		return err
	} else {
		_, err = h.client.Call(req)
		return err
	}

}
