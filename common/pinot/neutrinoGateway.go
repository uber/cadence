package pinot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	serviceNeutrino  = "neutrino"
	neutrinoURL      = "http://localhost:5436/v1/statement"
	HeaderRPCCaller  = "RPC-Caller"
	HeaderRPCService = "RPC-Service"
)

type Gateway struct {
	httpClient *http.Client
}

// NewGateway New returns a Gateway
func NewGateway() *Gateway {
	return &Gateway{
		httpClient: &http.Client{},
	}
}

// Query sends over the given query to Neutrino and return result
func (g *Gateway) Query(query string) (*queryResponse, error) {
	var queryResp queryResponse

	//req, err := http.NewRequestWithContext(ctx, http.MethodPost, neutrinoURL, strings.NewReader(query))
	// TODO: need to do a research about the difference of these 2 methods. Do we really need ctx?

	req, err := http.NewRequest(http.MethodPost, neutrinoURL, strings.NewReader(query))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request; %v", err)
	}

	req.Header.Set(HeaderRPCService, serviceNeutrino)
	req.Header.Set(HeaderRPCCaller, "ceres")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request; %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get an OK status code instead got %d", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&queryResp)
	if err != nil {
		return nil, fmt.Errorf("failed to read response as JSON; %v", err)
	}

	if queryResp.Error != nil {
		return nil, fmt.Errorf("failed to run query; %s", queryResp.Error.Message)
	}

	return &queryResp, nil
}

type queryResponse struct {
	Columns []struct {
		Name string `json:"name"`
	} `json:"columns"`
	Data  [][]interface{} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	}
}

func (q *queryResponse) rows() []map[string]string {
	rows := make([]map[string]string, len(q.Data))

	for i := range q.Data {
		row := make(map[string]string)
		for j := range q.Columns {
			row[q.Columns[j].Name] = fmt.Sprintf("%v", q.Data[i][j])
		}
		rows[i] = row
	}
	return rows
}
