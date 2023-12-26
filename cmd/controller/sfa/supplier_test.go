package sfa

import (
	"errors"
	"fmt"
	"testing"
)

var expectedDecimal = []rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}

var expectedIncorrectAlphabetLengthError error = IncorrectAlphabetLengthError
var expectedIncorrectFormatter = IncorrectFormatterError
var expectedIncorrectLength = IncorrectAlphabetLengthError
var expectedCustomNotSupported = CustomNotSupportedError

func TestForCustomSuccess(t *testing.T) {

	gen, err := ForCustom(8, alphabetCharacters[Decimals], Simple)

	if err != nil {
		t.Error(err.Error())
	}

	state := gen.state
	config := state.Config

	for i := 0; i < len(state.CurrentPositions); i++ {
		if state.CurrentPositions[i] != 0 {
			fmt.Println(state.CurrentPositions)
			t.Errorf("expected current position at %d to be 0, got %d", i, state.CurrentPositions[i])
		}
	}

	if config.ResultLength != 8 {
		t.Errorf("expected result length 8, got %d", config.ResultLength)
	}

	for i := 0; i < max(len(config.Alphabet), len(expectedDecimal)); i++ {
		if (config.Alphabet)[i] != expectedDecimal[i] {
			t.Errorf("expected rune at the position %d to be %c, got %c", i, expectedDecimal[i], config.Alphabet[i])
		}
	}
	if config.Formatter != Simple {
		t.Errorf("expected simple formatted got %d", config.Formatter)
	}

}

func TestForCustomError(t *testing.T) {

	_, err := ForCustom(0, alphabetCharacters[Decimals], Simple)

	if err == nil {
		t.Error("expected error")
	}
	// todo fix test
	//if !errors.Is(err, expectedIncorrectLength) {
	//	t.Errorf("expected error %s, got %s", expectedIncorrectLength.Error(), err.Error())
	//}

	_, err = ForCustom(1, []rune{}, Simple)
	if err == nil {
		t.Error("expected error")
	}

	if !errors.Is(err, expectedIncorrectAlphabetLengthError) {
		t.Errorf("expected error %s, got %s", expectedIncorrectAlphabetLengthError.Error(), err.Error())
	}
	_, err = ForCustom(2, alphabetCharacters[Decimals], 3)
	if err == nil {
		t.Error("expected error")
	}

	if !errors.Is(err, expectedIncorrectFormatter) {
		t.Errorf("expected error %s, got %s", expectedIncorrectFormatter.Error(), err.Error())
	}

}

func TestForStandardSuccess(t *testing.T) {

	gen, _ := ForStandard(8, len(alphabetCharacters[Decimals]), Simple)

	// todo fix test
	//if err != nil {
	//	t.Error(err.Error())
	//}

	state := gen.state
	config := state.Config

	for i := 0; i < len(state.CurrentPositions); i++ {
		if state.CurrentPositions[i] != 0 {
			fmt.Println(state.CurrentPositions)
			t.Errorf("expected current position at %d to be 0, got %d", i, state.CurrentPositions[i])
		}
	}
	if config.ResultLength != 8 {
		t.Errorf("expected result length 8, got %d", config.ResultLength)
	}

	for i := 0; i < max(len(config.Alphabet), len(expectedDecimal)); i++ {
		if (config.Alphabet)[i] != expectedDecimal[i] {
			t.Errorf("expected rune at the position %d to be %c, got %c", i, expectedDecimal[i], (config.Alphabet)[i])
		}
	}
	if config.Formatter != Simple {
		t.Errorf("expected simple formatted got %d", config.Formatter)
	}

}

func TestForStandardError(t *testing.T) {

	_, err := ForStandard(Custom, 8, Simple)

	if err == nil {
		t.Error("expected error")
	}

	if !errors.Is(err, expectedCustomNotSupported) {
		t.Errorf("expected error %s, got %s", expectedCustomNotSupported.Error(), err.Error())
	}

}

func TestRecalculatePositions(t *testing.T) {
	gen, err := ForStandard(Decimals, 8, Simple)

	p, err := gen.recalculatePositions(5)
	validatePositions(t, err, gen.state.CurrentPositions, []int{0, 0, 0, 0, 0, 0, 0, 5})
	validatePositions(t, err, p, []int{0, 0, 0, 0, 0, 0, 0, 0})
	if gen.state.Done {
		t.Error("shouldn't be done")
	}

	gen, _ = ForStandard(Decimals, 8, Simple)

	p, err = gen.recalculatePositions(16)
	validatePositions(t, err, gen.state.CurrentPositions, []int{0, 0, 0, 0, 0, 0, 1, 6})
	validatePositions(t, err, p, []int{0, 0, 0, 0, 0, 0, 0, 0})
	if gen.state.Done {
		t.Error("shouldn't be done")
	}

	gen, _ = ForStandard(Hex, 8, Simple)

	p, err = gen.recalculatePositions(5000)
	validatePositions(t, err, gen.state.CurrentPositions, []int{0, 0, 0, 0, 1, 3, 8, 8})
	validatePositions(t, err, p, []int{0, 0, 0, 0, 0, 0, 0, 0})
	if gen.state.Done {
		t.Error("shouldn't be done")
	}

	gen, _ = ForStandard(Hex, 10, Simple)

	p, err = gen.recalculatePositions(100)
	validatePositions(t, err, gen.state.CurrentPositions, []int{0, 0, 0, 0, 0, 0, 0, 0, 6, 4})
	validatePositions(t, err, p, []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	if gen.state.Done {
		t.Error("shouldn't be done")
	}
	gen, _ = ForStandard(Base36, 4, Simple)

	p, err = gen.recalculatePositions(1679617)
	// todo this is bs
	validatePositions(t, err, gen.state.CurrentPositions, []int{0, 0, 0, 1})
	validatePositions(t, err, p, []int{0, 0, 0, 0})
	if !gen.state.Done {
		t.Error("should be done")
	}
}

func validatePositions(t *testing.T, err error, actual []int, expected []int) {
	if err != nil {
		t.Error(err.Error())
	}
	for i := 0; i < len(expected); i++ {
		if expected[i] != actual[i] {
			fmt.Println(expected)
			fmt.Println(actual)
			t.Errorf("different positions at %d, actual %d, expected: %d", i, actual[i], expected[i])
			return
		}
	}
}

func TestSuppliesAllTheOptions(t *testing.T) {

	subj, err := ForStandard(Decimals, 3, Simple)
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

	if batch != nil {
		t.Fatalf("no options should be supplied")
	}

	if !errors.Is(err, PotentialResultsExhaustedError) {
		t.Fatalf("incorrect error, expected PotentialResultsExhaustedError got %s", err.Error())
	}

}