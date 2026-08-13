/**
 * Единое правило пароля для браузера и API.
 * Печатный ASCII гарантирует одинаковый ввод на обычной и защищённой
 * мобильной клавиатуре; пробелы и национальные раскладки не допускаются.
 */
export const PASSWORD_REQUIREMENTS = [
  {
    id: 'length',
    label: 'От 8 до 72 символов',
    test: (password: string) => password.length >= 8 && password.length <= 72,
  },
  {
    id: 'alphabet',
    label: 'Только латинские символы без пробелов',
    test: (password: string) => password.length > 0 && /^[\x21-\x7e]+$/.test(password),
  },
  {
    id: 'letter',
    label: 'Минимум одна латинская буква',
    test: (password: string) => /[A-Za-z]/.test(password),
  },
  {
    id: 'digit',
    label: 'Минимум одна цифра',
    test: (password: string) => /\d/.test(password),
  },
  {
    id: 'special',
    label: 'Минимум один спецсимвол: ! @ # $ % и другие',
    test: (password: string) => /[^A-Za-z0-9]/.test(password),
  },
] as const

export function isValidPassword(password: string): boolean {
  return PASSWORD_REQUIREMENTS.every((requirement) => requirement.test(password))
}

export const PASSWORD_HINT =
  '8–72 символа: латинские буквы, цифры и минимум один спецсимвол'
