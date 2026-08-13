import { useMemo, useState, type FormEvent } from 'react'
import {
  AVATAR_ORDER,
  avatarUrl,
  type AvatarId,
} from '../../../entities/settings'
import {
  ApiError,
  ApiErrorCode,
  useLoginMutation,
  useRegisterMutation,
} from '../../../shared/api'
import { SCENE_ROOT } from '../../../shared/config/assets'
import { isValidEmail } from '../../../shared/lib/email'
import { isValidPassword, PASSWORD_REQUIREMENTS } from '../../../shared/lib/password'
import './auth.css'

type AuthMode = 'login' | 'register'

export function AuthPage({ serviceUnavailable = false }: { serviceUnavailable?: boolean }) {
  const [mode, setMode] = useState<AuthMode>('login')
  const [email, setEmail] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [password, setPassword] = useState('')
  const [confirmation, setConfirmation] = useState('')
  const [avatar, setAvatar] = useState<AvatarId>('profile')
  const [showPassword, setShowPassword] = useState(false)
  const [passwordFocused, setPasswordFocused] = useState(false)
  const [clientError, setClientError] = useState('')
  const login = useLoginMutation()
  const register = useRegisterMutation()
  const pending = login.isPending || register.isPending
  const requestError = login.error ?? register.error
  const errorMessage = clientError || authErrorMessage(requestError)

  const title = mode === 'login' ? 'Вход в аккаунт' : 'Создание игрока'
  const subtitle =
    mode === 'login'
      ? 'Продолжите обучение и борьбу за место в рейтинге'
      : 'Выберите имя и сохраните прогресс на сервере'

  const canSubmit = useMemo(() => {
    if (pending || !isValidEmail(email) || !isValidPassword(password)) return false
    if (mode === 'register') {
      return displayName.trim().length >= 2 && password === confirmation
    }
    return true
  }, [confirmation, displayName, email, mode, password, pending])

  const switchMode = (next: AuthMode) => {
    setMode(next)
    setClientError('')
    login.reset()
    register.reset()
  }

  const submit = (event: FormEvent) => {
    event.preventDefault()
    setClientError('')
    if (mode === 'register') {
      if (!isValidPassword(password)) {
        setClientError('Выполните все требования к паролю.')
        return
      }
      if (password !== confirmation) {
        setClientError('Пароли не совпадают.')
        return
      }
      register.mutate({ email: email.trim(), password, displayName, avatar })
      return
    }
    login.mutate({ email: email.trim(), password })
  }

  return (
    <main className='auth-page'>
      <div className='auth-page__veil' />
      <section className='auth-shell' aria-labelledby='auth-title'>
        <header className='auth-brand'>
          <img src={`${SCENE_ROOT}/brand-shield.png`} alt='' className='auth-brand__shield pixel-art' />
          <div className='auth-brand__copy'>
            <strong>АНТИСКАМ</strong>
            <span>ТРЕНАЖЁР</span>
          </div>
        </header>

        <div className='auth-heading'>
          <h1 id='auth-title'>{title}</h1>
          <p>{subtitle}</p>
        </div>

        <form className='auth-card' onSubmit={submit} noValidate>
          <div className='auth-tabs' role='tablist' aria-label='Способ входа'>
            <button
              type='button'
              role='tab'
              aria-selected={mode === 'login'}
              className={mode === 'login' ? 'auth-tabs__item auth-tabs__item--active' : 'auth-tabs__item'}
              onClick={() => switchMode('login')}
            >
              ВОЙТИ
            </button>
            <button
              type='button'
              role='tab'
              aria-selected={mode === 'register'}
              className={mode === 'register' ? 'auth-tabs__item auth-tabs__item--active' : 'auth-tabs__item'}
              onClick={() => switchMode('register')}
            >
              РЕГИСТРАЦИЯ
            </button>
          </div>

          {mode === 'register' && (
            <fieldset className='auth-avatar-picker'>
              <legend>ВАШ АВАТАР</legend>
              <div className='auth-avatar-picker__items'>
                {AVATAR_ORDER.map((candidate) => (
                  <button
                    key={candidate}
                    type='button'
                    className={candidate === avatar ? 'auth-avatar auth-avatar--active' : 'auth-avatar'}
                    aria-label={`Выбрать аватар ${AVATAR_ORDER.indexOf(candidate) + 1}`}
                    aria-pressed={candidate === avatar}
                    onClick={() => setAvatar(candidate)}
                  >
                    <img src={avatarUrl(candidate)} alt='' className='pixel-art' />
                  </button>
                ))}
              </div>
            </fieldset>
          )}

          <label className='auth-field'>
            <span>ЭЛЕКТРОННАЯ ПОЧТА</span>
            <input
              type='email'
              autoComplete='email'
              inputMode='email'
              maxLength={254}
              value={email}
              placeholder='Например, name@gmail.com'
              onChange={(event) => setEmail(event.target.value)}
            />
            {mode === 'register' && <small>Почта будет использоваться для входа в аккаунт</small>}
          </label>

          {mode === 'register' && (
            <label className='auth-field'>
              <span>ИМЯ В РЕЙТИНГЕ</span>
              <input
                type='text'
                autoComplete='nickname'
                minLength={2}
                maxLength={24}
                value={displayName}
                placeholder='Как вас увидят другие игроки'
                onChange={(event) => setDisplayName(event.target.value)}
              />
            </label>
          )}

          <label className='auth-field auth-field--password'>
            <span>ПАРОЛЬ</span>
            <span className='auth-password'>
              <input
                type={showPassword ? 'text' : 'password'}
                autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
                minLength={8}
                maxLength={72}
                value={password}
                placeholder='Например, Antiscam1!'
                aria-describedby={mode === 'register' ? 'password-requirements' : undefined}
                onFocus={() => setPasswordFocused(true)}
                onBlur={() => setPasswordFocused(false)}
                onChange={(event) => setPassword(event.target.value)}
              />
              <button
                type='button'
                aria-label={showPassword ? 'Скрыть пароль' : 'Показать пароль'}
                onClick={() => setShowPassword((shown) => !shown)}
              >
                {showPassword ? '●' : '◉'}
              </button>
            </span>
            {mode === 'register' && passwordFocused && (
              <div
                className='password-requirements'
                id='password-requirements'
                role='status'
                aria-live='polite'
              >
                <strong>Надёжный пароль содержит:</strong>
                <ul>
                  {PASSWORD_REQUIREMENTS.map((requirement) => {
                    const completed = requirement.test(password)
                    return (
                      <li
                        key={requirement.id}
                        className={completed ? 'password-requirement password-requirement--done' : 'password-requirement'}
                      >
                        <span aria-hidden='true'>{completed ? '✓' : '○'}</span>
                        {requirement.label}
                      </li>
                    )
                  })}
                </ul>
              </div>
            )}
          </label>

          {mode === 'register' && (
            <label className='auth-field'>
              <span>ПОВТОРИТЕ ПАРОЛЬ</span>
              <input
                type={showPassword ? 'text' : 'password'}
                autoComplete='new-password'
                minLength={8}
                maxLength={72}
                value={confirmation}
                placeholder='Введите пароль ещё раз'
                onChange={(event) => setConfirmation(event.target.value)}
              />
            </label>
          )}

          {(errorMessage || serviceUnavailable) && (
            <p className='auth-error' role='alert'>
              {errorMessage || 'Сервер временно недоступен. Попробуйте ещё раз.'}
            </p>
          )}

          <button className='auth-submit' type='submit' disabled={!canSubmit}>
            {pending ? 'ПОДКЛЮЧАЕМ…' : mode === 'login' ? 'ВОЙТИ' : 'СОЗДАТЬ ИГРОКА'}
          </button>

          <p className='auth-card__note'>
            Прогресс, миссии и рейтинг будут привязаны к этому аккаунту.
          </p>
        </form>
      </section>
    </main>
  )
}

function authErrorMessage(error: Error | null): string {
  if (!(error instanceof ApiError)) return error ? 'Не удалось связаться с сервером.' : ''
  switch (error.code) {
    case ApiErrorCode.invalidCredentials:
      return 'Неверная электронная почта или пароль.'
    case ApiErrorCode.emailTaken:
      return 'Эта почта уже используется. Войдите в аккаунт.'
    default:
      return error.message || 'Не удалось выполнить запрос.'
  }
}
