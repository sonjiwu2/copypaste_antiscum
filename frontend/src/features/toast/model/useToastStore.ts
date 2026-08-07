import { create } from 'zustand'

export interface ToastStatesType {
  toast: string | null
  timeoutId: number | null
}

export interface ToastActionsType {
  showToast: (message: string) => void
  dismiss: () => void
}

export interface ToastStore extends ToastStatesType {
  actions: ToastActionsType
  showToast: (message: string) => void
  dismiss: () => void
}

const defaultToastState: ToastStatesType = {
  toast: null,
  timeoutId: null,
}

export const useToastStore = create<ToastStore>((set, get) => ({
  ...defaultToastState,

  showToast: (message: string) => {
    const currentTimeout = get().timeoutId
    if (currentTimeout) {
      window.clearTimeout(currentTimeout)
    }

    const newTimeout = window.setTimeout(() => {
      set({ toast: null, timeoutId: null })
    }, 2600)

    set({ toast: message, timeoutId: newTimeout })
  },

  dismiss: () => {
    const currentTimeout = get().timeoutId
    if (currentTimeout) {
      window.clearTimeout(currentTimeout)
    }
    set({ toast: null, timeoutId: null })
  },

  actions: {
    showToast: (message: string) => get().showToast(message),
    dismiss: () => get().dismiss(),
  },
}))
