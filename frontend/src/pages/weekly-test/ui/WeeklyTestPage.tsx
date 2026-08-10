import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useUserProgressStore } from '../../../entities/user-progress'
import {
  useCheckWeeklyTestAnswerMutation,
  useSubmitWeeklyTestMutation,
  useWeeklyTestQuery,
  type WeeklyTest,
  type WeeklyTestAnswer,
} from '../../../shared/api'
import { ART_ROOT } from '../../../shared/config/assets'
import './weekly-test.css'

type TestDraft = {
  questionIndex: number
  answers: Record<string, number>
  lives: number
}

export function WeeklyTestPage({ onBack }: { onBack: () => void }) {
  const testQuery = useWeeklyTestQuery()
  const recordCompletion = useUserProgressStore((state) => state.recordWeeklyTestCompletion)
  const submitMutation = useSubmitWeeklyTestMutation()
  const test = submitMutation.data ?? testQuery.data

  // Завершённый тест: черновик больше не нужен, результат уходит в прогресс.
  useEffect(() => {
    if (!test?.result) return

    removeDraft(test.testId)
    recordCompletion(test.weekStart, test.result.score, test.result.earnedXp)
  }, [recordCompletion, test])

  const submitAnswers = useCallback(
    (finalAnswers: Record<string, number>) => {
      if (!test) {
        return Promise.resolve()
      }

      return submitMutation.mutateAsync({
        testId: test.testId,
        answers: answersForRequest(test, finalAnswers),
      })
    },
    [submitMutation, test]
  )

  if (testQuery.isPending) {
    return <WeeklyTestLoading onBack={onBack} />
  }

  if (testQuery.isError || !test) {
    return (
      <WeeklyTestError
        error={testQuery.error?.message}
        onBack={onBack}
        onRetry={testQuery.refetch}
      />
    )
  }

  if (test.completed && test.result) {
    return <WeeklyTestResultView test={test} onBack={onBack} />
  }

  return (
    <WeeklyExam
      // Ключ пересоздаёт экзамен при смене теста: черновик читается заново, а
      // ответы прошлой недели не протекают в новую попытку.
      key={test.testId}
      test={test}
      submitting={submitMutation.isPending}
      submitFailed={submitMutation.isError}
      onSubmit={submitAnswers}
      onBack={onBack}
    />
  )
}

interface WeeklyExamProps {
  test: WeeklyTest
  submitting: boolean
  submitFailed: boolean
  onSubmit: (answers: Record<string, number>) => Promise<unknown>
  onBack: () => void
}

/**
 * Сам экзамен: вопрос, варианты, жизни и таймер.
 *
 * Компонент монтируется под ключом теста, поэтому сохранённый черновик
 * поднимается прямо в начальном состоянии — отдельный эффект восстановления
 * не нужен и лишнего перерисовывания на старте не возникает.
 */
