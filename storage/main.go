package storage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	stdLog "log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"time"
)

var token string

func SetToken(s string) {
	token = s
}

type ServerResponse struct {
	Status string `json:"status"`
	Data   struct {
		Server string `json:"server"`
	} `json:"data"`
}

type UploadResponse struct {
	Status string `json:"status"`
	Data   struct {
		Code      string `json:"code"`
		AdminCode string `json:"adminCode"`
		FileName  string `json:"fileName"`
		MD5       string `json:"md5"`
	} `json:"data"`
}

func getAvailableServer() string {
	endpoint := "https://api.gofile.io/getServer"

	res, err := http.Get(endpoint)
	if err != nil {
		stdLog.Println(err)
		return ""
	}
	defer res.Body.Close()

	respData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		stdLog.Println(err)
		return ""
	}

	s := new(ServerResponse)
	if err := json.Unmarshal(respData, s); err != nil {
		stdLog.Println(err)
		return ""
	}

	if s.Status == "ok" {
		return s.Data.Server
	}
	stdLog.Println(s)
	return ""
}

func Upload(file *os.File) (string, error) {
	makeEndpoint := func(server string) string {
		return fmt.Sprintf("https://%s.gofile.io/uploadFile", server)
	}

	server := getAvailableServer()
	if server == "" {
		return "", errors.New("no available server")
	}

	b, w := prepareData(
		file,
		map[string]string{
			// "token": token,
			// currently email used to identify
			"email": token,
			// seems like expire field doesn't work but lets set it for future usage
			"expire": strconv.FormatInt(time.Now().Add(time.Hour*72).Unix(), 10),
		},
	)

	if b == nil || w == nil {
		return "", fmt.Errorf("something went wrong while preparing data to send")
	}

	r, err := http.NewRequest(http.MethodPost, makeEndpoint(server), b)
	if err != nil {
		return "", fmt.Errorf("can't create request: %s", err)
	}

	r.Header.Set("Content-Type", w.FormDataContentType())

	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return "", fmt.Errorf("can't do request: %s", err)
	}
	defer res.Body.Close()

	respData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("can't read server response: %s", err)
	}

	uploadInfo := new(UploadResponse)
	if err := json.Unmarshal(respData, uploadInfo); err != nil {
		return "", fmt.Errorf("can't decode server response: %s", err)
	}

	if uploadInfo.Status == "ok" {
		return fmt.Sprintf("gofile.io/d/%s", uploadInfo.Data.Code), nil
	}
	stdLog.Println(uploadInfo)
	return "", fmt.Errorf("something's wrong: %v", uploadInfo)
}

func prepareData(file *os.File, params map[string]string) (*bytes.Buffer, *multipart.Writer) {
	b := new(bytes.Buffer)
	w := multipart.NewWriter(b)
	defer func() {
		_ = w.Close()
	}()

	if err := addFile(w, file); err != nil {
		stdLog.Println(err)
		return nil, nil
	}

	if len(params) > 0 {
		for name, value := range params {
			if err := addField(w, name, value); err != nil {
				stdLog.Println(err)
				return nil, nil
			}
		}
	}

	return b, w
}

func addFile(w *multipart.Writer, file *os.File) error {
	fw, err := w.CreateFormFile("file", file.Name())
	if err != nil {
		return err
	}

	if _, err := io.Copy(fw, file); err != nil {
		return err
	}
	return nil
}

func addField(w *multipart.Writer, name, value string) error {
	ew, err := w.CreateFormField(name)
	if err != nil {
		return err
	}

	if _, err := ew.Write([]byte(value)); err != nil {
		return err
	}
	return nil
}
