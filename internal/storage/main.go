package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

type Storage struct {
	server string
	token  string
	logger *zap.Logger
}

func New(token string, logger *zap.Logger) *Storage {
	return &Storage{
		token:  token,
		logger: logger,
		server: "https://api.gofile.io/getServer",
	}
}

func (s *Storage) getAvailableServer() string {
	res, err := http.Get(s.server)
	if err != nil {
		s.logger.Error("server unavailable", zap.Error(err))
		return ""
	}
	defer res.Body.Close()

	respData, err := io.ReadAll(res.Body)
	if err != nil {
		s.logger.Error("can't read server response", zap.Error(err))
		return ""
	}

	resp := new(ServerResponse)
	if err := json.Unmarshal(respData, resp); err != nil {
		s.logger.Error("can't unmarshal server response", zap.Error(err))
		return ""
	}

	if resp.Status == "ok" {
		return resp.Data.Server
	}

	s.logger.Error("status not ok", zap.String("status", resp.Status))
	return ""
}

func (s *Storage) Upload(fileLocation string) (string, error) {
	server := s.getAvailableServer()
	if server == "" {
		s.logger.Error("no available server")
		return "", fmt.Errorf("no available server")
	}
	endpoint := fmt.Sprintf("https://%s.gofile.io/uploadFile", server)

	f, err := os.Open(fileLocation)
	if err != nil {
		s.logger.Error("can't open file", zap.Error(err))
		return "", fmt.Errorf("can't open file")
	}

	defer func() {
		_ = f.Close()
	}()

	data, w := prepareData(
		f,
		map[string]string{
			"token": s.token,
			// expire not really useful, no need to use it
			//"expire": strconv.FormatInt(time.Now().Add(time.Hour*72).Unix(), 10),
		},
		s.logger,
	)

	if data == nil || w == nil {
		s.logger.Error("couldn't prepare data to loading, exiting..")
		return "", fmt.Errorf("something went wrong while preparing data to send")
	}

	r, err := http.NewRequest(http.MethodPost, endpoint, data)
	if err != nil {
		s.logger.Error("can't prepare request", zap.Error(err))
		return "", fmt.Errorf("can't send request")
	}

	r.Header.Set("Content-Type", w.FormDataContentType())

	res, err := http.DefaultClient.Do(r)
	if err != nil {
		s.logger.Error("can't send request", zap.Error(err))
		return "", fmt.Errorf("can't send request")
	}
	defer res.Body.Close()

	respData, err := io.ReadAll(res.Body)
	if err != nil {
		s.logger.Error("can't read server's response", zap.Error(err))
		return "", fmt.Errorf("can't upload file")
	}

	uploadInfo := new(UploadResponse)
	if err := json.Unmarshal(respData, uploadInfo); err != nil {
		s.logger.Error("can't parse server's response", zap.Error(err))
		return "", fmt.Errorf("can't upload file")
	}

	if uploadInfo.Status == "ok" {
		return fmt.Sprintf("gofile.io/d/%s", uploadInfo.Data.Code), nil
	}

	s.logger.Error("status not ok", zap.Any("response", uploadInfo))
	return "", fmt.Errorf("can't upload file")
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

func prepareData(file *os.File, params map[string]string, logger *zap.Logger) (*bytes.Buffer, *multipart.Writer) {
	b := new(bytes.Buffer)
	w := multipart.NewWriter(b)
	defer func() {
		_ = w.Close()
	}()

	if err := multipartFile(w, file); err != nil {
		logger.Error("can't prepare file", zap.Error(err))
		return nil, nil
	}

	if len(params) > 0 {
		for name, value := range params {
			if err := multipartField(w, name, value); err != nil {
				logger.Error("can't prepare fields", zap.Error(err))
				return nil, nil
			}
		}
	}

	return b, w
}