function WeeklyExam({ test, submitting, submitFailed, onSubmit, onBack }: WeeklyExamProps) {
  const checkMutation = useCheckWeeklyTestAnswerMutation()

  const [draft] = useState(() => loadDraft(test))
  const [questionIndex, setQuestionIndex] = useState(draft.questionIndex)
  const [answers, setAnswers] = useState<Record<string, number>>(draft.answers)
  const [lives, setLives] = useState(draft.lives)
  const [helpOpen, setHelpOpen] = useState(false)
  const [now, setNow] = useState(() => Date.now())
  const finishingRef = useRef(false)

  useEffect(() => {
    saveDraft(test.testId, { questionIndex, answers, lives })
  }, [answers, lives, questionIndex, test.testId])

  useEffect(() => {
    const interval = window.setInterval(() => setNow(Date.now()), 250)

    return () => window.clearInterval(interval)
  }, [])

  const finishExam = useCallback(
    async (finalAnswers: Record<string, number>) => {
      // Экзамен завершается один раз: время и последний ответ могут сработать
      // одновременно, а второй запрос сервер уже не примет.
      if (finishingRef.current) return

      finishingRef.current = true
      try {
        await onSubmit(finalAnswers)
      } catch {
        finishingRef.current = false
      }
    },
    [onSubmit]
  )

  const remainingSeconds = Math.max(0, Math.ceil((Date.parse(test.expiresAt) - now) / 1000))

  useEffect(() => {
    if (remainingSeconds === 0) {
      void finishExam(answers)
    }
  }, [answers, finishExam, remainingSeconds])

  const safeQuestionIndex = Math.min(questionIndex, test.questions.length - 1)
  const question = test.questions[safeQuestionIndex]
  const selectedIndex = answers[question.id]
  const isLast = safeQuestionIndex === test.questions.length - 1
  const busy = checkMutation.isPending || submitting

  const handleNext = async () => {
    if (selectedIndex === undefined || busy) return

    try {
      const checked = await checkMutation.mutateAsync({
        testId: test.testId,
        questionId: question.id,
        optionIndex: selectedIndex,
      })
      const nextAnswers = { ...answers, [question.id]: selectedIndex }
      const nextLives = checked.correct ? lives : Math.max(0, lives - 1)
      setAnswers(nextAnswers)
      setLives(nextLives)

      if (isLast || nextLives === 0) {
        await finishExam(nextAnswers)
        return
      }

      setQuestionIndex((current) => Math.min(current + 1, test.questions.length - 1))
    } catch {
      // Текст ошибки выводится под вариантами; выбранный ответ можно отправить повторно.
    }
  }

  return (
    <main className='weekly-test-page'>
      <section className='weekly-exam' aria-labelledby='weekly-exam-title'>
        <header className='weekly-exam__header'>
          <button
            type='button'
            className='weekly-square-button'
            onClick={onBack}
            aria-label='На главную'
          >
            ←
          </button>
          <div>
            <h1 id='weekly-exam-title'>ЕЖЕНЕДЕЛЬНЫЙ ТЕСТ</h1>
            <span className={`weekly-test-source weekly-test-source--${test.source}`}>
              {sourceLabel(test.source)}
            </span>
          </div>
          <button
            type='button'
            className='weekly-square-button'
            onClick={() => setHelpOpen((value) => !value)}
            aria-label='Правила теста'
            aria-expanded={helpOpen}
          >
            ?
          </button>
        </header>

        {helpOpen && (
          <aside className='weekly-help'>
            Ответьте на {test.rules.questionCount} ситуаций за{' '}
            {Math.round(test.rules.durationSeconds / 60)} минут. Третья ошибка завершает экзамен.
            Для награды нужно минимум {test.rules.passingCorrect} правильных ответов.
          </aside>
        )}

        <div
          className='weekly-progress'
          aria-label={`Пройден вопрос ${safeQuestionIndex + 1} из ${test.rules.questionCount}`}
        >
          <span
            style={{ width: `${((safeQuestionIndex + 1) / test.rules.questionCount) * 100}%` }}
          />
        </div>
        <div className='weekly-exam__meta'>
          <span>
            Вопрос {safeQuestionIndex + 1} из {test.rules.questionCount}
          </span>
          <span
            className={
              remainingSeconds <= 60 ? 'weekly-timer weekly-timer--danger' : 'weekly-timer'
            }
          >
            ◷ Осталось {formatClock(remainingSeconds)}
          </span>
        </div>

        <article className='weekly-question-card'>
          <span className='weekly-question-card__number'>Вопрос {safeQuestionIndex + 1}</span>
          <h2>{question.prompt}</h2>

          <div className='weekly-options' role='radiogroup' aria-label='Варианты ответа'>
            {question.options.map((option, optionIndex) => (
              <button
                key={`${question.id}-${optionIndex}`}
                type='button'
                role='radio'
                aria-checked={selectedIndex === optionIndex}
                className={
                  selectedIndex === optionIndex
                    ? 'weekly-option weekly-option--selected'
                    : 'weekly-option'
                }
                disabled={busy}
                onClick={() =>
                  setAnswers((current) => ({ ...current, [question.id]: optionIndex }))
                }
              >
                <span className='weekly-option__radio' aria-hidden='true' />
                <span>{option}</span>
              </button>
            ))}
          </div>

          {(checkMutation.isError || submitFailed) && (
            <p className='weekly-question__error'>
              Не удалось проверить ответ. Проверьте соединение и попробуйте ещё раз.
            </p>
          )}
        </article>

        <section className='weekly-lives' aria-label={`Осталось жизней: ${lives}`}>
          <div className='weekly-lives__icon' aria-hidden='true'>
            ♥
          </div>
          <div>
            <span>Жизни</span>
            <div className='weekly-lives__hearts' aria-hidden='true'>
              {Array.from({ length: test.rules.maxLives }, (_, index) => (
                <span className={index < lives ? '' : 'weekly-heart--lost'} key={index}>
                  ♥
                </span>
              ))}
            </div>
          </div>
          <p>Потеряешь {test.rules.maxLives} жизни — тест завершится.</p>
        </section>

        <button
          type='button'
          className='weekly-next-button'
          disabled={selectedIndex === undefined || busy || remainingSeconds === 0}
          onClick={() => void handleNext()}
        >
          {busy ? 'ПРОВЕРЯЕМ...' : isLast ? 'ЗАВЕРШИТЬ' : 'ДАЛЕЕ'}
        </button>
      </section>
    </main>
  )
}

