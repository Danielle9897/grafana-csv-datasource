# Adding a backend to a Grafana Data Source

In the previous part of this guide, we looked at how to get started with writing data source plugins for Grafana. For many data sources, integrating a custom data source can be done completely in the Grafana browser client. For others, you might want the plugin to be able to continue running even after closing your browser window, such as alerting, or authentication.

Luckily, Grafana has support for _backend plugins_, which lets your data source plugin communicate with a process running on the server.

Last time, we started writing a data source plugin that would read CSV files. Let's see how a backend plugin lets you read a file on the server and return the data back to the client.

## Requirements

- Go

## Creating the backend plugin

Create another directory `backend` in your project root directory, containing a `main.go` file. The Grafana backend plugin system uses the [go-plugin](https://github.com/hashicorp/go-plugin) library from [Hashicorp](https://www.hashicorp.com/), and while communication happens over gRPC, here we'll use a Go client library that wraps the boilerplate for you.

```go
// backend/main.go
package main

import (
	"context"
	"log"
	"os"

	gf "github.com/github.com/marcusolsson/grafana-csv-datasource/pkg/grafana"
)

type CSVDatasource struct {
	logger *log.Logger
}

func (d *CSVDatasource) ID() string {
	return "<your-github-handle>-csv-datasource"
}

func (d *CSVDatasource) Query(ctx context.Context, tr gf.TimeRange, ds gf.Datasource, queries []gf.Query) ([]gf.QueryResult, error) {
	return []gf.QueryResult{}, nil
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
```

Let's leave the implementation of the `Query` method for now to see how to make Grafana discover our backend plugin.

### Building a binary for our backend plugin

In order for Grafana to discover our plugin, we have to build a binary with the following suffixes, depending on the machine you're using to run Grafana:

```
_linux_amd64
_darwin_amd64
_windows_amd64.exe
```

I'm running Grafana on my MacBook, so I'll go ahead and build a Darwin binary:

```
go build -o ./dist/csv-datasource_darwin_amd64 ./backend
```

The binary needs to be bundled into `./dist` directory together with the frontend assets.

Next, we'll have to let Grafana know it should look for a backend plugin by updating the `plugin.json`.

```json
// src/plugin.json
{
  "backend": true,
  "executable": "csv-datasource",
}
```

Where the executable is the name of the binary, with the suffix removed.

Restart Grafana and verify that your plugin is running:

```
$ ps aux | grep csv-datasource
```

## Calling the backend from the client

Let's make the `testDatasource` call our backend to make sure it's responding correctly.

```ts
// src/CSVDataSource.ts

// Is this necessary? Maybe should be part of @grafana/ui?
interface Request {
  queries: any[];
  from?: string;
  to?: string;
}

...

testDatasource() {
  const requestData: Request = {
    from: '5m',
    to: 'now',
    queries: [
      {
        datasourceId: this.id,
      },
    ],
  };

  // How to avoid hard-coding the URL?
  const url = 'http://localhost:3000/api/tsdb/query'

  return fetch(url, {
    method: 'post',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(requestData),
  })
    .then((response: any) => {
      if (response.status === 200) {
        return { status: 'success', message: 'Data source is working', title: 'Success' };
      } else {
        return { status: 'failed', message: 'Data source is not working', title: 'Error' };
      }
    })
    .catch((error: any) => {
      return { status: 'failed', message: 'Data source is not working', title: 'Error' };
    });
}
```

Confirm that the client is able to call our backend plugin by hitting **Save & Test** on your data source. It should give you a green message saying _Data source is working_.

Now that all the plumbing is done, let's start implementing the query method for both the frontend, and the backend plugins.

```
query(options: DataQueryRequest<CSVQuery>): Promise<DataQueryResponse> {

  // Is this needed? Could this be simplified?
  const queries = options.targets.map((target: any) => {
    return {
      queryType: 'query',
      target: target.target,
      refId: target.refId,
      hide: target.hide,
      type: target.type,
      datasourceId: this.id,
    };
  });

  const requestData: Request = {
    queries: queries,
  };

  if (options.range) {
    requestData.from = options.range.from.valueOf().toString();
    requestData.to = options.range.to.valueOf().toString();
  }

  return fetch(url, {
    method: 'post',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(requestData),
  })
    .then((response: any) => response.json())
    .then((response: any) => {
      const res: any = [];

      // This will be look better once backend starts returning data frames.
      _.forEach(response.results, r => {
        _.forEach(r.series, s => {
          res.push({ target: s.name, datapoints: s.points });
        });
        _.forEach(r.tables, t => {
          t.type = 'table';
          t.refId = r.refId;
          res.push(t);
        });
      });

      response.data = res;

      return response;
    });
}
```

That's it for the frontend! Next, let's have a look at implementing the backend.

```
type CSVQuery struct {
	RefID  string `json:"refId"`
	Fields string `json:"fields"`
}

type CSVOptions struct {
	Path string `json:"path"`
}

type CSVDatasource struct {
	logger *log.Logger
}

func (d *CSVDatasource) ID() string {
	return "marcusolsson-csv-datasource"
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
```

_Note:_ I've left the implementation for `parseCSV` out of this article for brevity purposes, but feel free to check it out in full on [Github](https://github.com/marcusolsson/grafana-csv-datasource/blob/master/pkg/main.go).

## Next steps
That's it! If you've made it this far, you should have a fully fledged data source plugin for Grafana, complete with backend support. You should have a pretty good feeling of where to change the implementation to support your own data source, but here are a few pointers:

- Update `CSVConfigEditor` to expose the configuration options for your specific data source.
- Update `CSVQueryEditor` to support the query model used to fetch data from your data source. 
- Construct the query inside `CSVDataSource.query` in the frontend.
- Implement the `CSVDatasource.Query` in the backend to make the outgoing requests to your data source.