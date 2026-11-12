/**
 * Динамический sitemap.xml: статические страницы во всех локалях. Отдаётся
 * Nitro-сервером Nuxt по https://example.com/sitemap.xml. Добавляйте новые
 * страницы в STATIC_PATHS, а динамические (из API) — через $fetch к
 * config.apiInternalUrl.
 */
const LOCALES = [
  { code: 'ru', prefix: '' },
  { code: 'ky', prefix: '/ky' },
  { code: 'en', prefix: '/en' },
];

// Пути относительно префикса локали. '' -> корень.
const STATIC_PATHS = [''];

function xmlEscape(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&apos;');
}

export default defineEventHandler((event) => {
  const config = useRuntimeConfig(event);
  const siteUrl = String(config.public.siteUrl ?? 'https://example.com').replace(/\/$/, '');

  const urls = LOCALES.flatMap((locale) =>
    STATIC_PATHS.map((path) => `${siteUrl}${locale.prefix}${path || '/'}`),
  );

  const entries = urls.map((loc) => `  <url><loc>${xmlEscape(loc)}</loc></url>`).join('\n');
  const body = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${entries}
</urlset>
`;

  setHeader(event, 'content-type', 'application/xml; charset=utf-8');
  return body;
});
