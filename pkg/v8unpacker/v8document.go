package v8unpacker

import (
	"bytes"
	"compress/flate"
	"fmt"
)

const blockHeaderLength = 31

type v8blockHeader struct {
	DocumentLength v8address
	BlockLength    v8address
	NextBlock      v8address
}

// Читает документ по блокам

func ReadDocument(reader Reader, begin v8address) []byte {
	documentData := make([]byte, 0)
	documentLength := v8address(0)

	next := begin
	for {
		header, data := ReadBlock(reader, next)

		if header.DocumentLength != 0 {
			documentLength = header.DocumentLength
		}

		if header.NextBlock == INT_MAX {
			documentData = append(documentData, data[:(documentLength-v8address(len(documentData)))]...)
			break
		}
		documentData = append(documentData, data...)

		next = header.NextBlock
	}
	return documentData
}

// Читает блок

func ReadBlock(reader Reader, begin v8address) (*v8blockHeader, []byte) {
	header := reader.ReadFragment(begin, blockHeaderLength)

	if header[0] != 13 {
		panic(fmt.Sprintf("! %d", header[0]))
	}

	blockHeader := new(v8blockHeader)

	blockHeader.DocumentLength = ConvertAddress(header[2:10])
	blockHeader.BlockLength = ConvertAddress(header[11:19])
	blockHeader.NextBlock = ConvertAddress(header[20:28])

	data := reader.ReadFragment(begin+blockHeaderLength, blockHeader.BlockLength)

	return blockHeader, data
}

// Читает и распаковывает документ содержимого

func ReadContentOld(reader Reader, begin v8address) string {
	data := ReadDocument(reader, begin)

	deflator := flate.NewReader(bytes.NewReader(data))
	buffer := make([]byte, 1024)
	out := make([]byte, 0)

	for {
		n, _ := deflator.Read(buffer)
		if n < len(buffer) {
			out = append(out, buffer[:n]...)
			break
		}
		out = append(out, buffer...)
	}

	return string(out)
}

func ReadContent(reader Reader, begin v8address, deflate bool) string {
	var content string
	data := ReadDocument(reader, begin)

	if deflate {
		reader := flate.NewReader(bytes.NewReader(data))
		buffer := make([]byte, 1024)
		out := make([]byte, 0)

		for {
			n, _ := reader.Read(buffer)
			if n < len(buffer) {
				out = append(out, buffer[:n]...)
				break
			}
			out = append(out, buffer...)
		}
		content = string(out)
	} else {
		content = string(data)
	}

	return content
}
