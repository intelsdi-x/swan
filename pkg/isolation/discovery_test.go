package isolation

import (
	"testing"
	"fmt"
)

func TestRemote(t *testing.T) {
	topol := topology{name: "Hello"}
	
	fmt.Printf(topol.name)

}
