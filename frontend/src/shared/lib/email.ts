/** Совпадает с публичными ограничениями серверной проверки email. */
export function isValidEmail(email: string): boolean {
  const value = email.trim()
  if (value.length < 5 || value.length > 254 || value.includes(' ')) return false

  const parts = value.split('@')
  if (parts.length !== 2) return false
  const [local, domain] = parts
  if (
    !local ||
    local.length > 64 ||
    local.startsWith('.') ||
    local.endsWith('.') ||
    local.includes('..') ||
    !/^[A-Za-z0-9.!#$%&'*+/=?^_`{|}~-]+$/.test(local)
  ) {
    return false
  }

  const labels = domain.split('.')
  return (
    domain.length <= 253 &&
    labels.length >= 2 &&
    labels.at(-1)!.length >= 2 &&
    labels.every(
      (label) =>
        label.length >= 1 &&
        label.length <= 63 &&
        !label.startsWith('-') &&
        !label.endsWith('-') &&
        /^[A-Za-z0-9-]+$/.test(label)
    )
  )
}
