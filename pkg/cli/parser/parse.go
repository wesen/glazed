package parser

import (
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
	"strings"
)

type Parser struct {
	parameterDefinitions []*parameters.ParameterDefinition
	argumentDefinitions  []*parameters.ParameterDefinition
}

type Option func(*Parser)

func WithParameterDefinitions(parameterDefinitions []*parameters.ParameterDefinition) Option {
	return func(p *Parser) {
		p.parameterDefinitions = parameterDefinitions
	}
}

func WithArgumentDefinitions(argumentDefinitions []*parameters.ParameterDefinition) Option {
	return func(p *Parser) {
		p.argumentDefinitions = argumentDefinitions
	}
}

func NewParser(options ...Option) *Parser {
	p := &Parser{}
	for _, option := range options {
		option(p)
	}
	return p
}

type ParseError struct {
	Message   string
	Parameter *parameters.ParameterDefinition
	Flag      string
}

type ParseResult struct {
	ParseErrors  []ParseError
	ParsedValues map[string]interface{}
	Arguments    []interface{}
}

const (
	StateFlags = iota
	StateFlagArgument
	StateArguments
)

const (
	ErrorUnknownFlag = iota
)

func (p *Parser) Parse(args []string) (*ParseResult, error) {
	ret := &ParseResult{}
	return ret, nil
}

// flagStrings stores the result of collecting command line argument into
// their respective flag values.
//
// Since we can have list flags, or even just parsing flags with multiple values and give proper feedback,
// we first go through all the flags and accumulate their values.
// When a flag is passed without value, we record a nil value to differentiate it from
// the empty string.
type flagStrings map[string][]*string

type collectResults struct {
	flagStrings flagStrings
	arguments   []string
	errors      []ParseError
}

// collectStrings parses the command line arguments and collects the string values for
// each flag into a hash map. For example, if the command line arguments are:
//
//	`--arg foo --arg bar --arg2 blop --arg2 blip --foo`
//
// then the result will be:
//
//	arg = {"foo", "bar"}
//	arg2 = {"blop", "blip"}
//	foo = {nil}
func (p *Parser) collectStrings(args []string) (*collectResults, error) {
	ret := &collectResults{
		flagStrings: flagStrings{},
		arguments:   []string{},
		errors:      []ParseError{},
	}

	state := StateFlags
	var pd_ *parameters.ParameterDefinition

	for _, arg := range args {
		switch state {
		case StateFlags:
			pd_ = nil
			if arg == "--" {
				state = StateArguments
				continue
			}
			if len(arg) > 2 && arg[0:2] == "--" {
				for _, pd := range p.parameterDefinitions {
					if pd.Name == arg[2:] {
						state = StateFlagArgument
						pd_ = pd

						v := strings.SplitN(arg, "=", 2)
						if len(v) != 2 {
							return nil, errors.New("Internal error: split of flag value failed")
						}
						ret.flagStrings[pd.Name] = append(ret.flagStrings[pd.Name], &v[1])
						break
					}
				}
				if pd_ == nil {
					ret.errors = append(ret.errors, ParseError{
						Message:   "Unknown flag: " + arg,
						Flag:      arg[2:],
						Parameter: nil,
					})
					state = StateFlags
				}
			} else if len(arg) > 1 && arg[0:1] == "-" {
				for _, pd := range p.parameterDefinitions {
					if pd.ShortFlag == arg[1:] {
						state = StateFlagArgument
						pd_ = pd

						if strings.Contains(arg, "=") {
							v := strings.SplitN(arg, "=", 2)
							if len(v) != 2 {
								return nil, errors.New("Internal error: split of flag value failed")
							}
							ret.flagStrings[pd.Name] = append(ret.flagStrings[pd.Name], &v[1])
						}

						break
					}
				}
				if pd_ == nil {
					ret.errors = append(ret.errors, ParseError{
						Message:   "Unknown flag: " + arg,
						Flag:      arg[1:],
						Parameter: nil,
					})
					state = StateFlags
				}
			} else {
				ret.arguments = append(ret.arguments, arg)
				state = StateArguments
			}

		case StateFlagArgument:
			if pd_ == nil {
				// should never happen
				return nil, errors.New("Internal error: pd_ is nil")
			}

			// just store the string for now
			ret.flagStrings[pd_.Name] = append(ret.flagStrings[pd_.Name], &arg)

		case StateArguments:
			ret.arguments = append(ret.arguments, arg)
		}
	}

	return ret, nil
}
