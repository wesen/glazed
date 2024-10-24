package settings

import (
	_ "embed"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/formatters/simple"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/middlewares/object"
	"github.com/go-go-golems/glazed/pkg/middlewares/row"
	"github.com/go-go-golems/glazed/pkg/middlewares/table"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io"
)

// Helpers for cobra commands

type GlazedParameterLayers struct {
	FieldsFiltersParameterLayer *FieldsFiltersParameterLayer `yaml:"fieldsFiltersParameterLayer"`
	OutputParameterLayer        *OutputParameterLayer        `yaml:"outputParameterLayer"`
	RenameParameterLayer        *RenameParameterLayer        `yaml:"renameParameterLayer"`
	ReplaceParameterLayer       *ReplaceParameterLayer       `yaml:"replaceParameterLayer"`
	SelectParameterLayer        *SelectParameterLayer        `yaml:"selectParameterLayer"`
	TemplateParameterLayer      *TemplateParameterLayer      `yaml:"templateParameterLayer"`
	JqParameterLayer            *JqParameterLayer            `yaml:"jqParameterLayer"`
	SortParameterLayer          *SortParameterLayer          `yaml:"sortParameterLayer"`
	SkipLimitParameterLayer     *SkipLimitParameterLayer     `yaml:"skipLimitParameterLayer"`
}

func (g *GlazedParameterLayers) MarshalYAML() (interface{}, error) {
	return &layers.ParameterLayerImpl{
		Name:        g.GetName(),
		Slug:        g.GetSlug(),
		Description: g.GetDescription(),
		Prefix:      g.GetPrefix(),
		ChildLayers: []layers.ParameterLayer{
			g.FieldsFiltersParameterLayer,
			g.OutputParameterLayer,
			g.RenameParameterLayer,
			g.ReplaceParameterLayer,
			g.SelectParameterLayer,
			g.TemplateParameterLayer,
			g.JqParameterLayer,
			g.SortParameterLayer,
		},
	}, nil
}

func (g *GlazedParameterLayers) GetName() string {
	return "Glazed Flags"
}

func (g *GlazedParameterLayers) GetSlug() string {
	return "glazed"
}

func (g *GlazedParameterLayers) GetDescription() string {
	return "Glazed flags"
}

func (g *GlazedParameterLayers) GetPrefix() string {
	return g.FieldsFiltersParameterLayer.GetPrefix()
}

func (g *GlazedParameterLayers) AddFlag(*parameters.ParameterDefinition) {
	panic("not supported me")
}

func (g *GlazedParameterLayers) GetParameterDefinitions() map[string]*parameters.ParameterDefinition {
	ret := make(map[string]*parameters.ParameterDefinition)
	for k, v := range g.RenameParameterLayer.GetParameterDefinitions() {
		ret[k] = v
	}

	for k, v := range g.OutputParameterLayer.GetParameterDefinitions() {
		ret[k] = v
	}

	for k, v := range g.SelectParameterLayer.GetParameterDefinitions() {
		ret[k] = v
	}

	for k, v := range g.TemplateParameterLayer.GetParameterDefinitions() {
		ret[k] = v
	}

	for k, v := range g.FieldsFiltersParameterLayer.GetParameterDefinitions() {
		ret[k] = v
	}

	for k, v := range g.ReplaceParameterLayer.GetParameterDefinitions() {
		ret[k] = v
	}

	for k, v := range g.JqParameterLayer.GetParameterDefinitions() {
		ret[k] = v
	}

	for k, v := range g.SortParameterLayer.GetParameterDefinitions() {
		ret[k] = v
	}

	for k, v := range g.SkipLimitParameterLayer.GetParameterDefinitions() {
		ret[k] = v
	}

	return ret
}

