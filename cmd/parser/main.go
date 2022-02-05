package main

import (
	"log"
	"os"

	"github.com/goboden/v8parser/pkg/v8parser"
)

func main() {
	ReadFile("./ext/Отчет партнера (франчайзи).epf")
	// ReadFile("./ext/upp_test__2022_01_26__09_10.cf")
}

func ReadFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer file.Close()

	v8parser.UnpackFile(file)

}
