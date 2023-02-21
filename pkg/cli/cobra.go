package cli

import (
	_ "embed"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Helpers for cobra commands

type FlagsDefaults struct {
	Output       *OutputFlagsDefaults
	Select       *SelectFlagsDefaults
	Rename       *RenameFlagsDefaults
	Template     *TemplateFlagsDefaults
	FieldsFilter *FieldsFilterFlagsDefaults
	Replace      *ReplaceFlagsDefaults
}

func NewFlagsDefaults() *FlagsDefaults {
	return &FlagsDefaults{
		Output:       NewOutputFlagsDefaults(),
		Select:       NewSelectFlagsDefaults(),
		Rename:       NewRenameFlagsDefaults(),
		Template:     NewTemplateFlagsDefaults(),
		FieldsFilter: NewFieldsFilterFlagsDefaults(),
		Replace:      NewReplaceFlagsDefaults(),
	}
}

// AddFlags adds all the glazed processing layer flags to a cobra.Command
//
// TODO(manuel, 2023-02-21) Interfacing directly with cobra to do glazed flags handling is deprecated
// As we are moving towards #150 we should do all of this through the parameter definitions instead.
//
// This could probably be modelled as something of a "Layer" class that can be added to a command.
func AddFlags(cmd *cobra.Command, defaults *FlagsDefaults) error {
	err := AddOutputFlags(cmd, defaults.Output)
	if err != nil {
		return err
	}
	err = AddSelectFlags(cmd, defaults.Select)
	if err != nil {
		return err
	}
	err = AddRenameFlags(cmd, defaults.Rename)
	if err != nil {
		return err
	}
	err = AddTemplateFlags(cmd, defaults.Template)
	if err != nil {
		return err
	}
	err = AddFieldsFilterFlags(cmd, defaults.FieldsFilter)
	if err != nil {
		return err
	}
	err = AddReplaceFlags(cmd, defaults.Replace)
	if err != nil {
		return err
	}

	cmds.SetFlagGroupOrder(cmd, []string{
		"output",
		"select",
		"template",
		"fields-filter",
		"rename",
		"replace",
	})

	return nil
}

func SetupProcessor(cmd *cobra.Command) (*cmds.GlazeProcessor, formatters.OutputFormatter, error) {
	outputSettings, err := ParseOutputFlags(cmd)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error parsing output flags")
	}

	templateSettings, err := ParseTemplateFlags(cmd)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error parsing template flags")
	}

	fieldsFilterSettings, err := ParseFieldsFilterFlags(cmd)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error parsing fields filter flags")
	}

	selectSettings, err := ParseSelectFlags(cmd)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error parsing select flags")
	}
	outputSettings.UpdateWithSelectSettings(selectSettings)
	fieldsFilterSettings.UpdateWithSelectSettings(selectSettings)
	templateSettings.UpdateWithSelectSettings(selectSettings)

	renameSettings, err := ParseRenameFlags(cmd)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error parsing rename flags")
	}

	replaceSettings, err := ParseReplaceFlags(cmd)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error parsing replace flags")
	}

	of, err := outputSettings.CreateOutputFormatter()
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error creating output formatter")
	}

	// rename middlewares run first because they are used to clean up column names
	// for the following middlewares too.
	// these following middlewares can create proper column names on their own
	// when needed
	err = renameSettings.AddMiddlewares(of)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error adding rename middlewares")
	}

	err = templateSettings.AddMiddlewares(of)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error adding template middlewares")
	}

	if (outputSettings.Output == "json" || outputSettings.Output == "yaml") && outputSettings.FlattenObjects {
		mw := middlewares.NewFlattenObjectMiddleware()
		of.AddTableMiddleware(mw)
	}
	fieldsFilterSettings.AddMiddlewares(of)

	err = replaceSettings.AddMiddlewares(of)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error adding replace middlewares")
	}

	var middlewares_ []middlewares.ObjectMiddleware
	if !templateSettings.UseRowTemplates && len(templateSettings.Templates) > 0 {
		ogtm, err := middlewares.NewObjectGoTemplateMiddleware(templateSettings.Templates)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "Could not process template argument")
		}
		middlewares_ = append(middlewares_, ogtm)
	}

	gp := cmds.NewGlazeProcessor(of, middlewares_)
	return gp, of, nil
}
