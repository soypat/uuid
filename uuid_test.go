package uuid

import (
	"encoding/json"
	"testing"
)

// will crash if
var uuid = NewGeneratorV4()

func TestUUID_JSONMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name string
		uuid UUID
	}{
		{
			name: "zero UUID",
			uuid: UUID{},
		},
		{
			name: "max UUID",
			uuid: Max(),
		},
		{
			name: "random UUID",
			uuid: UUID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.uuid)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			// Verify output is quoted hexadecimal string
			str := string(data)
			if str[0] != '"' || str[len(str)-1] != '"' {
				t.Errorf("expected quoted string, got %s", str)
			}

			// Verify it's the expected hex format (36 chars with hyphens + 2 quotes)
			if len(str) != 38 {
				t.Errorf("expected length 38, got %d: %s", len(str), str)
			}

			// Unmarshal back
			var got UUID
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			if got != tt.uuid {
				t.Errorf("round-trip failed: got %v, want %v", got, tt.uuid)
			}
		})
	}
	id := uuid.MustRandom()
	_ = id
	id = uuid.MustHashed(id, id[:])
}

func TestUUID_JSONUnmarshalFormats(t *testing.T) {
	// Both 32-char and 36-char formats should work
	expected := UUID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "with hyphens",
			input: `"6ba7b810-9dad-11d1-80b4-00c04fd430c8"`,
		},
		{
			name:  "without hyphens",
			input: `"6ba7b8109dad11d180b400c04fd430c8"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got UUID
			if err := json.Unmarshal([]byte(tt.input), &got); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}
			if got != expected {
				t.Errorf("got %v, want %v", got, expected)
			}
		})
	}
}

func TestUUID_JSONUnmarshalInvalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "invalid hex char", input: `"6ba7b810-9dad-11d1-80b4-00c04fd430cg"`},
		{name: "wrong length", input: `"6ba7b810-9dad-11d1-80b4"`},
		{name: "wrong hyphen position", input: `"6ba7b8109-dad-11d1-80b4-00c04fd430c8"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got UUID
			if err := json.Unmarshal([]byte(tt.input), &got); err == nil {
				t.Errorf("expected error for input %s", tt.input)
			}
		})
	}
}

func TestParse(t *testing.T) {
	expected := UUID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}

	tests := []struct {
		name  string
		input string
	}{
		{name: "32 chars", input: "6ba7b8109dad11d180b400c04fd430c8"},
		{name: "braces", input: "{6ba7b810-9dad-11d1-80b4-00c04fd430c8}"},
		{name: "urn lowercase", input: "urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8"},
		{name: "URN uppercase", input: "URN:UUID:6ba7b810-9dad-11d1-80b4-00c04fd430c8"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}
			if got != expected {
				t.Errorf("got %v, want %v", got, expected)
			}
		})
	}
}

func TestParseBytes(t *testing.T) {
	expected := UUID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}

	tests := []struct {
		name  string
		input string
	}{
		{name: "32 chars", input: "6ba7b8109dad11d180b400c04fd430c8"},
		{name: "36 chars with hyphens", input: "6ba7b810-9dad-11d1-80b4-00c04fd430c8"},
		{name: "braces", input: "{6ba7b810-9dad-11d1-80b4-00c04fd430c8}"},
		{name: "urn prefix", input: "urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8"},
		{name: "uppercase hex", input: "6BA7B810-9DAD-11D1-80B4-00C04FD430C8"},
		{name: "mixed case", input: "6Ba7b810-9DaD-11d1-80B4-00c04FD430c8"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBytes([]byte(tt.input))
			if err != nil {
				t.Fatalf("ParseBytes failed: %v", err)
			}
			if got != expected {
				t.Errorf("got %v, want %v", got, expected)
			}
		})
	}
}

func TestParse_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "too short", input: "6ba7b810-9dad-11d1"},
		{name: "too long", input: "6ba7b810-9dad-11d1-80b4-00c04fd430c8xxxx"},
		{name: "invalid hex", input: "6ba7b810-9dad-11d1-80b4-00c04fd430cg"},
		{name: "wrong brace format", input: "[6ba7b810-9dad-11d1-80b4-00c04fd430c8]"},
		{name: "unclosed brace", input: "{6ba7b810-9dad-11d1-80b4-00c04fd430c8"},
		{name: "invalid urn prefix", input: "urn:guid:6ba7b810-9dad-11d1-80b4-00c04fd430c8"},
		{name: "empty string", input: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			if err == nil {
				t.Errorf("expected error for input %q", tt.input)
			}
		})
	}
}

func TestNewHashed_Consistent(t *testing.T) {
	gen := NewGeneratorV4()
	space := UUID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	data := []byte("test data")

	// Generate multiple times with same inputs
	first := gen.MustHashed(space, data)
	for i := 0; i < 100; i++ {
		got := gen.MustHashed(space, data)
		if got != first {
			t.Fatalf("iteration %d: got %v, want %v", i, got, first)
		}
	}
}

func TestNewHashed_DifferentData(t *testing.T) {
	gen := NewGeneratorV4()
	space := UUID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}

	id1 := gen.MustHashed(space, []byte("data1"))
	id2 := gen.MustHashed(space, []byte("data2"))

	if id1 == id2 {
		t.Errorf("different data should produce different UUIDs: %v", id1)
	}
}

func TestNewHashed_DifferentSpace(t *testing.T) {
	gen := NewGeneratorV4()
	space1 := UUID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	space2 := UUID{0x6b, 0xa7, 0xb8, 0x11, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	data := []byte("same data")

	id1 := gen.MustHashed(space1, data)
	id2 := gen.MustHashed(space2, data)

	if id1 == id2 {
		t.Errorf("different space should produce different UUIDs: %v", id1)
	}
}

func TestNewHashed_VersionAndVariant(t *testing.T) {
	gen := NewGeneratorV4()
	space := UUID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}

	id := gen.MustHashed(space, []byte("test"))

	if id.Version() != 4 {
		t.Errorf("expected version 4, got %d", id.Version())
	}
	// Check variant bits (should be 10xxxxxx)
	if id[8]&0xc0 != 0x80 {
		t.Errorf("expected RFC 4122 variant, got %02x", id[8])
	}
}
