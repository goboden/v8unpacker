package main

import (
	"os"

	"github.com/goboden/v8parser/pkg/v8parser"
)

func main() {
	// fmt.Println(os.Getwd())

	// unpackFile("./ext/Отчет партнера (франчайзи).epf")
	// unpackFile("./ext/upp_test__2022_01_26__09_10.cf")

	ReadFile("./ext/Отчет партнера (франчайзи).epf")
	// ReadFile("./ext/upp_test__2022_01_26__09_10.cf")
}

func ReadFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		println(err.Error())
	}
	defer file.Close()

	v8parser.UnpackFile(file)

}
