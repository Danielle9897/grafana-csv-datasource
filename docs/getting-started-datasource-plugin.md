# Writing your first data source plugin for Grafana

Grafana has support for a wide range of data sources, like Prometheus, MySQL, or even Datadog. This means that it’s very likely that you can already visualize metrics from the systems you've already set up. In some cases though, you might already have an in-house metrics solution that you’d like to add to your Grafana dashboards. Luckily, Grafana supports _data source_ plugins, which lets you build a custom integration for your specific source of data.

In this guide, you’ll learn how to build a data source plugin for visualizing CSV files. At the end of this guide, you'll have a working example of a data source plugin, and understand how you can extend it to support the specific use-case you have in mind.

## Preparations

First of all, we'll create a project directory for our plugin.

### Requirements

- NodeJS
- yarn

### Directory structure

Create a new directory and initialize your module. Also, go ahead and create a `src` directory within your module, to hold our source code. 

```
mkdir grafana-csv-datasource
cd grafana-csv-datasource
yarn init
mkdir src
```

## grafana-toolkit

Tooling for modern web development can be tricky to wrap your head around. While you certainly could write you own webpack configuration, for this guide, I'm going to use _grafana-toolkit_. 

grafana-toolkit is a CLI application that aims to simplify Grafana plugin development. We'll focus on writing the code, and the toolkit will build and test it for us.

First, let’s install grafana-toolkit in our active project:

```
yarn add @grafana/toolkit --dev
```

Next, add @grafana/ui which holds common UI components for Grafana. We're going to be using some of them for our own components.

```
yarn add @grafana/ui --dev
```

The toolkit requires every plugin to provide a README as well and a LICENSE. If you haven't yet decided which license you want to release your plugin under, for now, just make sure the files are present.

```
touch README.md LICENSE
```

## `plugin.json` and `module.ts`

Every plugin you create will require at least two files, `plugin.json`, and `module.ts`.

`plugin.json` contains information about your plugin, and tells Grafana about what capabilities it needs. We'll create a basic `plugin.json` that you can update with your own information.

```js
// src/plugin.json
{
  "id": "<your-github-handle>-csv-datasource",
  "name": "CSV file",
  "type": "datasource",
  "metrics": true,

  "info": {
    "description": "A datasource for loading CSV files",
    "author": {
      "name": "...",
      "url": "..."
    },
    "keywords": [],
    "version": "1.0.0",
    "updated": "2019-09-10"
  },

  "dependencies": {
    "grafanaVersion": "3.x.x",
    "plugins": []
  }
}
```

`module.ts` is the entry point for your plugin, and where you should export the implementation for your plugin. Your `module.ts` will look differently depending on the type of plugin you're building. 

For data source plugins, you'll want your `module.ts` to export an implementation of the `DataSourcePlugin` from the `@grafana/ui` module.

```ts
// src/module.ts
import { DataSourcePlugin } from '@grafana/ui';

import { CSVDataSource, CSVOptions, CSVQuery } from './CSVDataSource';

export const plugin = new DataSourcePlugin<CSVDataSource, CSVQuery, CSVOptions>(CSVDataSource);
```

### Defining our data source

Let's create the actual implementation for the data source. At this point, it won't be doing much, but we'll make sure to change that in a bit, so bear with me for now.

```ts
// src/CSVDataSource.ts
import { DataQueryRequest, DataQueryResponse, DataSourceApi, DataSourceInstanceSettings, DataQuery, DataSourceJsonData } from '@grafana/ui';

export interface CSVQuery extends DataQuery {
}

export interface CSVOptions extends DataSourceJsonData {
}

export class CSVDataSource extends DataSourceApi<CSVQuery, CSVOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<CSVOptions>) {
    super(instanceSettings);
  }

  query(options: DataQueryRequest<CSVQuery>): Promise<DataQueryResponse> {
    return Promise.resolve({ data: [] });
  }

  testDatasource() {
    return new Promise((resolve, reject) => {
      resolve({
        status: 'success',
        message: 'Yes',
      });
    });
  }
}
```

As you may have noticed in our  `module.ts`, we need to provide our data source implementation, as well as definitions for _queries_ and _options_. We'll update these along the way, but for now, let's settle with just the definitions.

> **Options and queries:** Options contains information about how you connect to your data source. It typically contains information like the hostname, port, or API keys. A query on the other hand is used when asking your data source for the data displayed by a specific panel.

By now, you should be able to run `grafana-toolkit plugin:build` successfully. Afterwards, you should have a directory called `dist` that contains the production assets for your plugin. The toolkit also generated a `.prettierrc.js` and a `tsconfig.json`, that will help you follow some of the conventions used when developing for Grafana.

## Trying out our new plugin

If you're developing on a Linux system, consider creating a symlink from the plugin directory of your current Grafana installation.

```
ln -s $(pwd) /var/lib/grafana/plugins/
```

By doing this you can test new changes by restarting Grafana.

