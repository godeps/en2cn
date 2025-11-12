package main

import (
	"fmt"
	"log"

	"github.com/godeps/en2cn"
)

func main() {
	engine, err := en2cn.NewEngine()
	if err != nil {
		log.Fatalf("init engine: %v", err)
	}

	words := []string{
		"hello",
		"coffee",
		"tiger",
		"banana",
		"tesla",
		"apple",
		"google",
		"microsoft",
		"openai",
	}
	for _, word := range words {
		result, err := engine.Convert(word)
		if err != nil {
			fmt.Printf("%s -> error: %v\n", word, err)
			continue
		}
		fmt.Printf("%s -> %s\n", word, result)
	}
}
