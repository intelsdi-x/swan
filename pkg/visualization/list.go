package visualization

// List is a convenience structure to encode a list of items attached with a
// certain label.
type List struct {
	elements []string
	label    string
}

// NewList is a constructor for a List structure with a certain list of
// elements attached to the specified label.
func NewList(elements []string, label string) *List {
	return &List{
		elements,
		label,
	}
}

// String returns a printable string representation of the list.
func (list *List) String() string {
	output := ""
	for _, value := range list.elements {
		output += list.label + value + "\n"
	}
	return output
}
