import { useState } from 'react'
import {
  avatarUrl,
  defaultSettingsForProfile,
  nextAvatar,
  useSettingsStore,
  type AppSettings,
} from '../../../entities/settings'
import { useUserProgressStore } from '../../../entities/user-progress'
import { useToastStore } from '../../../features/toast'
import { Glyph } from '../../../shared/ui/Glyph'
import { PixelIcon } from '../../../shared/ui/PixelIcon'
import { ART_ROOT } from '../../../shared/config/assets'
import { resetProfileData } from '../../../shared/api'
import { useNotificationsStore } from '../../../features/notifications'
import { SettingsCard } from './SettingsCard'
import { CheckboxRow, PendingRow, SliderRow, ToggleRow } from './SettingsRows'
import './settings.css'

/** Ключи в localStorage, которые сбрасывает кнопка «Сбросить фильтры». */
const FILTER_STORAGE_KEYS = ['antiscam.activeMission']

export function SettingsPage({ onToast }: { onToast?: (message: string) => void }) {
  const saved = useSettingsStore((state) => state.settings)
  const registeredAt = useSettingsStore((state) => state.registeredAt)
  const applySettings = useSettingsStore((state) => state.applySettings)
  const level = useUserProgressStore((state) => state.level ?? 1)
  const resetProgress = useUserProgressStore((state) => state.resetProgress)
  const showToast = useToastStore((state) => state.showToast)
  const notify = onToast ?? showToast

  // Черновик правок. Настройки вступают в силу по кнопке «Сохранить», поэтому
  // страница держит свою копию, а не пишет в store на каждое нажатие.
  const [draft, setDraft] = useState<AppSettings>(saved)
  const [renaming, setRenaming] = useState(false)
  const [resettingProgress, setResettingProgress] = useState(false)

  const isDirty = Object.keys(draft).some(
    (key) => draft[key as keyof AppSettings] !== saved[key as keyof AppSettings]
  )

  const change = <K extends keyof AppSettings>(key: K, value: AppSettings[K]) =>
    setDraft((current) => ({ ...current, [key]: value }))

  const handleSave = () => {
    applySettings(draft)
    setRenaming(false)
    notify('Настройки сохранены.')
  }

  const handleCancel = () => {
    setDraft(saved)
    setRenaming(false)
  }

  const handleClearFilters = () => {
    for (const key of FILTER_STORAGE_KEYS) {
      try {
        localStorage.removeItem(key)
      } catch {
        // Без хранилища сбрасывать нечего: фильтры и так живут до перезагрузки.
      }
    }
    notify('Фильтры каталога сброшены.')
  }

  const handleClearProgress = async () => {
    // Действие необратимо, поэтому подтверждается отдельно.
    if (!window.confirm('Удалить весь прогресс обучения? Действие необратимо.')) {
      return
    }

    setResettingProgress(true)
    try {
      await resetProfileData()
      resetProgress()
      useNotificationsStore.getState().resetNotifications()
      clearWeeklyTestDrafts()
      notify('Весь прогресс, недельный тест и отсчёт удалены.')
      window.setTimeout(() => window.location.assign('/'), 250)
    } catch {
      notify('Не удалось удалить серверный прогресс. Попробуйте ещё раз.')
      setResettingProgress(false)
    }
  }

  const registrationDate = new Date(registeredAt).toLocaleDateString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
  })

  return (
    <main className='settings-layout'>
      <div className='settings-grid'>
        <SettingsCard icon='settings' title='ОБЩИЕ' tone='blue'>
          <label className='select-row'>
            <span>Язык</span>
            <select
              value={draft.language}
              onChange={(event) =>
                change('language', event.target.value as AppSettings['language'])
              }
            >
              <option value='ru'>Русский</option>
            </select>
          </label>

          <PendingRow
            label='Анимации интерфейса'
            description='Управление плавными переходами появится в следующей версии'
          />
          <ToggleRow
            label='Показывать советы дня'
            description='Показывать полезный совет на главной'
            checked={draft.showDailyTip}
            onChange={(value) => change('showDailyTip', value)}
          />

          <p className='settings-note'>
            <PixelIcon name='star' size={18} />
            Настройки хранятся в этом браузере и не передаются на сервер.
          </p>
        </SettingsCard>

        <SettingsCard icon='sound-on' title='ЗВУК' tone='purple'>
          <SliderRow
            icon='sound-off'
            label='Общая громкость'
            value={draft.masterVolume}
            onChange={(value) => change('masterVolume', value)}
          />
          <SliderRow
            icon='new'
            label='Эффекты'
            value={draft.effectsVolume}
            disabled={draft.masterVolume === 0}
            onChange={(value) => change('effectsVolume', value)}
          />
          <SliderRow
            icon='music'
            label='Музыка'
            value={draft.musicVolume}
            disabled={draft.masterVolume === 0 || draft.muteMenuMusic}
            onChange={(value) => change('musicVolume', value)}
          />
          <CheckboxRow
            label='Отключить музыку в меню'
            checked={draft.muteMenuMusic}
            onChange={(value) => change('muteMenuMusic', value)}
          />
        </SettingsCard>

        <SettingsCard icon='bell' title='УВЕДОМЛЕНИЯ' tone='orange'>
          <ToggleRow
            label='Ежедневные напоминания'
            description='Напоминать о ежедневных целях'
            checked={draft.notifyDaily}
            onChange={(value) => change('notifyDaily', value)}
          />
          <ToggleRow
            label='Новые миссии'
            description='Уведомлять о появлении новых миссий'
            checked={draft.notifyNewMissions}
            onChange={(value) => change('notifyNewMissions', value)}
          />
          <ToggleRow
            label='Прогресс'
            description='Сообщать о новом уровне и полученных наградах'
            checked={draft.notifyProgress}
            onChange={(value) => change('notifyProgress', value)}
          />

          <p className='settings-note'>
            <PixelIcon name='star' size={18} />
            Уведомления показываются внутри приложения.
          </p>
        </SettingsCard>

        <SettingsCard icon='listing' title='ИНТЕРФЕЙС' tone='blue'>
          <ToggleRow
            label='Тёмная тема'
            description='Тёмное оформление вместо светлого'
            checked={draft.theme === 'dark'}
            onChange={(value) => change('theme', value ? 'dark' : 'light')}
          />
          <PendingRow
            label='Масштаб интерфейса'
            description='Увеличение текста и элементов управления'
          />
          <PendingRow
            label='Высокая контрастность'
            description='Более заметные границы и контрастные цвета'
          />

          <p className='settings-note'>
            <PixelIcon name='star' size={18} />
            Тема применяется ко всему приложению и сохраняется в этом браузере.
          </p>
        </SettingsCard>

        <SettingsCard icon='account' title='АККАУНТ' tone='green'>
          <div className='account-row'>
            <img src={avatarUrl(draft.avatar)} alt='' className='account-row__avatar pixel-art' />
            <div className='account-row__info'>
              {renaming ? (
                <label className='account-row__rename'>
                  <span className='visually-hidden'>Имя игрока</span>
                  <input
                    type='text'
                    maxLength={24}
                    value={draft.playerName}
                    autoFocus
                    onChange={(event) => change('playerName', event.target.value)}
                  />
                </label>
              ) : (
                <strong>{draft.playerName || 'Без имени'}</strong>
              )}
              <span>Уровень {level}</span>
              <small className='account-row__registered'>
                <img src={`${ART_ROOT}/weekly-calendar.png`} alt='' className='pixel-art' />
                Дата регистрации: {registrationDate}
              </small>
            </div>
          </div>
          <div className='account-actions'>
            <button type='button' onClick={() => change('avatar', nextAvatar(draft.avatar))}>
              <Glyph name='pencil' size={16} />
              Изменить аватар
            </button>
            <button type='button' onClick={() => setRenaming((value) => !value)}>
              <Glyph name='key' size={16} />
              {renaming ? 'Готово' : 'Сменить имя'}
            </button>
          </div>
        </SettingsCard>

        <SettingsCard icon='database' title='ДАННЫЕ' tone='red'>
          <button className='data-button' type='button' onClick={handleClearFilters}>
            <PixelIcon name='reset' size={18} />
            Сбросить фильтры
          </button>
          <p className='data-hint'>Сбросить фильтры каталога на стандартные значения</p>

          <button
            className='data-button data-button--danger'
            type='button'
            disabled={resettingProgress}
            onClick={() => void handleClearProgress()}
          >
            <Glyph name='trash' size={18} />
            {resettingProgress ? 'Удаляем данные...' : 'Очистить прогресс'}
          </button>
          <p className='data-hint'>Удалить весь прогресс обучения</p>

          <p className='settings-warning'>
            <Glyph name='warning' size={18} />
            Эти действия необратимы. Пожалуйста, будьте осторожны.
          </p>
        </SettingsCard>
      </div>

      <footer className='settings-footer'>
        <button
          type='button'
          className='settings-footer__reset'
          onClick={() => setDraft((current) => defaultSettingsForProfile(current))}
        >
          <PixelIcon name='reset' size={18} />
          По умолчанию
        </button>
        <button
          type='button'
          className='settings-footer__save'
          disabled={!isDirty}
          onClick={handleSave}
        >
          <PixelIcon name='save' size={18} />
          Сохранить
        </button>
        <button
          type='button'
          className='settings-footer__cancel'
          disabled={!isDirty}
          onClick={handleCancel}
        >
          <PixelIcon name='close' size={18} />
          Отмена
        </button>
        <p className='settings-footer__hint'>Изменения применяются после сохранения</p>
      </footer>
    </main>
  )
}

function clearWeeklyTestDrafts() {
  try {
    for (let index = localStorage.length - 1; index >= 0; index -= 1) {
      const key = localStorage.key(index)
      if (key?.startsWith('antiscam.weekly_test.')) {
        localStorage.removeItem(key)
      }
    }
  } catch {
    // Серверный профиль уже очищен; недоступный localStorage не мешает сбросу.
  }
}
