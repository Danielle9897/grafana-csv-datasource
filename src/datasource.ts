import _ from 'lodash';

interface Request {
  queries: any[];
  from?: string;
  to?: string;
}

export class GenericDatasource {
  id: string;
  path: string;

  /** @ngInject */
  constructor(instanceSettings: any, $q: any, private backendSrv: any, private templateSrv: any) {
    this.path = instanceSettings.path;
    this.id = instanceSettings.id;
  }

  query(options: any) {
    console.log(this.path);
    const targets = _.map(options.targets, (target: any) => {
      return {
        queryType: 'query',
        target: this.templateSrv.replace(target.target, options.scopedVars, 'regex'),
        refId: target.refId,
        hide: target.hide,
        type: target.type || 'timeserie',
        datasourceId: this.id,
      };
    });

    const requestData: Request = {
      queries: targets,
    };

    if (options.range) {
      requestData.from = options.range.from.valueOf().toString();
      requestData.to = options.range.to.valueOf().toString();
    }

    return this.backendSrv
      .datasourceRequest({
        url: '/api/tsdb/query',
        method: 'POST',
        data: requestData,
      })
      .then((response: any) => {
        const res: any = [];
        _.forEach(response.data.results, r => {
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
    return this.backendSrv
      .datasourceRequest({
        url: '/api/tsdb/query',
        method: 'POST',
        data: {
          from: '5m',
          to: 'now',
          queries: [
            {
              datasourceId: this.id,
            },
          ],
        },
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

  annotationQuery(options: any) {}

  metricFindQuery(query: any) {}
}
