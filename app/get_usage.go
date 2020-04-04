package app

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/g-wilson/runtime/hand"
)

type GetUsageRequest struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

type GetUsageResponse struct {
	*GetUsageRequest
	MinFileSize  int64 `json:"min_file_size"`
	MaxFileSize  int64 `json:"max_file_size"`
	MeanFileSize int64 `json:"mean_file_size"`
	SumFileSize  int64 `json:"sum_file_size"`
	RequestCount int64 `json:"request_count"`
}

var queryStr = `fields file_size, handler_duration
| filter file_size > 0
| stats count(), sum(file_size), avg(file_size), min(file_size), max(file_size)`

func (a *App) GetUsage(ctx context.Context, req *GetUsageRequest) (*GetUsageResponse, error) {
	q, err := a.cloudwatch.StartQueryWithContext(ctx, &cloudwatchlogs.StartQueryInput{
		QueryString:  strPointer(queryStr),
		StartTime:    timeEpochPointer(req.StartTime),
		EndTime:      timeEpochPointer(req.EndTime),
		LogGroupName: strPointer(a.logGroup),
	})
	if err != nil {
		return nil, hand.Wrap("query_error", err)
	}

	results, err := getQueryResults(ctx, a.cloudwatch, q.QueryId)
	if err != nil {
		return nil, err
	}

	if len(results.Results) < 1 {
		return &GetUsageResponse{}, nil
	}

	var asInt = make([]int64, len(results.Results[0]))
	for i, col := range results.Results[0] {
		v, err := strconv.Atoi(*col.Value)
		if err != nil {
			return nil, fmt.Errorf("error parsing result: %v", err)
		}
		asInt[i] = int64(v)
	}

	return &GetUsageResponse{
		GetUsageRequest: req,
		RequestCount:    asInt[0],
		SumFileSize:     asInt[1],
		MeanFileSize:    asInt[2],
		MinFileSize:     asInt[3],
		MaxFileSize:     asInt[4],
	}, nil
}

func timeEpochPointer(t time.Time) *int64 {
	s := int64(t.Unix())
	return &s
}

func getQueryResults(ctx context.Context, svc *cloudwatchlogs.CloudWatchLogs, queryID *string) (*cloudwatchlogs.GetQueryResultsOutput, error) {
	iteration := 0
	var err error

LOOP:
	for {
		queryReq, resp := svc.GetQueryResultsRequest(&cloudwatchlogs.GetQueryResultsInput{QueryId: queryID})

		err = queryReq.Send()
		if err != nil {
			break LOOP
		}

		switch *resp.Status {
		case "Cancelled", "Failed", "Timeout", "Unknown":
			err = fmt.Errorf("log query failed with status %s", *resp.Status)
			break LOOP

		case "Complete":
			return resp, nil

		case "Running", "Scheduled":
			if iteration > 5 {
				err = errors.New("log query timed out")
				break LOOP
			}
			time.Sleep((100 * time.Millisecond) + (time.Duration(iteration) * time.Second))
			iteration++
			continue LOOP

		default:
			err = errors.New("log query failed with unknown status")
			break LOOP
		}
	}

	return nil, err
}
