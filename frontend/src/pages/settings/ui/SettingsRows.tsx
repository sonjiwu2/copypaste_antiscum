import { useId } from 'react'
import { PixelIcon, type PixelIconName } from '../../../shared/ui/PixelIcon'

export interface ToggleRowProps {
  label: string
  description?: string
  checked: boolean
  disabled?: boolean
  onChange: (checked: boolean) => void
}

/**
 * Строка с переключателем.
 *
 * Под оформлением остаётся обычный чекбокс: он доступен с клавиатуры и читалка
 * сама сообщает состояние, чего не даёт `div` с обработчиком клика.
 */
export function ToggleRow({ label, description, checked, disabled, onChange }: ToggleRowProps) {
  const descriptionId = useId()

  return (
    <label className={`toggle-row${disabled ? ' toggle-row--disabled' : ''}`}>
      <span className='toggle-row__copy'>
        <strong>{label}</strong>
        {description && <small id={descriptionId}>{description}</small>}
      </span>
      <input
        type='checkbox'
        className='toggle-row__input'
        checked={checked}
        disabled={disabled}
        aria-describedby={description ? descriptionId : undefined}
        onChange={(event) => onChange(event.target.checked)}
      />
      <span className='toggle-row__switch' aria-hidden='true' />
    </label>
  )
}

export interface SliderRowProps {
  icon: PixelIconName
  label: string
  value: number
  disabled?: boolean
  onChange: (value: number) => void
}

/** Строка с ползунком громкости: иконка, подпись, шкала и текущее значение. */
export function SliderRow({ icon, label, value, disabled, onChange }: SliderRowProps) {
  return (
    <div className={`slider-row${disabled ? ' slider-row--disabled' : ''}`}>
      <PixelIcon name={icon} size={24} className='slider-row__icon' />
      <label className='slider-row__control'>
        <span>{label}</span>
        <input
          type='range'
          min={0}
          max={100}
          step={5}
          value={value}
          disabled={disabled}
          onChange={(event) => onChange(Number(event.target.value))}
        />
      </label>
      <output className='slider-row__value'>{value}%</output>
    </div>
  )
}

export interface PendingRowProps {
  label: string
  description?: string
}

/**
 * Строка настройки, которой пока нет.
 *
 * Показывается вместо переключателя: неактивный тумблер выглядит как поломка,
 * а замок и подпись честно сообщают, что настройка появится позже.
 */
export function PendingRow({ label, description }: PendingRowProps) {
  return (
    <div className='pending-row'>
      <span className='pending-row__copy'>
        <strong>{label}</strong>
        {description && <small>{description}</small>}
      </span>
      <span className='pending-row__badge'>
        <PixelIcon name='locked' size={16} />
        В разработке
      </span>
    </div>
  )
}

export interface CheckboxRowProps {
  label: string
  checked: boolean
  onChange: (checked: boolean) => void
}

export function CheckboxRow({ label, checked, onChange }: CheckboxRowProps) {
  return (
    <label className='checkbox-row'>
      <input
        type='checkbox'
        checked={checked}
        onChange={(event) => onChange(event.target.checked)}
      />
      <span>{label}</span>
    </label>
  )
}
