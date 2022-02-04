package v8parser

import (
	"fmt"
	"strings"
)

type ListItem struct {
	value    string
	children []ListItem
}

type List []ListItem

func NewList() List {
	list := make([]ListItem, 0)
	return list
}

func ReadList(data string) {
	data = strings.Trim(data, string(rune(65279)))
	// data = strings.Trim(data, "{")
	// data = strings.Trim(data, "}")

	list := NewList()

	ch := ReadSection(data, 0, list)
	<-ch

	println(">", len(list))
}

func ReadSection(data string, level int, list List) <-chan string {

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
				inCh := ReadSection(data, level, list)
				data = <-inCh
				continue
			}

			if sym == '}' {
				data = data[1:]
				outCh <- data
				break
			}

			if sym != '\n' && sym != '\r' {
				dlm := delimiterPos(data)
				if dlm > 0 {
					value := data[0:dlm]
					fmt.Println(value)
					data = data[dlm:]

					item := new(ListItem)
					item.value = value
					list = append(list, *item)

					continue
				}
			}

			data = data[1:]
		}

		close(outCh)
	}()

	return outCh
}

func delimiterPos(data string) int {
	dlmA := strings.Index(data, ",")
	dlmB := strings.Index(data, "}")
	if dlmA >= 0 || dlmB >= 0 {
		if dlmA >= 0 && dlmA < dlmB {
			return dlmA
		}
		return dlmB
	}
	return -1
}