func (g *GlazedParameterLayers) AddFlagsToCobraCommand(cmd *cobra.Command) error {
	err := g.OutputParameterLayer.AddFlagsToCobraCommand(cmd)
	if err != nil {
		return err
	}
	err = g.FieldsFiltersParameterLayer.AddFlagsToCobraCommand(cmd)
	if err != nil {
		return err
	}
	err = g.SelectParameterLayer.AddFlagsToCobraCommand(cmd)
	if err != nil {
		return err
	}
	err = g.TemplateParameterLayer.AddFlagsToCobraCommand(cmd)
	if err != nil {
		return err
	}
	err = g.RenameParameterLayer.AddFlagsToCobraCommand(cmd)
	if err != nil {
		return err
	}
	err = g.ReplaceParameterLayer.AddFlagsToCobraCommand(cmd)
	if err != nil {
		return err
	}
	err = g.JqParameterLayer.AddFlagsToCobraCommand(cmd)
	if err != nil {
		return err
	}
	err = g.SortParameterLayer.AddFlagsToCobraCommand(cmd)
	if err != nil {
		return err
	}
	err = g.SkipLimitParameterLayer.AddFlagsToCobraCommand(cmd)
	if err != nil {
		return err
	}

	return nil
}

func (g *GlazedParameterLayers) ParseFlagsFromCobraCommand(cmd *cobra.Command) (map[string]interface{}, error) {
	ps, err := g.OutputParameterLayer.ParseFlagsFromCobraCommand(cmd)
	if err != nil {
		return nil, err
	}
	ps_, err := g.SelectParameterLayer.ParseFlagsFromCobraCommand(cmd)
	if err != nil {
		return nil, err
	}
	for k, v := range ps_ {
		ps[k] = v
	}
	ps_, err = g.RenameParameterLayer.ParseFlagsFromCobraCommand(cmd)
	if err != nil {
		return nil, err
	}
	for k, v := range ps_ {
		ps[k] = v
	}
	ps_, err = g.TemplateParameterLayer.ParseFlagsFromCobraCommand(cmd)
	if err != nil {
		return nil, err
	}
	for k, v := range ps_ {
		ps[k] = v
	}
	ps_, err = g.FieldsFiltersParameterLayer.ParseFlagsFromCobraCommand(cmd)
	if err != nil {
		return nil, err
	}
	for k, v := range ps_ {
		ps[k] = v
	}
	ps_, err = g.ReplaceParameterLayer.ParseFlagsFromCobraCommand(cmd)
	if err != nil {
		return nil, err
	}
	for k, v := range ps_ {
		ps[k] = v
	}
	ps_, err = g.JqParameterLayer.ParseFlagsFromCobraCommand(cmd)
	if err != nil {
		return nil, err
	}
	for k, v := range ps_ {
		ps[k] = v
	}
	ps_, err = g.SortParameterLayer.ParseFlagsFromCobraCommand(cmd)
	if err != nil {
		return nil, err
	}
	for k, v := range ps_ {
		ps[k] = v
	}
	ps_, err = g.SkipLimitParameterLayer.ParseFlagsFromCobraCommand(cmd)
	if err != nil {
		return nil, err
	}
	for k, v := range ps_ {
		ps[k] = v
	}

	return ps, nil
}

func (g *GlazedParameterLayers) ParseFlagsFromJSON(m map[string]interface{}, onlyProvided bool) (map[string]interface{}, error) {
	ps, err := g.OutputParameterLayer.ParseFlagsFromJSON(m, onlyProvided)
	if err != nil {
		return nil, err
	}
	ps_, err := g.SelectParameterLayer.ParseFlagsFromJSON(m, onlyProvided)
	if err != nil {
		return nil, err
	}
	for k, v := range ps_ {
		ps[k] = v
	}
	ps_, err = g.RenameParameterLayer.ParseFlagsFromJSON(m, onlyProvided)
	if err != nil {
		return nil, err
	}
	for k, v := range ps_ {
		ps[k] = v
	}
	ps_, err = g.TemplateParameterLayer.ParseFlagsFromJSON(m, onlyProvided)
	if err != nil {
		return nil, err
	}
	for k, v := range ps_ {
		ps[k] = v
	}
	ps_, err = g.FieldsFiltersParameterLayer.ParseFlagsFromJSON(m, onlyProvided)
	if err != nil {
		return nil, err
	}
	for k, v := range ps_ {
		ps[k] = v
	}
	ps_, err = g.ReplaceParameterLayer.ParseFlagsFromJSON(m, onlyProvided)
	if err != nil {
		return nil, err
	}
	for k, v := range ps_ {
		ps[k] = v
	}
	ps_, err = g.JqParameterLayer.ParseFlagsFromJSON(m, onlyProvided)
	if err != nil {
		return nil, err
	}
	for k, v := range ps_ {
		ps[k] = v
	}
	ps_, err = g.SortParameterLayer.ParseFlagsFromJSON(m, onlyProvided)
	if err != nil {
		return nil, err
	}
	for k, v := range ps_ {
		ps[k] = v
	}
	ps_, err = g.SkipLimitParameterLayer.ParseFlagsFromJSON(m, onlyProvided)
	if err != nil {
		return nil, err
	}
	for k, v := range ps_ {
		ps[k] = v
	}

	return ps, nil

}

