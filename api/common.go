package api

import (
	"dnscheck/dbs"
	"fmt"
)

type GetConfigRequest struct {
	ClientId string
}

type SetResultsRequest dbs.ResultInfo

func withPrefix(url string) string {
	s := fmt.Sprintf("/api/v1%s", url)
	return s
}
