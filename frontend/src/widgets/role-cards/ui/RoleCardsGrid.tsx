import { RoleCard } from './RoleCard'

export function RoleCardsGrid({
  onNavigate,
}: {
  onNavigate: (label: string, role: 'buyer' | 'seller') => void
}) {
  return (
    <div className='role-grid'>
      <RoleCard
        role='buyer'
        title='ПОКУПАТЕЛЬ'
        description='Узнай, как покупать безопасно и избегать частых схем обмана.'
        badge='Проверенный покупатель'
        missions={24}
        completed={18}
        accuracy='92%'
        onAction={() => onNavigate('MISSIONS', 'buyer')}
      />
      <RoleCard
        role='seller'
        title='ПРОДАВЕЦ'
        description='Узнай, как продавать безопасно и защищать свои объявления.'
        badge='Проверенный продавец'
        missions={22}
        completed={16}
        accuracy='89%'
        onAction={() => onNavigate('MISSIONS', 'seller')}
      />
    </div>
  )
}
