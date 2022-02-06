package v8unpacker

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
)

const (
	сontainerHeaderLength = v8address(16)

	META_TYPE_NOT_SUPP = 0
	META_TYPE_EXT_PROC = 1
	META_TYPE_EXT_REPT = 2
	META_TYPE_CONFIG   = 3
)

type fileIndex map[string]v8address

type Container struct {
	reader Reader
	index  fileIndex
	header []byte
}

type RootContainer struct {
	Container
	Metadata ListTree
	MetaType int
}

func (c *RootContainer) MetadataType() int {
	metadataID := c.Metadata.Get(3).Get(0).Value()

	switch metadataID {
	case ExtProcID:
		return META_TYPE_EXT_PROC
	case ExtReptID:
		return META_TYPE_EXT_REPT
	case ConfID:
		return META_TYPE_CONFIG
	}
	return META_TYPE_NOT_SUPP
}

func (c *Container) FileAsContent(name string, deflate bool) string {
	address := c.index[name]
	content := ReadContent(c.reader, address, deflate)

	return content
}

func (c *Container) FileAsListTree(name string, deflate bool) *ListTree {
	content := c.FileAsContent(name, deflate)
	list, err := NewListTree().Load(content)
	if err != nil {
		log.Fatal(err.Error())
	}

	return list
}

func (c *Container) SaveFile(name string, filename string) {
	content := c.FileAsContent(name, true)

	err := os.WriteFile(filename, []byte(content), 0666)
	if err != nil {
		fmt.Println(err.Error(), filename)
	}

}

func (c *Container) PrintIndex() {
	for key := range c.index {
		fmt.Println(key)
	}
}

func (c *RootContainer) formsID() string {
	switch c.MetaType {
	case META_TYPE_EXT_PROC:
		return ExtProcForms
	case META_TYPE_EXT_REPT:
		return ExtProcForms
	default:
		return ""
	}
}

func (c *RootContainer) GetForms() (map[string]string, error) {
	formsID := c.formsID()
	intList := c.Metadata.Get(3).Get(1)
	if !intList.isValue {
		for i := 0; i < intList.Length(); i++ {
			sect := intList.Get(i)
			if sect.isValue {
				continue
			}

			if sect.Get(0).Value() == formsID && sect.Length() > 2 {
				// forms := make(map[string]string, 0)
				for n := 2; n < sect.Length(); n++ {
					indexName := sect.Get(n).Value()
					fmt.Println("---------- ", indexName)
					// c.FileAsListTree(indexName).Print()

					formName := c.FileAsListTree(indexName, true).Get(1).Get(1).Get(1).Get(1).Get(2).Value()
					fmt.Println(formName)

					// fmt.Println(c.FileAsContent(indexName + ".0"))
					// c.FileAsListTree(indexName + ".0").Print()

					reader := NewBytesReader([]byte(c.FileAsContent(indexName+".0", true)))
					formsC := ReadContainer(reader)
					formsC.PrintIndex()

				}
			}
		}
	}
	return nil, nil
}

func ReadContainer(reader Reader) *Container {
	cont := new(Container)

	cont.reader = reader
	cont.header = ReadHeader(reader)
	cont.index = ReadIndex(reader)

	return cont
}

func ReadRootContainer(reader Reader) *RootContainer {
	cont := new(RootContainer)

	cont.reader = reader
	cont.header = ReadHeader(reader)
	cont.index = ReadIndex(reader)

	rootList := cont.FileAsListTree("root", true)
	metadata := cont.FileAsListTree(rootList.Get(1).Value(), true)
	cont.Metadata = *metadata

	cont.MetaType = cont.MetadataType()

	return cont
}

// Читает заголовок контейнера

func ReadHeader(reader Reader) []byte {
	headerBegin := v8address(0)

	header := reader.ReadFragment(headerBegin, сontainerHeaderLength)
	return header
}

// Читает оглавление контейнера

func ReadIndex(reader Reader) fileIndex {
	data := ReadDocument(reader, сontainerHeaderLength)
	index := make(fileIndex, 0)

	length := v8address(len(data))
	for i := v8address(0); i < length; i += 12 {
		attributes := v8address(binary.LittleEndian.Uint32(data[i:(i + 4)]))
		content := v8address(binary.LittleEndian.Uint32(data[(i + 4):(i + 8)]))

		attrData := ReadDocument(reader, attributes)
		filename := ConvertFilename(attrData[20:])

		index[filename] = content
	}
	return index
}
