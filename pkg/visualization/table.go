package visualization

import (
	"github.com/olekukonko/tablewriter"
	"os"
)

// Table is a convenience structure to encode headers, rows and columns to be
// printed with olekukonko's tablewriter.
type Table struct {
	headers []string
	data    [][]string
}

// NewTable is a constructor for a Table structure with the provided headers
// and table cells.
func NewTable(headers []string, data [][]string) *Table {
	return &Table{
		headers,
		data,
	}
}

// Draw renders the text table to stdout.
func (table *Table) Draw() {
	output := tablewriter.NewWriter(os.Stdout)
	output.SetHeader(table.headers)
	for _, v := range table.data {
		output.Append(v)
	}
	output.Render()
}
