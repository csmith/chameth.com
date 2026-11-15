package atproto

import (
	"time"
)

const base32 = "234567abcdefghijklmnopqrstuvwxyz"

// TID returns the timestamp identifier for the given time and clock ID,
// suitable as a record key. The same inputs always give the same TID.
func TID(t time.Time, clockID uint) string {
	// > The top bit is always 0
	// > The next 53 bits represent microseconds since the UNIX epoch. 53 bits is chosen as the
	// > maximum safe integer precision in a 64-bit floating point number, as used by Javascript.
	// > The final 10 bits are a random "clock identifier."
	value := (uint64(t.UnixMicro()&0x1F_FFFF_FFFF_FFFF) << 10) | uint64(clockID&0x3FF)
	out := make([]byte, 13)
	for i := range 13 {
		out[12-i] = base32[value&0x1F]
		value >>= 5
	}
	return string(out)
}
