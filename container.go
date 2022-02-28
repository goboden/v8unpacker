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

func (c *Container) ReadFile(name string, deflate bool) (*File, error) {
	address, ok := c.index[name]
	if !ok {
		return nil, fmt.Errorf("file %s not found in index", name)
	}

	data := readContent(c.reader, address, deflate)
	isContainer := binary.LittleEndian.Uint32(data[:4]) == intMax
	file := &File{Name: name, Data: data, IsContainer: isContainer}

	return file, nil
}

func (c *Container) GetIndex() fileIndex {
	return c.index
}

func (c *Container) InIndex(name string) bool {
	_, ok := c.index[name]
	return ok
}

func (c *RootContainer) ReadFormNames() ([]string, error) {
	formsID, ok := metadata[c.metadataID][mdtForms]
	if !ok {
		return nil, fmt.Errorf("forms id not found, metadata id: %s", c.metadataID)
	}

	intList, err := c.Metadata.Get(3, 1)
	if err != nil {
		return nil, fmt.Errorf("find forms section: %s", err.Error())
	}

	if intList.isValue {
		return nil, fmt.Errorf("find forms section: wrong metadata format")
	}

	for i := 0; i < intList.Length(); i++ {
		sect, err := intList.Get(i)
		if err != nil {
			return nil, fmt.Errorf("find forms section: %s", err.Error())
		}

		if sect.isValue {
			continue
		}

		if val, err := sect.GetValue(0); err == nil && val == formsID {
			forms := make([]string, 0)

			for n := 2; n < sect.Length(); n++ {
				name, err := sect.GetValue(n)
				if err != nil {
					break
				}

				forms = append(forms, name)
			}

			return forms, nil
		}
	}

	return nil, fmt.Errorf("find forms section: not found")
}

func (c *RootContainer) ReadForm(filename string) (*Form, error) {
	_, ok := c.index[filename]
	if !ok {
		return nil, fmt.Errorf("file %s not found in index", filename)
	}

	file, err := c.ReadFile(filename, true)
	if err != nil {
		return nil, err
	}

	file0, err := c.ReadFile(filename+".0", true)
	if err != nil {
		return nil, err
	}

	form := &Form{DescriptionFile: file, MainFile: file0}
	form.Read()

	return form, nil
}

func (c *RootContainer) ReadObjectModule() (string, error) {
	filename, err := c.Metadata.GetValue(3, 1, 1, 3, 1, 1, 2)
	if err != nil {
		return "", fmt.Errorf("find object module: %s", err.Error())
	}

	filename += ".0"
	if !c.InIndex(filename) {
		return "", nil
	}

	file, err := c.ReadFile(filename, true)
	if err != nil {
		return "", err
	}

	module, err := ObjectModuleText(file)
	if err != nil {
		return "", err
	}

	return module, nil
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

	root, err := cont.ReadFile("root", true)
	if err != nil {
		log.Fatal(err)
	}

	rootList, err := root.AsListTree()
	if err != nil {
		log.Fatal(err)
	}

	metaFileName, err := rootList.GetValue(1)
	if err != nil {
		log.Fatal(err)
	}

	metaFile, err := cont.ReadFile(metaFileName, true)
	if err != nil {
		log.Fatal(err)
	}

	metadata, err := metaFile.AsListTree()
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
