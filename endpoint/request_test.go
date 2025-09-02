package endpoint_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lansfy/gonkex/runner"

	"github.com/stretchr/testify/require"
)

// part of tests for request was implemented in mocks tests

func TestUploadFiles(t *testing.T) {
	srv := testServerUpload(t)
	defer srv.Close()

	runner.RunWithTesting(t, srv.URL, &runner.RunWithTestingOpts{
		TestsDir: "testdata/upload-files",
	})
}

func TestMultipartFormData(t *testing.T) {
	srv := testServerMultipartFormData(t)
	defer srv.Close()

	runner.RunWithTesting(t, srv.URL, &runner.RunWithTestingOpts{
		TestsDir: "testdata/multipart/form-data.yaml",
	})
}

type response struct {
	Status            string `json:"status"`
	File1Name         string `json:"file_1_name"`
	File1Content      string `json:"file_1_content"`
	File2Name         string `json:"file_2_name"`
	File2Content      string `json:"file_2_content"`
	FieldsTestName    string `json:"fields_test_name"`
	FieldsTestContent string `json:"fields_test_content"`
}

func testServerUpload(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := response{
			Status: "OK",
		}

		resp.File1Name, resp.File1Content = formFile(t, r, "file1")
		resp.File2Name, resp.File2Content = formFile(t, r, "file2")

		resp.FieldsTestName, resp.FieldsTestContent = "fieldTest", r.FormValue("fieldTest")

		respData, err := json.Marshal(resp)
		require.NoError(t, err)

		w.Header().Set("Content-Type", "application/json")

		_, err = w.Write(respData)
		require.NoError(t, err)
	}))
}

func formFile(t *testing.T, r *http.Request, field string) (string, string) {
	file, header, err := r.FormFile(field)
	require.NoError(t, err)

	defer func() { _ = file.Close() }()

	contents, err := io.ReadAll(file)
	require.NoError(t, err)

	return header.Filename, string(contents)
}

type multipartResponse struct {
	ContentTypeHeader  string `json:"content_type_header"`
	RequestBodyContent string `json:"request_body_content"`
}

func testServerMultipartFormData(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		r.Body = io.NopCloser(bytes.NewReader(body))

		resp := multipartResponse{
			ContentTypeHeader:  r.Header.Get("Content-Type"),
			RequestBodyContent: string(body),
		}

		respData, err := json.Marshal(resp)
		require.NoError(t, err)

		w.Header().Set("Content-Type", "application/json")

		_, err = w.Write(respData)
		require.NoError(t, err)
	}))
}
