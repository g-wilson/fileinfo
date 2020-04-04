package app

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/g-wilson/runtime/logger"
	"github.com/mostlygeek/go-exiftool"
	"github.com/sirupsen/logrus"
	"github.com/vansante/go-ffprobe"
)

type App struct {
	logger *logrus.Entry

	maxFileSize int64
	exiftool    *exiftool.Stayopen
	logGroup    string
	cloudwatch  *cloudwatchlogs.CloudWatchLogs
}

func New() (*App, error) {
	appLogger := logger.Create("fileinfo", os.Getenv("LOG_FORMAT"), os.Getenv("LOG_LEVEL"))

	awsConfig := aws.NewConfig().WithRegion(os.Getenv("AWS_REGION"))
	awsSession := session.Must(session.NewSession())

	maxFileSize, err := strconv.Atoi(os.Getenv("MAX_FILE_SIZE"))
	if err != nil {
		maxFileSize = 11e6 // 11 MB, allow for e.g. 10.1 MiB
	}

	ffprobe.SetFFProbeBinPath(os.Getenv("FFPROBE_BIN_PATH"))
	_, err = ffprobe.GetProbeData("", 1*time.Second)
	if err == ffprobe.ErrBinNotFound {
		return nil, fmt.Errorf("exiftool bin not found: %w", err)
	}

	et, err := exiftool.NewStayOpen(os.Getenv("EXIFTOOL_BIN_PATH"), "-json")
	if err != nil {
		return nil, fmt.Errorf("exiftool init error: %w", err)
	}

	app := &App{
		logger:      appLogger,
		exiftool:    et,
		maxFileSize: int64(maxFileSize),
		logGroup:    "/aws/lambda/fileinfo",
		cloudwatch:  cloudwatchlogs.New(awsSession, awsConfig),
	}

	return app, nil
}

func strPointer(str string) *string {
	return &str
}
