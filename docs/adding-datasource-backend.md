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
  // ...

  "backend": true,
  "executable": "csv-datasource",

  // ...
}
```

Where the executable is the name of the binary, with the suffix removed.

Restart Grafana and verify that your plugin is running:

```
$ ps aux | grep cdv-datasource
```