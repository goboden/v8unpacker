package v8unpacker

const objectModuleName = "МодульОбъекта"

func FindModules(container *RootContainer) (map[string]string, error) {
	modules := make(map[string]string, 0)

	forms, _ := container.ReadFormNames()
	for _, name := range forms {
		form, _ := container.ReadForm(name)

		modules[form.Name] = form.Module
	}

	module, err := container.ReadObjectModule()
	if err != nil {
		return modules, err
	}

	if module != "" {
		modules[objectModuleName] = module
	}

	return modules, nil
}

func ObjectModuleText(file *File) (string, error) {
	if v8address(len(file.Data)) > сontainerHeaderLength {
		con := ReadContainer(NewBytesReader(file.Data))
		module, err := con.ReadFile("text", false)
		if err != nil {
			return "", err
		}
		return module.AsString(), nil
	}
	return "", nil
}
