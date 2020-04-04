package app

import (
	"context"
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/g-wilson/runtime/logger"
	"github.com/sirupsen/logrus"
	"github.com/vansante/go-ffprobe"
)

const (
	AnalyzerTypeMD5     = "md5"
	AnalyzerTypeSHA1    = "sha1"
	AnalyzerTypeEXIF    = "exif"
	AnalyzerTypeFFProbe = "ffprobe"
)

type ReadFileRequest struct {
	URL       string        `json:"url"`
	Analyzers AnalyzersList `json:"analyzers"`
}

type ReadFileResponse struct {
	Size      int64  `json:"size"`
	SizeHuman string `json:"size_human"`
	Mimetype  string `json:"mimetype"`

	DigestMD5  string `json:"digest_md5,omitempty"`
	DigestSHA1 string `json:"digest_sha1,omitempty"`

	EXIF    map[string]interface{} `json:"exif,omitempty"`
	FFProbe *ffprobe.ProbeData     `json:"ffprobe,omitempty"`
}

type AnalyzersList []string

func (l AnalyzersList) Has(in string) bool {
	for _, str := range l {
		if str == in {
			return true
		}
	}

	return false
}

func (a *App) ReadFile(ctx context.Context, req *ReadFileRequest) (*ReadFileResponse, error) {
	handlerStartedAt := time.Now()
	reqLogger := logger.FromContext(ctx)

	file, err := NewFileFromURL(ctx, req.URL, a.maxFileSize)
	if err != nil {
		if err.Error() == ErrFileTooLarge.Error() {
			reqLogger.Update(reqLogger.Entry().WithField("file_size", a.maxFileSize))
		}
		return nil, err
	}
	defer file.Cleanup()

	reqLogger.Update(reqLogger.Entry().WithField("file_dl_duration", time.Now().Sub(handlerStartedAt).String()))

	response := &ReadFileResponse{
		Size:      file.Size,
		SizeHuman: humanize.Bytes(uint64(file.Size)),
	}

	infoStartedAt := time.Now()

	err = a.doSerially(ctx, file, req, response)

	reqLogger.Update(reqLogger.Entry().WithFields(logrus.Fields{
		"file_size":          response.Size,
		"file_info_duration": time.Now().Sub(infoStartedAt).String(),
	}))

	return response, err
}

func (a *App) doSerially(ctx context.Context, file *File, req *ReadFileRequest, response *ReadFileResponse) error {
	file.mutex.Lock()
	defer file.mutex.Unlock()

	reqLog := logger.FromContext(ctx).Entry()

	mimetype, err := Mimetype(ctx, file.file)
	if err != nil {
		return fmt.Errorf("analyzer error (mimetype): %w", err)
	}

	response.Mimetype = mimetype

	if req.Analyzers.Has(AnalyzerTypeMD5) {
		digestmd5, err := DigestMD5(ctx, file.file)
		if err != nil {
			return fmt.Errorf("analyzer error (md5): %w", err)
		}

		response.DigestMD5 = digestmd5
	}

	if req.Analyzers.Has(AnalyzerTypeSHA1) {
		digestsha1, err := DigestSHA1(ctx, file.file)
		if err != nil {
			return fmt.Errorf("analyzer error (sha1): %w", err)
		}

		response.DigestSHA1 = digestsha1
	}

	if req.Analyzers.Has(AnalyzerTypeEXIF) {
		exifInfo, err := EXIF(ctx, file.file, a.exiftool)
		if err == nil {
			response.EXIF = exifInfo
		} else {
			response.EXIF = nil
			reqLog.WithError(err).Warn("analyzer error (exif)")
		}
	}

	if req.Analyzers.Has(AnalyzerTypeFFProbe) {
		ffprobeInfo, err := FFProbe(ctx, file.file)
		if err == nil {
			response.FFProbe = ffprobeInfo
		} else {
			response.FFProbe = nil
			reqLog.WithError(err).Warn("analyzer error (ffprobe)")
		}
	}

	return nil
}
