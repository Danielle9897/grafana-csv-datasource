package grafana

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/grafana/grafana-plugin-model/go/datasource"
	plugin "github.com/hashicorp/go-plugin"
)

type TimeRange struct {
	From time.Time
	To   time.Time
}

type Datasource struct {
	ID       int64
	OrgID    int64
	Name     string
	Type     string
	URL      string
	JsonData json.RawMessage
}

type Point struct {
	Timestamp time.Time
	Value     float64
}

type DataFrame struct {
	Name   string
	Tags   map[string]string
	Points []Point
}

type Query struct {
	RefID         string
	MaxDataPoints int64
	Interval      time.Duration
	ModelJson     json.RawMessage
}

type QueryResult struct {
	Error      string
	RefID      string
	MetaJson   string
	DataFrames []DataFrame
}

type DatasourcePlugin interface {
	Plugin

	Query(ctx context.Context, tr TimeRange, ds Datasource, queries []Query) ([]QueryResult, error)
}

type datasourcePlugin struct {
	plugin.NetRPCUnsupportedPlugin

	plugin DatasourcePlugin
}

func (p *datasourcePlugin) Query(ctx context.Context, req *datasource.DatasourceRequest) (*datasource.DatasourceResponse, error) {
	fmt.Fprintln(os.Stderr, "%+v", req)

	tr := TimeRange{
		From: time.Unix(0, req.TimeRange.FromEpochMs*int64(time.Millisecond)),
		To:   time.Unix(0, req.TimeRange.FromEpochMs*int64(time.Millisecond)),
	}

	dsi := Datasource{
		ID:       req.Datasource.Id,
		OrgID:    req.Datasource.OrgId,
		Name:     req.Datasource.Name,
		Type:     req.Datasource.Type,
		URL:      req.Datasource.Url,
		JsonData: json.RawMessage(req.Datasource.JsonData),
	}

	var queries []Query
	for _, q := range req.Queries {
		queries = append(queries, Query{
			RefID:         q.RefId,
			MaxDataPoints: q.MaxDataPoints,
			Interval:      time.Duration(q.IntervalMs) * time.Millisecond,
			ModelJson:     []byte(q.ModelJson),
		})
	}

	fmt.Fprintf(os.Stderr, "%+v", queries)

	results, err := p.plugin.Query(ctx, tr, dsi, queries)
	if err != nil {
		return nil, err
	}

	var respResults []*datasource.QueryResult

	for _, res := range results {
		tss := []*datasource.TimeSeries{}

		for _, df := range res.DataFrames {
			pts := []*datasource.Point{}
			for _, p := range df.Points {
				pts = append(pts, &datasource.Point{
					Timestamp: p.Timestamp.UnixNano() / int64(time.Millisecond),
					Value:     p.Value,
				})
			}
			tss = append(tss, &datasource.TimeSeries{
				Name:   df.Name,
				Tags:   df.Tags,
				Points: pts,
			})
		}

		respResults = append(respResults, &datasource.QueryResult{
			Error:    res.Error,
			RefId:    res.RefID,
			MetaJson: res.MetaJson,
			Series:   tss,
		})
	}

	return &datasource.DatasourceResponse{
		Results: respResults,
	}, nil
}
