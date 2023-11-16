package data_test

import (
	"context"
	"encoding/json"
	"reflect"
	"strconv"
	"testing"
	"time"

	"permission/conf"
	"permission/helpers"

	"permission/pkg/golib/v2/zlog"
)

func TestEs_Info(t *testing.T) {
	// Ping the ElasticSearch server to get e.g. the version number
	addr := "http://10.116.252.14:9200"
	if a, exist := conf.RConf.Elastic["demo"]; exist {
		addr = a.Addr
	}

	info, code, err := helpers.ElasticClient.Ping(addr).Do(ctx)
	if err != nil {
		t.Error("[TestEs_Info] error: ", err.Error())
		return
	}

	zlog.Debug(ctx, "ElasticSearch returned with code:", code, " and version:", info.Version.Number)

	// Getting the ES version number is quite common, so there's a shortcut
	esVersion, err := helpers.ElasticClient.ElasticsearchVersion(addr)
	if err != nil {
		t.Error("[TestEs_Info] error: ", err.Error())
		return
	}

	zlog.Infof(ctx, "ElasticSearch version: %s", esVersion)
}

// Tweet is a structure used for serializing/deserialize data in ElasticSearch.
type Tweet struct {
	User     string                `json:"user"`
	Message  string                `json:"message"`
	Retweets int                   `json:"retweets"`
	Image    string                `json:"image,omitempty"`
	Created  time.Time             `json:"created,omitempty"`
	Tags     []string              `json:"tags,omitempty"`
	Location string                `json:"location,omitempty"`
	Suggest  *elastic.SuggestField `json:"suggest_field,omitempty"`
}

func TestEs_Insert(t *testing.T) {
	// Index a tweet (using JSON serialization)
	tweet1 := Tweet{User: "olivere", Message: "Take Five", Retweets: 0}
	put1, err := helpers.ElasticClient.Index().
		Index("twitter").
		Id("1").
		BodyJson(tweet1).
		Do(ctx)
	if err != nil {
		t.Error("[TestEs_Insert] error: ", err.Error())
		return
	}

	zlog.Infof(ctx, "Indexed tweet %s to index %s, type %s", put1.Id, put1.Index, put1.Type)

	// Index a second tweet (by string)
	tweet2 := `{"user" : "olivere", "message" : "It's a Raggy Waltz"}`
	put2, err := helpers.ElasticClient.Index().
		Index("twitter").
		Type("tweet").
		Id("2").
		BodyString(tweet2).
		Do(ctx)
	if err != nil {
		t.Error("[TestEs_Insert] error: ", err.Error())
		return
	}
	zlog.Infof(ctx, "Indexed tweet %s to index %s, type %s", put2.Id, put2.Index, put2.Type)
}

func TestEs_Bulk(t *testing.T) {
	var doCtx context.Context
	if ctx != nil {
		doCtx = ctx
	} else {
		doCtx = context.Background()
	}

	bulkRequest := helpers.ElasticClient.Bulk()
	for i := 0; i < 100; i++ {
		tweet := Tweet{User: "olivere", Message: "this is message: " + strconv.Itoa(i)}
		req := elastic.NewBulkIndexRequest().Index("twitter").Type("tweet").Id("_id").Doc(tweet)
		// req := elastic.NewBulkUpdateRequest().Index("twitter").Type("tweet").Id("_id").Doc(tweet).DocAsUpsert(true)
		bulkRequest = bulkRequest.Add(req)
	}
	bulkResponse, err := bulkRequest.Do(doCtx)
	if err != nil {
		t.Error("[TestEs_Bulk] error: ", err.Error())
		return
	}

	if bulkResponse == nil || bulkResponse.Errors {
		t.Error("bulkResponse got errors")
		return
	}
	for _, vs := range bulkResponse.Items {
		for k, v := range vs {
			zlog.Infof(ctx, "k = %s and v = %v", k, v)
		}
	}
}

func TestEs_Query(t *testing.T) {
	// Get tweet with specified ID
	get1, err := helpers.ElasticClient.Get().
		Index("twitter").
		// Type("tweet").
		Id("1").
		Do(ctx)
	if err != nil {
		t.Error("[TestEs_Query]  error: ", err.Error())
		return
	}
	if get1.Found {
		zlog.Infof(ctx, "Got document %s in version %d from index %s, type %s", get1.Id, get1.Version, get1.Index, get1.Type)
	}

	termQuery := elastic.NewTermQuery("user", "olivere")
	searchResult, err := helpers.ElasticClient.Search().
		Index("twitter"). // search in index "twitter"
		Query(termQuery). // specify the query
		// Sort("user", true). // sort by "user" field, ascending
		From(0).Size(10). // take documents 0-9
		Pretty(true).     // pretty print request and response JSON
		Do(ctx)           // execute
	if err != nil {
		t.Error("[TestEs_Query]  error: ", err.Error())
		return
	}

	// searchResult is of type SearchResult and returns hits, suggestions,
	// and all kinds of other information from ElasticSearch.
	zlog.Infof(ctx, "Query took %d milliseconds", searchResult.TookInMillis)

	// Each is a convenience function that iterates over hits in a search result.
	// It makes sure you don't need to check for nil values in the response.
	// However, it ignores errors in serialization. If you want full control
	// over iterating the hits, see below.
	var ttyp Tweet
	for _, item := range searchResult.Each(reflect.TypeOf(ttyp)) {
		if t, ok := item.(Tweet); ok {
			zlog.Infof(ctx, "Tweet by %s: %s", t.User, t.Message)
		}
	}
	// TotalHits is another convenience function that works even when something goes wrong.
	zlog.Infof(ctx, "Found a total of %d tweets", searchResult.TotalHits())

	// Here's how you iterate through results with full control over each step.
	if searchResult.Hits.TotalHits.Value > 0 {
		zlog.Infof(ctx, "Found a total of %d tweets", searchResult.Hits.TotalHits)

		// Iterate through results
		for _, hit := range searchResult.Hits.Hits {
			// hit.Index contains the name of the index
			// Deserialize hit.Source into a Tweet (could also be just a map[string]interface{}).
			var t Tweet
			err := json.Unmarshal(hit.Source, &t)
			if err != nil {
				// Deserialization failed
				zlog.Error(ctx, "Tweet by %s got error: %s", t.User, err.Error())
				continue
			}

			// Work with tweet
			zlog.Infof(ctx, "Tweet by %s: %s", t.User, t.Message)
		}
	} else {
		// No hits
		zlog.Info(ctx, "Found no tweets")
	}
}

func TestEs_Update(t *testing.T) {
	// Update a tweet by the update API of ElasticSearch.
	// We just increment the number of retweets.
	update, err := helpers.ElasticClient.Update().Index("twitter").Id("1").
		Script(elastic.NewScriptInline("ctx._source.retweets += params.num").Lang("painless").Param("num", 1)).
		Upsert(map[string]interface{}{"retweets": 0}).
		Do(ctx)
	if err != nil {
		t.Error("[TestEs_Update]  error: ", err.Error())
		return
	}
	zlog.Infof(ctx, "New version of tweet %q is now %d", update.Id, update.Version)
}

func TestEs_Delete(t *testing.T) {
	// Delete an index.
	deleteIndex, err := helpers.ElasticClient.DeleteIndex("twitter").Do(ctx)
	if err != nil {
		t.Error("[TestEs_Delete]  error: ", err.Error())
		return
	}

	if !deleteIndex.Acknowledged {
		zlog.Warn(ctx, "Not acknowledged")
	}
}
