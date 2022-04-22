package esw

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/olivere/elastic/v7"
	"github.com/opensearch-project/opensearch-go"
	log "github.com/sirupsen/logrus"
)

const (
	ContentTypePlain = "text/plain"
	ContentTypeHTML  = "text/html"
)

type (
	Wrapper struct {
		config           Config
		OpensearchClient *opensearch.Client
		ESClient         *elastic.Client
	}

	Config struct {
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		Host     string `mapstructure:"host"`
	}
)

// exported functions
// use for bulk request
func (w *Wrapper) NewBulkRequest() *BulkRequest {
	return newBulkRequest(w)
}

func readApiResponse(body io.ReadCloser) ([]byte, error) {
	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("error while parsing body - %v", err)
	}
	return bodyBytes, nil
}

func newOpensearchClient(config Config) (*opensearch.Client, error) {
	return opensearch.NewClient(opensearch.Config{
		Addresses: []string{config.Host},
		Username:  config.Username,
		Password:  config.Password,
	})
}

func newESClient(config Config) (*elastic.Client, error) {
	return elastic.NewClient(
		elastic.SetURL(config.Host),
		elastic.SetBasicAuth(config.Username, config.Password),
		elastic.SetGzip(true),
		elastic.SetSniff(false),
		elastic.SetErrorLog(log.StandardLogger()),
		elastic.SetHealthcheck(false),
	)
}

func NewWrapper(config Config) (*Wrapper, error) {
	opensearchClient, err := newOpensearchClient(config)
	if err != nil {
		return nil, err
	}
	esClient, err := newESClient(config)
	if err != nil {
		return nil, err
	}
	return &Wrapper{
		config:           config,
		OpensearchClient: opensearchClient,
		ESClient:         esClient,
	}, nil
}

func QueryDataForAPI(esWrapper Wrapper, index string, query *elastic.BoolQuery, is_detail bool, return_fields []string) ([]map[string]interface{}, error) {
	client := esWrapper.ESClient
	documents := []map[string]interface{}{}

	// Return first match to limit response
	searchResult, err := client.Search().Index(index+"*").Query(query).Size(1000).Sort("UPDATE_DATETIME", true).Do(context.TODO())
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if searchResult.Hits == nil {
		log.Error("expected SearchResult.Hits != nil; got nil")
		return nil, err
	}

	for _, hit := range searchResult.Hits.Hits {
		item := make(map[string]interface{})
		err := json.Unmarshal(hit.Source, &item)
		if err != nil {
			log.Errorf("error fetched infra data %v", err)
			return nil, err
		}

		returndoc := make(map[string]interface{})

		if is_detail {
			documents = append(documents, item)
		} else {
			for _, val := range return_fields {
				returndoc[val] = item[val]
			}
			documents = append(documents, returndoc)
		}
	}

	for _, d := range documents {
		fmt.Printf("fetched infra data %v\n", d)
	}

	return documents, nil
}

func QueryScrollDataForAPI(esWrapper Wrapper, index string, query *elastic.BoolQuery, is_detail bool, return_fields []string, sort_field string) ([]map[string]interface{}, error) {
	client := esWrapper.ESClient
	documents := []map[string]interface{}{}

	scroller := client.Scroll(index+"*").Query(query).Size(1000).Sort(sort_field, true)

	for {
		data, err := scroller.Do(context.TODO())
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		for _, hit := range data.Hits.Hits {
			item := make(map[string]interface{})
			err := json.Unmarshal(hit.Source, &item)
			if err != nil {
				log.Errorf("error fetched infra data %v", err)
				return nil, err
			}

			returndoc := make(map[string]interface{})

			if is_detail {
				documents = append(documents, item)
			} else {
				for _, val := range return_fields {
					returndoc[val] = item[val]
				}
				documents = append(documents, returndoc)
			}
		}
	}
	return documents, nil
}
