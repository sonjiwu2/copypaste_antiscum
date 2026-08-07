import { useState, useEffect, useMemo, useCallback, useRef } from 'react'
import { MISSIONS, scenariosToMissions } from '../../../entities/mission'
import { useUserProgressStore } from '../../../entities/user-progress'
import { useScenarios, useStartAttemptMutation, useSubmitChoiceMutation } from '../../../shared/api'
import type { AttemptNode, Consequence } from '../../../shared/api/types'
import { useToastStore } from '../../../features/toast'
import {
  loadProgress,
  buildMessages,
  DEMO_STEPS,
  TOTAL_STEPS,
  type SavedProgress,
  type ChatMessage,
} from '../lib/missionPlayHelpers'

export function useMissionPlay(missionId: string, onToast?: (message: string) => void) {
  const triggerToast = useCallback(
    (msg: string) => {
      if (onToast) {
        onToast(msg)
      } else {
        useToastStore.getState().showToast(msg)
      }
    },
    [onToast]
  )

  const { scenarios } = useScenarios()

  const mission = useMemo(() => {
    if (scenarios && scenarios.length > 0) {
      const allMissions = scenariosToMissions(scenarios)
      const found = allMissions.find((item) => item.id === missionId)
      if (found) return found
    }
    return MISSIONS.find((item) => item.id === missionId) ?? MISSIONS[0]
  }, [scenarios, missionId])

  // Состояние бэкенда
  const [attemptId, setAttemptId] = useState<string | null>(null)
  const [revealedNodes, setRevealedNodes] = useState<AttemptNode[]>([])
  const [currentNodeId, setCurrentNodeId] = useState<string>('')
  const [score, setScore] = useState<number>(100)
  const [status, setStatus] = useState<'in_progress' | 'completed'>('in_progress')
  const [lastConsequence, setLastConsequence] = useState<Consequence | null>(null)
  const [progress, setProgress] = useState<SavedProgress>(() => loadProgress(mission.id))

  const startMutation = useStartAttemptMutation()
  const submitMutation = useSubmitChoiceMutation()
  const startedRef = useRef(false)

  // Стартуем попытку на бэкенде при выборе сценария
  useEffect(() => {
    if (!startedRef.current && missionId) {
      startedRef.current = true
      startMutation.mutate(missionId, {
        onSuccess: (data) => {
          if (data && Array.isArray(data.revealedNodes)) {
            setAttemptId(data.attemptId)
            setRevealedNodes(data.revealedNodes)
            setCurrentNodeId(data.currentNodeId ?? '')
            setScore(data.score ?? 100)
            setStatus(data.status ?? 'in_progress')
          }
        },
        onError: () => {
          // Игнорируем сетевые ошибки, продолжая в демо-режиме
        },
      })
    }
  }, [missionId])

  // Сообщения чата из раскрытых узлов бэкенда (или fallback)
  const messages: ChatMessage[] = useMemo(() => {
    if (revealedNodes.length > 0) {
      const list: ChatMessage[] = []
      revealedNodes.forEach((node, index) => {
        if (node.text) {
          list.push({
            id: node.id ?? `node-${index}`,
            side: (node.sender as 'seller' | 'buyer') ?? (index % 2 === 0 ? 'seller' : 'buyer'),
            text: node.text,
            time: `10:${String(30 + index * 2).padStart(2, '0')}`,
          })
        }
      })
      return list
    }
    return buildMessages(progress)
  }, [revealedNodes, progress])

  // Текущий узел принятия решения с бэкенда
  const currentDecisionNode = useMemo(() => {
    return (
      revealedNodes.find((node) => node.id === currentNodeId && node.type === 'decision') ??
      revealedNodes.slice().reverse().find((node) => node.type === 'decision')
    )
  }, [revealedNodes, currentNodeId])

  // Реальные варианты ответа с бэкенда
  const apiChoices = useMemo(() => {
    if (currentDecisionNode?.choices && currentDecisionNode.choices.length > 0) {
      return currentDecisionNode.choices.map((c, i) => ({
        id: c.id,
        text: c.label,
        tone: (i === 0 ? 'safe' : i === 1 ? 'neutral' : 'risky') as 'safe' | 'neutral' | 'risky',
      }))
    }
    return null
  }, [currentDecisionNode])

  // Обработчик выбора ответа
  const chooseResponse = useCallback(
    async (choiceIdOrIndex: string | number) => {
      // 1. Бэкенд режим
      if (attemptId && currentDecisionNode && apiChoices) {
        let choiceObj = apiChoices[0]
        if (typeof choiceIdOrIndex === 'number') {
          choiceObj = apiChoices[choiceIdOrIndex] ?? apiChoices[0]
        } else {
          choiceObj = apiChoices.find((c) => c.id === choiceIdOrIndex) ?? apiChoices[0]
        }

        if (!choiceObj) return

        // Показываем сообщение пользователя в чате мгновенно
        const userNode: AttemptNode = {
          id: `user-choice-${Date.now()}`,
          type: 'message',
          sender: 'buyer',
          text: choiceObj.text,
        }

        setRevealedNodes((prev) => [...prev, userNode])
        setProgress((prev) => ({
          ...prev,
          step: prev.step + 1,
          choices: [...prev.choices, typeof choiceIdOrIndex === 'number' ? choiceIdOrIndex : 0],
        }))

        try {
          const transition = await submitMutation.mutateAsync({
            attemptId,
            nodeId: currentDecisionNode.id,
            choiceId: choiceObj.id,
          })

          if (transition && Array.isArray(transition.revealedNodes)) {
            setRevealedNodes((prev) => [...prev, ...transition.revealedNodes])
            setCurrentNodeId(transition.currentNodeId)
            setScore(transition.score)
            setStatus(transition.status)
            setLastConsequence(transition.consequence)

            if (transition.consequence) {
              triggerToast(`${transition.consequence.title}: ${transition.consequence.explanation}`)
            }

            if (transition.status === 'completed') {
              const userStore = useUserProgressStore.getState()
              userStore.recordMissionCompletion(mission.id, transition.score, mission.xp)
              triggerToast(`Сценарий завершён! Финальный счёт: ${transition.score}/100`)
            }
          }
          return
        } catch {
          // Игнорируем сетевые ошибки
        }
      }

      // 2. Демо-режим (fallback)
      const idx = typeof choiceIdOrIndex === 'number' ? choiceIdOrIndex : 0
      const stepIndex = Math.max(0, Math.min(progress.step, TOTAL_STEPS - 1))
      const stepObj = DEMO_STEPS[stepIndex] ?? DEMO_STEPS[0]
      const choice = stepObj.choices[idx] ?? stepObj.choices[0]
      const next: SavedProgress = {
        ...progress,
        step: progress.step + 1,
        risk: Math.max(0, Math.min(8, progress.risk + choice.risk)),
        xp: progress.xp + choice.xp,
        choices: [...progress.choices, idx],
      }
      setProgress(next)

      if (next.step >= TOTAL_STEPS) {
        setStatus('completed')
        const userStore = useUserProgressStore.getState()
        userStore.recordMissionCompletion(mission.id, 90, mission.xp)
        triggerToast(`${mission.title} завершён!`)
      }
    },
    [attemptId, currentDecisionNode, apiChoices, progress, submitMutation, triggerToast, mission]
  )

  const [tipsOpen, setTipsOpen] = useState(false)
  const [elapsed, setElapsed] = useState(0)

  useEffect(() => {
    const updateElapsed = () =>
      setElapsed(Math.max(0, Math.floor((Date.now() - progress.startedAt) / 1000)))
    updateElapsed()
    const timer = window.setInterval(updateElapsed, 1000)
    return () => window.clearInterval(timer)
  }, [progress.startedAt])

  const restartMission = useCallback(() => {
    startedRef.current = false
    setRevealedNodes([])
    setAttemptId(null)
    setStatus('in_progress')
    setScore(100)
    setProgress({ step: 0, risk: 2, xp: 0, choices: [], startedAt: Date.now() })
    startMutation.mutate(missionId, {
      onSuccess: (data) => {
        if (data && Array.isArray(data.revealedNodes)) {
          setAttemptId(data.attemptId)
          setRevealedNodes(data.revealedNodes)
          setCurrentNodeId(data.currentNodeId)
          setScore(data.score)
          setStatus(data.status)
        }
      },
    })
    triggerToast('Сценарий перезапущен.')
  }, [missionId, startMutation, triggerToast])

  const elapsedText = `${String(Math.floor(elapsed / 60)).padStart(2, '0')}:${String(elapsed % 60).padStart(2, '0')}`

  const currentStep = useMemo(() => {
    if (apiChoices && apiChoices.length > 0) {
      return {
        objective: currentDecisionNode?.prompt ?? 'Проанализируйте сообщение продавца.',
        choices: apiChoices,
      }
    }
    const stepIndex = Math.max(0, Math.min(progress.step, TOTAL_STEPS - 1))
    return DEMO_STEPS[stepIndex] ?? DEMO_STEPS[0]
  }, [apiChoices, currentDecisionNode, progress.step])

  const completed = status === 'completed' || progress.step >= TOTAL_STEPS
  const progressPercent = completed ? 100 : Math.min(100, Math.round(((progress.step + 1) / TOTAL_STEPS) * 100))
  const accuracy = score
  const riskLabel = score >= 80 ? 'LOW RISK' : score >= 50 ? 'MEDIUM RISK' : 'HIGH RISK'

  return {
    mission,
    attemptId,
    progress,
    completed,
    tipsOpen,
    setTipsOpen,
    messages,
    currentStep,
    progressPercent,
    accuracy,
    riskLabel,
    elapsedText,
    chooseResponse,
    restartMission,
    lastConsequence,
    isSubmitting: submitMutation.isPending,
  }
}
