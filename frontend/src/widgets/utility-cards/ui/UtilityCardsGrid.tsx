import { useSettingsStore } from '../../../entities/settings'
import { currentLocalDateKey, useUserProgressStore } from '../../../entities/user-progress'
import { useToastStore } from '../../../features/toast'
import { UtilityCard, type UtilityCardProps } from './UtilityCard'

type CardConfig = Omit<UtilityCardProps, 'onAction'> & {
  /** Куда ведёт кнопка. Строка — раздел приложения, иначе показывается подсказка. */
  target: string | { toast: string }
}

const CARDS: CardConfig[] = [
  {
    kind: 'tip',
    title: 'СОВЕТ ДНЯ',
    description:
      'Никогда не передавай смс-коды и пароли никому, даже если они представились поддержкой.',
    action: 'Подробнее',
    art: 'tip-lock.png',
    target: 'RULES',
  },
  {
    kind: 'sos',
    title: 'АНТИСКАМ SOS',
    description:
      'Уже попался? Получи пошаговый план: останови перевод, защити аккаунты и сохрани доказательства.',
    action: 'В РАЗРАБОТКЕ',
    art: 'shield.png',
    artCollection: 'icons',
    pending: true,
    target: { toast: 'Раздел «Антискам SOS» пока в разработке.' },
  },
  {
    kind: 'weekly',
    title: 'ЕЖЕНЕДЕЛЬНЫЙ ТЕСТ',
    description: 'Проверь свои знания и поднимись в рейтинге.',
    action: 'Пройти тест',
    art: 'weekly-calendar.png',
    target: 'WEEKLY_TEST',
  },
  {
    kind: 'rewards',
    title: 'НАГРАДЫ',
    description: 'Копи XP, открывай бейджи и особые награды.',
    action: 'Смотреть награды',
    art: 'rewards-gift.png',
    target: 'REWARDS',
  },
]

export function UtilityCardsGrid({
  onNavigate,
  onToast,
}: {
  onNavigate: (label: string) => void
  onToast?: (message: string) => void
}) {
  const showToast = useToastStore((state) => state.showToast)
  const showDailyTip = useSettingsStore((state) => state.settings.showDailyTip)
  const lastDailyRewardDate = useUserProgressStore((state) => state.lastDailyRewardDate)
  const rewardClaimed = lastDailyRewardDate === currentLocalDateKey()
  const notify = onToast ?? showToast

  const cards = showDailyTip ? CARDS : CARDS.filter((card) => card.kind !== 'tip')

  return (
    <div className='utility-grid'>
      {cards.map(({ target, ...card }) => (
        <UtilityCard
          key={card.kind}
          {...card}
          description={
            card.kind === 'rewards' && rewardClaimed
              ? 'Награда на сегодня уже собрана. Возвращайся завтра!'
              : card.description
          }
          action={card.kind === 'rewards' && rewardClaimed ? 'Получено сегодня' : card.action}
          claimed={card.kind === 'rewards' && rewardClaimed}
          onAction={() => (typeof target === 'string' ? onNavigate(target) : notify(target.toast))}
        />
      ))}
    </div>
  )
}
