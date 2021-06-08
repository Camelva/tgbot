package storage

import (
	"io"
	"mime/multipart"
	"os"
)

func multipartField(w *multipart.Writer, name, value string) error {
	ew, err := w.CreateFormField(name)
	if err != nil {
		return err
	}

	if _, err := ew.Write([]byte(value)); err != nil {
		return err
	}
	return nil
}

func multipartFile(w *multipart.Writer, file *os.File) error {
	fw, err := w.CreateFormFile("file", file.Name())
	if err != nil {
		return err
	}

	if _, err := io.Copy(fw, file); err != nil {
		return err
	}
	return nil
}
