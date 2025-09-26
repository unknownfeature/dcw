package sfa

import (
	"context"
	"errors"
	"github.com/unknownfeature/dcw/cmd/common"
	"github.com/unknownfeature/dcw/cmd/common/dto"
	"golang.org/x/sync/semaphore"
	"log"
)

type Worker struct {
	semaphore   *semaphore.Weighted
	context     context.Context
	transformer dto.ResponseTransformer[[]string]
	verifier    common.Predicate[string]
}

func NewWorker(semaphore *semaphore.Weighted,
	context context.Context,
	transformer dto.ResponseTransformer[[]string],
	verifier common.Predicate[string]) common.Function[[]byte, *dto.Request[string]] {
	return &Worker{semaphore, context, transformer, verifier}
}
func (w *Worker) Apply(work []byte) (*dto.Request[string], error) {
	defer w.semaphore.Release(1)
	resp, err := w.transformer.BytesToResponse(work)

	if err != nil {
		log.Printf("can't process response %s because of %s", string(work), err.Error())
		return nil, err
	}
	// special response (kind of crutch) which tells the client that the result is found and we can relax
	if resp.Done {
		return nil, errors.New(dto.Done)
	}

	input := resp.Body

	for i := 0; i < len(input); i++ {
		current := input[i]
		var res bool
		res, err = w.verifier.Test(current)

		if err != nil {
			return nil, err
		}
		if res {
			return &dto.Request[string]{Type: dto.Result, Body: current}, nil
		}
	}

	return nil, nil
}
