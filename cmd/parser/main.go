package main

import (
	"log"
	"os"

	"github.com/goboden/v8parser/pkg/v8unpacker"
)

func main() {
	ReadFile("./ext/Test.epf")
	// ReadFile("./ext/Test.cf")

	// ReadFile("./ext/a8209a4d-89a4-426f-9c30-265be508663e.txt")
	// ReadFile("./ext/49c2d0f1-6445-4ef1-b995-e7d7f332b1c2.txt")
}

func ReadFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err.Error())
	}
	reader := v8unpacker.NewFileReader(file)
	defer file.Close()

	cont := v8unpacker.ReadRootContainer(reader)

	// cont.Metadata.Print()
	// cont.Metadata.Get(3).Print()

	// cont.Metadata.Get(3).Get(1).Print()

	// forms, _ := cont.GetForms()
	// for key, value := range forms {
	// 	fmt.Println(key)
	// 	fmt.Println(value)
	// }
	// cont.GetForms()

	fn := cont.Metadata.GetChain(3, 1, 1, 3, 1, 1, 2).Value()
	m := cont.FileAsContent(fn+".0", true)
	bReader := v8unpacker.NewBytesReader([]byte(m))
	contM := v8unpacker.ReadContainer(bReader)
	// fmt.Println(contM.FileAsContent("text", false))
	contM.SaveFile("text", "M.bsl", false)

	// cont := v8unpacker.ReadContainer(reader)
	// cont.PrintIndex()
	// fmt.Println(cont.FileAsContent("module", false))
}
