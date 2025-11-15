package streamlive

import "math/rand"

const defaultRandomAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func PseudorandomString(length int) string {
	return PseudorandomStringWithAlphabet(length, defaultRandomAlphabet)
}

func PseudorandomStringWithAlphabet(length int, alphabet string) string {
	b := make([]byte, length)
	m := len(alphabet)

	for i := range b {
		b[i] = alphabet[rand.Intn(m)]
	}

	return string(b)
}
