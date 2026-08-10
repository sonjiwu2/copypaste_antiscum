import { useState, useRef, useEffect, useCallback } from 'react'

export function useHeaderProfileMenu() {
  const [profileOpen, setProfileOpen] = useState(false)
  const profileRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const closeProfile = (event: MouseEvent) => {
      if (!profileRef.current?.contains(event.target as Node)) setProfileOpen(false)
    }
    const closeWithEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setProfileOpen(false)
    }
    document.addEventListener('mousedown', closeProfile)
    document.addEventListener('keydown', closeWithEscape)
    return () => {
      document.removeEventListener('mousedown', closeProfile)
      document.removeEventListener('keydown', closeWithEscape)
    }
  }, [])

  const toggleProfile = useCallback(() => {
    setProfileOpen((value) => !value)
  }, [])

  const closeMenu = useCallback(() => {
    setProfileOpen(false)
  }, [])

  return {
    profileOpen,
    profileRef,
    toggleProfile,
    closeMenu,
  }
}
