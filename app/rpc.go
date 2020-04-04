package app

import (
	"github.com/g-wilson/runtime/rpcservice"
	"github.com/xeipuuv/gojsonschema"
)

func (a *App) RPCService() *rpcservice.Service {
	return rpcservice.NewService(a.logger).
		AddMethod("read_file", a.ReadFile, gojsonschema.NewStringLoader(`{
			"type": "object",
			"additionalProperties": false,
			"required": [ "url", "analyzers" ],
			"properties": {
				"url": {
					"type": "string",
					"minLength": 1
				},
				"analyzers":{
					"type": "array",
					"minItems": 1
				}
			}
		}`)).
		AddMethod("get_usage", a.GetUsage, gojsonschema.NewStringLoader(`{
			"type": "object",
			"additionalProperties": false,
			"required": [ "start_time", "end_time" ],
			"properties": {
				"start_time": {
					"type": "string",
					"format": "date-time"
				},
				"end_time": {
					"type": "string",
					"format": "date-time"
				}
			}
		}`))
}
