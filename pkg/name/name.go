package name

import (
	"math/rand"
	"time"

	petname "github.com/dustinkirkland/golang-petname"
)

func GenerateName() string {
	// generate some entropy
	rand.Seed(time.Now().UTC().UnixNano())
	return petname.Generate(3, "-")
}
