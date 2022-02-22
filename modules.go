package v8unpacker

import (
	"encoding/binary"
	"fmt"
	"regexp"
)

const objectModuleName = "МодульОбъекта"

func FindModules(container *RootContainer) (map[string]string, error) {
	modules := make(map[string]string, 0)

	formsSection, err := findFormsSection(container)
	if err != nil {
		return nil, fmt.Errorf("get modules: %s", err.Error())
	}

	forms, err := container.findFormsInSection(formsSection)
	if err != nil {
		return nil, fmt.Errorf("get modules: %s", err.Error())
	}

	for k, v := range forms {
		modules[k] = v
	}

	module, err := container.findObjectModule()
	if err != nil {
		return modules, nil
	}
	modules[objectModuleName] = module

	return modules, nil
}

func formName(data string) string {
	re := regexp.MustCompile(`\{\d,[\d],(\w{8}-\w{4}-\w{4}-\w{4}-\w{12})\},"(\S+?)",\n?`)
	found := re.FindStringSubmatch(data)
	if found == nil {
		return ""
	}

	if len(found) < 3 {
		return ""
	}

	return found[2]
}

func readModuleFromContainer(fileContent string) (string, bool, error) {
	if binary.LittleEndian.Uint32([]byte(fileContent)[:4]) == intMax {
		reader := NewBytesReader([]byte(fileContent))
		formContainer := ReadContainer(reader)
		formModule, err := formContainer.FileAsContent("module", false)
		if err != nil {
			return "", false, err
		}

		return formModule, true, nil
	}

	return "", false, nil
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
