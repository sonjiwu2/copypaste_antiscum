// Node 22+ объявляет собственный localStorage, который без ключа
// --localstorage-file бросает SecurityError при первом же обращении и
// перекрывает хранилище страницы. Тестам достаточно хранилища в памяти:
// оно ведёт себя как Storage и живёт в пределах одного файла тестов.
class MemoryStorage implements Storage {
  private entries = new Map<string, string>()

  get length(): number {
    return this.entries.size
  }

  key(index: number): string | null {
    return [...this.entries.keys()][index] ?? null
  }

  getItem(key: string): string | null {
    return this.entries.get(key) ?? null
  }

  setItem(key: string, value: string): void {
    this.entries.set(key, String(value))
  }

  removeItem(key: string): void {
    this.entries.delete(key)
  }

  clear(): void {
    this.entries.clear()
  }
}

for (const name of ['localStorage', 'sessionStorage']) {
  Object.defineProperty(globalThis, name, {
    configurable: true,
    writable: true,
    value: new MemoryStorage(),
  })
}
