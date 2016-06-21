package visualization

import (
	"fmt"
)

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

// PrintList prints elements from list.
func PrintList(list *List) {
	for _, value := range list.elements {
		fmt.Println(list.label + value)
	}
}
