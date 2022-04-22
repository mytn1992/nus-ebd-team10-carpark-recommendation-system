package esw

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"github.com/opensearch-project/opensearch-go/opensearchutil"
	log "github.com/sirupsen/logrus"
)

type BulkRequest struct {
	wrapper *Wrapper
	BodyStr string
}

type BulkRequestBody struct {
	Index string `json:"_index,omitempty"`
	Id    string `json:"_id,omitempty"`
}

func (r *BulkRequest) addRequest(action string, b BulkRequestBody, doc interface{}) error {
	jDoc, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	lines := []string{}
	req := map[string]BulkRequestBody{}
	req[action] = b
	bb, err := json.Marshal(req)
	if err != nil {
		return err
	}
	lines = append(lines, string(bb))
	lines = append(lines, string(jDoc))
	r.BodyStr += strings.Join(lines, "\n") + "\n"
	return nil
}

func (r *BulkRequest) AddIndexRequest(index string, id string, doc interface{}) error {
	b := BulkRequestBody{
		Index: index,
		Id:    id,
	}
	return r.addRequest("index", b, doc)
}

func (r *BulkRequest) Do() (*opensearchutil.BulkIndexerResponse, error) {
	bulkReq := opensearchapi.BulkRequest{
		Body: strings.NewReader(r.BodyStr),
	}
	resp, err := bulkReq.Do(context.Background(), r.wrapper.OpensearchClient)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bodyBytes, err := readApiResponse(resp.Body)
	if err != nil {
		return nil, err
	}

	res := opensearchutil.BulkIndexerResponse{}
	json.Unmarshal(bodyBytes, &res)
	if res.Took == 0 && !res.HasErrors {
		log.Info(string(bodyBytes))
	}
	return &res, nil
}

func newBulkRequest(w *Wrapper) *BulkRequest {
	return &BulkRequest{wrapper: w}
}
