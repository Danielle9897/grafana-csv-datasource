# CSV Data Source plugin for Grafana

An example of a Data Source plugin for [Grafana](https://www.grafana.com).

__Note:__ This repository exists for experimental purposes only, and is currently not fit for use.

## Directory structure

`cmd/backend` holds the source code for the backend plugin (Go).

`src` holds the source code for the frontend plugin (Typescript).

## Developing

Build frontend plugin:

```
yarn build
```

Build backend plugin:

```
make build-darwin
```
