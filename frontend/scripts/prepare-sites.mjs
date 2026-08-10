/**
 * Готовит собранный клиент к публикации на Cloudflare Workers.
 *
 * Приложение — SPA с маршрутами вида /missions, поэтому воркер отдаёт
 * index.html на любой неизвестный HTML-запрос: иначе прямая ссылка на
 * внутренний раздел возвращала бы 404.
 */
import { mkdir, writeFile } from 'node:fs/promises'

const worker = `export default {
  async fetch(request, env) {
    const response = await env.ASSETS.fetch(request)

    if (
      response.status === 404 &&
      request.method === 'GET' &&
      (request.headers.get('accept') || '').includes('text/html')
    ) {
      const indexUrl = new URL('/index.html', request.url)
      return env.ASSETS.fetch(new Request(indexUrl, request))
    }

    return response
  },
}
`

await mkdir(new URL('../dist/server/', import.meta.url), { recursive: true })
await writeFile(new URL('../dist/server/index.js', import.meta.url), worker, 'utf8')
