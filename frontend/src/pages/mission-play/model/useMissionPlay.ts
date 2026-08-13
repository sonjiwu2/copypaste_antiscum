import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { scenariosToMissions, type Mission } from '../../../entities/mission'
import { earnedXpFor, useUserProgressStore } from '../../../entities/user-progress'
import {
  isAttemptGone,
  newIdempotencyKey,
  useAttemptQuery,
  useScenarios,
  useStartAttemptMutation,
  useSubmitChoiceMutation,
} from '../../../shared/api'
import type { Attempt, AttemptNode, Scenario } from '../../../shared/api'
import { useToastStore } from '../../../features/toast'
import {
  buildEvents,
  buildMessages,
  isAttemptVersionOutdated,
  readStoredAttemptId,
  riskGradeOf,
  riskMarkerPercent,
  writeStoredAttemptId,
  type ChatMessage,
  type WorldEvent,
} from '../lib/missionPlayHelpers'

const FALLBACK_OBJECTIVE = 'Прочитайте переписку и оцените, чем рискует сделка.'

export function useMissionPlay(missionId: string, onToast?: (message: string) => void) {
  const showToast = useToastStore((state) => state.showToast)
  const triggerToast = useCallback(
    (message: string) => (onToast ?? showToast)(message),
    [onToast, showToast]
  )

  const { scenarios } = useScenarios()
  const { mission, scenario } = useMissionScenario(scenarios, missionId)

  const { attempt, isLoading, error, isSubmitting, submitChoice, restart } = useAttemptLifecycle(
    missionId,
    scenario?.version,
    triggerToast
  )

  const elapsedText = useElapsedTime(attempt?.startedAt)

  const completed = attempt?.status === 'completed'
  const score = attempt?.score ?? 100

  const messages = useMemo<ChatMessage[]>(
    () => (attempt ? buildMessages(attempt) : []),
    [attempt]
  )

  // Факты обстановки идут отдельной хроникой: система не участник переписки.
  const events = useMemo<WorldEvent[]>(() => (attempt ? buildEvents(attempt) : []), [attempt])

  const currentDecision = useMemo(
    () => currentDecisionNode(attempt),
    [attempt]
  )

  // Варианты отдаются без оценки: подсветка безопасного ответа заранее
  // лишила бы сценарий смысла.
  const choices = useMemo(
    () => (currentDecision?.choices ?? []).map((choice) => ({ id: choice.id, text: choice.label })),
    [currentDecision]
  )

  // Признаки мошенничества раскрываются только по уже сделанным выборам.
  const signals = useMemo(
    () => (attempt?.decisions ?? []).map((decision) => decision.consequence),
    [attempt]
  )

  const chooseResponse = useCallback(
    async (choiceIndex: number) => {
      const choice = choices[choiceIndex]
      if (!attempt || !currentDecision || !choice) {
        return
      }

      const transition = await submitChoice(attempt.attemptId, currentDecision.id, choice.id)
      if (!transition) {
        return
      }

      triggerToast(`${transition.consequence.title}: ${transition.consequence.explanation}`)

      if (transition.status === 'completed') {
        useUserProgressStore
          .getState()
          .recordMissionCompletion(attempt.scenario.id, transition.score, mission.xp)
        triggerToast(`Сценарий завершён. Результат: ${transition.score} из 100.`)
      }
    },
    [attempt, choices, currentDecision, mission.xp, submitChoice, triggerToast]
  )

  // Во время прохождения пользователь стоит на очередном решении, а после
  // завершения счётчик показывает, сколькими решениями сценарий закончился.
  const totalSteps = scenario?.maxDecisions ?? 1
  const decisionsMade = attempt?.decisions.length ?? 0
  const step = Math.min(completed ? decisionsMade : decisionsMade + 1, totalSteps)

  return {
    mission,
    // Роль берётся из попытки: каталог может быть ещё не загружен, а перепутать
    // стороны сделки нельзя — от этого зависит вся раскладка переписки.
    playerRole: attempt?.scenario.role ?? mission.role,
    started: Boolean(attempt),
    completed,
    outcome: terminalOutcome(attempt),
    messages,
    events,
    objective: currentDecision?.prompt ?? FALLBACK_OBJECTIVE,
    choices,
    signals,
    step,
    totalSteps,
    progressPercent: completed ? 100 : Math.round((decisionsMade / totalSteps) * 100),
    score,
    riskGrade: riskGradeOf(score),
    riskMarkerPercent: riskMarkerPercent(score),
    xpEarned: earnedXpFor(mission.xp, score),
    elapsedText,
    chooseResponse,
    restartMission: restart,
    isLoading,
    isSubmitting,
    error,
  }
}

/** Сценарий попытки и его карточка в каталоге миссий. */
function useMissionScenario(scenarios: Scenario[], missionId: string) {
  return useMemo(() => {
    const missions = scenariosToMissions(scenarios)

    return {
      mission: missions.find((item) => item.id === missionId) ?? placeholderMission(missionId),
      scenario: scenarios.find((item) => item.id === missionId),
    }
  }, [scenarios, missionId])
}

/**
 * Пока каталог не загружен, экран показывает заготовку карточки.
 * Прохождение при этом уже работает: оно опирается только на попытку.
 */
