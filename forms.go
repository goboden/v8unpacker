package v8unpacker

import (
	"regexp"
	"strings"
)

type Form struct {
	DescriptionFile *File
	MainFile        *File

	Name   string
	Module string
}

func (f *Form) Read() {
	f.Name = formName(f.DescriptionFile.AsString())

	if f.MainFile.IsContainer {
		reader := NewBytesReader(f.MainFile.Data)
		con := ReadContainer(reader)
		module, err := con.ReadFile("module", false)
		if err != nil {
			f.Module = ""
		} else {
			f.Module = module.AsString()
		}
	} else {
		tree, _ := f.MainFile.AsListTree()
		module, _ := tree.GetValue(2)
		module = strings.ReplaceAll(module, `""""`, `""`)
		module = strings.ReplaceAll(module, `""`, `"`)
		module = strings.Trim(module, `"`)
		f.Module = module
	}
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
