package main

import (
	"testing"

	"github.com/libp2p/go-libp2p/examples/testutils"
)

func TestMain(t *testing.T) {
	var h testutils.LogHarness
	h.ExpectPrefix("Okay, no connection from h1 to h3")
	h.ExpectPrefix("Meow! It worked!")
	h.Run(t, run)
}
