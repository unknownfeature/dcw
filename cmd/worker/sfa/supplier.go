package sfa

import (
	"context"
	"github.com/unknownfeature/dcw/cmd/common"
	"github.com/unknownfeature/dcw/cmd/common/dto"
	"golang.org/x/sync/semaphore"
)

type WorkSupplier struct {
	batchSize          int
	requestTransformer dto.RequestTransformer[int]
	semaphore          *semaphore.Weighted
	context            context.Context
}

func NewSupplier(batchSize int, requestTransformer dto.RequestTransformer[int], semaphore *semaphore.Weighted, context context.Context) common.Supplier[[]byte] {

	return &WorkSupplier{batchSize, requestTransformer, semaphore, context}
}

// generates next work request that goes to the server
func (s *WorkSupplier) Supply() ([]byte, error) {

	if err := s.semaphore.Acquire(s.context, 1); err != nil {
		return nil, err
	}

	req := dto.Request[int]{Type: dto.Work, Body: s.batchSize}

	return s.requestTransformer.RequestToBytes(req)
}
