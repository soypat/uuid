package uuid

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"hash"
	"io"
	"sync"
	"time"
)

var (
	errNilRandSource = errors.New("got nil random source for UUID Generator")
)

const variantrfc4122 = 0x80

type Generator struct {
	randsource  io.Reader
	versionBits uint8
	// Hash
	hashMu sync.Mutex
	hash   hash.Hash

	// time stuff.
	timeMu   sync.Mutex
	now      func() time.Time
	clockSeq uint16
	lastTime uint64
}

type GeneratorConfig struct {
	Time       func() time.Time
	RandSource io.Reader
	Hash       hash.Hash
	Version    Version
}

// NewGeneratorV4 returns a Generator configured for Version 4 (random) UUIDs
// using crypto/rand.Reader as the random source and MD5 as hash algorithm for NewHashed calls.
func NewGeneratorV4() *Generator {
	g := new(Generator)
	err := g.Init(GeneratorConfig{
		RandSource: rand.Reader,
		Version:    4,
		Hash:       md5.New(), // Default MD5, fast and nice.
	})
	if err != nil {
		panic("unreachable")
	}
	return g
}

func (g *Generator) SetRand(r io.Reader) {
	if r == nil {
		panic(errNilRandSource)
	}
	g.randsource = r
}

func (g *Generator) Init(cfg GeneratorConfig) error {
	if cfg.Version == 0 {
		cfg.Version = 4 // Default version.
	}
	if cfg.RandSource == nil {
		return errNilRandSource
	} else if cfg.Version.HasTime() && cfg.Time == nil {
		return errors.New("Generator version needs time source")
	} else if cfg.Version != 4 {
		return errors.New("UUID version not supported")
	} else if cfg.Hash == nil {
		return errors.New("nil hasher")
	}
	g.now = cfg.Time
	g.randsource = cfg.RandSource
	g.versionBits = uint8(cfg.Version) << 4
	g.hash = cfg.Hash
	return nil
}

func (g *Generator) MustHashed(space UUID, data []byte) UUID {
	uuid, err := g.NewHashed(space, data)
	if err != nil {
		panic(err)
	}
	return uuid
}

func (g *Generator) NewHashed(space UUID, data []byte) (uuid UUID, err error) {
	h := g.hash
	g.hashMu.Lock()
	defer g.hashMu.Unlock()
	h.Reset()
	_, err = h.Write(space[:])
	if err != nil {
		return uuid, err
	}
	_, err = h.Write(data)
	if err == nil {
		s := h.Sum(nil)
		copy(uuid[:], s)
		g.setVersionData(uuid[:])
	}
	return uuid, err
}

func (g *Generator) MustRandom() UUID {
	uuid, err := g.NewRandom()
	if err != nil {
		panic(err)
	}
	return uuid
}

func (g *Generator) NewRandom() (uuid UUID, err error) {
	err = g.random(uuid[:])
	if err == nil {
		g.setVersionData(uuid[:])
	}
	return uuid, err
}

// WriteRandHex generates random hexadecimal characters and encodes
// them into b byte buffer. No allocations performed.
func (g *Generator) WriteRandHex(b []byte) error {
	if len(b) == 0 {
		return nil
	}
	if len(b)%2 == 1 {
		var src [1]byte
		var dst [2]byte
		err := g.random(src[:])
		if err != nil {
			return err
		}
		hex.Encode(dst[:], src[:])
		b[len(b)-1] = dst[0]
		b = b[:len(b)-1]
	}
	binaryPart := b[len(b)/2:]
	err := g.random(binaryPart)
	if err != nil {
		return err
	}
	hex.Encode(b, binaryPart)
	return nil
}

func (g *Generator) setVersionData(uuid []byte) {
	// Only version 4 supported as of yet.
	uuid[6] = (uuid[6] & 0x0f) | g.versionBits
	uuid[8] = (uuid[8] & 0x3f) | variantrfc4122
}

// random completely fills slice b with random data.
func (g *Generator) random(b []byte) error {
	if g.randsource == nil {
		return errNilRandSource
	}
	_, err := io.ReadFull(g.randsource, b)
	return err
}

const _invalidx2d = 255

func init() {
	for i := range x2d {
		if x2d[i] == 0 && i != '0' {
			x2d[i] = _invalidx2d
		}
	}
}

var x2d = [...]byte{
	'0': 0,
	'1': 1,
	'2': 2,
	'3': 3,
	'4': 4,
	'5': 5,
	'6': 6,
	'7': 7,
	'8': 8,
	'9': 9,
	'a': 10, 'A': 10,
	'b': 11, 'B': 11,
	'c': 12, 'C': 12,
	'd': 13, 'D': 13,
	'e': 14, 'E': 14,
	'f': 15, 'F': 15,
}

// xtob converts hex characters x1 and x2 into a byte.
func xtob(x1, x2 byte) (_ byte, valid bool) {
	if x1 > 'f' || x2 > 'f' {
		return 0, false
	}
	b1 := x2d[x1]
	b2 := x2d[x2]
	return (b1 << 4) | b2, b1 != 255 && b2 != 255
}

// GetTime returns the current Time (100s of nanoseconds since 15 Oct 1582) and
// clock sequence as well as adjusting the clock sequence as needed.  An error
// is returned if the current time cannot be determined.
func (g *Generator) GetTime() (Time, uint16, error) {
	g.timeMu.Lock()
	defer g.timeMu.Unlock()
	return g.getTime()
}

func (g *Generator) getTime() (Time, uint16, error) {
	t := g.now()

	// If we don't have a clock sequence already, set one.
	if g.clockSeq == 0 {
		g.setClockSequence(-1)
	}
	now := uint64(t.UnixNano()/100) + g1582ns100

	// If time has gone backwards with this clock sequence then we
	// increment the clock sequence
	if now <= g.lastTime {
		g.clockSeq = ((g.clockSeq + 1) & 0x3fff) | 0x8000
	}
	g.lastTime = now
	return Time(now), g.clockSeq, nil
}

// ClockSequence returns the current clock sequence, generating one if not
// already set.  The clock sequence is only used for Version 1 UUIDs.
//
// The uuid package does not use global static storage for the clock sequence or
// the last time a UUID was generated.  Unless SetClockSequence is used, a new
// random clock sequence is generated the first time a clock sequence is
// requested by ClockSequence, GetTime, or NewUUID.  (section 4.2.1.1)
func (g *Generator) ClockSequence() int {
	defer g.timeMu.Unlock()
	g.timeMu.Lock()
	return g.clockSequence()
}

func (g *Generator) clockSequence() int {
	if g.clockSeq == 0 {
		g.setClockSequence(-1)
	}
	return int(g.clockSeq & 0x3fff)
}

// SetClockSequence sets the clock sequence to the lower 14 bits of seq.  Setting to
// -1 causes a new sequence to be generated.
func (g *Generator) SetClockSequence(seq int) {
	defer g.timeMu.Unlock()
	g.timeMu.Lock()
	g.setClockSequence(seq)
}

func (g *Generator) setClockSequence(seq int) {
	if seq == -1 {
		var b [2]byte
		g.random(b[:]) // clock sequence
		seq = int(b[0])<<8 | int(b[1])
	}
	oldSeq := g.clockSeq
	g.clockSeq = uint16(seq&0x3fff) | 0x8000 // Set our variant
	if oldSeq != g.clockSeq {
		g.lastTime = 0
	}
}