function WeeklyTestLoading({ onBack }: { onBack: () => void }) {
  return (
    <main className='weekly-test-page'>
      <section className='weekly-test-state'>
        <img
          src={`${ART_ROOT}/weekly-calendar.png`}
          alt=''
          className='pixel-art weekly-test-state__loading'
        />
        <h1>ИИ СОСТАВЛЯЕТ ЭКЗАМЕН</h1>
        <p>Groq готовит 20 новых антискам-ситуаций. Первая генерация может занять до минуты.</p>
        <button type='button' onClick={onBack}>
          НА ГЛАВНУЮ
        </button>
      </section>
    </main>
  )
}

function WeeklyTestError({
  error,
  onBack,
  onRetry,
}: {
  error?: string
  onBack: () => void
  onRetry: () => unknown
}) {
  return (
    <main className='weekly-test-page'>
      <section className='weekly-test-state'>
        <img src={`${ART_ROOT}/weekly-calendar.png`} alt='' className='pixel-art' />
        <h1>ТЕСТ НЕ ЗАГРУЗИЛСЯ</h1>
        <p>{error ?? 'Попробуйте ещё раз.'}</p>
        <div className='weekly-test-state__actions'>
          <button type='button' onClick={() => void onRetry()}>
            ПОВТОРИТЬ
          </button>
          <button type='button' onClick={onBack}>
            НА ГЛАВНУЮ
          </button>
        </div>
      </section>
    </main>
  )
}

function WeeklyTestResultView({ test, onBack }: { test: WeeklyTest; onBack: () => void }) {
  const result = test.result!
  const questions = useMemo(
    () => new Map(test.questions.map((question) => [question.id, question])),
    [test.questions]
  )

  return (
    <main className='weekly-test-page'>
      <div className='weekly-result-layout'>
        <section
          className={`weekly-result-card ${result.passed ? 'weekly-result-card--passed' : 'weekly-result-card--failed'}`}
        >
          <img src={`${ART_ROOT}/weekly-calendar.png`} alt='' className='pixel-art' />
          <span className='weekly-result-card__eyebrow'>
            {result.passed ? 'ЭКЗАМЕН СДАН' : 'ЭКЗАМЕН НЕ СДАН'}
          </span>
          <div className='weekly-result-card__score'>
            {result.score}
            <small>/100</small>
          </div>
          <h1>
            {result.passed
              ? 'ЗАЩИТА НА ВЫСОТЕ!'
              : result.timedOut
                ? 'ВРЕМЯ ВЫШЛО'
                : 'ПОПРОБУЙ ЕЩЁ РАЗ'}
          </h1>
          <p>
            Верно: {result.correct} из {result.total} · Осталось жизней: {result.livesLeft}
          </p>
          <strong>{result.passed ? `+${result.earnedXp} XP` : 'НАГРАДА НЕ НАЧИСЛЕНА'}</strong>
          <span>{sourceResultLabel(test.source)}</span>
          {test.nextAvailableAt && <WeeklyCountdown target={test.nextAvailableAt} />}
          <button type='button' onClick={onBack}>
            НА ГЛАВНУЮ
          </button>
        </section>

        <section className='weekly-review' aria-labelledby='weekly-review-title'>
          <h2 id='weekly-review-title'>РАЗБОР ОТВЕТОВ</h2>
          {result.review.map((review, index) => {
            const question = questions.get(review.questionId)
            if (!question) return null
            const selected =
              review.selectedIndex >= 0 ? question.options[review.selectedIndex] : 'Нет ответа'

            return (
              <article
                className={`weekly-review-card ${review.correct ? 'weekly-review-card--correct' : 'weekly-review-card--wrong'}`}
                key={review.questionId}
              >
                <h3>
                  {review.correct ? '✓' : '×'} {index + 1}. {question.prompt}
                </h3>
                <p>
                  <b>Ваш ответ:</b> {selected}
                </p>
                {!review.correct && (
                  <p>
                    <b>Правильный ответ:</b> {question.options[review.correctIndex]}
                  </p>
                )}
                <small>{review.explanation}</small>
              </article>
            )
          })}
        </section>
      </div>
    </main>
  )
}

