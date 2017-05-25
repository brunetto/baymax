package main

import (
	"log"
	"github.com/tmthrgd/go-bindata"
)

func main () {
	//data, err := ioutil.ReadFile("test.html")
	data, err := bindata.Asset("test.html")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(data))
}

