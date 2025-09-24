import { createDevApp } from '@backstage/dev-utils';
import { aegisPlugin, AegisPage } from '../src/plugin';

createDevApp()
  .registerPlugin(aegisPlugin)
  .addPage({
    element: <AegisPage />,
    title: 'Root Page',
    path: '/aegis',
  })
  .render();
