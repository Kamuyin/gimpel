import { defineConfig } from '@umijs/max';
import { join } from 'path';

export default defineConfig({
  antd: {},
  access: {},
  model: {},
  initialState: {},
  request: {},
  layout: {
    title: 'Gimpel',
    locale: false,
  },
  routes: [
    {
      path: '/',
      component: './index',
    },
  ],
  npmClient: 'pnpm',
  plugins: [require.resolve('@umijs/max-plugin-openapi')],
  openAPI: [
    {
      requestLibPath: "import { request } from '@umijs/max'",
      schemaPath: join(__dirname, '../openapi.json'),
      projectName: 'gimpel',
      mock: true,
    },
  ],
});
