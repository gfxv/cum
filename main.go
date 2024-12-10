package main

import (
	"bytes"
	_ "embed"
	"flag"
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
const defaultQRSize = 1024

const qrDir = "./qr"
const returnTmp = "./returned"

var (
	action     string
	in         string
	out        string
	qrSizeFlag int
)

func init() {
	flag.StringVar(&action, "a", "", "Specifies action to perform: encode or decode. Encode will take an input and convert it to mp4 format. Decode will attempt to convert provided video to oiginal format")
	flag.StringVar(&in, "in", "", "Path to input file")
	flag.StringVar(&out, "out", "", "Path to output file (will be created if not exists or overwritten if already exists)")
	flag.IntVar(&qrSizeFlag, "qrisize", defaultQRSize, "Define a size of QR Code. Can be omitted")
}

func main() {

	flag.Parse()

	switch strings.ToLower(action) {
	case "encode":
		if err := Encode(); err != nil {
			fmt.Printf("can't encode %s, err: %v", in, err)
		}
	case "decode":
		if err := Decode(); err != nil {
			fmt.Printf("can't decode %s, err: %v", in, err)
		}
	default:
		fmt.Println("unknown action")
		os.Exit(1)
	}

}

func Encode() error {
	if err := prepareDir(qrDir); err != nil {
		return err
	}
	defer cleanUpDir(qrDir)

	if err := generateQR(in); err != nil {
		return err
	}

	if err := generateVideo(); err != nil {
		return err
	}
	return nil
}

func Decode() error {
	if err := prepareDir(returnTmp); err != nil {
		return err
	}
	defer cleanUpDir(returnTmp)

	if err := splitVideo(); err != nil {
		return err
	}

	if err := loadFileFromImages(); err != nil {
		return err
	}
	return nil
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
		if err = qrcode.WriteFile(string(buffer[:n]), qrcode.Medium, defaultQRSize, qrPath); err != nil {
			return err
		}
		indx++
	}
	return nil
}

func generateVideo() error {
	flags := strings.Split(strings.Trim(string(ffmpegFlags), "\n"), " ")
	args := append(flags, out)
	cmd := exec.Command("ffmpeg", args...)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func splitVideo() error {
	cmd := exec.Command("ffmpeg", "-i", in, "-vf", "fps=20", "./returned/%d.png")
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
	// FIX: set capacity as len(dirs) * image_size, leave as hardcoded for now
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

	result, err := zbarimgDecode(data)
	if err != nil {
		return err
	}

	if _, err := buff.Write(result); err != nil {
		return err
	}
	return nil
}
