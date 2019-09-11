import React, { PureComponent } from 'react';

import { CSVDataSource } from './CSVDataSource';
import { CSVQuery, CSVOptions } from './types';

import { FormLabel, Select, QueryEditorProps } from '@grafana/ui';

type Props = QueryEditorProps<CSVDataSource, CSVQuery, CSVOptions>;

const options = [{ value: 'timestamp', label: 'timestamp' }, { value: 'value', label: 'value' }];

interface State {}

export class CSVQueryEditor extends PureComponent<Props, State> {
  state = {
    text: '',
  };

  onComponentDidMount() {}

  render() {
    const selected = options[0];

    return (
      <div>
        <div className="gf-form">
          <FormLabel width={4}>Fields</FormLabel>
          <Select width={12} options={options} value={selected} />
        </div>
      </div>
    );
  }
}
