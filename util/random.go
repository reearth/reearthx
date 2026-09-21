package util

import "crypto/rand"

const randomStringCharSet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// SecureRandomString returns a cryptographically secure random alphanumeric
// string of length n. Use it for credentials; lo.RandomString draws from
// math/rand and is predictable.
func SecureRandomString(n int) string {
	if n <= 0 {
		return ""
	}

	out := make([]byte, 0, n)
	buf := make([]byte, n)
	for len(out) < n {
		// crypto/rand.Read never returns an error and always fills buf.
		_, _ = rand.Read(buf)
		for _, b := range buf {
			// reject the tail so the modulo stays unbiased (62*4 == 248)
			if b >= 248 {
				continue
			}
			out = append(out, randomStringCharSet[b%byte(len(randomStringCharSet))])
			if len(out) == n {
				break
			}
		}
	}
	return string(out)
}
