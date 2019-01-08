package main

import (
	"errors"
	"strconv"
)

func add(a uint64, b uint64) uint64 {
	c := a + b

	if c < a {
		panic(errors.New("invalid addition"))
	}

	return c
}

func sub(a uint64, b uint64) uint64 {
	if b > a {
		panic(errors.New("invalid subtraction"))
	}

	return a - b
}

func s2u(s string) uint64 {
	u, err := strconv.ParseUint(s, 10, 64)

	checkError(err)

	return u
}

func u2s(u uint64) string {
	return strconv.FormatUint(u, 10)
}

func b2u(b []byte) uint64 {
	return s2u(string(b))
}

func u2b(u uint64) []byte {
	return []byte(u2s(u))
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
