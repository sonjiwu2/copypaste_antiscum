import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { parseArgs } from 'node:util'
import { PNG } from 'pngjs'

interface BBox {
  left: number
  top: number
  right: number
  bottom: number
}

type BBoxTuple = [number, number, number, number]

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const ROOT = path.resolve(currentDir, '..')
const DEFAULT_SOURCE = path.join(ROOT, 'ui reference', 'dowloads')
const DEFAULT_DESTINATION = path.join(ROOT, 'public', 'assets', 'anti-scam', 'provided')

const ASSET_MAPPINGS: Record<string, string> = {
  'ChatGPT Image 5 авг. 2026 г., 13_40_55.png': 'daily-trophy.png',
  'ChatGPT Image 5 авг. 2026 г., 13_41_00.png': 'weekly-calendar.png',
  'ChatGPT Image 5 авг. 2026 г., 13_41_05.png': 'rewards-gift.png',
  'ChatGPT Image 5 авг. 2026 г., 13_41_09.png': 'tip-lock.png',
  'ChatGPT Image 5 авг. 2026 г., 13_41_14.png': 'seller-role.png',
  'ChatGPT Image 5 авг. 2026 г., 13_41_19.png': 'buyer-role.png',
  'ChatGPT Image 5 авг. 2026 г., 13_41_24 (1).png': 'profile-avatar.png',
  'ChatGPT Image 5 авг. 2026 г., 13_41_25 (2).png': 'featured-flag.png',
  'ChatGPT Image 5 авг. 2026 г., 13_41_25 (3).png': 'featured-message.png',
  'ChatGPT Image 5 авг. 2026 г., 13_41_26 (4).png': 'featured-payment.png',
  'ChatGPT Image 5 авг. 2026 г., 13_41_26 (5).png': 'featured-listing.png',
  'ChatGPT Image 5 авг. 2026 г., 13_41_27 (6).png': 'featured-security.png',
  'ChatGPT Image 5 авг. 2026 г., 13_52_40 (1).png': 'leader-avatar-1.png',
  'ChatGPT Image 5 авг. 2026 г., 13_52_40 (2).png': 'leader-avatar-2.png',
  'ChatGPT Image 5 авг. 2026 г., 13_52_40 (3).png': 'leader-avatar-3.png',
  'ChatGPT Image 5 авг. 2026 г., 13_56_22.png': 'seller-character.png',
}

const MANUAL_BBOXES: Record<string, BBoxTuple> = {
  'ChatGPT Image 5 авг. 2026 г., 13_41_05.png': [425, 145, 1110, 850],
}

/**
 * Cleans near-transparent generation artifacts while preserving antialiased edges.
 */
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
 * Computes the visible bounding box of an image, applying dilation filter and padding.
 */
export function calculateVisibleBbox(png: PNG, alphaThreshold = 128, filterRadius = 4): BBoxTuple {
  const { width, height, data } = png
  const mask = new Uint8Array(width * height)

  for (let i = 0; i < width * height; i++) {
    mask[i] = data[i * 4 + 3] >= alphaThreshold ? 1 : 0
  }

  let minX = width
  let minY = height
  let maxX = -1
  let maxY = -1

  for (let y = 0; y < height; y++) {
    for (let x = 0; x < width; x++) {
      if (mask[y * width + x] === 1) {
        const x1 = Math.max(0, x - filterRadius)
        const x2 = Math.min(width - 1, x + filterRadius)
        const y1 = Math.max(0, y - filterRadius)
        const y2 = Math.min(height - 1, y + filterRadius)

        if (x1 < minX) minX = x1
        if (y1 < minY) minY = y1
        if (x2 > maxX) maxX = x2
        if (y2 > maxY) maxY = y2
      }
    }
  }

  if (maxX < 0 || maxY < 0) {
    throw new Error('The source image has no visible pixels')
  }

  const left = minX
  const top = minY
  const right = maxX + 1
  const bottom = maxY + 1

  const boxWidth = right - left
  const boxHeight = bottom - top
  const padding = Math.max(12, Math.round(Math.max(boxWidth, boxHeight) * 0.035))

  return [
    Math.max(0, left - padding),
    Math.max(0, top - padding),
    Math.min(width, right + padding),
    Math.min(height, bottom + padding),
  ]
}

/**
 * Crops a PNG image according to the specified bounding box.
 */
export function cropImage(png: PNG, bbox: BBoxTuple): PNG {
  const [left, top, right, bottom] = bbox
  const width = right - left
  const height = bottom - top

  const cropped = new PNG({ width, height })

  for (let y = 0; y < height; y++) {
    for (let x = 0; x < width; x++) {
      const srcX = left + x
      const srcY = top + y
      const srcIdx = (srcY * png.width + srcX) * 4
      const dstIdx = (y * width + x) * 4

      cropped.data[dstIdx] = png.data[srcIdx]
      cropped.data[dstIdx + 1] = png.data[srcIdx + 1]
      cropped.data[dstIdx + 2] = png.data[srcIdx + 2]
      cropped.data[dstIdx + 3] = png.data[srcIdx + 3]
    }
  }

  return cropped
}

export function processAssets(sourceDir: string, destDir: string): void {
  if (!fs.existsSync(destDir)) {
    fs.mkdirSync(destDir, { recursive: true })
  }

  for (const [sourceName, targetName] of Object.entries(ASSET_MAPPINGS)) {
    const sourcePath = path.join(sourceDir, sourceName)
    const targetPath = path.join(destDir, targetName)

    if (!fs.existsSync(sourcePath)) {
      console.warn(`Skipping missing source file: ${sourceName}`)
      continue
    }

    const fileBuffer = fs.readFileSync(sourcePath)
    const sourcePng = PNG.sync.read(fileBuffer)
    const cleanedPng = cleanTransparency(sourcePng)

    const bbox = MANUAL_BBOXES[sourceName] ?? calculateVisibleBbox(cleanedPng)
    const croppedPng = cropImage(cleanedPng, bbox)

    const outputBuffer = PNG.sync.write(croppedPng, { deflateLevel: 9 })
    fs.writeFileSync(targetPath, outputBuffer)

    const origSize = `${sourcePng.width}x${sourcePng.height}`
    const newSize = `${croppedPng.width}x${croppedPng.height}`
    console.log(
      `${targetName.padEnd(24)} ${origSize.padEnd(14)} -> ${newSize.padEnd(14)} bbox=[${bbox.join(', ')}]`
    )
  }
}

function main(): void {
  const { values } = parseArgs({
    options: {
      source: {
        type: 'string',
        short: 's',
        default: DEFAULT_SOURCE,
      },
      destination: {
        type: 'string',
        short: 'd',
        default: DEFAULT_DESTINATION,
      },
    },
  })

  const sourceDir = values.source ?? DEFAULT_SOURCE
  const destDir = values.destination ?? DEFAULT_DESTINATION

  processAssets(sourceDir, destDir)
}

if (process.argv[1] && fileURLToPath(import.meta.url) === path.resolve(process.argv[1])) {
  main()
}
