package v8unpacker

import (
	"fmt"
	"strings"
)

const (
	ExtProcID = "c3831ec8-d8d5-4f93-8a22-f9bfae07327f"
	ExtReptID = "e41aff26-25cf-4bb6-b6c1-3f478a75f374"
	ConfID    = "9cd510cd-abfc-11d4-9434-004095e12fc7" // ?

	ExtProcAttributes = "ec6bb5e5-b7a8-4d75-bec9-658107a699cf"
	ExtProcTables     = "2bcef0d1-0981-11d6-b9b8-0050bae0a95d"
	ExtProcForms      = "d5b0e5ed-256d-401c-9c36-f630cafd8a62"
	ExtProcTemplates  = "3daea016-69b7-4ed4-9453-127911372fe6"

	ExtReptAttributes = "7e7123e0-29e2-11d6-a3c7-0050bae0a776"
	ExtReptTables     = "b077d780-29e2-11d6-a3c7-0050bae0a776"
	ExtReptForms      = "a3b368c0-29e2-11d6-a3c7-0050bae0a776"
	ExtReptTemplates  = "3daea016-69b7-4ed4-9453-127911372fe6"
)

type ListTree struct {
	value    string
	elements []ListTree
	isValue  bool
}

func (l *ListTree) Load(data string) (*ListTree, error) {
	ReadListTreeData(data, l)
	return l, nil
}

func (l *ListTree) AppendValue(value string) {
	element := NewListTree()
	element.isValue = true
	element.value = value
	l.elements = append(l.elements, *element)
}

func (l *ListTree) AppendElement(element *ListTree) {
	l.elements = append(l.elements, *element)
}

func (l *ListTree) Length() int {
	return len(l.elements)
}

func (l *ListTree) Print(levels ...int) {
	if l.isValue {
		fmt.Printf("[%2d.%2d] %s%s\n", 0, 0, strings.Repeat(". ", 0), l.Value())
		return
	}

	level := 0
	if len(levels) != 0 {
		level = levels[0]
	} else {
		fmt.Printf("LEVELS: %d\n", len(l.elements))
	}

	next := level + 1

	for i, item := range l.elements {
		if item.isValue {
			fmt.Printf("[%2d.%2d] %s%s\n", level, i, strings.Repeat(". ", level), item.Value())
			continue
		}
		fmt.Printf("[%2d.%2d] %s%s\n", level, i, strings.Repeat(". ", level), "+")
		item.Print(next)
	}
}

func (l *ListTree) Get(index int) *ListTree {
	// if l.isValue {
	// 	return nil, errors.New("element is a value, not a list")
	// }
	// if index >= len(l.elements) {
	// 	return nil, fmt.Errorf("index %d is out of range %d", index, len(l.elements))
	// }
	// element := l.elements[index]
	// return &element, nil
	element := l.elements[index]
	return &element
}

func (l *ListTree) Value() string {
	// if !l.isValue {
	// 	return "", errors.New("element is a list, not a value")
	// }
	// return l.value, nil
	return l.value
}

func NewListTree() *ListTree {
	list := new(ListTree)
	list.value = ""
	list.isValue = false
	list.elements = make([]ListTree, 0)

	return list
}

func ReadListTreeData(data string, list *ListTree) {
	data = strings.Trim(data, string(rune(65279)))
	data = strings.Trim(data, "{")
	data = strings.Trim(data, "}")

	ch := ReadSection(data, 0, list)
	<-ch
}

func ReadSection(data string, level int, list *ListTree) <-chan string {

	level++
	outCh := make(chan string)

	go func() {
		for i := 0; len(data) > 0; i++ {
			sym := data[0]

			if sym == ',' {
				data = data[1:]
				continue
			}

			if sym == '{' {
				data = data[1:]

				element := NewListTree()

				inCh := ReadSection(data, level, element)
				data = <-inCh

				list.AppendElement(element)

				continue
			}

			if sym == '}' {
				data = data[1:]
				outCh <- data
				break
			}

			if sym != '\n' && sym != '\r' {
				d := nextDelimeter(data)
				if d > 0 {
					value := data[0:d]
					data = data[d:]

					list.AppendValue(value)

					continue
				}
			}

			data = data[1:]
		}
		close(outCh)
	}()

	return outCh
}

func nextDelimeter(data string) int {
	for i := 0; i < len(data); i++ {
		if data[i] == ',' || data[i] == '}' {
			return i
		}
	}

	return -1
}
