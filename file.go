package v8unpacker

import "log"

type File struct {
	Name        string
	Data        []byte
	IsContainer bool
}

func (f *File) AsString() string {
	return string(f.Data)
}

func (f *File) AsListTree() (*ListTree, error) {
	list, err := NewListTree().Load(f.AsString())
	if err != nil {
		log.Fatal(err.Error())
	}

	return list, nil
}
