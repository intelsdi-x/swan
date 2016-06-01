package visualization

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"os"
)

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

// PrintList prints elements from list.
func PrintList(list []string) {
	for _, value := range list {
		fmt.Println(value)
	}
}

// PrintExperimentMetadata prints elements from list.
func PrintExperimentMetadata(experimentMetadata *ExperimentMetadata) {
	fmt.Println("\nExperiment id: " + experimentMetadata.experimentID)
}