function WeeklyCountdown({ target }: { target: string }) {
  const [now, setNow] = useState(() => Date.now())
  useEffect(() => {
    const interval = window.setInterval(() => setNow(Date.now()), 1000)
    return () => window.clearInterval(interval)
  }, [])

  const seconds = Math.max(0, Math.ceil((Date.parse(target) - now) / 1000))
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const rest = seconds % 60

  return (
    <div className='weekly-countdown'>
      <small>СЛЕДУЮЩИЙ ТЕСТ ЧЕРЕЗ</small>
      <b>
        {days}д {twoDigits(hours)}:{twoDigits(minutes)}:{twoDigits(rest)}
      </b>
    </div>
  )
}

function answersForRequest(test: WeeklyTest, answers: Record<string, number>): WeeklyTestAnswer[] {
  return test.questions.flatMap((question) => {
    const optionIndex = answers[question.id]
    return optionIndex === undefined ? [] : [{ questionId: question.id, optionIndex }]
  })
}

function loadDraft(test: WeeklyTest): TestDraft {
  const empty = { questionIndex: 0, answers: {}, lives: test.rules.maxLives }
  try {
    const raw = localStorage.getItem(draftKey(test.testId))
    if (!raw) return empty
    const parsed = JSON.parse(raw) as Partial<TestDraft>
    const restoredLives = Number(parsed.lives)
    return {
      questionIndex: Math.max(
        0,
        Math.min(Number(parsed.questionIndex) || 0, test.questions.length - 1)
      ),
      answers: parsed.answers && typeof parsed.answers === 'object' ? parsed.answers : {},
      lives: Number.isFinite(restoredLives)
        ? Math.max(0, Math.min(restoredLives, test.rules.maxLives))
        : test.rules.maxLives,
    }
  } catch {
    return empty
  }
}

function saveDraft(testId: string, draft: TestDraft) {
  try {
    localStorage.setItem(draftKey(testId), JSON.stringify(draft))
  } catch {
    // Без localStorage экзамен работает до перезагрузки страницы.
  }
}

function removeDraft(testId: string) {
  try {
    localStorage.removeItem(draftKey(testId))
  } catch {
    // Сохранённого черновика может не быть.
  }
}

const draftKey = (testId: string) => `antiscam.weekly_test.${testId}`
const twoDigits = (value: number) => String(value).padStart(2, '0')
const formatClock = (seconds: number) =>
  `${twoDigits(Math.floor(seconds / 60))}:${twoDigits(seconds % 60)}`

function sourceLabel(source: WeeklyTest['source']): string {
  if (source === 'groq') return '✦ СОЗДАНО GROQ AI'
  if (source === 'grok') return '✦ СОЗДАНО GROK'
  return 'РЕЗЕРВНЫЙ ТЕСТ'
}

function sourceResultLabel(source: WeeklyTest['source']): string {
  if (source === 'groq') return '20 ситуаций созданы Groq AI'
  if (source === 'grok') return '20 ситуаций созданы Grok'
  return 'Использован проверенный резервный экзамен'
}