func (g *GlazedParameterLayers) InitializeParameterDefaultsFromStruct(s interface{}) error {
	err := g.OutputParameterLayer.InitializeParameterDefaultsFromStruct(s)
	if err != nil {
		return err
	}
	err = g.FieldsFiltersParameterLayer.InitializeParameterDefaultsFromStruct(s)
	if err != nil {
		return err
	}

	err = g.SelectParameterLayer.InitializeParameterDefaultsFromStruct(s)
	if err != nil {
		return err
	}
	err = g.TemplateParameterLayer.InitializeParameterDefaultsFromStruct(s)
	if err != nil {
		return err
	}
	err = g.RenameParameterLayer.InitializeParameterDefaultsFromStruct(s)
	if err != nil {
		return err
	}
	err = g.ReplaceParameterLayer.InitializeParameterDefaultsFromStruct(s)
	if err != nil {
		return err
	}
	err = g.JqParameterLayer.InitializeParameterDefaultsFromStruct(s)
	if err != nil {
		return err
	}
	err = g.SortParameterLayer.InitializeParameterDefaultsFromStruct(s)
	if err != nil {
		return err
	}
	err = g.SkipLimitParameterLayer.InitializeParameterDefaultsFromStruct(s)
	if err != nil {
		return err
	}
	return nil
}

type GlazeParameterLayerOption func(*GlazedParameterLayers) error

