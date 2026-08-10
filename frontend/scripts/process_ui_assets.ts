/**
 * Готовит пиксельные ассеты для приложения.
 *
 * Исходники лежат в `ui reference` в том виде, в котором их выдал генератор:
 * большой холст, полупрозрачный мусор по краям и вес в сотни килобайт.
 * Скрипт обрезает пустые поля и уменьшает картинку до размера, который реально
 * нужен интерфейсу, — иначе иконка 22 px тянула бы 400 KB трафика.
 *
 * Запуск: npm run process-assets
 */
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { parseArgs } from 'node:util'
import { PNG } from 'pngjs'

type BBoxTuple = [left: number, top: number, right: number, bottom: number]

interface AssetGroup {
  /** Папка с исходниками относительно корня фронтенда. */
  source: string
  /** Папка публикации относительно корня фронтенда. */
  destination: string
  /** Максимальная сторона результата в пикселях. */
  maxSide: number
}

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const ROOT = path.resolve(currentDir, '..')

/*
 * Пути ведут в `ui reference/source` — там лежат только те исходники, из
 * которых собирается `public/assets`. Справочные материалы и неиспользуемые
 * варианты вынесены в соседние папки и конвейером не читаются.
 *
 * Сцена (`source/scene`) в список не входит: эти файлы публикуются как есть,
 * без обрезки и уменьшения.
 */
const ASSET_GROUPS: AssetGroup[] = [
  // Иконки интерфейса: самое крупное место показа — плитка миссии 58 px.
  { source: 'ui reference/source/icons', destination: 'public/assets/icons', maxSide: 128 },
  // Иллюстрации карточек и аватары: максимум 72 px в карточках главной страницы.
  { source: 'ui reference/source/art', destination: 'public/assets/art', maxSide: 192 },
  // Иллюстрации страницы правил: крупнее обычных иконок, но исходники 1280 px
  // в браузере не нужны.
  { source: 'ui reference/source/rules', destination: 'public/assets/rules', maxSide: 256 },
]

/**
 * Кадры, у которых автоматическая обрезка захватывает лишнее.
 * Ключ — имя файла, значение — рамка в координатах исходника.
 */
const MANUAL_BBOXES: Record<string, BBoxTuple> = {
  'rewards-gift.png': [425, 145, 1110, 850],
}

/** Убирает почти прозрачный мусор генерации, не задевая сглаженные края. */
export function cleanTransparency(png: PNG, threshold = 18): PNG {
  const { data } = png

  for (let i = 0; i < data.length; i += 4) {
    if (data[i + 3] < threshold) {
      data[i + 3] = 0
    }
  }

  return png
}

/**
 * Рамка видимой части картинки с небольшим запасом.
 *
 * Запас нужен, чтобы обрезка не срезала сглаженные края вплотную: иначе у
 * иконки появляется рваный контур.
 */
export function calculateVisibleBbox(png: PNG, alphaThreshold = 128, padding = 4): BBoxTuple {
  const { width, height, data } = png

  let minX = width
  let minY = height
  let maxX = -1
  let maxY = -1

  for (let y = 0; y < height; y++) {
    for (let x = 0; x < width; x++) {
      if (data[(y * width + x) * 4 + 3] < alphaThreshold) {
        continue
      }

      if (x < minX) minX = x
      if (y < minY) minY = y
      if (x > maxX) maxX = x
      if (y > maxY) maxY = y
    }
  }

  if (maxX < 0) {
    throw new Error('В исходной картинке нет видимых пикселей')
  }

  const extra = Math.max(padding, Math.round(Math.max(maxX - minX, maxY - minY) * 0.035))

  return [
    Math.max(0, minX - extra),
    Math.max(0, minY - extra),
    Math.min(width, maxX + 1 + extra),
    Math.min(height, maxY + 1 + extra),
  ]
}

export function cropImage(png: PNG, [left, top, right, bottom]: BBoxTuple): PNG {
  const cropped = new PNG({ width: right - left, height: bottom - top })

  for (let y = 0; y < cropped.height; y++) {
    for (let x = 0; x < cropped.width; x++) {
      const source = ((top + y) * png.width + (left + x)) * 4
      const target = (y * cropped.width + x) * 4

      cropped.data.set(png.data.subarray(source, source + 4), target)
    }
  }

  return cropped
}

