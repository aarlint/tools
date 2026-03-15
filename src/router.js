import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  { path: '/', redirect: '/md' },
  { path: '/md', component: () => import('./tools/markdown/MarkdownTool.vue'), meta: { name: 'Markdown', icon: 'M↓', group: 'Text' } },
  { path: '/json', component: () => import('./tools/json/JsonTool.vue'), meta: { name: 'JSON', icon: '{}', group: 'Text' } },
  { path: '/diff', component: () => import('./tools/diff/DiffTool.vue'), meta: { name: 'Diff', icon: '±', group: 'Text' } },
  { path: '/regex', component: () => import('./tools/regex/RegexTool.vue'), meta: { name: 'Regex', icon: '.*', group: 'Text' } },
  { path: '/csv', component: () => import('./tools/csv/CsvTool.vue'), meta: { name: 'CSV', icon: '⊞', group: 'Data' } },
  { path: '/sql', component: () => import('./tools/sql/SqlTool.vue'), meta: { name: 'SQL', icon: 'Q>', group: 'Data' } },
  { path: '/base64', component: () => import('./tools/base64/Base64Tool.vue'), meta: { name: 'Base64', icon: 'b6', group: 'Encode' } },
  { path: '/jwt', component: () => import('./tools/jwt/JwtTool.vue'), meta: { name: 'JWT', icon: 'Jt', group: 'Encode' } },
  { path: '/url', component: () => import('./tools/url/UrlTool.vue'), meta: { name: 'URL', icon: '%u', group: 'Encode' } },
  { path: '/hash', component: () => import('./tools/hash/HashTool.vue'), meta: { name: 'Hash', icon: '#h', group: 'Encode' } },
  { path: '/color', component: () => import('./tools/color/ColorTool.vue'), meta: { name: 'Color', icon: '◉', group: 'Dev' } },
  { path: '/epoch', component: () => import('./tools/epoch/EpochTool.vue'), meta: { name: 'Epoch', icon: '⏱', group: 'Dev' } },
  { path: '/uuid', component: () => import('./tools/uuid/UuidTool.vue'), meta: { name: 'UUID', icon: 'id', group: 'Dev' } },
  { path: '/cron', component: () => import('./tools/cron/CronTool.vue'), meta: { name: 'Cron', icon: '⏰', group: 'Dev' } },
  { path: '/qr', component: () => import('./tools/qr/QrTool.vue'), meta: { name: 'QR', icon: '▣', group: 'Dev' } },
  { path: '/s/:id', component: () => import('./views/ShareView.vue'), meta: { hidden: true } },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export { routes }
export default router
