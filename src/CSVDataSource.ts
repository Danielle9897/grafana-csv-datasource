import _ from 'lodash';

import { DataQueryRequest, DataQueryResponse, DataSourceApi, DataSourceInstanceSettings, DataQuery, DataSourceJsonData } from '@grafana/ui';
import { getBackendSrv } from '@grafana/runtime';

export interface CSVQuery extends DataQuery {
  fields?: string;
}

export interface CSVOptions extends DataSourceJsonData {
  path?: string;
}

interface Request {
  queries: any[];
  from?: string;
  to?: string;
}

export class CSVDataSource extends DataSourceApi<CSVQuery, CSVOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<CSVOptions>) {
    super(instanceSettings);
  }

  query(options: DataQueryRequest<CSVQuery>): Promise<DataQueryResponse> {
    const requestData: Request = {
      queries: options.targets.map((target: any) => {
        return {
          datasourceId: this.id,
          refId: target.refId,
          fields: target.fields,
        };
      }),
    };

    if (options.range) {
      requestData.from = options.range.from.valueOf().toString();
      requestData.to = options.range.to.valueOf().toString();
    }

    return getBackendSrv()
      .post('/api/tsdb/query', requestData)
      .then((response: any) => {
        const res: any = [];

        _.forEach(response.results, r => {
          _.forEach(r.series, s => {
            res.push({ target: s.name, datapoints: s.points });
          });
        });

        response.data = res;

        return response;
      });
  }

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

    return getBackendSrv()
      .post('/api/tsdb/query', requestData)
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
}
