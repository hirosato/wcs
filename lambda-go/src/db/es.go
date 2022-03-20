package db

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/hirosato/wcs/env"
	"github.com/hirosato/wcs/model"
)

var es *elasticsearch.Client

func init() {
	config := elasticsearch.Config{
		Addresses: []string{env.GetEsUrl()},
	}
	// if !env.IsLocal {
	config.Transport = NewAmazonESTransport()
	// }

	es, _ = elasticsearch.NewClient(config)
}

func PutEsPainting(painting *model.Painting) error {
	log.Printf("EVENT: put ES start")
	paintingByte, err := json.Marshal(painting)
	if err != nil {
		return err
	}
	paintingReader := bytes.NewReader(paintingByte)
	req := esapi.IndexRequest{
		Body:       paintingReader,
		Index:      "wcs",
		DocumentID: painting.GetId(),
		Pretty:     true,
	}
	res, err := req.Do(context.Background(), es)
	log.Printf("EVENT: put ES req.Do done")
	if err != nil {
		log.Fatal(err)
		return err
	}
	row := make([]byte, 1000)
	res.Body.Read(row)
	fmt.Println(string(row))
	//{"Message":"User: anonymous is not authorized to perform: es:ESHttpPut"}
	//https://blog.linkode.co.jp/entry/2020/04/22/093502
	if res.IsError() {
		log.Fatal("Something went wrong %i", res.Status())
		return errors.New("something went wrong:" + res.Status())
	}
	defer res.Body.Close()
	log.Printf("EVENT: access to ES end")
	return nil
}

type searchResult struct {
	Hits resultHits `json:"hits"`
}

// ResultHits represents the result of the search hits
type resultHits struct {
	Hits []struct {
		Source model.Painting `json:"_source"`
	} `json:"hits"`
}

func ListWaterColorSite(offset int) []model.Painting {
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	}
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}
	log.Printf("EVENT: access to ES start")
	res, err := es.Search(
		es.Search.WithContext(context.TODO()),
		es.Search.WithIndex("wcs"),
		es.Search.WithBody(&buf),
		es.Search.WithPretty(),
		es.Search.WithFrom(offset),
		es.Search.WithSize(10),
		es.Search.WithSort("timestamp.keyword:desc"),
	)
	log.Printf("EVENT: access to ES end")
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}

	var esres searchResult
	err = json.NewDecoder(res.Body).Decode(&esres)
	if err != nil {
		log.Fatalf("Error decoding result: %s", err)
	}
	result := []model.Painting{}
	for i := 0; i < len(esres.Hits.Hits); i++ {
		result = append(result, esres.Hits.Hits[i].Source)
	}
	defer res.Body.Close()
	return result
}

// get
/*
GET /painting/_search
{
  "from": 0,
  "size": 10,
  "query": {
    "match_all": {}
  },
  "sort": [
    {
      "Timestamp.keyword": {
        "order": "desc"
      }
    }
  ]
}
*/
