import { UtilityCard } from './UtilityCard'
import { useToastStore } from '../../../features/toast'

export function UtilityCardsGrid({
  onNavigate,
  onToast,
}: {
  onNavigate: (label: string) => void
  onToast?: (msg: string) => void
}) {
  const showToast = useToastStore((state) => state.showToast)
  const handleToast = onToast ?? showToast

  return (
    <div className='utility-grid'>
      <UtilityCard
        kind='tip'
        title='СОВЕТ ДНЯ'
        description='Никогда не передавай смс-коды и пароли никому, даже если они представились поддержкой.'
        action='Подробнее'
        asset='tip-lock.png'
        onAction={() => onNavigate('RULES')}
      />
      <UtilityCard
        kind='daily'
        title='ЕЖЕДНЕВНЫЙ ЧЕЛЛЕНДЖ'
        description='Пройди ежедневное задание и получи дополнительный XP!'
        action='Начать'
        asset='daily-trophy.png'
        onAction={() => onNavigate('MISSIONS')}
      />
      <UtilityCard
        kind='weekly'
        title='ЕЖЕНЕДЕЛЬНЫЙ ТЕСТ'
        description='Проверь свои знания и поднимись в рейтинге.'
        action='Пройти тест'
        asset='weekly-calendar.png'
        onAction={() => handleToast('Выбран еженедельный тест.')}
      />
      <UtilityCard
        kind='rewards'
        title='НАГРАДЫ И ДОСТИЖЕНИЯ'
        description='Копи XP, открывай бейджи и особые награды.'
        action='Смотреть награды'
        asset='rewards-gift.png'
        onAction={() => onNavigate('PROGRESS')}
      />
    </div>
  )
}
