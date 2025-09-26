package util

import (
	"encoding/json"
	"errors"
	"github.com/unknownfeature/dcw/cmd/common"
	"os"
)

// just a deduplication of unmarshalling logic
func ReadToStruct[T any](fileLocation string, constructor common.SupplierFunc[*T]) (*T, error) {
	if fileLocation == "" {
		return nil, errors.New("state file can't be empty")
	}

	obj := constructor()

	if _, err := os.Stat(fileLocation); errors.Is(err, os.ErrNotExist) {
		return obj, err
	}
	content, e := os.ReadFile(fileLocation)
	if e != nil {
		return obj, e
	}

	err := json.Unmarshal(content, obj)
	return obj, err

}
