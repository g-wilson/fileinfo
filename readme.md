# Fileinfo

This application solves a common problem in HTTP API backends - reading metadata from user-uploaded content.

There are several issues with implementing a solution which reads a file into memory within your own application. It is advantageous to abstract this from your business logic by letting a service such as S3 handle file upload and storage. You can use this application as a microservice to read metadata by passing it a URL to the file. With S3 this would be a signed URL.

It is build using my [runtime](https://github.com/g-wilson/runtime) library for easy deployment on Lambda behind API Gateway.

It sues FFProbe and Exiftool as dependencies for some "analyzer" types. On Lambda they should be deployed as Lambda Layers.

### Configuration

`dotenv` is used locally as a configuration file to inject environment variables.

Here is what you need for local development:

```
ENV=local
HTTP_PORT=3030
LOG_LEVEL=debug
LOG_FORMAT=text
AWS_REGION=eu-west-1

# Fileinfo
FFPROBE_BIN_PATH=/usr/local/bin/ffprobe
EXIFTOOL_BIN_PATH=/usr/local/bin/exiftool
MAX_FILE_SIZE=11000000
```

### Deploying a lambda

Deployment is manual:

```
GOARCH=amd64 GOOS=linux go build ./cmd/main.go && zip fileinfo.zip main
```

Then upload the zip file to the lambda UI.
