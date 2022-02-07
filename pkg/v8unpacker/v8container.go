package v8unpacker

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"strings"
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
	metadataID, err := c.Metadata.GetValue(3, 0)
	if err != nil {
		log.Fatal(err)
	}

	switch metadataID {
	case extProcID:
		return META_TYPE_EXT_PROC
	case extReptID:
		return META_TYPE_EXT_REPT
	case confID:
		return META_TYPE_CONFIG
	}
	return META_TYPE_NOT_SUPP
}

func (c *Container) FileAsContent(name string, deflate bool) string {
	address := c.index[name]
	content := readContent(c.reader, address, deflate)

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

func (c *Container) SaveFile(name string, filename string, deflate bool) {
	content := c.FileAsContent(name, deflate)

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
		return extProcForms
	case META_TYPE_EXT_REPT:
		return extReptForms
	default:
		return ""
	}
}

func (c *RootContainer) GetForms() (map[string]string, error) {
	formsID := c.formsID()
	intList, err := c.Metadata.Get(3, 1)
	if err != nil {
		log.Fatal(err)
	}

	if !intList.isValue {
		for i := 0; i < intList.Length(); i++ {
			sect, err := intList.Get(i)
			if err != nil {
				log.Fatal(err)
			}
			if sect.isValue {
				continue
			}

			if val, err := sect.GetValue(0); err != nil && val == formsID && sect.Length() > 2 {
				forms := make(map[string]string, 0)
				for n := 2; n < sect.Length(); n++ {
					indexName, err := sect.GetValue(n)
					if err != nil {
						log.Fatal(err)
					}

					formName, err := c.FileAsListTree(indexName, true).GetValue(1, 1, 1, 1, 2)
					if err != nil {
						log.Fatal(err)
					}
					formName = strings.Trim(formName, `"`)

					reader := NewBytesReader([]byte(c.FileAsContent(indexName+".0", true)))
					formsC := ReadContainer(reader)

					formModule := formsC.FileAsContent("module", false)
					forms[formName] = formModule
					formsC.SaveFile("module", formName+".bsl", false)
				}
				return forms, nil
			}
		}
	}
	return nil, nil
}

func ReadContainer(reader Reader) *Container {
	cont := new(Container)

	cont.reader = reader
	cont.header = readHeader(reader)
	cont.index = readIndex(reader)

	return cont
}

func ReadRootContainer(reader Reader) *RootContainer {
	cont := new(RootContainer)

	cont.reader = reader
	cont.header = readHeader(reader)
	cont.index = readIndex(reader)

	rootList := cont.FileAsListTree("root", true)
	metaFileName, err := rootList.GetValue(1)
	if err != nil {
		log.Fatal(err)
	}
	metadata := cont.FileAsListTree(metaFileName, true)
	cont.Metadata = *metadata

	cont.MetaType = cont.MetadataType()

	return cont
}

// Читает заголовок контейнера

func readHeader(reader Reader) []byte {
	headerBegin := v8address(0)

	header := reader.ReadFragment(headerBegin, сontainerHeaderLength)
	return header
}

func readIndex(reader Reader) fileIndex {
	data := readDocument(reader, сontainerHeaderLength)
	index := make(fileIndex, 0)

	length := v8address(len(data))
	for i := v8address(0); i < length; i += 12 {
		attributes := v8address(binary.LittleEndian.Uint32(data[i:(i + 4)]))
		content := v8address(binary.LittleEndian.Uint32(data[(i + 4):(i + 8)]))

		attrData := readDocument(reader, attributes)
		filename := convertFilename(attrData[20:])

		index[filename] = content
	}
	return index
}
