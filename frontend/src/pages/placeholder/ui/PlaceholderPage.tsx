import { PixelIcon } from '../../../shared/ui/PixelIcon'
import { AppRoute } from '../../../widgets/header'

export function PlaceholderPage({
  route,
  onBack,
}: {
  route: Exclude<AppRoute, 'home' | 'missions' | 'play'>
  onBack: () => void
}) {
  const title = route.toUpperCase()
  return (
    <main className='route-placeholder'>
      <PixelIcon
        name={route === 'progress' ? 'chart' : route === 'rules' ? 'book' : 'gear'}
        size={54}
      />
      <h1>{title}</h1>
      <p>This section is connected to navigation and ready for its detailed build.</p>
      <button type='button' onClick={onBack}>
        Back to Missions
      </button>
    </main>
  )
}
