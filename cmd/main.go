package main

/*

This file is the Lambda-specific entry-point.

All dependencies and config should be handled in the constructor function of the application.

*/

import (
	fileinfo "fileinfo/app"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	app, err := fileinfo.New()
	if err != nil {
		panic(err)
	}

	lambda.Start(app.RPCService().WrapAPIGatewayHTTP())
}
