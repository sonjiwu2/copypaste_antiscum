import { HeroBanner } from '../../../widgets/hero-banner'
import { RoleCardsGrid } from '../../../widgets/role-cards'
import { UtilityCardsGrid } from '../../../widgets/utility-cards'
import { FeaturedMissionsPanel } from '../../../widgets/featured-missions'

export function HomePage({
  onNavigate,
}: {
  onNavigate: (label: string, role?: 'buyer' | 'seller' | 'both') => void
  onToast?: (message: string) => void
}) {
  return (
    <main className='home-layout'>
      <div className='home-main'>
        <HeroBanner onAction={() => onNavigate('MISSIONS', 'both')} />
        <RoleCardsGrid onNavigate={onNavigate} />
        <UtilityCardsGrid onNavigate={onNavigate} />
      </div>
      <aside className='home-sidebar'>
        <FeaturedMissionsPanel onAction={() => onNavigate('MISSIONS', 'both')} />
      </aside>
    </main>
  )
}

