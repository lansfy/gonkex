package endpoint

import (
	"bytes"
	"fmt"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lansfy/gonkex/models"
)

const (
	headerContentType       = "Content-Type"
	headerHost              = "Host"
	headerMultipartFormData = "multipart/form-data"
	headerApplicationJSON   = "application/json"
)

func NewRequest(host string, test models.TestInterface) (*http.Request, string, error) {
	if test.GetForm() != nil {
		return newMultipartRequest(host, test)
	}
	return newCommonRequest(host, test)
}

func newMultipartRequest(host string, test models.TestInterface) (*http.Request, string, error) {
	buff := &bytes.Buffer{}
	w := multipart.NewWriter(buff)

	err := addBoundary(test.ContentType(), w)
	if err != nil {
		return nil, "", err
	}

	form := test.GetForm()

	err = addFormFields(form.GetFields(), w)
	if err != nil {
		return nil, "", err
	}

	err = addFiles(form.GetFiles(), w)
	if err != nil {
		return nil, "", err
	}

	_ = w.Close()

	req, err := makeRequest(test, buff, host)
	if err != nil {
		return nil, "", err
	}

	// this is necessary, it will contain boundary
	req.Header.Set(headerContentType, w.FormDataContentType())

	return req, buff.String(), nil
}

func addBoundary(contentTypeValue string, w *multipart.Writer) error {
	if contentTypeValue == "" {
		return nil
	}

	contentType, params, err := mime.ParseMediaType(contentTypeValue)
	if err != nil {
		return fmt.Errorf("parse %s '%s': %w", headerContentType, contentTypeValue, err)
	}
	if contentType != headerMultipartFormData {
		return fmt.Errorf(
			"form support only %s '%s' ('%s' provided)",
			headerContentType, headerMultipartFormData, contentType,
		)
	}

	boundary := params["boundary"]
	if boundary == "" {
		return nil
	}
	err = w.SetBoundary(boundary)
	if err != nil {
		return fmt.Errorf("set custom boundary '%s': %w", boundary, err)
	}
	return nil
}

func addFiles(files map[string]string, w *multipart.Writer) error {
	for name, path := range files {
		err := addFile(path, w, name)
		if err != nil {
			return err
		}
	}

	return nil
}

func addFile(filename string, w *multipart.Writer, fieldname string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	fw, err := w.CreateFormFile(fieldname, filepath.Base(filename))
	if err != nil {
		return err
	}

	_, err = fw.Write(content)
	return err
}

func addFormFields(fields map[string]string, w *multipart.Writer) error {
	fieldNames := []string{}
	for n := range fields {
		fieldNames = append(fieldNames, n)
	}
	sort.Strings(fieldNames)

	for _, name := range fieldNames {
		err := w.WriteField(name, fields[name])
		if err != nil {
			return err
		}
	}

	return nil
}

func newCommonRequest(host string, test models.TestInterface) (*http.Request, string, error) {
	body := test.GetRequest()
	req, err := makeRequest(test, bytes.NewBuffer([]byte(body)), host)
	if err != nil {
		return nil, "", err
	}

	if req.Header.Get(headerContentType) == "" {
		req.Header.Set(headerContentType, headerApplicationJSON)
	}

	return req, body, nil
}

func makeRequest(test models.TestInterface, body *bytes.Buffer, host string) (*http.Request, error) {
	req, err := http.NewRequest(
		strings.ToUpper(test.GetMethod()),
		host+test.Path()+test.ToQuery(),
		body,
	)
	if err != nil {
		return nil, err
	}

	for k, v := range test.Headers() {
		if strings.EqualFold(k, headerHost) {
			req.Host = v
		} else {
			req.Header.Add(k, v)
		}
	}

	for k, v := range test.Cookies() {
		req.AddCookie(&http.Cookie{
			Name:  k,
			Value: v,
		})
	}

	return req, nil
}
