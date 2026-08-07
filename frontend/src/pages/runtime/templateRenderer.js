import { renderValue } from './vdom.js'

const exactMustache = /^\s*{{\s*([^{}]*?)\s*}}\s*$/
const anyMustache = /{{\s*([\s\S]*?)\s*}}/g
const expressionCache = new Map()

function compileExpression(expression) {
  if (!expressionCache.has(expression)) {
    expressionCache.set(expression, new Function('ctx', `with (ctx) { return (${expression}); }`))
  }
  return expressionCache.get(expression)
}

function evaluate(expression, context) {
  try {
    return compileExpression(expression)(context)
  } catch (error) {
    console.error('[template] Failed expression:', expression, error)
    return undefined
  }
}

function exactExpression(value) {
  const match = String(value || '').match(exactMustache)
  return match ? match[1] : null
}

function interpolate(value, context) {
  return String(value).replace(anyMustache, (_, expression) => {
    const result = evaluate(expression, context)
    return result === null || result === undefined ? '' : String(result)
  })
}

function extendContext(parent, key, value, index) {
  const child = Object.create(parent)
  child[key] = value
  child.$index = index
  return child
}

function processTextNode(node, context) {
  const source = node.textContent || ''
  if (!source.includes('{{')) return

  const exact = exactExpression(source)
  if (exact !== null) {
    const result = evaluate(exact, context)
    node.replaceWith(renderValue(result))
    return
  }

  const fragment = document.createDocumentFragment()
  let cursor = 0
  source.replace(anyMustache, (whole, expression, offset) => {
    if (offset > cursor) fragment.append(document.createTextNode(source.slice(cursor, offset)))
    const result = evaluate(expression, context)
    fragment.append(renderValue(result))
    cursor = offset + whole.length
    return whole
  })
  if (cursor < source.length) fragment.append(document.createTextNode(source.slice(cursor)))
  node.replaceWith(fragment)
}

function installPseudoStates(element, baseCss, hoverCss, activeCss, focusCss) {
  let hovered = false
  let active = false
  let focused = false

  const apply = () => {
    const extras = []
    if (hovered && hoverCss) extras.push(hoverCss)
    if (focused && focusCss) extras.push(focusCss)
    if (active && activeCss) extras.push(activeCss)
    element.style.cssText = [baseCss, ...extras].filter(Boolean).join(';')
  }

  if (hoverCss) {
    element.addEventListener('mouseenter', () => {
      hovered = true
      apply()
    })
    element.addEventListener('mouseleave', () => {
      hovered = false
      active = false
      apply()
    })
  }
  if (activeCss) {
    element.addEventListener('pointerdown', () => {
      active = true
      apply()
    })
    window.addEventListener(
      'pointerup',
      () => {
        active = false
        apply()
      },
      { once: true }
    )
  }
  if (focusCss) {
    element.addEventListener('focus', () => {
      focused = true
      apply()
    })
    element.addEventListener('blur', () => {
      focused = false
      apply()
    })
  }
}

function bindEvent(element, attributeName, handler) {
  if (typeof handler !== 'function') return
  const normalized = attributeName.toLowerCase()
  let eventName = normalized.slice(2)
  if (
    eventName === 'change' &&
    element.matches('input:not([type="checkbox"]):not([type="radio"]), textarea')
  ) {
    eventName = 'input'
  }
  element.addEventListener(eventName, handler)
}

function processAttributes(element, context) {
  const pendingProperties = []
  let hoverCss = ''
  let activeCss = ''
  let focusCss = ''

  for (const attribute of Array.from(element.attributes)) {
    const originalName = attribute.name
    const name = originalName.toLowerCase()
    const rawValue = attribute.value

    if (name.startsWith('hint-placeholder-')) {
      element.removeAttribute(originalName)
      continue
    }

    if (name === 'style-hover' || name === 'style-active' || name === 'style-focus') {
      const css = interpolate(rawValue, context)
      if (name === 'style-hover') hoverCss = css
      if (name === 'style-active') activeCss = css
      if (name === 'style-focus') focusCss = css
      element.removeAttribute(originalName)
      continue
    }

    const expression = exactExpression(rawValue)

    if (name.startsWith('on') && expression !== null) {
      bindEvent(element, name, evaluate(expression, context))
      element.removeAttribute(originalName)
      continue
    }

    if (name === 'ref' && expression !== null) {
      const ref = evaluate(expression, context)
      if (ref && typeof ref === 'object') ref.current = element
      element.removeAttribute(originalName)
      continue
    }

    if (name === 'value' && expression !== null && 'value' in element) {
      pendingProperties.push(() => {
        const value = evaluate(expression, context)
        element.value = value === null || value === undefined ? '' : value
      })
      element.removeAttribute(originalName)
      continue
    }

    if ((name === 'disabled' || name === 'checked') && expression !== null) {
      pendingProperties.push(() => {
        element[name] = Boolean(evaluate(expression, context))
      })
      element.removeAttribute(originalName)
      continue
    }

    if (rawValue.includes('{{')) {
      const result =
        expression !== null ? evaluate(expression, context) : interpolate(rawValue, context)

      if (result === false || result === null || result === undefined) {
        element.removeAttribute(originalName)
      } else if (result === true) {
        element.setAttribute(originalName, '')
      } else {
        element.setAttribute(originalName, String(result))
      }
    }
  }

  const baseCss = element.style.cssText
  if (hoverCss || activeCss || focusCss) {
    installPseudoStates(element, baseCss, hoverCss, activeCss, focusCss)
  }

  return pendingProperties
}

