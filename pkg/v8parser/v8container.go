package v8parser

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
)

const сontainerHeaderLength = v8address(16)

type fileIndex map[string]v8address

type Container struct {
	file     *os.File
	index    fileIndex
	header   []byte
	metadata ListTree
}

func (c *Container) MetadataType() string {
	metadataID := c.metadata.Get(3).Get(0).Value()

	switch metadataID {
	case ExtProcID:
		return "Внешняя обработка"
	case ExtReptID:
		return "Внешний отчет"
	}
	return fmt.Sprintf("Неизвестен [%s]", metadataID)
}

func (c *Container) FileAsContent(name string) string {
	address := c.index[name]
	content := ReadContent(c.file, address)

	return content
}

func (c *Container) FileAsListTree(name string) *ListTree {
	content := c.FileAsContent(name)
	list, err := NewListTree().Load(content)
	if err != nil {
		log.Fatal(err.Error())
	}

	return list
}

func ReadContainer(file *os.File) Container {
	container := new(Container)

	container.file = file
	container.header = ReadHeader(file)
	container.index = ReadIndex(file)

	rootList := container.FileAsListTree("root")
	metadata := container.FileAsListTree(rootList.Get(1).Value())
	container.metadata = *metadata

	return *container
}

// Читает заголовок контейнера

func ReadHeader(file *os.File) []byte {
	headerBegin := v8address(0)

	header := ReadFileFragment(file, headerBegin, сontainerHeaderLength)
	return header
}

// Читает оглавление контейнера

func ReadIndex(file *os.File) fileIndex {
	data := ReadDocument(file, сontainerHeaderLength)
	index := make(fileIndex, 0)

	length := v8address(len(data))
	for i := v8address(0); i < length; i += 12 {
		attributes := v8address(binary.LittleEndian.Uint32(data[i:(i + 4)]))
		content := v8address(binary.LittleEndian.Uint32(data[(i + 4):(i + 8)]))

		attrData := ReadDocument(file, attributes)
		filename := ConvertFilename(attrData[20:])

		index[filename] = content
	}
	return index
}
