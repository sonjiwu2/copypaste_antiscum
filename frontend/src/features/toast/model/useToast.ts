import { useEffect } from 'react'
import { useToastStore } from './useToastStore'

export interface UseToastReturn {
  toast: string | null
  showToast: (message: string) => void
  dismiss: () => void
}

export function useToast(): UseToastReturn {
  const toast = useToastStore((state) => state.toast)
  const showToast = useToastStore((state) => state.showToast)
  const dismiss = useToastStore((state) => state.dismiss)

  useEffect(() => {
    return () => {
      const timeoutId = useToastStore.getState().timeoutId
      if (timeoutId) {
        window.clearTimeout(timeoutId)
        useToastStore.setState({ timeoutId: null })
      }
    }
  }, [])

  return { toast, showToast, dismiss }
}
