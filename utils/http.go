package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

type IHttpClient interface {
	Get(url string, responseBody any, opts ...*HttpOptions) (*HttpResponse, error)
	Put(url string, body *bytes.Buffer, responseBody any, opts ...*HttpOptions) (*HttpResponse, error)
}

type HttpClient struct{}

type HttpResponse struct {
	StatusCode         int
	Body               *[]byte
	ContentType        *string
	ContentDisposition *string
	Length             *int64
}

type HttpOptions struct {
	Headers           *map[string]string
	BasicAuthUser     *string
	BasicAuthPassword *string
	ContentType       string
}

func (h *HttpClient) Get(url string, responseBody any, opts ...*HttpOptions) (*HttpResponse, error) {
	options := HttpOptions{}
	if len(opts) > 0 {
		options = *opts[0]
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if options.Headers != nil {
		for key, value := range *options.Headers {
			req.Header.Set(key, value)
		}
	}

	if options.BasicAuthUser != nil && options.BasicAuthPassword != nil {
		req.SetBasicAuth(*options.BasicAuthUser, *options.BasicAuthPassword)
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	r := HttpResponse{
		StatusCode: response.StatusCode,
	}

	if (response.ContentLength == -1 || response.ContentLength > 0) && responseBody != nil {
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}

		if response.StatusCode > 299 {
			zap.L().Warn("Got error code with response from http request", zap.String("response", string(bodyBytes)))
			return &HttpResponse{
				StatusCode: response.StatusCode,
			}, nil
		}

		r.Body = &bodyBytes

		contentType := response.Header.Get("Content-Type")
		if len(contentType) > 0 {
			r.ContentType = &contentType
		}

		contentDisposition := response.Header.Get("Content-Disposition")
		if len(contentDisposition) > 0 {
			r.ContentDisposition = &contentDisposition
		}

		r.Length = &response.ContentLength

		switch {
		case strings.Contains(contentType, "application/json"):
			if err := json.Unmarshal(bodyBytes, responseBody); err != nil {
				return nil, err
			}
			break
		case strings.Contains(contentType, "plain/text") || strings.Contains(contentType, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"):
			s := string(bodyBytes)
			responseBody = &s
			break
		default:
			zap.L().Warn("Unsupported response type, not decoding", zap.String("type", contentType))
			break
		}
	}

	return &r, nil
}

func (h *HttpClient) Put(url string, body *bytes.Buffer, responseBody any, opts ...*HttpOptions) (*HttpResponse, error) {
	options := HttpOptions{}
	if len(opts) > 0 {
		options = *opts[0]
	}

	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", options.ContentType)

	if options.Headers != nil {
		for key, value := range *options.Headers {
			req.Header.Set(key, value)
		}
	}

	if options.BasicAuthUser != nil && options.BasicAuthPassword != nil {
		req.SetBasicAuth(*options.BasicAuthUser, *options.BasicAuthPassword)
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	r := HttpResponse{
		StatusCode: response.StatusCode,
	}

	if (response.ContentLength == -1 || response.ContentLength > 0) && responseBody != nil {
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}

		if response.StatusCode > 299 {
			zap.L().Warn("Got error code with response from http request", zap.String("response", string(bodyBytes)))
			return &HttpResponse{
				StatusCode: response.StatusCode,
			}, nil
		}

		r.Body = &bodyBytes

		contentType := response.Header.Get("Content-Type")
		if len(contentType) > 0 {
			r.ContentType = &contentType
		}

		contentDisposition := response.Header.Get("Content-Disposition")
		if len(contentDisposition) > 0 {
			r.ContentDisposition = &contentDisposition
		}

		r.Length = &response.ContentLength

		switch {
		case strings.Contains(contentType, "application/json"):
			if err := json.Unmarshal(bodyBytes, responseBody); err != nil {
				return nil, err
			}
			break
		case strings.Contains(contentType, "plain/text"):
			s := string(bodyBytes)
			responseBody = &s
			break
		default:
			zap.L().Warn("Unsupported response type, not decoding", zap.String("type", contentType))
			break
		}
	}

	return &r, nil
}
