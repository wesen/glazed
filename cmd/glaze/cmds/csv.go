package cmds

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/helpers"
	"github.com/pkg/errors"
	"os"
)

type CsvCommand struct {
	description *cmds.CommandDescription
}

func NewCsvCommand() (*CsvCommand, error) {
	glazedParameterLayer, err := cli.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	return &CsvCommand{
		description: cmds.NewCommandDescription(
			"csv",
			cmds.WithShort("Format CSV files"),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"input-files",
					parameters.ParameterTypeStringList,
					parameters.WithRequired(true),
				),
			),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"delimiter",
					parameters.ParameterTypeString,
					parameters.WithHelp("delimiter to use"),
					parameters.WithDefault(","),
				),
				parameters.NewParameterDefinition(
					"comment",
					parameters.ParameterTypeString,
					parameters.WithHelp("comment character to use"),
					parameters.WithDefault("#"),
				),
				parameters.NewParameterDefinition(
					"fields-per-record",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("number of fields per record (negative to disable)"),
					parameters.WithDefault(0),
				),
				parameters.NewParameterDefinition(
					"trim-leading-space",
					parameters.ParameterTypeBool,
					parameters.WithHelp("trim leading space"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"lazy-quotes",
					parameters.ParameterTypeBool,
					parameters.WithHelp("allow lazy quotes"),
					parameters.WithDefault(false),
				),
			),
			cmds.WithLayers(
				glazedParameterLayer,
			),
		),
	}, nil
}

func (c *CsvCommand) Description() *cmds.CommandDescription {
	return c.description
}

func (c *CsvCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	gp cmds.Processor,
) error {
	inputFiles, ok := ps["input-files"].([]string)
	if !ok {
		return errors.New("input-files argument is not a string list")
	}

	comma, _ := ps["delimiter"].(string)
	if len(comma) != 1 {
		return errors.New("delimiter must be a single character")
	}
	commaRune := rune(comma[0])

	comment, _ := ps["comment"].(string)
	if len(comment) != 1 {
		return errors.New("comment must be a single character")
	}
	commentRune := rune(comment[0])

	fieldsPerRecord, _ := ps["fields-per-record"].(int)
	trimLeadingSpace, _ := ps["trim-leading-space"].(bool)
	lazyQuotes, _ := ps["lazy-quotes"].(bool)

	options := []helpers.ParseCSVOption{
		helpers.WithComma(commaRune),
		helpers.WithComment(commentRune),
		helpers.WithFieldsPerRecord(fieldsPerRecord),
		helpers.WithTrimLeadingSpace(trimLeadingSpace),
		helpers.WithLazyQuotes(lazyQuotes),
	}

	for _, arg := range inputFiles {
		if arg == "-" {
			arg = "/dev/stdin"
		}

		// open arg and create a reader
		f, err := os.Open(arg)
		if err != nil {
			return errors.Wrap(err, "could not open file")
		}
		defer f.Close()

		s, err := helpers.ParseCSV(f, options...)
		if err != nil {
			return errors.Wrap(err, "could not parse CSV file")
		}

		for _, row := range s {
			err = gp.ProcessInputObject(row)
			if err != nil {
				return errors.Wrap(err, "could not process CSV row")
			}
		}
	}

	return nil
}