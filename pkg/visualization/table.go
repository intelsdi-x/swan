package visualization

import (
	"github.com/olekukonko/tablewriter"
	"os"
)

// Table is a model for data.
type Table struct {
	headers []string
	data    [][]string
}

// NewTable creates new model of data representation.
func NewTable(headers []string, data [][]string) *Table {
	return &Table{
		headers,
		data,
	}
}

// DrawTable draws a struct with headers and data rows.
func DrawTable(table *Table) error {
	output := tablewriter.NewWriter(os.Stdout)
	output.SetHeader(table.headers)
	for _, v := range table.data {
		output.Append(v)
	}
	output.Render()
	return nil
}
