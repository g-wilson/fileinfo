package handler

import (
	"context"
	"fmt"
	"time"

	"fileinfo/internal/file"

	"github.com/dustin/go-humanize"
	"github.com/g-wilson/runtime/ctxlog"
	"github.com/mostlygeek/go-exiftool"
	"github.com/sirupsen/logrus"
	"github.com/vansante/go-ffprobe"
)

const (
	AnalyzerTypeMD5     = "md5"
	AnalyzerTypeSHA1    = "sha1"
	AnalyzerTypeEXIF    = "exif"
	AnalyzerTypeFFProbe = "ffprobe"
)

type Function struct {
	maxFileSize int64
	exiftool    *exiftool.Stayopen
}

type Request struct {
	URL       string        `json:"url"`
	Analyzers AnalyzersList `json:"analyzers"`
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

type Response struct {
	Size      int64  `json:"size"`
	SizeHuman string `json:"size_human"`
	Mimetype  string `json:"mimetype"`

	DigestMD5  string `json:"digest_md5,omitempty"`
	DigestSHA1 string `json:"digest_sha1,omitempty"`

	EXIF    map[string]interface{} `json:"exif,omitempty"`
	FFProbe *ffprobe.ProbeData     `json:"ffprobe,omitempty"`
}

func (f *Function) Do(ctx context.Context, req *Request) (*Response, error) {
	handlerStartedAt := time.Now()
	reqLog := ctxlog.FromContext(ctx)

	fileObj, err := file.NewFromURL(ctx, req.URL, f.maxFileSize)
	if err != nil {
		if err.Error() == file.ErrFileTooLarge.Error() {
			reqLog.Update(reqLog.Entry().WithField("file_size", f.maxFileSize))
		}
		return nil, err
	}
	defer fileObj.Cleanup()

	reqLog.Update(reqLog.Entry().WithField("file_dl_duration", time.Now().Sub(handlerStartedAt).String()))

	response := &Response{
		Size:      fileObj.Size(),
		SizeHuman: humanize.Bytes(uint64(fileObj.Size())),
	}

	infoStartedAt := time.Now()

	mimetype, err := Mimetype(ctx, fileObj.File())
	if err != nil {
		return nil, fmt.Errorf("analyzer error (mimetype): %w", err)
	}

	response.Mimetype = mimetype

	if req.Analyzers.Has(AnalyzerTypeMD5) {
		digestmd5, err := DigestMD5(ctx, fileObj.File())
		if err != nil {
			return nil, fmt.Errorf("analyzer error (md5): %w", err)
		}

		response.DigestMD5 = digestmd5
	}

	if req.Analyzers.Has(AnalyzerTypeSHA1) {
		digestsha1, err := DigestSHA1(ctx, fileObj.File())
		if err != nil {
			return nil, fmt.Errorf("analyzer error (sha1): %w", err)
		}

		response.DigestSHA1 = digestsha1
	}

	if req.Analyzers.Has(AnalyzerTypeEXIF) {
		exifInfo, err := EXIF(ctx, fileObj.File(), f.exiftool)
		if err == nil {
			response.EXIF = exifInfo
		} else {
			response.EXIF = nil
			reqLog.Entry().WithError(err).Warn("analyzer error (exif)")
		}
	}

	if req.Analyzers.Has(AnalyzerTypeFFProbe) {
		ffprobeInfo, err := FFProbe(ctx, fileObj.File())
		if err == nil {
			response.FFProbe = ffprobeInfo
		} else {
			response.FFProbe = nil
			reqLog.Entry().WithError(err).Warn("analyzer error (ffprobe)")
		}
	}

	reqLog.Update(reqLog.Entry().WithFields(logrus.Fields{
		"file_size":          response.Size,
		"file_info_duration": time.Now().Sub(infoStartedAt).String(),
	}))

	return response, err
}
