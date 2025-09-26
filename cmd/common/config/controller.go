package config

import "github.com/unknownfeature/dcw/cmd/util"

type Alphabet int

const (
	Decimals Alphabet = iota
	Hex
	Uuid
	Base36
	Base64
	Custom
)

var (
	alphabetCharacters = map[Alphabet][]rune{
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

type ControllerConfig[T StringFromAlphabetCustomConfig] struct {
	Workers               int                            `json:"workers"`
	MaxSendRetries        int                            `json:"maxSendRetries"`
	MaxSendRetriesTtsSec  int                            `json:"maxSendRetriesTtsSec"`
	PrecomputeChannelSize int                            `json:"precomputeChannelSize"`
	CustomConfig          StringFromAlphabetCustomConfig `json:"customConfig"`
}

type StringFromAlphabetCustomConfig struct {
	Alphabet  Alphabet  `json:"alphabet"`
	ResLength int       `json:"resLength"`
	Formatter Formatter `json:"formatter"`
}

func ReadControllerConfig[T StringFromAlphabetCustomConfig]() (*ControllerConfig[T], error) {

	return util.ReadToStruct[ControllerConfig[T]](configNames[Controller], func() *ControllerConfig[T] { return &ControllerConfig[T]{} })

}
