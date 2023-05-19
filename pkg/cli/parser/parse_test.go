package parser

import (
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/stretchr/testify/require"
	"testing"
)

func NewSimplePD(name string) *parameters.ParameterDefinition {
	return parameters.NewParameterDefinition(name, parameters.ParameterTypeString)
}

func NewShortPD(name string, short string) *parameters.ParameterDefinition {
	return parameters.NewParameterDefinition(
		name,
		parameters.ParameterTypeString,
		parameters.WithShortFlag(short),
	)
}

func TestCollectStrings_SingleFlagWithValue(t *testing.T) {
	args := []string{"--flag", "value"}

	p := NewParser(WithParameterDefinitions([]*parameters.ParameterDefinition{
		NewSimplePD("flag"),
	}))

	ret, err := p.Parse(args)
	require.NoError(t, err)

	require.Equal(t, map[string]interface{}{
		"flag": "value",
	}, ret.ParsedValues)
}

func TestCollectStrings_MultipleFlagsWithValues(t *testing.T)                 {}
func TestCollectStrings_SingleFlagWithoutValue(t *testing.T)                  {}
func TestCollectStrings_MultipleFlagsWithoutValues(t *testing.T)              {}
func TestCollectStrings_MixedFlagsWithAndWithoutValues(t *testing.T)          {}
func TestCollectStrings_FlagWithEmptyValue(t *testing.T)                      {}
func TestCollectStrings_FlagWithQuotedValue(t *testing.T)                     {}
func TestCollectStrings_InvalidFlag(t *testing.T)                             {}
func TestCollectStrings_MissingFlagValue(t *testing.T)                        {}
func TestCollectStrings_FlagWithHyphenInValue(t *testing.T)                   {}
func TestCollectStrings_DuplicateFlagsWithValues(t *testing.T)                {}
func TestCollectStrings_FlagWithNumberValue(t *testing.T)                     {}
func TestCollectStrings_FlagWithSpecialCharactersValue(t *testing.T)          {}
func TestCollectStrings_UnrecognizedFlag(t *testing.T)                        {}
func TestCollectStrings_FlagWithValueAndAdditionalParameters(t *testing.T)    {}
func TestCollectStrings_SingleShortFlagWithValue(t *testing.T)                {}
func TestCollectStrings_MultipleShortFlagsWithValues(t *testing.T)            {}
func TestCollectStrings_ShortAndLongFlagsWithValues(t *testing.T)             {}
func TestCollectStrings_ShortAndLongFlagsMixedOrder(t *testing.T)             {}
func TestCollectStrings_DuplicateShortAndLongFlagsCombined(t *testing.T)      {}
func TestCollectStrings_FlagWithHyphenValueAndSeparator(t *testing.T)         {}
func TestCollectStrings_FlagTerminatorDoubleHyphenSeparator(t *testing.T)     {}
func TestCollectStrings_PositionalArgsAfterSeparator(t *testing.T)            {}
func TestCollectStrings_UnknownFlagsAfterSeparator(t *testing.T)              {}
func TestCollectStrings_NonFlagArgumentsBeforeAndAfterSeparator(t *testing.T) {}
