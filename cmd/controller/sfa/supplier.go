package sfa

import (
	"encoding/json"
	"github.com/unknownfeature/dcw/cmd/common/config"
	"github.com/unknownfeature/dcw/cmd/util"
	"log"
	"math"
	"sort"
	"sync"
)

var (
	alphabetCharacters = map[config.Alphabet][]rune{
		config.Decimals: []rune("0123456789"),
		config.Hex:      []rune("0123456789abcdef"),
		config.Base36:   []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"),
		config.Base64:   []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"),
	}
)

var (
	formattersFunctions = map[config.Formatter]func([]rune) (string, error){
		config.Simple: ToStringFromRunes,
		config.Uuid4:  ToUuid4StringFromRunes,
	}
)

type Config struct {
	Alphabet     []rune           `json:"alphabet"`
	ResultLength int              `json:"resultLength"`
	Formatter    config.Formatter `json:"formatter"`
}

type State struct {
	Config Config `json:"config"`

	// total number of all possible results (|Alphabet| ^ ResultLength)
	Total int `json:"total"`
	// currently sent to workers
	Current int `json:"current"`
}

type Supplier struct {
	state             *State
	stateLock         *sync.RWMutex
	precomputeChannel chan string
}

// todo actually add state persistance
const StateFile = "/home/sfa_gen.json"

func ForCustom(precomputeChannelSize int, resultLength int, alphabet []rune, formatter config.Formatter) (*Supplier, error) {

	if resultLength <= 0 {
		return nil, IncorrectResultLengthError
	}
	if alphabet == nil || len(alphabet) == 0 {
		return nil, IncorrectAlphabetLengthError
	}
	if int(formatter) >= len(formattersFunctions) {
		return nil, IncorrectFormatterError
	}
	stateAlphabet := append([]rune(nil), alphabet...)
	// Sort the alphabet to ensure deterministic and canonical generation order.
	sort.Slice(stateAlphabet, func(i, j int) bool {
		return stateAlphabet[i] < stateAlphabet[j]
	})
	// Initialize state: positions start at 0, Total is calculated as N^L.
	state := &State{Config: Config{stateAlphabet, resultLength, formatter}, Total: int(math.Pow(float64(len(stateAlphabet)), float64(resultLength)))}
	return StringFromAlphabetGeneratorFromState(precomputeChannelSize, state)

}

func ForStandard(precomputeChannelSize int, alphabet config.Alphabet, resultLength int, formatter config.Formatter) (*Supplier, error) {

	if alphabet == config.Custom {
		return nil, CustomNotSupportedError
	}
	return ForCustom(precomputeChannelSize, resultLength, alphabetCharacters[alphabet], formatter)
}

func Resume(precomputeChannelSize int, stateFileLocation string) (*Supplier, error) {

	res, err := util.ReadToStruct[State](stateFileLocation, func() *State { return &State{} })
	if err != nil {
		return nil, err
	}
	return StringFromAlphabetGeneratorFromState(precomputeChannelSize, res)
}

func StringFromAlphabetGeneratorFromState(precomputeChannelSize int, state *State) (*Supplier, error) {

	supl := &Supplier{state, &sync.RWMutex{}, make(chan string, precomputeChannelSize)}

	go supl.generate(make([]rune, state.Config.ResultLength), state.Config.ResultLength-1)
	return supl, nil
}

// Apply requests a batch of strings. It is the primary generation entry point.
func (g *Supplier) Apply(batchSize int) ([]string, error) {

	// todo refactor
	g.stateLock.Lock()
	if g.state.Total == g.state.Current {
		g.stateLock.Unlock()
		return nil, PotentialResultsExhaustedError
	}
	g.stateLock.Unlock()

	var err error = nil
	res := make([]string, 0)

	for i := 0; i < batchSize; i++ {
		g.stateLock.Lock()
		if g.state.Total > g.state.Current {
			g.state.Current++
			res = append(res, <-g.precomputeChannel)
			g.stateLock.Unlock()
		} else {
			g.stateLock.Unlock()
			err = PotentialResultsExhaustedError
			break
		}

	}

	return res, err
}

func (g *Supplier) CurrentState() ([]byte, error) {
	g.stateLock.RLock() // Use RLock since only reading state for serialization
	res, e := json.Marshal(g.state)
	g.stateLock.RUnlock()
	return res, e
}

func (g *Supplier) generate(template []rune, startWordPosition int) {
	if startWordPosition == 0 {
		for ap := 0; ap < len(g.state.Config.Alphabet); ap++ {
			template[startWordPosition] = g.state.Config.Alphabet[ap]

			strRes, err := formattersFunctions[g.state.Config.Formatter](template)
			if err != nil {
				// literally should not be possible
				log.Fatal(err)
			}
			g.precomputeChannel <- strRes

		}

	} else {
		for ap := 0; ap < len(g.state.Config.Alphabet); ap++ {
			template[startWordPosition] = g.state.Config.Alphabet[ap]
			g.generate(template, startWordPosition-1)
		}
	}
}
