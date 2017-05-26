package main

import (
	"github.com/rogpeppe/rog-go/reverse"
	"os"
	"log"
	"fmt"
)

func main () {
	file, err := os.Open("test.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := reverse.NewScanner(file)
	count := 0
	for scanner.Scan() {
		if count >= 10 {
			break
		}
		line := scanner.Text()
		fmt.Println(line)
		count++
	}
}
