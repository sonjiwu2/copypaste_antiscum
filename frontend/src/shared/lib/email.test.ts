import { describe, expect, it } from 'vitest'
import { isValidEmail } from './email'

describe('isValidEmail', () => {
  it.each([
    ['dmitriy@example.com', true],
    ['Dmitriy.Novikov+game@example.co.uk', true],
    ['кириллица@example.com', false],
    ['without-at.example.com', false],
    ['name@localhost', false],
    ['name..double@example.com', false],
  ])('проверяет %s', (email, expected) => {
    expect(isValidEmail(email)).toBe(expected)
  })
})
