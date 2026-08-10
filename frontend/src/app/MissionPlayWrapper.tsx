import { Component, type ErrorInfo, type ReactNode } from 'react'
import { useParams } from 'react-router-dom'
import { MissionPlayPage } from '../pages/mission-play'

interface ErrorBoundaryProps {
  children: ReactNode
  onExit: () => void
}

interface ErrorBoundaryState {
  hasError: boolean
}

/**
 * Ловит сбой отрисовки сценария.
 *
 * Без границы ошибок исключение в одном экране гасит всё приложение, и
 * пользователь остаётся на пустой странице без выхода в каталог.
 */
class MissionErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  state: ErrorBoundaryState = { hasError: false }

  static getDerivedStateFromError(): ErrorBoundaryState {
    return { hasError: true }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Сбой экрана прохождения:', error, errorInfo)
  }

  render() {
    if (!this.state.hasError) {
      return this.props.children
    }

    return (
      <main className='mission-play-layout mission-play-layout--message'>
        <section className='play-panel'>
          <h2>Не удалось показать сценарий</h2>
          {/* Текст ошибки не показываем: он ничего не говорит пользователю. */}
          <p>Попробуйте открыть миссию заново из каталога.</p>
          <button
            className='exit-mission-button'
            type='button'
            onClick={() => {
              this.setState({ hasError: false })
              this.props.onExit()
            }}
          >
            ← К списку миссий
          </button>
        </section>
      </main>
    )
  }
}

/** Достаёт id миссии из адреса и оборачивает экран границей ошибок. */
export function MissionPlayWrapper({
  onExit,
  onToast,
}: {
  onExit: () => void
  onToast?: (message: string) => void
}) {
  const { missionId } = useParams<{ missionId: string }>()

  if (!missionId) {
    return (
      <main className='mission-play-layout mission-play-layout--message'>
        <section className='play-panel'>
          <h2>Миссия не выбрана</h2>
          <button className='exit-mission-button' type='button' onClick={onExit}>
            ← К списку миссий
          </button>
        </section>
      </main>
    )
  }

  return (
    <MissionErrorBoundary onExit={onExit}>
      <MissionPlayPage missionId={missionId} onExit={onExit} onToast={onToast} />
    </MissionErrorBoundary>
  )
}