Let's see if our plugin gets picked up by Grafana! Open Grafana in your browser, navigate to Configuration -> Plugins, and type "csv" to find your plugin. The details view gives your users instructions on how to use the plugin.

Next, navigate to Configuration -> Data Sources, type "csv", and select your data source. For now, this view is going not going to show much, but we should still be able to click the "Save & Test" button. Hopefully, you see the message we configured in `testDatasource`.

## Development workflow

As an exercise, try changing the message returned by  `testDatasource` , rebuild the assets with grafana-toolkit, and restart Grafana. Click the "Save & Test" button again to verify your change was effective.

## Adding data source settings

For most data sources, you'll want to give your users the ability to configure things like hostname or authentication method. Although our CSV example doesn't require authentication at this point, we might want to set the path to the CSV file. We can accomplish this by adding an _options editor_.

```tsx
// src/CSVConfigEditor.tsx
import React, { PureComponent, ChangeEvent } from 'react';

import { DataSourcePluginOptionsEditorProps, DataSourceSettings, FormField } from '@grafana/ui';

import { CSVOptions } from './CSVDataSource';

type CSVSettings = DataSourceSettings<CSVOptions>;

interface Props extends DataSourcePluginOptionsEditorProps<CSVSettings> {}

interface State {}

export class CSVConfigEditor extends PureComponent<Props, State> {
  componentDidMount() {}

  onPathChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      path: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  render() {
    const { options } = this.props;
    const { jsonData } = options;

    return (
      <div className="gf-form-group">
        <div className="gf-form">
          <FormField label="Path" value={jsonData.path || ''} onChange={this.onPathChange} />
        </div>
      </div>
    );
  }
}
```

If you've been developing using ReactJS before, you'll likely have seen something similar. Grafana editors are indeed ReactJS components. If you haven't, for now, focus on the `render` method. It returns the template for the data source settings view. You'll likely be changing this to expose the options required by your own data source.

> _Note:_ You might have noticed that we added the path to an object called `jsonData`. This object is automatically persisted for you, and will be made available to your data source implementation as well.

We also need to add the config editor to our data source by updating `module.ts`:

```ts
// src/module.ts
import { DataSourcePlugin } from '@grafana/ui';

import { CSVDataSource } from './CSVDataSource';
import { CSVConfigEditor } from './CSVConfigEditor';
import { CSVOptions, CSVQuery } from './types';

export const plugin = new DataSourcePlugin<CSVDataSource, CSVQuery, CSVOptions>(CSVDataSource)
  .setConfigEditor(CSVConfigEditor);
```

Build your assets, restart Grafana, and check out the configuration for our data source. There should now be a text field where you can configure the path.

## Querying your data source

Most likely you want your users to be able to select the data they're interested in. For MySQL and PostgreSQL this would be SQL queries, while Prometheus has its own query language, called PromQL. Let's add query support for our plugin, using a custom _query editor_.

There's a lot we can do when it comes to querying CSV data, like filtering rows based on string values in one of our fields. Let's keep it simple for now though, by letting the user supply a comma separated list of fields to visualize.

Create a new ReactJS component called `CSVQueryEditor`, and have it return a `FormField` from the `@grafana/ui` package.

```tsx
// src/CSVQueryEditor.tsx
import React, { PureComponent, ChangeEvent } from 'react';

import { FormField, QueryEditorProps } from '@grafana/ui';

import { CSVDataSource, CSVQuery, CSVOptions } from './CSVDataSource';

type Props = QueryEditorProps<CSVDataSource, CSVQuery, CSVOptions>;

interface State {}

export class CSVQueryEditor extends PureComponent<Props, State> {
  onComponentDidMount() {}

  onFieldsChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, fields: event.target.value });
  };

  render() {
    const { query } = this.props;
    const { fields } = query;

    return (
      <div className="gf-form">
        <FormField label="Fields" value={fields || ''} onChange={this.onFieldsChange} />
      </div>
    );
  }
}
```

Finally, let's configure our data source to use the query editor:

```ts
// src/module.ts
import { DataSourcePlugin } from '@grafana/ui';

import { CSVDataSource } from './CSVDataSource';
import { CSVConfigEditor } from './CSVConfigEditor';
import { CSVQueryEditor } from './CSVQueryEditor';
import { CSVOptions, CSVQuery } from './types';

export const plugin = new DataSourcePlugin<CSVDataSource, CSVQuery, CSVOptions>(CSVDataSource)
  .setConfigEditor(CSVConfigEditor)
  .setQueryEditor(CSVQueryEditor);
```

When configuring your panel, you should have something like this:

![Query editor](./query-editor.png)

## Next up

If your data source is available from JavaScript in your browser, you should already have the tools you need. Construct the requests in the `query` method, parse the response and return it to Grafana.

While many data sources can be implemented solely as frontend plugins, some require a backend to help with persistance, caching, and authentication.

In the second part of this guide, we'll look at adding a backend plugin that will read and parse the CSV file, and return it as a Grafana _data frame_. Stay tuned!
