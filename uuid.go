package uuid

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"slices"
)

var (
	errInvalidURNPrefix = errors.New("invalid URN prefix")
	errInvalidFormat    = errors.New("invalid UUID format")
	errInvalidLength    = errors.New("invalid length")

	max = UUID{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
	}
)

// UUID is a 128 bit Universal Unique IDentifier as defined in RFC 4122.
// Since UUID is a concrete, non-pointer type, all representable values of UUID are
// technically valid UUIDs. It is up to user to check if versioning/variant/time information is valid.
type UUID [16]byte

// Version represents a UUID's version. This value is numerical in representation i.e:
// Version of 4 corresponds to UUID Version 4.
type Version byte

// Zero returns the zero UUID (all 0's).
func Zero() UUID { return UUID{} }

// Max returns the maximum UUID (all ff's).
func Max() UUID { return max }

// Time returns the time in 100s of nanoseconds since 15 Oct 1582 encoded in
// uuid.  The time is only defined for version 1, 2, 6 and 7 UUIDs.
func (uuid UUID) Time() Time {
	var t Time
	switch uuid.Version() {
	case 6:
		time := binary.BigEndian.Uint64(uuid[:8]) // Ignore uuid[6] version b0110
		t = Time(time)
	case 7:
		time := binary.BigEndian.Uint64(uuid[:8])
		t = Time((time>>16)*10000 + g1582ns100)
	default: // forward compatible
		time := int64(binary.BigEndian.Uint32(uuid[0:4]))
		time |= int64(binary.BigEndian.Uint16(uuid[4:6])) << 32
		time |= int64(binary.BigEndian.Uint16(uuid[6:8])&0xfff) << 48
		t = Time(time)
	}
	return t
}

func Parse(s string) (uuid UUID, err error) {
	switch len(s) {
	case 32, 36, 36 + 2, 36 + 9:
		return ParseBytes([]byte(s))
	}
	return uuid, errInvalidFormat
}

var urnPfx = []byte("urn:uuid:")

func ParseBytes(b []byte) (uuid UUID, err error) {
	if len(b) < 32 {
		return uuid, errInvalidLength
	}
	// First trim excess characters.
	switch len(b) {
	case 32, 36:
		err = uuid.UnmarshalText(b)
		return uuid, err
	case 36 + 2:
		if b[0] == '{' && b[37] == '}' {
			err = uuid.UnmarshalText(b[1:37])
			return uuid, err
		}
	case 36 + 9:
		if !bytes.EqualFold(b[:9], urnPfx) {
			return uuid, errInvalidURNPrefix
		}
		err = uuid.UnmarshalText(b[9:])
		return uuid, err
	}
	return uuid, errInvalidFormat
}

// String returns the string form of uuid, xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
// , or "" if uuid is invalid.
func (uuid UUID) String() string {
	var buf [36]byte
	encodeHex(buf[:], uuid)
	return string(buf[:])
}

// IsZero checks if the UUID is all zeros.
// Can be used to check if uninitialized but is not a validity check for some versions.
func (uuid UUID) IsZero() bool {
	return uuid == (UUID{})
}

// IsMax checks if the UUID is all 0xff's (all max value byte).
func (uuid UUID) IsMax() bool {
	return uuid == max
}

// Version returns the version bits of the uuid.
func (uuid UUID) Version() Version {
	return Version(uuid[6] >> 4)
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (uuid *UUID) UnmarshalText(data []byte) error {
	s := data
	switch len(s) {
	case 32: // xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
		var ok bool
		for i := range uuid {
			uuid[i], ok = xtob(s[i*2], s[i*2+1])
			if !ok {
				return errInvalidFormat
			}
		}
		return nil
	case 36: // xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
		if s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
			return errInvalidFormat
		}
		var ok bool
		for i, x := range [16]int{
			0, 2, 4, 6,
			9, 11,
			14, 16,
			19, 21,
			24, 26, 28, 30, 32, 34,
		} {
			uuid[i], ok = xtob(s[x], s[x+1])
			if !ok {
				return errInvalidFormat
			}
		}
		return nil
	default:
		return errInvalidLength
	}
}

// Encode36 encodes the 36 byte hexadecimal representation of the UUID
// into the first 36 bytes of the buffer.
// i.e: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
func (uuid UUID) Encode36(b []byte) error {
	if len(b) < 36 {
		return io.ErrShortBuffer
	}
	encodeHex(b[len(b)-36:], uuid)
	return nil
}

// Encode36 encodes the 32 byte hexadecimal representation of the UUID
// into the first 32 bytes of the buffer.
// i.e: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
func (uuid UUID) Encode32(b []byte) error {
	if len(b) < 32 {
		return io.ErrShortBuffer
	}
	hex.Encode(b[:32], uuid[:])
	return nil
}

// AppendText implements encoding.TextAppender.
func (uuid UUID) AppendText(b []byte) ([]byte, error) {
	b = slices.Grow(b, 36)
	b = b[:len(b)+36]
	encodeHex(b[len(b)-36:], uuid)
	return b, nil
}

// MarshalText implements encoding.TextMarshaler.
func (uuid UUID) MarshalText() ([]byte, error) {
	var buf [36]byte
	encodeHex(buf[:], uuid)
	return buf[:], nil
}

func encodeHex(dst []byte, uuid UUID) {
	hex.Encode(dst, uuid[:4])
	dst[8] = '-'
	hex.Encode(dst[9:13], uuid[4:6])
	dst[13] = '-'
	hex.Encode(dst[14:18], uuid[6:8])
	dst[18] = '-'
	hex.Encode(dst[19:23], uuid[8:10])
	dst[23] = '-'
	hex.Encode(dst[24:], uuid[10:])
}

// ID returns the id for a Version 2 UUID. IDs are only defined for Version 2 UUIDs.
func (uuid UUID) IDv2() uint32 {
	return binary.BigEndian.Uint32(uuid[0:4])
}

// HasTime checks if the UUID version has [Time] field. See [UUID.Time].
func (v Version) HasTime() bool {
	switch v {
	case 1, 2, 6, 7:
		return true
	}
	return false
}
