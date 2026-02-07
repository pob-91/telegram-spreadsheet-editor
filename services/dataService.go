package services

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"telegram-spreadsheet-editor/model"
	"telegram-spreadsheet-editor/utils"

	"go.uber.org/zap"
)

type IDataService interface {
	GetSpreadsheet(source model.SpreadsheetSource) (io.Reader, error)
	WriteSpreadsheet(source model.SpreadsheetSource, sheet io.Reader) error
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

func (s *NCDataService) GetSpreadsheet(source model.SpreadsheetSource) (io.Reader, error) {
	ncSource, err := getSource(source)
	if err != nil {
		return nil, err
	}

	fileUrl, err := getFileUrl(ncSource)
	if err != nil {
		return nil, err
	}

	password, exists := os.LookupEnv(ncSource.PasswordEnv)
	if !exists {
		zap.L().Error("Expected to find nextcloud user password", zap.String("var", ncSource.PasswordEnv))
		return nil, fmt.Errorf("Expected to find nextcloud user password")
	}

	opts := utils.HttpOptions{}
	if len(ncSource.User) > 0 && len(password) > 0 {
		opts.BasicAuthUser = &ncSource.User
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

func (s *NCDataService) WriteSpreadsheet(source model.SpreadsheetSource, sheet io.Reader) error {
	ncSource, err := getSource(source)
	if err != nil {
		return err
	}

	fileUrl, err := getFileUrl(ncSource)
	if err != nil {
		return err
	}

	b, err := io.ReadAll(sheet)
	if err != nil {
		zap.L().DPanic("Failed to read sheet", zap.Error(err))
		return fmt.Errorf("Failed to read sheet")
	}

	password, exists := os.LookupEnv(ncSource.PasswordEnv)
	if !exists {
		zap.L().Error("Expected to find nextcloud user password.", zap.String("var", ncSource.PasswordEnv))
		return fmt.Errorf("Expected to find nextcloud user password")
	}

	opts := utils.HttpOptions{
		ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	}

	if len(ncSource.User) > 0 && len(password) > 0 {
		opts.BasicAuthUser = &ncSource.User
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

func getSource(source model.SpreadsheetSource) (*model.NextcloudSpreadsheetSource, error) {
	if source == nil {
		zap.L().DPanic("Spreadsheet source is nil.")
		return nil, fmt.Errorf("Spreadsheet source is nil")
	}
	ncSource, ok := source.(*model.NextcloudSpreadsheetSource)
	if !ok {
		zap.L().Error("Expected nextcloud spreadsheet source.", zap.String("type", source.GetType()))
		return nil, fmt.Errorf("Expected nextcloud spreadsheet source")
	}

	return ncSource, nil
}

func getFileUrl(source *model.NextcloudSpreadsheetSource) (string, error) {
	u, err := url.Parse(source.BaseUrl)
	if err != nil {
		zap.L().DPanic("Failed to parse base url", zap.String("url", source.BaseUrl), zap.Error(err))
		return "", fmt.Errorf("URL error")
	}

	u.Path = path.Join(u.Path, source.FilePath)

	return u.String(), nil
}
