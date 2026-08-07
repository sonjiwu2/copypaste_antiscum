export class DCLogic {
  constructor(props = {}) {
    this.props = props
    this.__renderer = null
    this.__mounted = false
  }

  setState(update, callback) {
    const previousState = { ...(this.state || {}) }
    const patch = typeof update === 'function' ? update(this.state || {}, this.props) : update

    if (patch && typeof patch === 'object') {
      this.state = { ...(this.state || {}), ...patch }
    }

    if (!this.__renderer) {
      callback?.()
      return
    }

    this.__renderer.render({
      previousState,
      afterRender: callback,
      callDidUpdate: this.__mounted,
    })
  }
}
