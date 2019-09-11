package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	gf "github.com/marcusolsson/grafana-csv-datasource/pkg/grafana"
)

type CSVDatasource struct {
	logger *log.Logger
}

func (d *CSVDatasource) ID() string {
	return "marcusolsson-csv-datasource"
}

type JsonOptions struct {
	Path string `json:"path"`
}

func (d *CSVDatasource) Query(ctx context.Context, tr gf.TimeRange, ds gf.Datasource, queries []gf.Query) ([]gf.QueryResult, error) {
	var opts JsonOptions
	if err := json.Unmarshal(ds.JsonData, &opts); err != nil {
		return nil, err
	}

	f, err := os.Open(opts.Path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	pts := []gf.Point{}
	for _, rec := range records {
		t, _ := time.Parse(time.RFC3339, rec[0])

		f, _ := strconv.ParseFloat(rec[1], 64)

		pts = append(pts, gf.Point{
			Timestamp: t,
			Value:     f,
		})
	}

	res := []gf.QueryResult{}

	for _, q := range queries {
		res = append(res, gf.QueryResult{
			RefID: q.RefID,
			DataFrames: []gf.DataFrame{
				{
					Name:   "Values",
					Points: pts,
				},
			},
		})
	}

	d.logger.Printf("%+v", res)

	return res, nil
}

func main() {
	logger := log.New(os.Stderr, "", 0)

	g := gf.New()

	g.Register(&CSVDatasource{
		logger: logger,
	})

	if err := g.Run(); err != nil {
		logger.Fatal(err)
	}
}
