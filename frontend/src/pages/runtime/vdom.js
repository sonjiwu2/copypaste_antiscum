const SVG_NS = 'http://www.w3.org/2000/svg'
const SVG_TAGS = new Set([
  'svg',
  'rect',
  'path',
  'circle',
  'line',
  'polyline',
  'polygon',
  'g',
  'defs',
  'clipPath',
  'text',
])

function flatten(value, out = []) {
  if (Array.isArray(value)) {
    value.forEach((item) => flatten(item, out))
  } else if (value !== null && value !== undefined && value !== false && value !== true) {
    out.push(value)
  }
  return out
}

export const React = {
  createRef() {
    return { current: null }
  },
  createElement(type, props, ...children) {
    return {
      __dcVNode: true,
      type,
      props: props || {},
      children: flatten(children),
    }
  },
}

function setStyle(element, style) {
  if (!style) return
  if (typeof style === 'string') {
    element.style.cssText = style
    return
  }
  Object.entries(style).forEach(([key, value]) => {
    if (value === null || value === undefined) return
    try {
      element.style[key] = String(value)
    } catch {
      element.style.setProperty(key, String(value))
    }
  })
}

function setProp(element, key, value) {
  if (key === 'key' || value === undefined || value === null || value === false) return
  if (key === 'style') {
    setStyle(element, value)
    return
  }
  if (key === 'className') {
    element.setAttribute('class', value)
    return
  }
  if (/^on[A-Z]/.test(key) && typeof value === 'function') {
    element.addEventListener(key.slice(2).toLowerCase(), value)
    return
  }
  if (key === 'children') return
  if (value === true) {
    element.setAttribute(key, '')
    return
  }
  element.setAttribute(key, String(value))
}

export function renderValue(value, parentNamespace = null) {
  const fragment = document.createDocumentFragment()
  flatten(value).forEach((item) => {
    if (item instanceof Node) {
      fragment.append(item)
      return
    }
    if (item && item.__dcVNode) {
      const tag = item.type
      const useSvg = parentNamespace === SVG_NS || SVG_TAGS.has(tag)
      const el = useSvg ? document.createElementNS(SVG_NS, tag) : document.createElement(tag)
      Object.entries(item.props || {}).forEach(([key, val]) => setProp(el, key, val))
      const childNs = useSvg ? SVG_NS : null
      item.children.forEach((child) => el.append(renderValue(child, childNs)))
      fragment.append(el)
      return
    }
    fragment.append(document.createTextNode(String(item)))
  })
  return fragment
}
