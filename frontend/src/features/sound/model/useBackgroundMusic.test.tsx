import { act, cleanup, renderHook, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { DEFAULT_SETTINGS, useSettingsStore } from '../../../entities/settings'
import { useBackgroundMusic } from './useBackgroundMusic'

class AudioMock {
  src = ''
  loop = false
  preload = ''
  volume = 1
  paused = true
  play = vi.fn(async () => {
    this.paused = false
  })
  pause = vi.fn(() => {
    this.paused = true
  })
}

describe('useBackgroundMusic', () => {
  let audio: AudioMock

  beforeEach(() => {
    localStorage.clear()
    useSettingsStore.setState({ settings: DEFAULT_SETTINGS })
    audio = new AudioMock()
    vi.stubGlobal(
      'Audio',
      vi.fn(function AudioConstructor() {
        return audio
      })
    )
  })

  afterEach(() => {
    cleanup()
    vi.unstubAllGlobals()
  })

  it('зацикливает трек и применяет сохранённую громкость', async () => {
    renderHook(() => useBackgroundMusic())

    await waitFor(() => expect(audio.play).toHaveBeenCalled())
    expect(audio.loop).toBe(true)
    expect(audio.preload).toBe('auto')
    expect(audio.volume).toBeCloseTo(0.32)
  })

  it('останавливает музыку после включения выключателя', async () => {
    renderHook(() => useBackgroundMusic())
    await waitFor(() => expect(audio.play).toHaveBeenCalled())

    act(() => {
      useSettingsStore.getState().applySettings({ muteMenuMusic: true })
    })

    expect(audio.pause).toHaveBeenCalled()
    expect(audio.volume).toBe(0)
  })
})
