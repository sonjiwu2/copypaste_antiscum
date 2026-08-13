import { describe, expect, it } from 'vitest'
import { isValidPassword } from './password'

describe('isValidPassword', () => {
  it.each([
    ['Antiscam1!', true],
    ['Пароль123!', false],
    ['Antiscam1', false],
    ['Antiscam!', false],
    ['12345678!', false],
    ['Antiscam 1!', false],
  ])('проверяет %s', (password, expected) => {
    expect(isValidPassword(password)).toBe(expected)
  })
})
