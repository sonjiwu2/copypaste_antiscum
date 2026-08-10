import { useUserProgressStore } from '../../../entities/user-progress'
import { useToastStore } from '../../../features/toast'
import { useMissionPlay } from '../model/useMissionPlay'
import { ConversationPanel } from './ConversationPanel'
import { InsightsSidebar } from './InsightsSidebar'
import { MissionResultDialog } from './MissionResultDialog'
import { MissionSidebar } from './MissionSidebar'
import './mission-play.css'

export interface MissionPlayPageProps {
  missionId: string
  onExit: () => void
  onToast?: (message: string) => void
}

export function MissionPlayPage({ missionId, onExit, onToast }: MissionPlayPageProps) {
  const showToast = useToastStore((state) => state.showToast)
  const notify = onToast ?? showToast

  const completedCount = useUserProgressStore(
    (state) => Object.keys(state.completedMissions ?? {}).length
  )
  const achievementsCount = useUserProgressStore(
    (state) => (state.unlockedAchievements ?? []).length
  )
  const streakDays = useUserProgressStore((state) => state.streakDays ?? 1)

  const play = useMissionPlay(missionId, notify)

  if (play.error) {
    return (
      <PlayMessage
        title='Сценарий не запустился'
        text='Сервер не ответил на запрос старта. Проверьте, что бэкенд запущен.'
        onExit={onExit}
      />
    )
  }

  if (play.isLoading || !play.started) {
    return <PlayMessage title='Загружаем сценарий…' onExit={onExit} />
  }

  return (
    <main className='mission-play-layout'>
      <MissionSidebar
        mission={play.mission}
        step={play.step}
        totalSteps={play.totalSteps}
        progressPercent={play.progressPercent}
        objective={play.objective}
        riskGrade={play.riskGrade}
        riskMarkerPercent={play.riskMarkerPercent}
        completedCount={completedCount}
        achievementsCount={achievementsCount}
        streakDays={streakDays}
        onExit={onExit}
      />

      <ConversationPanel
        playerRole={play.playerRole}
        missionId={play.mission.id}
        messages={play.messages}
        choices={play.choices}
        disabled={play.completed || play.isSubmitting}
        onChoose={play.chooseResponse}
      />

      <InsightsSidebar
        events={play.events}
        signals={play.signals}
        elapsedText={play.elapsedText}
        xpEarned={play.xpEarned}
        score={play.score}
        onShowTip={notify}
      />

      {play.completed && (
        <MissionResultDialog
          outcome={play.outcome}
          xpEarned={play.xpEarned}
          score={play.score}
          riskGrade={play.riskGrade}
          onRestart={play.restartMission}
          onExit={onExit}
        />
      )}
    </main>
  )
}

/** Экран прохождения до появления попытки: загрузка или сбой запуска. */
function PlayMessage({
  title,
  text,
  onExit,
}: {
  title: string
  text?: string
  onExit: () => void
}) {
  return (
    <main className='mission-play-layout mission-play-layout--message'>
      <section className='play-panel'>
        <h2>{title}</h2>
        {text && <p>{text}</p>}
        <button className='exit-mission-button' type='button' onClick={onExit}>
          ← К списку миссий
        </button>
      </section>
    </main>
  )
}
