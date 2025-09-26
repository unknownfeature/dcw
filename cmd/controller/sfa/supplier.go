package sfa

import (
	"encoding/json"
	"github.com/unknownfeature/dcw/cmd/common/config"
	"github.com/unknownfeature/dcw/cmd/util"
	"math"
	"slices"
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
	// stores the state of generation: for each position in the result -> index of the rune in the alphabet
	// This array acts as a multi-digit counter in base |Alphabet|.
	CurrentPositions []int `json:"currentPositions"`
	// total number of all possible results (|Alphabet| ^ ResultLength)
	Total int `json:"total"`
	// currently results generated (the number of "counts" performed)
	Current int `json:"current"`
}

type Supplier struct {
	state     *State
	stateLock *sync.RWMutex
}

// todo actually add state persistance
const StateFile = "/home/sfa_gen.json"

func ForCustom(resultLength int, alphabet []rune, formatter config.Formatter) (*Supplier, error) {

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
	state := &State{Config: Config{stateAlphabet, resultLength, formatter}, CurrentPositions: make([]int, resultLength), Total: int(math.Pow(float64(len(stateAlphabet)), float64(resultLength)))}
	return StringFromAlphabetGeneratorFromState(state)

}

func ForStandard(alphabet config.Alphabet, resultLength int, formatter config.Formatter) (*Supplier, error) {

	if alphabet == config.Custom {
		return nil, CustomNotSupportedError
	}
	return ForCustom(resultLength, alphabetCharacters[alphabet], formatter)
}

func Resume(stateFileLocation string) (*Supplier, error) {

	res, err := util.ReadToStruct[State](stateFileLocation, func() *State { return &State{} })
	if err != nil {
		return nil, err
	}
	return StringFromAlphabetGeneratorFromState(res)
}

func StringFromAlphabetGeneratorFromState(state *State) (*Supplier, error) {

	return &Supplier{state, &sync.RWMutex{}}, nil
}

// Apply requests a batch of strings. It is the primary generation entry point.
func (g *Supplier) Apply(batchSize int) ([]string, error) {

	// todo enable this(temporary disabled due to bugs)
	// The commented out code suggests the original intent was to use an arithmetic
	// position advancement (jump ahead), but it currently relies on recursion.

	currentPositions, err := g.recalculatePositions(batchSize)

	if err != nil {
		return nil, err
	}

	template := make([]rune, g.state.Config.ResultLength)
	chunk := make([]string, 0)

	g.stateLock.Lock() // Lock for writing (modifying state/current positions)
	defer g.stateLock.Unlock()

	// if we have reached the end
	if util.AllEqual(g.state.CurrentPositions, len(g.state.Config.Alphabet)-1) {
		return nil, PotentialResultsExhaustedError
	}
	// Start the recursive generation algorithm (depth-first search/base-N counting)
	_, _, err = g.generateBatch(&chunk, template, batchSize, 0, currentPositions)

	return chunk, err
}

func (g *Supplier) CurrentState() ([]byte, error) {
	g.stateLock.RLock() // Use RLock since only reading state for serialization
	res, e := json.Marshal(g.state)
	g.stateLock.RUnlock()
	return res, e
}

// this function implements a multi-digit counter in a custom base (|Alphabet|) through a depth-first approach
func (g *Supplier) generateBatch(res *[]string, current []rune, remaining int, depth int, currentIndices []int) (bool, int, error) {

	// Check 1: Stop if we iterated over all the positions
	if util.AllEqual(currentIndices, len(g.state.Config.Alphabet)-1) {
		return false, 0, PotentialResultsExhaustedError
	}
	// Check 2: Stop if the requested batch is full.
	if remaining == 0 {
		return false, remaining, nil
	}

	alphabetLength := len(g.state.Config.Alphabet)

	// BASE CASE: If depth equals string length, a full string has been formed.
	if depth == len(current) {

		strRes, err := formattersFunctions[g.state.Config.Formatter](current)
		if err != nil {
			return false, remaining, err
		}
		*res = append(*res, strRes)
		// Signal a successful generation and decrease the remaining count.
		return true, remaining - 1, nil
	}

	counter := remaining
	times := 0
	carryover := false

	// RECURSIVE STEP: Iterate through possible characters at the current position ('digit').
	for times < alphabetLength && counter > 0 && g.state.Total > g.state.Current {
		// Set the character at the current depth, starting from the last saved index.
		// The modulo ensures we wrap around the alphabet when incrementing.
		current[depth] = g.state.Config.Alphabet[(times+currentIndices[depth])%alphabetLength]

		// RECURSIVE CALL: Generate the rest of the string by increasing the depth (moving right).
		newCarryover, newLeft, err := g.generateBatch(res, current, counter, depth+1, currentIndices)

		counter = newLeft
		carryover = carryover || newCarryover // Track if any inner call caused a rollover/carry.
		if err != nil {
			break
		}
		times++
	}

	// CARRYOVER LOGIC: Acts like the ripple-carry in a multi-digit counter.
	if carryover {
		oldVal := currentIndices[depth]
		newVal := oldVal + times               // Sum of the old index + number of times this position looped.
		adjustedVal := newVal % alphabetLength // The new index at this position after loop and rollover.
		currentIndices[depth] = adjustedVal
		// If the new total count (newVal) is different from the adjusted index (adjustedVal),
		// it means a full rollover occurred, and we signal a carry to the position (digit) on the left.
		return newVal != adjustedVal, counter, nil
	}
	return false, counter, nil

}

// updatePositions implements the arithmetic advancement for the multi-digit counter.
func (g *Supplier) updatePositions(positions []int, log int, total int, index int) int {

	aplhabetlength := len(g.state.Config.Alphabet)

	if index == len(positions) {
		return 0
	}

	newLog := log
	adjustIndex := len(positions)-index == log
	newSum := total
	newCarryover := 0
	if adjustIndex {
		iteration := int(math.Pow(float64(aplhabetlength), float64(log)))
		newSum = total % iteration
		newCarryover = total / iteration
		newLog = newLog - 1
	}

	carryover := g.updatePositions(positions, newLog, newSum, index+1)
	newValue := positions[index] + carryover
	if index == len(positions)-1 {
		newValue += newSum
	}
	positions[index] = int(math.Min(float64(newValue), float64(aplhabetlength-1)))
	if positions[index] < newValue && newCarryover == 0 {
		newCarryover++
	}
	return newCarryover
}

// this function calculates the "jump" in positions
// this ius needed to make this generator insanely optimal and not make workers to wait for the previous generation to be over
// positions are only needed for the generating function to know where pick up the generation from
func (g *Supplier) recalculatePositions(batchSize int) ([]int, error) {

	g.stateLock.Lock()
	defer g.stateLock.Unlock()

	alphabetLength := len(g.state.Config.Alphabet)

	// this is needed in order to understand how many positions will the batch fully rotate
	log := int(math.Log10(float64(batchSize)) / math.Log10(float64(alphabetLength)))

	oldPositions := slices.Clone(g.state.CurrentPositions)

	_ = g.updatePositions(g.state.CurrentPositions, int(math.Min(float64(log), float64(g.state.Config.ResultLength))), batchSize, 0)
	return oldPositions, nil
}