function placeholderMission(missionId: string): Mission {
  return {
    id: missionId,
    title: 'Сценарий загружается…',
    description: '',
    difficulty: 'medium',
    role: 'buyer',
    category: 'payment',
    xp: 15,
    section: 'featured',
    status: 'in-progress',
    tone: 'blue',
    icon: 'role-buyer',
  }
}

/**
 * Ведёт попытку сценария: восстанавливает начатую, начинает новую и применяет
 * выборы. Состояние живёт в кэше запроса, поэтому второй копии его нет.
 */
function useAttemptLifecycle(
  missionId: string,
  currentScenarioVersion: number | undefined,
  triggerToast: (message: string) => void
) {
  const [attemptId, setAttemptId] = useState<string | null>(() => readStoredAttemptId(missionId))
  const attemptQuery = useAttemptQuery(attemptId ?? undefined)
  const submitMutation = useSubmitChoiceMutation()

  const startMutation = useStartAttemptMutation({
    onSuccess: (attempt) => {
      writeStoredAttemptId(missionId, attempt.attemptId)
      setAttemptId(attempt.attemptId)
    },
    onError: () => triggerToast('Не удалось начать сценарий. Проверьте соединение с сервером.'),
  })

  // Помнит сценарий, для которого старт уже запрошен. Без этого неудачный
  // запрос повторялся бы на каждый повторный рендер.
  const startRequestedFor = useRef<string | null>(null)

  // Незавершённая попытка закреплена за старой версией JSON и намеренно не
  // меняется на сервере. Когда каталог сообщает о новой версии, начинаем новую
  // попытку: иначе пользователь продолжит видеть исправленный сценарий без
  // добавленных реплик до самого финала старого прохождения.
  useEffect(() => {
    const restored = attemptQuery.data
    if (
      !isAttemptVersionOutdated(restored, currentScenarioVersion) ||
      startMutation.isPending ||
      startRequestedFor.current === missionId
    ) {
      return
    }

    startRequestedFor.current = missionId
    writeStoredAttemptId(missionId, null)
    setAttemptId(null)
    startMutation.mutate(missionId, {
      onSuccess: () => triggerToast('Сценарий обновлён и начат заново с полным диалогом.'),
    })
  }, [attemptQuery.data, currentScenarioVersion, missionId, startMutation, triggerToast])

  // Сохранённая попытка могла быть удалена на сервере или принадлежать другому
  // профилю: такая попытка не восстанавливается, сценарий начинается заново.
  const activeAttemptId = isAttemptGone(attemptQuery.error) ? null : attemptId

  useEffect(() => {
    if (activeAttemptId || startRequestedFor.current === missionId) {
      return
    }

    startRequestedFor.current = missionId
    writeStoredAttemptId(missionId, null)
    startMutation.mutate(missionId)
  }, [activeAttemptId, missionId, startMutation])

  const refetchAttempt = attemptQuery.refetch

  const submitChoice = useCallback(
    async (currentAttemptId: string, nodeId: string, choiceId: string) => {
      try {
        return await submitMutation.mutateAsync({
          attemptId: currentAttemptId,
          nodeId,
          choiceId,
          idempotencyKey: newIdempotencyKey(),
        })
      } catch {
        // Сервер выбор не принял. Перечитываем попытку, чтобы экран показывал
        // то состояние, которое сервер считает текущим.
        await refetchAttempt()
        triggerToast('Выбор не сохранён. Попробуйте ещё раз.')

        return null
      }
    },
    [refetchAttempt, submitMutation, triggerToast]
  )

  const restart = useCallback(() => {
    writeStoredAttemptId(missionId, null)
    // Старая попытка больше не должна кормить экран и таймер:
    // до ответа сервера показываем запуск, затем берём новый `startedAt`.
    setAttemptId(null)
    startMutation.mutate(missionId, {
      onSuccess: () => triggerToast('Сценарий начат заново.'),
    })
  }, [missionId, startMutation, triggerToast])

  return {
    attempt: activeAttemptId ? attemptQuery.data : undefined,
    isLoading: attemptQuery.isLoading || startMutation.isPending,
    error: startMutation.error,
    isSubmitting: submitMutation.isPending,
    submitChoice,
    restart,
  }
}

function currentDecisionNode(attempt?: Attempt): AttemptNode | undefined {
  return attempt?.revealedNodes.find(
    (node) => node.id === attempt.currentNodeId && node.type === 'decision'
  )
}

function terminalOutcome(attempt?: Attempt) {
  return attempt?.revealedNodes.find((node) => node.id === attempt.currentNodeId)?.outcome
}

/** Время с начала попытки. Отсчёт ведётся от серверной отметки старта. */
function useElapsedTime(startedAt?: string): string {
  const [now, setNow] = useState(() => Date.now())

  useEffect(() => {
    const timer = window.setInterval(() => setNow(Date.now()), 1000)

    return () => window.clearInterval(timer)
  }, [])

  if (!startedAt) {
    return '00:00'
  }

  const seconds = Math.max(0, Math.floor((now - new Date(startedAt).getTime()) / 1000))

  return `${String(Math.floor(seconds / 60)).padStart(2, '0')}:${String(seconds % 60).padStart(2, '0')}`
}
