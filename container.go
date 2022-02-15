package v8unpacker

import (
	"encoding/binary"
	"fmt"
	"log"
	"strings"
)

const (
	сontainerHeaderLength = v8address(16)

	mdTypeNotSupported  = 0
	mdTypeExtProcedure  = 1
	mdTypeExtReport     = 2
	mdTypeConfiguration = 3
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
	case idExtProcedure:
		return mdTypeExtProcedure
	case idExtReport:
		return mdTypeExtReport
	case idConfiguration:
		return mdTypeConfiguration
	}
	return mdTypeNotSupported
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

func (c *Container) PrintIndex() {
	for key := range c.index {
		fmt.Println(key)
	}
}

func (c *RootContainer) formsID() string {
	switch c.MetaType {
	case mdTypeExtProcedure:
		return extProcForms
	case mdTypeExtReport:
		return extReptForms
	default:
		return ""
	}
}

func (c *RootContainer) GetModules() (map[string]string, error) {
	modules := make(map[string]string, 0)

	formsSection, err := c.findFormsSection()
	if err != nil {
		return nil, fmt.Errorf("get modules: %s", err.Error())
	}

	forms, err := c.findFormsInSection(formsSection)
	if err != nil {
		return nil, fmt.Errorf("get modules: %s", err.Error())
	}

	for k, v := range forms {
		modules[k] = v
	}

	module, err := c.findObjectModule()
	if err != nil {
		// return nil, fmt.Errorf("get modules: %s", err.Error())
		return modules, nil
	}
	modules["МодульОбъекта"] = module

	return modules, nil
}

func (c *RootContainer) findFormsSection() (*ListTree, error) {
	formsID := c.formsID()
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

		if val, err := sect.GetValue(0); err == nil && val == formsID && sect.Length() > 2 {
			return sect, nil
		}
	}

	return nil, fmt.Errorf("find forms section: not found")
}

func (c *RootContainer) findFormsInSection(section *ListTree) (map[string]string, error) {

	forms := make(map[string]string, 0)

	for n := 2; n < section.Length(); n++ {
		indexName, err := section.GetValue(n)
		if err != nil {
			break
		}

		formName, formModule, err := c.findFormData(indexName)
		if err != nil {
			return nil, fmt.Errorf("find forms in section: %s", err.Error())
		}

		forms[formName] = formModule
	}

	return forms, nil
}

func (c *RootContainer) findFormData(nameInIndex string) (string, string, error) {
	formData, err := c.FileAsListTree(nameInIndex, true)
	if err != nil {
		return "", "", err
	}

	formName, isManaged, err := formInfo(formData)
	if err != nil {
		// return "", "", fmt.Errorf("find form data: %s", err.Error())
		formName = nameInIndex
		isManaged = false
	} else {
		formName = strings.Trim(formName, `"`)
	}

	if !isManaged {
		formContent, err := c.FileAsContent(nameInIndex+".0", true)
		if err != nil {
			return "", "", err
		}
		reader := NewBytesReader([]byte(formContent))
		formContainer := ReadContainer(reader)
		formModule, err := formContainer.FileAsContent("module", false)
		if err != nil {
			return "", "", err
		}

		return formName, formModule, nil
	}

	formListTree, err := c.FileAsListTree(nameInIndex+".0", true)
	if err != nil {
		return "", "", err
	}
	formModule, err := formListTree.GetValue(2)
	if err != nil {
		return "", "", err
	}
	formModule = strings.ReplaceAll(formModule, `""""`, `""`)
	formModule = strings.ReplaceAll(formModule, `""`, `"`)
	formModule = strings.Trim(formModule, `"`)

	return formName, formModule, nil
}

func (c *RootContainer) findObjectModule() (string, error) {
	nameInIndex, err := c.Metadata.GetValue(3, 1, 1, 3, 1, 1, 2)
	if err != nil {
		return "", fmt.Errorf("find object module: %s", err.Error())
	}

	moduleData, err := c.FileAsContent(nameInIndex+".0", true)
	if err != nil {
		return "", err
	}

	moduleText, err := readObjectModuleText(moduleData)
	if err != nil {
		return "", err
	}

	return moduleText, nil
}

func readObjectModuleText(moduleData string) (string, error) {
	if v8address(len(moduleData)) > сontainerHeaderLength {
		moduleContainer := ReadContainer(NewBytesReader([]byte(moduleData)))
		moduleText, err := moduleContainer.FileAsContent("text", false)
		if err != nil {
			return "", err
		}
		return moduleText, nil
	}
	return "", nil
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

	cont.MetaType = cont.MetadataType()

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

func formInfo(list *ListTree) (string, bool, error) {
	var isManaged bool

	formName, err := list.GetValue(1, 1, 1, 1, 2)
	if err != nil {
		return "", false, fmt.Errorf("form info: %s", err.Error())
	}

	isManagedStr, err := list.GetValue(1, 1, 1, 3)
	if err != nil {
		return "", false, fmt.Errorf("form info: %s", err.Error())
	}

	if isManagedStr == "1" {
		isManaged = true
	}

	return formName, isManaged, nil
}
