package uuid

import "time"

// A Time represents a time as the number of 100's of nanoseconds since 15 Oct
// 1582.
type Time int64

const (
	lillian    = 2299160          // Julian day of 15 Oct 1582
	unix       = 2440587          // Julian day of 1 Jan 1970
	epoch      = unix - lillian   // Days between epochs
	g1582      = epoch * 86400    // seconds between epochs
	g1582ns100 = g1582 * 10000000 // 100s of a nanoseconds between epochs
)

// UnixTime converts t the number of seconds and nanoseconds using the Unix
// epoch of 1 Jan 1970.
func (t Time) UnixTime() (sec, nsec int64) {
	sec = int64(t - g1582ns100)
	nsec = (sec % 10000000) * 100
	sec /= 10000000
	return sec, nsec
}

// GoTime is a convenience method to convert UUID time to Go's standard library [time.Time].
func (t Time) GoTime() time.Time {
	return time.Unix(t.UnixTime())
}
