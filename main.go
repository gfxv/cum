package main

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/skip2/go-qrcode"
)

//go:embed ffmpeg-flags
var ffmpegFlags string

const bufferSize = 2048
const qrSize = 1024
const filename = "test.txt"
const outputFile = "output.mp4"
const qrDir = "./qr"

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
