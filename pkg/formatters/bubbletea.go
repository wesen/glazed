package formatters

import (
	"encoding/json"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

// TODO(2023-02-17, manuel) Refactor the formatters so that they don't all have to do the middleware and row thing

type BubbleTeaOutputFormatter struct {
	Table       *types.Table
	middlewares []middlewares.TableMiddleware
}

func (J *BubbleTeaOutputFormatter) AddRow(row types.Row) {
	J.Table.Rows = append(J.Table.Rows, row)
}

func (f *BubbleTeaOutputFormatter) SetColumnOrder(columns []types.FieldName) {
	f.Table.Columns = columns
}

func (J *BubbleTeaOutputFormatter) AddTableMiddleware(mw middlewares.TableMiddleware) {
	J.middlewares = append(J.middlewares, mw)
}

func (J *BubbleTeaOutputFormatter) AddTableMiddlewareInFront(mw middlewares.TableMiddleware) {
	J.middlewares = append([]middlewares.TableMiddleware{mw}, J.middlewares...)
}

func (J *BubbleTeaOutputFormatter) AddTableMiddlewareAtIndex(i int, mw middlewares.TableMiddleware) {
	J.middlewares = append(J.middlewares[:i], append([]middlewares.TableMiddleware{mw}, J.middlewares[i:]...)...)
}

func (J *BubbleTeaOutputFormatter) Output() (string, error) {
	for _, middleware := range J.middlewares {
		newTable, err := middleware.Process(J.Table)
		if err != nil {
			return "", err
		}
		J.Table = newTable
	}

	// TODO(manuel, 2022-11-21) We should build a custom BubbleTeaMarshal for Table
	var rows []map[string]interface{}
	for _, row := range J.Table.Rows {
		rows = append(rows, row.GetValues())
	}
	jsonBytes, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func NewBubbleTeaOutputFormatter(outputAsObjects bool) *BubbleTeaOutputFormatter {
	return &BubbleTeaOutputFormatter{
		Table:       types.NewTable(),
		middlewares: []middlewares.TableMiddleware{},
	}
}
