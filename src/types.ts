import { DataQuery, DataSourceJsonData } from '@grafana/ui';

export interface CSVQuery extends DataQuery {
  fields?: string;
}

export interface CSVOptions extends DataSourceJsonData {
  path?: string;
}