function processNode(node, context) {
  if (node.nodeType === Node.TEXT_NODE) {
    processTextNode(node, context)
    return
  }
  if (node.nodeType !== Node.ELEMENT_NODE) return

  const tag = node.tagName.toLowerCase()

  if (tag === 'sc-if') {
    const expression = exactExpression(node.getAttribute('value'))
    const visible = expression !== null && Boolean(evaluate(expression, context))
    const fragment = document.createDocumentFragment()
    if (visible) {
      Array.from(node.childNodes).forEach((child) => {
        const clone = child.cloneNode(true)
        fragment.append(clone)
        processNode(clone, context)
      })
    }
    node.replaceWith(fragment)
    return
  }

  if (tag === 'sc-for') {
    const expression = exactExpression(node.getAttribute('list'))
    const list = expression === null ? [] : evaluate(expression, context)
    const alias = node.getAttribute('as') || 'item'
    const fragment = document.createDocumentFragment()
    Array.from(list || []).forEach((item, index) => {
      const localContext = extendContext(context, alias, item, index)
      Array.from(node.childNodes).forEach((child) => {
        const clone = child.cloneNode(true)
        fragment.append(clone)
        processNode(clone, localContext)
      })
    })
    node.replaceWith(fragment)
    return
  }

  const pendingProperties = processAttributes(node, context)
  Array.from(node.childNodes).forEach((child) => processNode(child, context))
  pendingProperties.forEach((apply) => apply())
}

function rememberUiState(root) {
  const active = document.activeElement
  const focusKey =
    active && root.contains(active)
      ? active.getAttribute('id') ||
        active.getAttribute('aria-label') ||
        active.getAttribute('name')
      : null

  const scroll = new Map()
  root
    .querySelectorAll('[data-screen-label], [style*="overflow-y:auto"], main')
    .forEach((element, index) => {
      scroll.set(index, { top: element.scrollTop, left: element.scrollLeft })
    })

  return { focusKey, scroll, windowX: window.scrollX, windowY: window.scrollY }
}

function restoreUiState(root, snapshot) {
  requestAnimationFrame(() => {
    if (snapshot.focusKey) {
      const escaped = CSS.escape(snapshot.focusKey)
      const target = root.querySelector(
        `#${escaped}, [aria-label="${escaped}"], [name="${escaped}"]`
      )
      target?.focus({ preventScroll: true })
    }
    root
      .querySelectorAll('[data-screen-label], [style*="overflow-y:auto"], main')
      .forEach((element, index) => {
        const position = snapshot.scroll.get(index)
        if (position) {
          element.scrollTop = position.top
          element.scrollLeft = position.left
        }
      })
    window.scrollTo(snapshot.windowX, snapshot.windowY)
  })
}

export class TemplateRenderer {
  constructor({ root, template, component }) {
    this.root = root
    this.template = template
    this.component = component
    component.__renderer = this
  }

  mount() {
    this.render({ callDidUpdate: false })
    this.component.__mounted = true
    this.component.componentDidMount?.()
  }

  unmount() {
    this.component.componentWillUnmount?.()
    this.root.replaceChildren()
  }

  render({ previousState = null, afterRender = null, callDidUpdate = false } = {}) {
    const snapshot = rememberUiState(this.root)
    const context = this.component.renderVals()
    const tpl = document.createElement('template')
    tpl.innerHTML = this.template
    const fragment = tpl.content.cloneNode(true)
    Array.from(fragment.childNodes).forEach((node) => processNode(node, context))
    this.root.replaceChildren(fragment)
    restoreUiState(this.root, snapshot)
    if (callDidUpdate) this.component.componentDidUpdate?.(this.component.props, previousState)
    afterRender?.()
  }
}
