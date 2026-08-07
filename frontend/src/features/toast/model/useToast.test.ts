import { describe, beforeEach, afterEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useToast } from './useToast'
import { useToastStore } from './useToastStore'

describe('useToast custom hook', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    useToastStore.setState({ toast: null, timeoutId: null })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('should initialize with null toast', () => {
    const { result } = renderHook(() => useToast())

    expect(result.current.toast).toBeNull()
  })

  it('should show toast message when showToast is called', () => {
    const { result } = renderHook(() => useToast())

    act(() => {
      result.current.showToast('Test notification')
    })

    expect(result.current.toast).toBe('Test notification')
  })

  it('should automatically clear toast after 2600ms', () => {
    const { result } = renderHook(() => useToast())

    act(() => {
      result.current.showToast('Self-destructing toast')
    })

    expect(result.current.toast).toBe('Self-destructing toast')

    act(() => {
      vi.advanceTimersByTime(2600)
    })

    expect(result.current.toast).toBeNull()
  })

  it('should reset timer when showToast is called multiple times', () => {
    const { result } = renderHook(() => useToast())

    act(() => {
      result.current.showToast('First message')
    })

    act(() => {
      vi.advanceTimersByTime(2000)
    })

    // Send second toast before first one expires
    act(() => {
      result.current.showToast('Second message')
    })

    expect(result.current.toast).toBe('Second message')

    // Advance 2000ms (total 4000ms from start, but only 2000ms since second message)
    act(() => {
      vi.advanceTimersByTime(2000)
    })

    // Toast should still be visible because timer was reset
    expect(result.current.toast).toBe('Second message')

    // Advance remaining 600ms
    act(() => {
      vi.advanceTimersByTime(600)
    })

    expect(result.current.toast).toBeNull()
  })

  it('should dismiss toast manually when dismiss is called', () => {
    const { result } = renderHook(() => useToast())

    act(() => {
      result.current.showToast('Dismiss me')
    })

    expect(result.current.toast).toBe('Dismiss me')

    act(() => {
      result.current.dismiss()
    })

    expect(result.current.toast).toBeNull()
  })

  it('should clear active timeout on unmount', () => {
    const clearTimeoutSpy = vi.spyOn(window, 'clearTimeout')
    const { result, unmount } = renderHook(() => useToast())

    act(() => {
      result.current.showToast('Unmounting soon')
    })

    unmount()

    expect(clearTimeoutSpy).toHaveBeenCalled()
    clearTimeoutSpy.mockRestore()
  })
})
