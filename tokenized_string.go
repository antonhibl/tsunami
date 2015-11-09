package main

import (
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type tokenizedString struct {
	base   string
	random *rand.Rand
	source rand.Source
}

func NewTokenizedString(base string) *tokenizedString {
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)
	return &tokenizedString{
		base:   base,
		random: random,
		source: source,
	}
}

func RandomDigit() string {
	return strconv.Itoa(random.Intn(9))
}

func RandomLetter() string {
	charSet := "abcdefghijklmnopqrstuvwxyz"
	return string(charSet[random.Intn(len(charSet))])
}

func RandomULetter() string {
	charSet := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	return string(charSet[random.Intn(len(charSet))])
}

func (this *tokenizedString) String() string {
	toSend := this.base
	digitTokenCount := strings.Count(this.base, "{D}")

	for i := 0; i < digitTokenCount; i++ {
		toSend = strings.Replace(toSend, "{D}", RandomDigit(), 1)
	}

	letterTokenCount := strings.Count(this.base, "{l}")
	for i := 0; i < letterTokenCount; i++ {
		toSend = strings.Replace(toSend, "{l}", RandomLetter(), 1)
	}

	uLetterTokenCount := strings.Count(this.base, "{L}")
	for i := 0; i < uLetterTokenCount; i++ {
		toSend = strings.Replace(toSend, "{L}", RandomULetter(), 1)
	}

	return toSend
}
