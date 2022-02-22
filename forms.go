package v8unpacker

import (
	"fmt"
	"strings"
)

func findFormsSection(container *RootContainer) (*ListTree, error) {
	formsID, ok := metadata[container.metadataID][mdtForms]
	if !ok {
		return nil, fmt.Errorf("forms id not found, metadata id: %s", container.metadataID)
	}

	intList, err := container.Metadata.Get(3, 1)
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
	formData, err := c.FileAsContent(nameInIndex, true)
	if err != nil {
		return "", "", err
	}

	formName := formName(formData)

	fileContent, err := c.FileAsContent(nameInIndex+".0", true)
	if err != nil {
		return "", "", err
	}

	formModule, ok, err := readModuleFromContainer(fileContent)
	if err != nil {
		return "", "", err
	}

	if ok {
		return formName, formModule, nil
	}

	formListTree, err := c.FileAsListTree(nameInIndex+".0", true)
	if err != nil {
		return "", "", err
	}
	formModule, err = formListTree.GetValue(2)
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
