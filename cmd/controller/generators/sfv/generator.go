package sfv

import (
	"encoding/json"
	"errors"
	"github.com/skirkyn/dcw/cmd/controller/generators"
	"math"
	"os"
	"sort"
	"sync"
)

type Vocabulary int

const (
	Decimals Vocabulary = iota
	Hex
	Uuid
	Base36
	Base64
	Custom
)

var (
	vocabularyCharacters = map[Vocabulary][]rune{
		Decimals: []rune("0123456789"),
		Hex:      []rune("0123456789abcdef"),
		Base36:   []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"),
		Base64:   []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"),
	}
)

type Formatter int

const (
	Simple Formatter = iota
	Uuid4
)

var (
	formattersFunctions = map[Formatter]func([]rune) (string, error){
		Simple: ToStringFromRunes,
		Uuid4:  ToUuid4StringFromRunes,
	}
)

type Config struct {
	Vocabulary   []rune    `json:"vocabulary"`
	ResultLength int       `json:"resultLength"`
	Formatter    Formatter `json:"formatter"`
}

type State struct {
	Config           Config `json:"config"`
	CurrentPositions []int  `json:"currentPositions"`
	Done             bool   `json:"done"`
}

type Generator struct {
	state     *State
	stateLock *sync.RWMutex
}

const StateFile = "/home/sfa_gen.json"

func ForCustom(resultLength int, vocabulary []rune, formatter Formatter) (generators.Generator[int, []string], error) {

	if resultLength <= 0 {
		return nil, IncorrectResultLengthError
	}
	if vocabulary == nil || len(vocabulary) == 0 {
		return nil, IncorrectVocabularyLengthError
	}
	if int(formatter) >= len(formattersFunctions) {
		return nil, IncorrectFormatterError
	}
	stateVocabulary := append([]rune(nil), vocabulary...)
	sort.Slice(stateVocabulary, func(i, j int) bool {
		return stateVocabulary[i] < stateVocabulary[j]
	})
	state := State{Config: Config{stateVocabulary, resultLength, formatter}}
	return StringFromVocabularyGeneratorFromState(state)

}

func ForStandard(vocabulary Vocabulary, resultLength int, formatter Formatter) (generators.Generator[int, []string], error) {

	if vocabulary == Custom {
		return nil, CustomNotSupportedError
	}
	return ForCustom(resultLength, vocabularyCharacters[vocabulary], formatter)
}

func Resume(stateFileLocation string) (generators.Generator[int, []string], error) {
	if stateFileLocation == "" {
		return nil, errors.New("state file can't be empty")
	}
	if _, err := os.Stat(stateFileLocation); errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	content, e := os.ReadFile(stateFileLocation)
	if e != nil {
		return nil, e
	}
	state := State{}
	e = json.Unmarshal(content, &state)
	if e != nil {
		return nil, e
	}

	return StringFromVocabularyGeneratorFromState(state)
}

func StringFromVocabularyGeneratorFromState(state State) (generators.Generator[int, []string], error) {

	return &Generator{&state, &sync.RWMutex{}}, nil
}

func (g *Generator) Next(batchSize int) ([]string, error) {

	currentPositions, err := g.recalculatePositions(batchSize)

	if err != nil {
		return nil, err
	}

	template := make([]rune, g.state.Config.ResultLength)
	chunk := make([]string, batchSize)
	err = g.generateBatch(&chunk, template, batchSize, 0, currentPositions)

	return chunk, err
}

func (g *Generator) CurrentState() ([]byte, error) {
	g.stateLock.RLock()
	res, e := json.Marshal(g.state)
	g.stateLock.RUnlock()
	return res, e
}

func (g *Generator) generateBatch(res *[]string, current []rune, batchSize int, depth int, currentIndices []int) error {

	if len(*res) == batchSize {
		return nil
	}
	vocabularyLength := len(g.state.Config.Vocabulary)

	if depth == vocabularyLength {
		strRes, err := formattersFunctions[g.state.Config.Formatter](current)
		if err != nil {
			return err
		}
		*res = append(*res, strRes)
	}

	for i := currentIndices[depth]; i < vocabularyLength; i++ {
		current[depth] = g.state.Config.Vocabulary[i]
		currentIndices[depth] = i
		err := g.generateBatch(res, current, batchSize, depth+1, currentIndices)
		if err != nil {
			return err
		}
	}
	currentIndices[depth] = 0
	return nil
}

func (g *Generator) updatePositions(positions *[]int, log int, sum int, index int) int {
	positionsDeref := *positions
	vocabLength := len(g.state.Config.Vocabulary)

	if index == len(positionsDeref) {
		return 0
	}

	newLog := log
	adjustIndex := len(positionsDeref)-index == log
	newSum := sum
	newCarryover := 0
	if adjustIndex {
		iteration := int(math.Pow(float64(vocabLength), float64(log)))
		newSum = sum % iteration
		newCarryover = sum / iteration
		newLog = newLog - 1
	}

	carryover := g.updatePositions(positions, newLog, newSum, index+1)
	newValue := positionsDeref[index] + carryover
	if index == len(positionsDeref)-1 {
		newValue += newSum
	}
	positionsDeref[index] = int(math.Min(float64(newValue), float64(vocabLength)))

	return newCarryover
}

func (g *Generator) recalculatePositions(batchSize int) ([]int, error) {

	g.stateLock.Lock()

	if g.state.Done {
		return nil, PotentialResultsExhaustedError
	}
	vocabularyLength := len(g.state.Config.Vocabulary)
	log := int(math.Log10(float64(batchSize)) / math.Log10(float64(vocabularyLength)))

	oldPositions := append(make([]int, g.state.Config.ResultLength), g.state.CurrentPositions...)
	newPositions := append(make([]int, g.state.Config.ResultLength), g.state.CurrentPositions...)

	carryover := g.updatePositions(&newPositions, int(math.Min(float64(log), float64(g.state.Config.ResultLength))), batchSize, 0)

	if carryover > 0 {
		g.state.Done = true
	}

	g.state.CurrentPositions = newPositions

	g.stateLock.Unlock()

	return oldPositions, nil
}