/**
 * Уменьшает картинку усреднением по площади.
 *
 * Пиксель-арт нарисован крупными блоками, поэтому усреднение сохраняет
 * границы блоков и, в отличие от выбора одного пикселя, не «съедает» тонкий
 * контур в один пиксель.
 */
export function downscale(png: PNG, maxSide: number): PNG {
  const scale = maxSide / Math.max(png.width, png.height)
  if (scale >= 1) {
    return png
  }

  const resized = new PNG({
    width: Math.max(1, Math.round(png.width * scale)),
    height: Math.max(1, Math.round(png.height * scale)),
  })

  for (let y = 0; y < resized.height; y++) {
    const sourceTop = Math.floor((y * png.height) / resized.height)
    const sourceBottom = Math.max(
      sourceTop + 1,
      Math.floor(((y + 1) * png.height) / resized.height)
    )

    for (let x = 0; x < resized.width; x++) {
      const sourceLeft = Math.floor((x * png.width) / resized.width)
      const sourceRight = Math.max(
        sourceLeft + 1,
        Math.floor(((x + 1) * png.width) / resized.width)
      )

      let alpha = 0
      let red = 0
      let green = 0
      let blue = 0

      for (let sourceY = sourceTop; sourceY < sourceBottom; sourceY++) {
        for (let sourceX = sourceLeft; sourceX < sourceRight; sourceX++) {
          const index = (sourceY * png.width + sourceX) * 4
          // Цвет взвешивается прозрачностью: полностью прозрачные пиксели
          // хранят произвольный цвет и иначе осветлили бы края.
          const weight = png.data[index + 3] / 255

          red += png.data[index] * weight
          green += png.data[index + 1] * weight
          blue += png.data[index + 2] * weight
          alpha += weight
        }
      }

      const pixels = (sourceBottom - sourceTop) * (sourceRight - sourceLeft)
      const target = (y * resized.width + x) * 4

      if (alpha > 0) {
        resized.data[target] = Math.round(red / alpha)
        resized.data[target + 1] = Math.round(green / alpha)
        resized.data[target + 2] = Math.round(blue / alpha)
      }

      resized.data[target + 3] = Math.round((alpha / pixels) * 255)
    }
  }

  return resized
}

function processGroup({ source, destination, maxSide }: AssetGroup): void {
  const sourceDir = path.join(ROOT, source)
  const destinationDir = path.join(ROOT, destination)

  if (!fs.existsSync(sourceDir)) {
    console.warn(`Пропущена группа: нет папки ${source}`)

    return
  }

  fs.mkdirSync(destinationDir, { recursive: true })

  const files = fs
    .readdirSync(sourceDir)
    .filter((file) => file.toLowerCase().endsWith('.png'))
    .sort()

  console.log(`\n${source} -> ${destination} (max ${maxSide}px)`)

  for (const file of files) {
    const original = PNG.sync.read(fs.readFileSync(path.join(sourceDir, file)))
    const cleaned = cleanTransparency(original)
    const bbox = MANUAL_BBOXES[file] ?? calculateVisibleBbox(cleaned)
    const result = downscale(cropImage(cleaned, bbox), maxSide)
    const output = PNG.sync.write(result, { deflateLevel: 9 })

    fs.writeFileSync(path.join(destinationDir, file), output)

    console.log(
      `  ${file.padEnd(28)} ${`${original.width}x${original.height}`.padEnd(11)} -> ` +
        `${`${result.width}x${result.height}`.padEnd(9)} ${(output.length / 1024).toFixed(1)} KB`
    )
  }
}

function main(): void {
  const { values } = parseArgs({
    options: {
      source: { type: 'string', short: 's' },
      destination: { type: 'string', short: 'd' },
      'max-side': { type: 'string', short: 'm' },
    },
  })

  // Явная пара путей обрабатывает одну папку, без аргументов — все группы.
  if (values.source && values.destination) {
    processGroup({
      source: values.source,
      destination: values.destination,
      maxSide: Number(values['max-side'] ?? 128),
    })

    return
  }

  ASSET_GROUPS.forEach(processGroup)
}

if (process.argv[1] && fileURLToPath(import.meta.url) === path.resolve(process.argv[1])) {
  main()
}
