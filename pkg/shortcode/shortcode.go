package shortcode

import "math/rand"

type Generator interface {
	NextID() string
}

type shortCodeGenerator struct {
	minLength int
}

const chars = "abcdefghijklmnopqrstuvwsyzABCDEFJHIJKLMNOKPRSTUVWSVZ0123456789"

func (s *shortCodeGenerator) NextID() string {
	length := len(chars)
	id := make([]byte, s.minLength)
	for i := 0; i < s.minLength; i++ {
		id[i] = chars[rand.Intn(length)]
	}
	return string(id)
}

func NewShortCodeGenerator(minLength int) Generator {
	return &shortCodeGenerator{minLength: minLength}
}
