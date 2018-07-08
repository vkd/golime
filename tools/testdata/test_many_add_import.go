package test

import (
	"fmt"
	"log"

	"github.com/pkg/errors"
)

func unusedMany() string {
	fmt.Printf("format")
	log.Printf("many")
	// hello from comments
	return errors.New("unused_many").Error()
}