func WithOutputParameterLayerOptions(options ...layers.ParameterLayerOptions) GlazeParameterLayerOption {
	return func(g *GlazedParameterLayers) error {
		for _, option := range options {
			err := option(g.OutputParameterLayer.ParameterLayerImpl)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WithSelectParameterLayerOptions(options ...layers.ParameterLayerOptions) GlazeParameterLayerOption {
	return func(g *GlazedParameterLayers) error {
		for _, option := range options {
			err := option(g.SelectParameterLayer.ParameterLayerImpl)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WithTemplateParameterLayerOptions(options ...layers.ParameterLayerOptions) GlazeParameterLayerOption {
	return func(g *GlazedParameterLayers) error {
		for _, option := range options {
			err := option(g.TemplateParameterLayer.ParameterLayerImpl)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WithRenameParameterLayerOptions(options ...layers.ParameterLayerOptions) GlazeParameterLayerOption {
	return func(g *GlazedParameterLayers) error {
		for _, option := range options {
			err := option(g.RenameParameterLayer.ParameterLayerImpl)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WithReplaceParameterLayerOptions(options ...layers.ParameterLayerOptions) GlazeParameterLayerOption {
	return func(g *GlazedParameterLayers) error {
		for _, option := range options {
			err := option(g.ReplaceParameterLayer.ParameterLayerImpl)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WithFieldsFiltersParameterLayerOptions(options ...layers.ParameterLayerOptions) GlazeParameterLayerOption {
	return func(g *GlazedParameterLayers) error {
		for _, option := range options {
			err := option(g.FieldsFiltersParameterLayer.ParameterLayerImpl)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WithJqParameterLayerOptions(options ...layers.ParameterLayerOptions) GlazeParameterLayerOption {
	return func(g *GlazedParameterLayers) error {
		for _, option := range options {
			err := option(g.JqParameterLayer.ParameterLayerImpl)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WithSortParameterLayerOptions(options ...layers.ParameterLayerOptions) GlazeParameterLayerOption {
	return func(g *GlazedParameterLayers) error {
		for _, option := range options {
			err := option(g.SortParameterLayer.ParameterLayerImpl)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WithSkipLimitParameterLayerOptions(options ...layers.ParameterLayerOptions) GlazeParameterLayerOption {
	return func(g *GlazedParameterLayers) error {
		for _, option := range options {
			err := option(g.SkipLimitParameterLayer.ParameterLayerImpl)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func NewGlazedParameterLayers(options ...GlazeParameterLayerOption) (*GlazedParameterLayers, error) {
	fieldsFiltersParameterLayer, err := NewFieldsFiltersParameterLayer()
	if err != nil {
		return nil, err
	}
	outputParameterLayer, err := NewOutputParameterLayer()
	if err != nil {
		return nil, err
	}
	renameParameterLayer, err := NewRenameParameterLayer()
	if err != nil {
		return nil, err
	}
	replaceParameterLayer, err := NewReplaceParameterLayer()
	if err != nil {
		return nil, err
	}
	selectParameterLayer, err := NewSelectParameterLayer()
	if err != nil {
		return nil, err
	}
	templateParameterLayer, err := NewTemplateParameterLayer()
	if err != nil {
		return nil, err
	}
	jqParameterLayer, err := NewJqParameterLayer()
	if err != nil {
		return nil, err
	}
	sortParameterLayer, err := NewSortParameterLayer()
	if err != nil {
		return nil, err
	}
	skipLimitParameterLayer, err := NewSkipLimitParameterLayer()
	if err != nil {
		return nil, err
	}
	ret := &GlazedParameterLayers{
		FieldsFiltersParameterLayer: fieldsFiltersParameterLayer,
		OutputParameterLayer:        outputParameterLayer,
		RenameParameterLayer:        renameParameterLayer,
		ReplaceParameterLayer:       replaceParameterLayer,
		SelectParameterLayer:        selectParameterLayer,
		TemplateParameterLayer:      templateParameterLayer,
		JqParameterLayer:            jqParameterLayer,
		SortParameterLayer:          sortParameterLayer,
		SkipLimitParameterLayer:     skipLimitParameterLayer,
	}

	for _, option := range options {
		err := option(ret)
		if err != nil {
			return nil, err
		}
	}

	return ret, nil
}

func SetupRowOutputFormatter(ps map[string]interface{}) (formatters.RowOutputFormatter, error) {
	outputSettings, err := NewOutputFormatterSettings(ps)
	if err != nil {
		return nil, err
	}

	of, err := outputSettings.CreateRowOutputFormatter()
	if err != nil {
		return nil, err
	}

	return of, nil
}

func SetupTableOutputFormatter(ps map[string]interface{}) (formatters.TableOutputFormatter, error) {
	selectSettings, err := NewSelectSettingsFromParameters(ps)
	if err != nil {
		return nil, err
	}

	outputSettings, err := NewOutputFormatterSettings(ps)
	if err != nil {
		return nil, err
	}

	var of formatters.TableOutputFormatter
	if selectSettings.SelectField != "" {
		of = simple.NewSingleColumnFormatter(
			selectSettings.SelectField,
			simple.WithSeparator(selectSettings.SelectSeparator),
			simple.WithOutputFile(outputSettings.OutputFile),
			simple.WithOutputMultipleFiles(outputSettings.OutputMultipleFiles),
			simple.WithOutputFileTemplate(outputSettings.OutputFileTemplate),
		)
	} else {
		of, err = outputSettings.CreateTableOutputFormatter()
		if err != nil {
			return nil, errors.Wrapf(err, "Error creating output formatter")
		}
	}
	return of, nil

}

// SetupTableProcessor processes all the glazed flags out of ps and returns a TableProcessor
// configured with all the necessary middlewares except for the output formatter.
//
// DO(manuel, 2023-06-30) It would be good to used a parsedLayer here, if we ever refactor that part
func SetupTableProcessor(ps map[string]interface{}, options ...middlewares.TableProcessorOption) (*middlewares.TableProcessor, error) {
	// TODO(manuel, 2023-03-06): This is where we should check that flags that are mutually incompatible don't clash
	//
	// See: https://github.com/go-go-golems/glazed/issues/199
	templateSettings, err := NewTemplateSettings(ps)
	if err != nil {
		return nil, err
	}
	selectSettings, err := NewSelectSettingsFromParameters(ps)
	if err != nil {
		return nil, err
	}
	renameSettings, err := NewRenameSettingsFromParameters(ps)
	if err != nil {
		return nil, err
	}
	fieldsFilterSettings, err := NewFieldsFilterSettings(ps)
	if err != nil {
		return nil, err
	}
	replaceSettings, err := NewReplaceSettingsFromParameters(ps)
	if err != nil {
		return nil, err
	}
	jqSettings, err := NewJqSettingsFromParameters(ps)
	if err != nil {
		return nil, err
	}
	sortSettings, err := NewSortSettingsFromParameters(ps)
	if err != nil {
		return nil, err
	}
	outputSettings, err := NewOutputFormatterSettings(ps)
	if err != nil {
		return nil, err
	}
	skipLimitSettings, err := NewSkipLimitSettingsFromParameters(ps)
	if err != nil {
		return nil, err
	}

	templateSettings.UpdateWithSelectSettings(selectSettings)

	gp := middlewares.NewTableProcessor(options...)

	// rename middlewares run first because they are used to clean up column names
	// for the following middlewares too.
	// these following middlewares can create proper column names on their own
	// when needed
	err = renameSettings.AddMiddlewares(gp)
	if err != nil {
		return nil, errors.Wrapf(err, "Error adding rename middlewares")
	}

	err = templateSettings.AddMiddlewares(gp)
	if err != nil {
		return nil, errors.Wrapf(err, "Error adding template middlewares")
	}

	if (outputSettings.Output == "json" || outputSettings.Output == "yaml") && outputSettings.FlattenObjects {
		mw := row.NewFlattenObjectMiddleware()
		gp.AddRowMiddlewareInFront(mw)
	}
	fieldsFilterSettings.AddMiddlewares(gp)

	err = replaceSettings.AddMiddlewares(gp)
	if err != nil {
		return nil, errors.Wrapf(err, "Error adding replace middlewares")
	}

	var middlewares_ []middlewares.ObjectMiddleware
	if !templateSettings.UseRowTemplates && len(templateSettings.Templates) > 0 {
		ogtm, err := object.NewTemplateMiddleware(templateSettings.Templates)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not process template argument")
		}
		middlewares_ = append(middlewares_, ogtm)
	}

	jqObjectMiddleware, jqTableMiddleware, err := NewJqMiddlewaresFromSettings(jqSettings)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not create jq middlewares")
	}

	if jqObjectMiddleware != nil {
		middlewares_ = append(middlewares_, jqObjectMiddleware)
	}

	if jqTableMiddleware != nil {
		gp.AddTableMiddleware(jqTableMiddleware)
	}

	// NOTE(manuel, 2023-03-20): We need to figure out how to order middlewares, on the command line.
	// This is not possible with cobra, which doesn't have ordering of flags, and adding that
	// to the API that we currently use (which is a unordered hashmap, and parsed layers that lose the positioning)
	// is not trivial.
	sortSettings.AddMiddlewares(gp)

	if skipLimitSettings.Skip != 0 || skipLimitSettings.Limit != 0 {
		gp.AddRowMiddleware(&row.SkipLimitMiddleware{
			Skip:  skipLimitSettings.Skip,
			Limit: skipLimitSettings.Limit,
		})
	}

	gp.AddObjectMiddleware(middlewares_...)

	return gp, nil
}

// SetupProcessorOutput creates a new Output middleware (either row or table, depending on the format
// and the stream flag set in ps) and adds it to the TableProcessor. Additional middlewares required by ]
// the chosen output format might be added as well (for example, flattening rows when using table-oriented
// output formats).
//
// It also returns the output formatter that was created.
func SetupProcessorOutput(gp *middlewares.TableProcessor, ps map[string]interface{}, w io.Writer) (formatters.OutputFormatter, error) {
	// first, try to get a row updater
	rowOf, err := SetupRowOutputFormatter(ps)

	if rowOf != nil {
		err = rowOf.RegisterRowMiddlewares(gp)
		if err != nil {
			return nil, err
		}
		gp.AddRowMiddleware(row.NewOutputMiddleware(rowOf, w))
		return rowOf, nil
	} else {
		if _, ok := err.(*ErrorRowFormatUnsupported); !ok {
			return nil, err
		}

		of, err := SetupTableOutputFormatter(ps)
		if err != nil {
			return nil, err
		}
		err = of.RegisterTableMiddlewares(gp)
		if err != nil {
			return nil, err
		}

		gp.AddTableMiddleware(table.NewOutputMiddleware(of, w))

		return of, nil
	}
}
