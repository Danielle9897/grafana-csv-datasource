package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	gf "github.com/marcusolsson/grafana-csv-datasource/pkg/grafana"
)

type CSVDatasource struct {
	logger *log.Logger
}

func (d *CSVDatasource) ID() string {
	return "marcusolsson-csv-datasource"
}

type CSVQuery struct {
	RefID  string `json:"refId"`
	Fields string `json:"fields"`
}

type CSVOptions struct {
	Path string `json:"path"`
}

func (d *CSVDatasource) Query(ctx context.Context, tr gf.TimeRange, ds gf.Datasource, queries []gf.Query) ([]gf.QueryResult, error) {
	var opts CSVOptions
	if err := json.Unmarshal(ds.JsonData, &opts); err != nil {
		return nil, err
	}

	var res []gf.QueryResult

	for _, q := range queries {
		var query CSVQuery
		if err := json.Unmarshal(q.ModelJson, &query); err != nil {
			return nil, err
		}

		fields := strings.Split(query.Fields, ",")

		frame, err := parseCSV(opts.Path, fields)
		if err != nil {
			return nil, err
		}

		res = append(res, gf.QueryResult{
			RefID:      query.RefID,
			DataFrames: []gf.DataFrame{frame},
		})
	}

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

func parseCSV(path string, fields []string) (gf.DataFrame, error) {
	if len(fields) < 2 {
		return gf.DataFrame{}, errors.New("requires at least 2 fields")
	}

	f, err := os.Open(path)
	if err != nil {
		return gf.DataFrame{}, err
	}
	defer f.Close()

	reader := csv.NewReader(f)

	records, err := reader.ReadAll()
	if err != nil {
		return gf.DataFrame{}, err
	}

	columns := make(map[string][]string)

	var header []string
	for _, rec := range records {
		if len(header) == 0 {
			header = rec
		} else {
			for i, val := range rec {
				columns[header[i]] = append(columns[header[i]], val)
			}
		}
	}

	rows := rowsFromCols(columns, fields)

	pts := []gf.Point{}
	for _, row := range rows {
		if len(row) != 2 {
			continue
		}

		t, err := time.Parse(time.RFC3339, row[0])
		if err != nil {
			return gf.DataFrame{}, nil
		}

		f, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			return gf.DataFrame{}, nil
		}

		pts = append(pts, gf.Point{
			Timestamp: t,
			Value:     f,
		})
	}

	return gf.DataFrame{
		Name:   fields[1],
		Points: pts,
	}, nil
}

func rowsFromCols(cols map[string][]string, fields []string) [][]string {
	rows := make([][]string, len(cols[fields[0]]))

	for _, f := range fields {
		for j, v := range cols[f] {
			rows[j] = append(rows[j], v)
		}
	}
	return rows
}
