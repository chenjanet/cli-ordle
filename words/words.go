package words

import (
	"math/rand"
	"time"
)

// create & seed generator
var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func RandomWord() (string, error) {
	index := r.Intn(len(words))
	return words[index], nil
}

func IsValidGuess(guess string) bool {
	if len(guess) != 5 {
		return false
	}
	high := len(valid) - 1
	low := 0
	for low <= high {
		mid := (high-low)/2 + low
		if guess == valid[mid] {
			return true
		} else if guess < valid[mid] {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}
	return false
}
