package visualization

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

// List is a model for data.
type List struct {
	elements []string
	label    string
}

// NewList creates new model of data representation.
func NewList(elements []string, label string) *List {
	return &List{
		elements,
		label,
	}
}

// ExperimentMetadata is a model for data.
type ExperimentMetadata struct {
	experimentID string
}

// NewExperimentMetadata creates new model of data representation.
func NewExperimentMetadata(ID string) *ExperimentMetadata {
	return &ExperimentMetadata{
		ID,
	}
}
