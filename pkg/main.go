package main

import (
	"context"
	"encoding/csv"
	"log"
	"os"
	"strconv"
	"time"

	gf "github.com/marcusolsson/my-plugin/pkg/grafana"
)

type JsonDatasource struct {
	logger *log.Logger
}

func (d *JsonDatasource) ID() string {
	return "my-backend-datasource"
}

func (d *JsonDatasource) Query(ctx context.Context, tr gf.TimeRange, ds gf.Datasource, queries []gf.Query) ([]gf.QueryResult, error) {
	f, err := os.Open("/tmp/sample.csv")
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

	g.Register(&JsonDatasource{
		logger: logger,
	})

	if err := g.Run(); err != nil {
		logger.Fatal(err)
	}
}
