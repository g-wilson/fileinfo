package app

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/g-wilson/runtime/hand"
)

var (
	ErrFileTooLarge   = hand.New("file_too_large")
	ErrRequestTimeout = hand.New("request_timeout")
	ErrInvalidURL     = hand.New("invalid_url")
	ErrNoResponseBody = hand.New("no_response_body")
)

type File struct {
	file  *os.File
	mutex sync.Mutex

	Size int64
}

func NewFileFromURL(ctx context.Context, url string, maxSizeBytes int64) (*File, error) {
	body, err := doRequest(ctx, url)
	if err != nil {
		if _, ok := err.(hand.E); !ok {
			return nil, fmt.Errorf("cannot fetch remote file: %w", err)
		}
		return nil, err
	}
	defer body.Close()

	tmpFile, err := ioutil.TempFile(os.TempDir(), "fileinfo-")
	if err != nil {
		return nil, fmt.Errorf("cannot create temporary file: %w", err)
	}

	f := &File{
		file: tmpFile,
	}

	bytesWritten, err := io.CopyN(tmpFile, body, maxSizeBytes)
	if err != nil && err != io.EOF {
		_ = f.Cleanup()
		return nil, fmt.Errorf("error saving file to disk: %w", err)
	} else if err == nil {
		return nil, ErrFileTooLarge
	}

	f.Size = bytesWritten

	return f, nil
}

func (f *File) Cleanup() error {
	err := f.file.Close()
	if err != nil {
		return fmt.Errorf("error closing temp file: %w", err)
	}

	err = os.Remove(f.file.Name())
	if err != nil {
		return fmt.Errorf("error removing temp file: %w", err)
	}

	return nil
}

func doRequest(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, ErrInvalidURL
	}

	req.Header = http.Header{
		"User-Agent": []string{"fileinfo-reader (+https://github.com/g-wilson/fileinfo)"},
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if netErr, ok := err.(net.Error); ok {
			if netErr.Timeout() {
				return nil, ErrRequestTimeout
			}
		}

		return nil, err
	}

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		if res.StatusCode == http.StatusNoContent || res.Body == nil {
			return nil, ErrNoResponseBody
		}

		return res.Body, nil
	}

	statusText := http.StatusText(res.StatusCode)
	if statusText == "" {
		statusText = "unknown"
	}

	statusParts := strings.Fields(statusText)

	for i := range statusParts {
		statusParts[i] = strings.ToLower(statusParts[i])
	}

	return nil, hand.New("request_failed").WithMeta(hand.M{
		"status_text": strings.Join(statusParts, "_"),
		"http_status": res.StatusCode,
	})
}
