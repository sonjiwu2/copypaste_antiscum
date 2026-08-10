import { PixelIcon, type PixelIconName } from '../../../shared/ui/PixelIcon'
import type { AppRoute } from '../../../widgets/header'

type PlaceholderRoute = Exclude<AppRoute, 'home' | 'missions' | 'play' | 'weekly-test' | 'rewards'>

const SECTIONS: Record<PlaceholderRoute, { icon: PixelIconName; title: string; text: string }> = {
  progress: {
    icon: 'progress',
    title: 'ПРОГРЕСС',
    text: 'Раздел появится вместе со статистикой навыков.',
  },
  rules: {
    icon: 'rules',
    title: 'ПРАВИЛА',
    text: 'Здесь соберём разбор схем обмана и правила безопасной сделки.',
  },
  settings: {
    icon: 'settings',
    title: 'НАСТРОЙКИ',
    text: 'Звук, уведомления и сброс прогресса появятся здесь.',
  },
}

export function PlaceholderPage({
  route,
  onBack,
}: {
  route: PlaceholderRoute
  onBack: () => void
}) {
  const section = SECTIONS[route]

  return (
    <main className='route-placeholder'>
      <PixelIcon name={section.icon} size={54} />
      <h1>{section.title}</h1>
      <p>{section.text}</p>
      <button type='button' onClick={onBack}>
        К списку миссий
      </button>
    </main>
  )
}
