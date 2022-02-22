package v8unpacker

import (
	"encoding/binary"
	"fmt"
	"log"
)

const (
	сontainerHeaderLength = v8address(16)
)

type fileIndex map[string]v8address

type Container struct {
	reader Reader
	index  fileIndex
	header []byte
}

type RootContainer struct {
	Container
	Metadata   ListTree
	metadataID string
}

func (c *Container) FileAsContent(name string, deflate bool) (string, error) {
	address, ok := c.index[name]
	if !ok {
		return "", fmt.Errorf("file %s not found in index", name)
	}

	content := readContent(c.reader, address, deflate)

	return content, nil
}

func (c *Container) FileAsListTree(name string, deflate bool) (*ListTree, error) {
	content, err := c.FileAsContent(name, deflate)
	if err != nil {
		return nil, err
	}
	list, err := NewListTree().Load(content)
	if err != nil {
		log.Fatal(err.Error())
	}

	return list, nil
}

func (c *Container) GetIndex() fileIndex {
	return c.index
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

	rootList, err := cont.FileAsListTree("root", true)
	if err != nil {
		log.Fatal(err)
	}

	metaFileName, err := rootList.GetValue(1)
	if err != nil {
		log.Fatal(err)
	}
	metadata, err := cont.FileAsListTree(metaFileName, true)
	if err != nil {
		log.Fatal(err)
	}
	cont.Metadata = *metadata

	metadataID, err := cont.Metadata.GetValue(3, 0)
	if err != nil {
		log.Fatal(err)
	}
	cont.metadataID = metadataID

	return cont
}

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
