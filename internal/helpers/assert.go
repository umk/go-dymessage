package helpers

import (
	"fmt"
	"github.com/umk/go-dymessage"
)

// DataTypesMustMatch checks whether the dynamic entity and message definition
// types match, assuming that the entity has been created for this particular
// type. If they doesn't match, the function panics.
func DataTypesMustMatch(e *dymessage.Entity, def *dymessage.MessageDef) {
	if e.DataType != def.DataType {
		message := fmt.Sprintf(
			"dynamic entity must have the data type %d, but found %d",
			def.DataType, e.DataType)
		panic(message)
	}
}
