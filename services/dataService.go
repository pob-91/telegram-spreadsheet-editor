package services

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"nextcloud-spreadsheet-editor/utils"
	"os"
	"path"

	"go.uber.org/zap"
)

type IDataService interface {
	GetSpreadsheet() (io.Reader, error)
}

type NCDataService struct {
	Http utils.IHttpClient
}

const (
	BASE_URL_KEY       string = "NEXTCLOUD_WEBDAV_BASE_URL"
	XLSX_FILE_PATH_KEY string = "XLSX_FILE_PATH"
	USER_KEY           string = "NEXTCLOUD_USER"
	PASSWORD_KEY       string = "NEXTCLOUD_PASSWORD"
)

func (s *NCDataService) GetSpreadsheet() (io.Reader, error) {
	baseUrl := os.Getenv(BASE_URL_KEY)
	filePath := os.Getenv(XLSX_FILE_PATH_KEY)

	u, err := url.Parse(baseUrl)
	if err != nil {
		zap.L().DPanic("Failed to parse base url", zap.String("url", baseUrl), zap.Error(err))
		return nil, fmt.Errorf("URL error")
	}

	user := os.Getenv(USER_KEY)
	password := os.Getenv(PASSWORD_KEY)

	u.Path = path.Join(u.Path, user, filePath)

	// TODO: Modify this to not need to send the response string and just get the bytes
	var responseString string
	response, err := s.Http.Get(u.String(), &responseString, &utils.HttpOptions{
		BasicAuthUser:     &user,
		BasicAuthPassword: &password,
	})
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
