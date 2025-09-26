package sfa

import (
	"errors"
	"github.com/unknownfeature/dcw/cmd/common/config"
	"testing"
)

var expectedDecimal = []rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}

var expectedIncorrectAlphabetLengthError error = IncorrectAlphabetLengthError
var expectedIncorrectFormatter = IncorrectFormatterError
var expectedIncorrectLength = IncorrectAlphabetLengthError
var expectedCustomNotSupported = CustomNotSupportedError

func TestForCustomSuccess(t *testing.T) {

	gen, err := ForCustom(1000, 8, alphabetCharacters[config.Decimals], config.Simple)

	if err != nil {
		t.Error(err.Error())
	}

	state := gen.state
	conf := state.Config

	if conf.ResultLength != 8 {
		t.Errorf("expected result length 8, got %d", conf.ResultLength)
	}

	for i := 0; i < max(len(conf.Alphabet), len(expectedDecimal)); i++ {
		if (conf.Alphabet)[i] != expectedDecimal[i] {
			t.Errorf("expected rune at the position %d to be %c, got %c", i, expectedDecimal[i], conf.Alphabet[i])
		}
	}
	if conf.Formatter != config.Simple {
		t.Errorf("expected simple formatted got %d", conf.Formatter)
	}

}

func TestForCustomError(t *testing.T) {

	_, err := ForCustom(1000, 0, alphabetCharacters[config.Decimals], config.Simple)

	if err == nil {
		t.Error("expected error")
	}
	// todo fix test
	//if !errors.Is(err, expectedIncorrectLength) {
	//	t.Errorf("expected error %s, got %s", expectedIncorrectLength.Error(), err.Error())
	//}

	_, err = ForCustom(1000, 1, []rune{}, config.Simple)
	if err == nil {
		t.Error("expected error")
	}

	if !errors.Is(err, expectedIncorrectAlphabetLengthError) {
		t.Errorf("expected error %s, got %s", expectedIncorrectAlphabetLengthError.Error(), err.Error())
	}
	_, err = ForCustom(1000, 2, alphabetCharacters[config.Decimals], 3)
	if err == nil {
		t.Error("expected error")
	}

	if !errors.Is(err, expectedIncorrectFormatter) {
		t.Errorf("expected error %s, got %s", expectedIncorrectFormatter.Error(), err.Error())
	}

}

func TestForStandardError(t *testing.T) {

	_, err := ForStandard(1000, config.Custom, 8, config.Simple)

	if err == nil {
		t.Error("expected error")
	}

	if !errors.Is(err, expectedCustomNotSupported) {
		t.Errorf("expected error %s, got %s", expectedCustomNotSupported.Error(), err.Error())
	}

}

//func TestRecalculatePositions(t *testing.T) {
//	gen, err := ForStandard(config.Decimals, 8, config.Simple)
//
//	p, err := gen.recalculatePositions(5)
//	validatePositions(t, err, gen.state.CurrentPositions, []int{0, 0, 0, 0, 0, 0, 0, 5})
//	validatePositions(t, err, p, []int{0, 0, 0, 0, 0, 0, 0, 0})
//
//	gen, _ = ForStandard(config.Decimals, 8, config.Simple)
//
//	p, err = gen.recalculatePositions(16)
//	validatePositions(t, err, gen.state.CurrentPositions, []int{0, 0, 0, 0, 0, 0, 1, 6})
//	validatePositions(t, err, p, []int{0, 0, 0, 0, 0, 0, 0, 0})
//
//	gen, _ = ForStandard(config.Hex, 8, config.Simple)
//
//	p, err = gen.recalculatePositions(5000)
//	validatePositions(t, err, gen.state.CurrentPositions, []int{0, 0, 0, 0, 1, 3, 8, 8})
//	validatePositions(t, err, p, []int{0, 0, 0, 0, 0, 0, 0, 0})
//
//	gen, _ = ForStandard(config.Hex, 10, config.Simple)
//
//	p, err = gen.recalculatePositions(100)
//	validatePositions(t, err, gen.state.CurrentPositions, []int{0, 0, 0, 0, 0, 0, 0, 0, 6, 4})
//	validatePositions(t, err, p, []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
//
//	gen, _ = ForStandard(config.Base36, 4, config.Simple)
//
//	p, err = gen.recalculatePositions(1679617)
//	// todo this is bs
//	validatePositions(t, err, gen.state.CurrentPositions, []int{0, 0, 0, 1})
//	validatePositions(t, err, p, []int{0, 0, 0, 0})
//
//}

func validatePositions(t *testing.T, err error, actual []int, expected []int) {
	if err != nil {
		t.Error(err.Error())
	}
	for i := 0; i < len(expected); i++ {
		if expected[i] != actual[i] {
			t.Errorf("different positions at %d, actual %d, expected: %d", i, actual[i], expected[i])
			return
		}
	}
}

func TestSuppliesAllTheOptions(t *testing.T) {

	subj, err := ForStandard(1000, config.Decimals, 3, config.Simple)
	if err != nil {
		t.Fatalf("error is not expected %s", err.Error())
	}

	batch, err := subj.Apply(1000)
	if err != nil {
		t.Fatalf("error is not expected %s", err.Error())
	}
	if len(batch) != 1000 {

		t.Fatalf("invalid batch size expected %d, actual %d", 1000, len(batch))
	}
	batch, err = subj.Apply(1000)

	if len(batch) != 0 {
		t.Fatalf("no options should be supplied")
	}

	if !errors.Is(err, PotentialResultsExhaustedError) {
		t.Fatalf("incorrect error, expected PotentialResultsExhaustedError got %s", err.Error())
	}

}

func TestSteps(t *testing.T) {

	subj, err := ForStandard(1000, config.Decimals, 4, config.Simple)
	if err != nil {
		t.Fatalf("error is not expected %s", err.Error())
	}
	batch, err := subj.Apply(10)
	counter := 1010

	for err == nil && counter > 0 {
		counter--
		if len(batch) == 0 {
			break
		}
		println(batch[len(batch)-1])
		batch, err = subj.Apply(10)
		if err != nil && err != PotentialResultsExhaustedError {
			t.Fatalf("error is not expected %s", err.Error())
		}

	}

	if counter != 10 {
		t.Fatalf("invalid counter expected 11, got %d", counter)
	}
}
