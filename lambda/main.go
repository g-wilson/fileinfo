package main

import (
	"fileinfo/handler"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	h, err := handler.Init()
	if err != nil {
		panic(err)
	}

	lambda.Start(h)
}
