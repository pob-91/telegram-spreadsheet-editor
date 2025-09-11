package services

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"telegram-spreadsheet-editor/utils"

	"go.uber.org/zap"
)

type IDataService interface {
	GetSpreadsheet() (io.Reader, error)
	WriteSpreadsheet(sheet io.Reader) error
}

type NCDataService struct {
	Http utils.IHttpClient
}

const (
	BASE_URL_KEY       string = "SHEET_BASE_URL"
	XLSX_FILE_PATH_KEY string = "XLSX_FILE_PATH"
	USER_KEY           string = "BASIC_AUTH_USER"
	PASSWORD_KEY       string = "BASIC_AUTH_PASSWORD"
)

func (s *NCDataService) GetSpreadsheet() (io.Reader, error) {
	fileUrl, err := getFileUrl()
	if err != nil {
		return nil, err
	}

	user := os.Getenv(USER_KEY)
	password := os.Getenv(PASSWORD_KEY)

	opts := utils.HttpOptions{}
	if len(user) > 0 && len(password) > 0 {
		opts.BasicAuthUser = &user
		opts.BasicAuthPassword = &password
	}

	// TODO: Modify this to not need to send the response string and just get the bytes
	var responseString string
	response, err := s.Http.Get(fileUrl, &responseString, &opts)
	if err != nil {
		zap.L().Error("Failed to download file", zap.Int("response", response.StatusCode), zap.Error(err))
		return nil, fmt.Errorf("Failed to download file")
	}

	if response.StatusCode != 200 {
		zap.L().Error("Non 200 response code", zap.Int("response", response.StatusCode))
		return nil, fmt.Errorf("Non 200 response code")
	}

	return bytes.NewReader(*response.Body), nil
}

func (s *NCDataService) WriteSpreadsheet(sheet io.Reader) error {
	fileUrl, err := getFileUrl()
	if err != nil {
		return err
	}

	b, err := io.ReadAll(sheet)
	if err != nil {
		zap.L().DPanic("Failed to read sheet", zap.Error(err))
		return fmt.Errorf("Failed to read sheet")
	}

	user := os.Getenv(USER_KEY)
	password := os.Getenv(PASSWORD_KEY)

	opts := utils.HttpOptions{
		ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	}

	if len(user) > 0 && len(password) > 0 {
		opts.BasicAuthUser = &user
		opts.BasicAuthPassword = &password
	}

	buffer := bytes.NewBuffer(b)
	response, err := s.Http.Put(fileUrl, buffer, nil, &opts)
	if err != nil {
		zap.L().Error("Failed to upload file", zap.Int("response", response.StatusCode), zap.Error(err))
		return fmt.Errorf("Failed to upload file")
	}

	return nil
}

// private
func getFileUrl() (string, error) {
	baseUrl := os.Getenv(BASE_URL_KEY)
	filePath := os.Getenv(XLSX_FILE_PATH_KEY)

	u, err := url.Parse(baseUrl)
	if err != nil {
		zap.L().DPanic("Failed to parse base url", zap.String("url", baseUrl), zap.Error(err))
		return "", fmt.Errorf("URL error")
	}

	u.Path = path.Join(u.Path, filePath)

	return u.String(), nil
}
