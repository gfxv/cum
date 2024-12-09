package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/skip2/go-qrcode"
)

const bufferSize = 2048
const qrSize = 1024
const filename = "test.txt"
const qrDir = "./qr"

func main() {

	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	buffer := make([]byte, bufferSize)
	indx := 1
	for {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			panic(err)
		}

		if n == 0 {
			break
		}

		fmt.Printf("read %d bytes: %s\n", n, string(buffer[:n]))
		qrName := fmt.Sprintf("%d-%s.png", indx, strings.Split(filename, ".")[0])
		qrPath := filepath.Join(qrDir, qrName)
		if err = qrcode.WriteFile(string(buffer), qrcode.Medium, qrSize, qrPath); err != nil {
			panic(err)
		}
		indx++
	}

}
