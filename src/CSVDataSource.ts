import _ from 'lodash';

import { DataQueryRequest, DataQueryResponse, DataSourceApi, DataSourceInstanceSettings } from '@grafana/ui';
import { CSVQuery, CSVOptions } from './types';

interface Request {
  queries: any[];
  from?: string;
  to?: string;
}

const url = 'http://localhost:3000/api/tsdb/query';

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

    return fetch(url, {
      method: 'post',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(requestData),
    })
      .then((response: any) => response.json())
      .then((response: any) => {
        console.log(response);
        const res: any = [];
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
        console.log(res);
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
}
