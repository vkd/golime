package test

import (
	"fmt"
	"log"
)

func unusedMany() string {
	fmt.Printf("format")
	log.Printf("many")
	var m = map[string]interface{}{
		""
	}
	return "unused_many"
}
