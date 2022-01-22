package handler

import (
	"embed"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/g-wilson/runtime/ctxlog"
	"github.com/g-wilson/runtime/http"
	"github.com/g-wilson/runtime/schema"
	"github.com/mostlygeek/go-exiftool"
	"github.com/vansante/go-ffprobe"
)

//go:embed *.json
var fs embed.FS

func Init() (http.Handler, error) {
	log := ctxlog.Create("fileinfo", os.Getenv("LOG_FORMAT"), os.Getenv("LOG_LEVEL"))

	maxFileSize, err := strconv.Atoi(os.Getenv("MAX_FILE_SIZE"))
	if err != nil {
		maxFileSize = 11e6 // 11 MB, allow for e.g. 10.1 MiB
	}

	ffprobe.SetFFProbeBinPath(os.Getenv("FFPROBE_BIN_PATH"))
	_, err = ffprobe.GetProbeData("", 1*time.Second)
	if err == ffprobe.ErrBinNotFound {
		return nil, fmt.Errorf("ffprobe bin not found: %w", err)
	}

	et, err := exiftool.NewStayOpen(os.Getenv("EXIFTOOL_BIN_PATH"), "-json")
	if err != nil {
		return nil, fmt.Errorf("exiftool init error: %w", err)
	}

	fn := &Function{
		exiftool:    et,
		maxFileSize: int64(maxFileSize),
	}

	handler, err := http.NewJSONHandler(fn.Do, schema.MustLoad(fs, "schema.json"))
	if err != nil {
		return nil, err
	}

	return http.WithMiddleware(
		handler,
		http.CreateRequestLogger(log),
		http.JSONErrorHandler,
	), nil
}
