import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { AuthPage } from './AuthPage'

const loginMutate = vi.fn()
const registerMutate = vi.fn()

vi.mock('../../../shared/api', () => ({
  ApiError: class extends Error {},
  ApiErrorCode: { invalidCredentials: 'INVALID_CREDENTIALS', emailTaken: 'EMAIL_TAKEN' },
  useLoginMutation: () => ({ mutate: loginMutate, reset: vi.fn(), isPending: false, error: null }),
  useRegisterMutation: () => ({ mutate: registerMutate, reset: vi.fn(), isPending: false, error: null }),
}))

describe('AuthPage', () => {
  beforeEach(() => {
    loginMutate.mockReset()
    registerMutate.mockReset()
  })

  it('входит по почте и паролю', async () => {
    const user = userEvent.setup()
    render(<AuthPage />)

    await user.type(screen.getByPlaceholderText('Например, name@gmail.com'), 'dmitriy@example.com')
    await user.type(screen.getByPlaceholderText('Например, Antiscam1!'), 'password1!')
    await user.click(screen.getByRole('button', { name: 'ВОЙТИ' }))

    expect(loginMutate).toHaveBeenCalledWith({ email: 'dmitriy@example.com', password: 'password1!' })
  })

  it('регистрирует имя и выбранный аватар', async () => {
    const user = userEvent.setup()
    render(<AuthPage />)

    await user.click(screen.getByRole('tab', { name: 'РЕГИСТРАЦИЯ' }))
    await user.click(screen.getByRole('button', { name: 'Выбрать аватар 3' }))
    await user.type(screen.getByPlaceholderText('Например, name@gmail.com'), 'dmitriy@example.com')
    await user.type(screen.getByPlaceholderText('Как вас увидят другие игроки'), 'Дмитрий')
    await user.type(screen.getByPlaceholderText('Например, Antiscam1!'), 'password1!')
    await user.type(screen.getByPlaceholderText('Введите пароль ещё раз'), 'password1!')
    await user.click(screen.getByRole('button', { name: 'СОЗДАТЬ ИГРОКА' }))

    expect(registerMutate).toHaveBeenCalledWith({
      email: 'dmitriy@example.com',
      password: 'password1!',
      displayName: 'Дмитрий',
      avatar: 'leader-2',
    })
  })

  it('показывает выполнение требований к паролю', async () => {
    const user = userEvent.setup()
    render(<AuthPage />)

    await user.click(screen.getByRole('tab', { name: 'РЕГИСТРАЦИЯ' }))
    const password = screen.getByPlaceholderText('Например, Antiscam1!')
    await user.click(password)

    expect(screen.getByText('Надёжный пароль содержит:')).toBeTruthy()
    await user.type(password, 'Antiscam1!')

    expect(document.querySelectorAll('.password-requirement--done')).toHaveLength(5)
  })
})
