package table

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/go-go-golems/glazed/pkg/helpers/templating"
	"github.com/go-go-golems/glazed/pkg/types"
	"sort"
	"strings"
	"text/template"
)

func PreserveColumnOrder(oldColumns []types.FieldName, newColumns map[types.FieldName]interface{}) []types.FieldName {
	seenRetColumns := map[types.FieldName]interface{}{}
	retColumns := []types.FieldName{}

	// preserve previous columns order as best as possible
	for _, column := range oldColumns {
		if _, ok := newColumns[column]; ok {
			retColumns = append(retColumns, column)
			seenRetColumns[column] = nil
		}
	}
	for key := range newColumns {
		if _, ok := seenRetColumns[key]; !ok {
			retColumns = append(retColumns, key)
			seenRetColumns[key] = nil
		}
	}
	return retColumns
}

type FlattenObjectMiddleware struct {
}

func NewFlattenObjectMiddleware() *FlattenObjectMiddleware {
	return &FlattenObjectMiddleware{}
}

func (fom *FlattenObjectMiddleware) Process(ctx context.Context, table *types.Table) (*types.Table, error) {
	ret := &types.Table{
		Columns: []types.FieldName{},
		Rows:    []types.Row{},
	}

	newColumns := map[types.FieldName]interface{}{}

	for _, row := range table.Rows {
		newValues := FlattenMapIntoColumns(row)
		newRow := newValues

		for pair := newValues.Oldest(); pair != nil; pair = pair.Next() {
			newColumns[pair.Key] = nil
		}
		ret.Rows = append(ret.Rows, newRow)
	}

	ret.Columns = PreserveColumnOrder(table.Columns, newColumns)

	return ret, nil
}

func FlattenMapIntoColumns(rows types.Row) types.Row {
	ret := types.NewMapRow()

	for pair := rows.Oldest(); pair != nil; pair = pair.Next() {
		key, value := pair.Key, pair.Value
		switch v := value.(type) {
		case types.Row:
			newColumns := FlattenMapIntoColumns(v)
			for pair_ := newColumns.Oldest(); pair_ != nil; pair_ = pair_.Next() {
				k, v := pair_.Key, pair_.Value
				ret.Set(fmt.Sprintf("%s.%s", key, k), v)
			}
		default:
			ret.Set(key, v)
		}
	}

	return ret
}

type PreserveColumnOrderMiddleware struct {
	columns []types.FieldName
}

func NewPreserveColumnOrderMiddleware(columns []types.FieldName) *PreserveColumnOrderMiddleware {
	return &PreserveColumnOrderMiddleware{
		columns: columns,
	}
}

func (scm *PreserveColumnOrderMiddleware) Process(ctx context.Context, table *types.Table) (*types.Table, error) {
	columnHash := map[types.FieldName]interface{}{}
	for _, column := range scm.columns {
		columnHash[column] = nil
	}

	table.Columns = PreserveColumnOrder(table.Columns, columnHash)
	return table, nil
}

type ReorderColumnOrderMiddleware struct {
	columns []types.FieldName
}

func NewReorderColumnOrderMiddleware(columns []types.FieldName) *ReorderColumnOrderMiddleware {
	return &ReorderColumnOrderMiddleware{
		columns: columns,
	}
}

func (scm *ReorderColumnOrderMiddleware) Process(ctx context.Context, table *types.Table) (*types.Table, error) {
	existingColumns := map[types.FieldName]interface{}{}
	for _, column := range table.Columns {
		existingColumns[column] = nil
	}

	seenColumns := map[types.FieldName]interface{}{}
	newColumns := []types.FieldName{}

	for _, column := range scm.columns {
		if strings.HasSuffix(column, ".") {
			for _, existingColumn := range table.Columns {
				if strings.HasPrefix(existingColumn, column) {
					if _, ok := seenColumns[existingColumn]; !ok {
						newColumns = append(newColumns, existingColumn)
						seenColumns[existingColumn] = nil
					}
				}
			}
		} else {
			if _, ok := seenColumns[column]; !ok {
				if _, ok := existingColumns[column]; ok {
					newColumns = append(newColumns, column)
					seenColumns[column] = nil
				}
			}

		}
	}

	for column := range existingColumns {
		if _, ok := seenColumns[column]; !ok {
			newColumns = append(newColumns, column)
			seenColumns[column] = nil
		}
	}

	table.Columns = newColumns

	return table, nil
}

