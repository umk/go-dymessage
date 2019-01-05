package memory

import (
	"encoding/binary"
	"unsafe"
)

var byteOrder binary.ByteOrder

// GetByteOrder gets the byte order for the system the application runs on.
func GetByteOrder() binary.ByteOrder { return byteOrder }

func init() {
	// Initialize the value, returned by GetByteOrder function.
	data := 1
	ptr := unsafe.Pointer(&data)
	if *(*byte)(ptr) == 0 {
		byteOrder = binary.BigEndian
	} else {
		byteOrder = binary.LittleEndian
	}
}
