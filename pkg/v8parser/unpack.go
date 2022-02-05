package v8parser

import (
	"encoding/binary"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode/utf16"
)

// https://infostart.ru/1c/articles/250142/
// https://infostart.ru/public/106310/
// https://github.com/e8tools/v8unpack

const INT_MAX = 2147483647

type v8address uint32

// Распаковывает файл контейнера

func UnpackFile(file *os.File) {
	container := ReadContainer(file)
	// metadataType := container.MetadataType()

	// fmt.Printf("Metadata Type: %s\n", metadataType)

	// container.metadata.Print()
	container.metadata.Get(3).Get(1).Print()
}

// Читает фрагмент файла по адресу

func ReadFileFragment(file *os.File, begin v8address, length v8address) []byte {
	file.Seek(int64(begin), 0)
	bufLength := v8address(length)

	buf := make([]byte, bufLength)
	for i := v8address(1); true; i++ {
		n, err := file.Read(buf)

		if n > 0 {
			return buf
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			println("ReadFileFragment. ", err.Error())
			break
		}
	}

	return buf
}

// Коневертирует имя файла из UTF16

func ConvertFilename(filenameUTF16 []byte) string {
	utf := make([]uint16, len(filenameUTF16)/2)
	for i := 0; i < len(filenameUTF16); i += 2 {
		utf[(i / 2)] = binary.LittleEndian.Uint16(filenameUTF16[i:])
	}

	filename := string(utf16.Decode(utf))
	filename = strings.TrimRight(filename, string([]byte{0, 0}))
	return filename
}

// Конвертирует адрес из заголовка блока

func ConvertAddress(b []byte) v8address {
	bytes := ""
	for _, v := range b {
		bytes += string(v)
	}
	i, err := strconv.ParseUint(bytes, 16, 32)
	if err != nil {
		return 0
	}
	return v8address(i)
}