type SortColumnsMiddleware struct {
}

func NewSortColumnsMiddleware() *SortColumnsMiddleware {
	return &SortColumnsMiddleware{}
}

func (scm *SortColumnsMiddleware) Process(ctx context.Context, table *types.Table) (*types.Table, error) {
	sort.Strings(table.Columns)
	return table, nil
}

type RowGoTemplateMiddleware struct {
	templates map[types.FieldName]*template.Template
	// this field is used to replace "." in keys before passing them to the template,
	// in order to avoid having to use the `index` template function to access fields
	// that contain a ".", which is frequent due to flattening.
	RenameSeparator string
}

// NewRowGoTemplateMiddleware creates a new RowGoTemplateMiddleware
// which is the simplest go template middleware.
//
// It will render the template for each row and return the result as a new column called with
// the given title.
//
// Because nested objects will be flattened to individual columns using the . separator,
// this will make fields inaccessible to the template. One way around this is to use
// {{ index . "field.subfield" }} in the template. Another is to pass a separator rename
// option.
//
// TODO(manuel, 2023-02-02) Add support for passing in custom funcmaps
// See #110 https://github.com/go-go-golems/glazed/issues/110
func NewRowGoTemplateMiddleware(
	templateStrings map[types.FieldName]string,
	renameSeparator string) (*RowGoTemplateMiddleware, error) {

	templates := map[types.FieldName]*template.Template{}
	for columnName, templateString := range templateStrings {
		tmpl, err := template.New("row").
			Funcs(sprig.TxtFuncMap()).
			Funcs(templating.TemplateFuncs).
			Parse(templateString)
		if err != nil {
			return nil, err
		}
		templates[columnName] = tmpl
	}

	return &RowGoTemplateMiddleware{
		templates:       templates,
		RenameSeparator: renameSeparator,
	}, nil
}

func (rgtm *RowGoTemplateMiddleware) Process(ctx context.Context, table *types.Table) (*types.Table, error) {
	ret := &types.Table{
		Columns: []types.FieldName{},
		Rows:    []types.Row{},
	}

	columnRenames := map[types.FieldName]types.FieldName{}
	existingColumns := map[types.FieldName]interface{}{}
	newColumns := map[types.FieldName]interface{}{}

	for _, columnName := range table.Columns {
		existingColumns[columnName] = nil
		ret.Columns = append(ret.Columns, columnName)
	}

	for _, row := range table.Rows {
		newRow := row

		templateValues := map[string]interface{}{}

		for pair := newRow.Oldest(); pair != nil; pair = pair.Next() {
			key, value := pair.Key, pair.Value

			if rgtm.RenameSeparator != "" {
				if _, ok := columnRenames[key]; !ok {
					columnRenames[key] = strings.ReplaceAll(key, ".", rgtm.RenameSeparator)
				}
			} else {
				columnRenames[key] = key
			}
			newKey := columnRenames[key]
			templateValues[newKey] = value
		}
		templateValues["_row"] = templateValues

		for columnName, tmpl := range rgtm.templates {
			var buf bytes.Buffer
			err := tmpl.Execute(&buf, templateValues)
			if err != nil {
				return nil, err
			}
			s := buf.String()

			// we need to handle the fact that some rows might not have all the keys, and thus
			// avoid counting columns as existing twice
			if _, ok := newColumns[columnName]; !ok {
				newColumns[columnName] = nil
				ret.Columns = append(ret.Columns, columnName)
			}
			newRow.Set(columnName, s)
		}

		ret.Rows = append(ret.Rows, newRow)
	}

	return ret, nil
}
