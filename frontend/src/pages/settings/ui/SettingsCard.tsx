import type { ReactNode } from 'react'
import { PixelIcon, type PixelIconName } from '../../../shared/ui/PixelIcon'

export interface SettingsCardProps {
  icon: PixelIconName
  title: string
  /** Цветовая тема карточки: совпадает с назначением блока. */
  tone?: 'blue' | 'purple' | 'orange' | 'green' | 'red'
  children: ReactNode
}

export function SettingsCard({ icon, title, tone = 'blue', children }: SettingsCardProps) {
  const headingId = `settings-${title.toLowerCase()}`

  return (
    <section className={`settings-card settings-card--${tone}`} aria-labelledby={headingId}>
      <div className='settings-card__heading'>
        <PixelIcon name={icon} size={22} />
        <h2 id={headingId}>{title}</h2>
      </div>
      <div className='settings-card__body'>{children}</div>
    </section>
  )
}
