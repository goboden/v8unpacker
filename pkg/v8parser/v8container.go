package v8parser

import (
	"encoding/binary"
	"os"
)

const сontainerHeaderLength = v8address(16)

type fileIndex map[string]v8address

// Читает заголовок контейнера

func ReadHeader(file *os.File) []byte {
	headerBegin := v8address(0)

	header := ReadFileFragment(file, headerBegin, сontainerHeaderLength)
	return header
}

// Читает оглавление контейнера

func ReadIndex(file *os.File) fileIndex {
	data := ReadDocument(file, сontainerHeaderLength)
	index := make(fileIndex, 0)

	length := v8address(len(data))
	for i := v8address(0); i < length; i += 12 {
		attributes := v8address(binary.LittleEndian.Uint32(data[i:(i + 4)]))
		content := v8address(binary.LittleEndian.Uint32(data[(i + 4):(i + 8)]))

		attrData := ReadDocument(file, attributes)
		filename := ConvertFilename(attrData[20:])

		index[filename] = content
	}
	return index
}
