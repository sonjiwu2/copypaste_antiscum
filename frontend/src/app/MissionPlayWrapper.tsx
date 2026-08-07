import React, { Component, ReactNode } from 'react'
import { useParams } from 'react-router-dom'
import { MissionPlayPage } from '../pages/mission-play'

interface ErrorBoundaryProps {
  children: ReactNode
  onExit: () => void
}

interface ErrorBoundaryState {
  hasError: boolean
  error?: Error
}

class MissionErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('MissionPlay Error:', error, errorInfo)
  }

  render() {
    if (this.state.hasError) {
      return (
        <div style={{ padding: '40px', textAlign: 'center', background: '#f8fbff', minHeight: '80vh' }}>
          <h2>Ошибка при загрузке сценария</h2>
          <p style={{ color: '#666', margin: '15px 0' }}>
            Произошёл сбой при отображении сценария: {this.state.error?.message}
          </p>
          <button
            type='button'
            style={{
              padding: '10px 20px',
              fontSize: '16px',
              cursor: 'pointer',
              background: '#1687e6',
              color: '#fff',
              border: 'none',
              borderRadius: '6px',
            }}
            onClick={() => {
              this.setState({ hasError: false })
              this.props.onExit()
            }}
          >
            ← Вернуться к каталогу миссий
          </button>
        </div>
      )
    }

    return this.props.children
  }
}

export function MissionPlayWrapper({
  onExit,
  onToast,
}: {
  onExit: () => void
  onToast?: (msg: string) => void
}) {
  const { missionId } = useParams<{ missionId: string }>()
  return (
    <MissionErrorBoundary onExit={onExit}>
      <MissionPlayPage
        missionId={missionId ?? 'buyer-fake-delivery'}
        onExit={onExit}
        onToast={onToast}
      />
    </MissionErrorBoundary>
  )
}
