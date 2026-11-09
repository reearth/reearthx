package pgxx

import "time"

// SetUnlockTimeout overrides the unlock timeout for a test and returns a
// function that restores it.
func SetUnlockTimeout(d time.Duration) func() {
	prev := unlockTimeout
	unlockTimeout = d
	return func() { unlockTimeout = prev }
}
