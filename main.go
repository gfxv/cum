package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unsafe"

	"github.com/skip2/go-qrcode"
)

//go:embed ffmpeg-flags
var ffmpegFlags string

const bufferSize = 2048
const qrSize = 1024
const filename = "test.txt"
const outputFile = "output.mp4"
const qrDir = "./qr"
const returnTmp = "./returned"

func main() {

	if err := prepareDir(qrDir); err != nil {
		fmt.Println(err)
		panic(err.Error())
	}
	defer cleanUpDir(qrDir)

	if err := generateQR(filename); err != nil {
		panic(err.Error())
	}

	if err := generateVideo(); err != nil {
		panic(err.Error())
	}

	if err := prepareDir(returnTmp); err != nil {
		panic(err)
	}
	defer cleanUpDir(returnTmp)

	if err := splitVideo(); err != nil {
		fmt.Println(err.Error())
		panic(err.Error())
	}

	if err := loadFileFromImages(); err != nil {
		fmt.Println(err)
		panic(err.Error())
	}
}

func prepareDir(dir string) error {
	if err := os.Mkdir(dir, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func cleanUpDir(dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	return nil
}

func generateQR(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	buffer := make([]byte, bufferSize)
	indx := 1
	for {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}

		if n == 0 {
			break
		}

		qrName := fmt.Sprintf("file-%d.png", indx)
		qrPath := filepath.Join(qrDir, qrName)
		if err = qrcode.WriteFile(string(buffer), qrcode.Medium, qrSize, qrPath); err != nil {
			return err
		}
		indx++
	}
	return nil
}

func generateVideo() error {
	flags := strings.Split(strings.Trim(string(ffmpegFlags), "\n"), " ")
	args := append(flags, outputFile)
	cmd := exec.Command("ffmpeg", args...)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func splitVideo() error {
	cmd := exec.Command("ffmpeg", "-i", "output.mp4", "-vf", "fps=20", "./returned/%d.png")
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func zbarimgDecode(data []byte) ([]byte, error) {
	cmd := exec.Command("zbarimg", "--quiet", "-Sdisable", "-Sqrcode.enable", "-")

	var out bytes.Buffer

	cmd.Stdin = bytes.NewBuffer(data)
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	result := strings.TrimSuffix(strings.TrimPrefix(out.String(), "QR-Code:"), "\n")
	return *(*[]byte)(unsafe.Pointer(&result)), nil
}

func nextEntryNameGenerator(total int) func() string {
	indx := 0
	return func() string {
		if indx >= total {
			return ""
		}
		indx++
		return fmt.Sprintf("%d.png", indx)
	}
}

func loadFileFromImages() error {
	dirs, err := os.ReadDir(returnTmp)
	if err != nil {
		return err
	}

	getNextEntryName := nextEntryNameGenerator(len(dirs))

	const returnedSize = 300000
	// FIX: set capacity as len(dirs) * image_size, leave with hardcoded for now
	buff := bytes.NewBuffer(make([]byte, 0, len(dirs)*returnedSize))
	for _, entry := range dirs {
		if entry.IsDir() {
			continue
		}

		if err := processEntry(getNextEntryName(), buff); err != nil {
			fmt.Printf("can't process entry %s, err: %v\n", entry.Name(), err)
			continue
		}
	}

	if err := os.WriteFile("returned-data.txt", buff.Bytes(), os.ModePerm); err != nil {
		return err
	}

	return nil
}

func processEntry(entryName string, buff *bytes.Buffer) error {
	path := filepath.Join(returnTmp, entryName)
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	fmt.Printf("processing %s\n", path)

	result, err := zbarimgDecode(data)
	if err != nil {
		return err
	}

	if _, err := buff.Write(result); err != nil {
		return err
	}
	return nil
}
