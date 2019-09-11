import { DataSourcePlugin } from '@grafana/ui';

import { CSVDataSource } from './CSVDataSource';
import { CSVConfigEditor } from './CSVConfigEditor';
import { CSVQueryEditor } from './CSVQueryEditor';
import { CSVOptions, CSVQuery } from './types';

export const plugin = new DataSourcePlugin<CSVDataSource, CSVQuery, CSVOptions>(CSVDataSource)
  .setConfigEditor(CSVConfigEditor)
  .setQueryEditor(CSVQueryEditor);
