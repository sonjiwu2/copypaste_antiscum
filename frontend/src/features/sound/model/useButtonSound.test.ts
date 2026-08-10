import { renderHook } from '@testing-library/react'
import { useButtonSound } from './useButtonSound'

describe('useButtonSound custom hook', () => {
  let mockGainNode: {
    gain: {
      setValueAtTime: ReturnType<typeof vi.fn>
      exponentialRampToValueAtTime: ReturnType<typeof vi.fn>
    }
    connect: ReturnType<typeof vi.fn>
  }
  let mockOscillatorNode: {
    type: string
    frequency: {
      setValueAtTime: ReturnType<typeof vi.fn>
    }
    connect: ReturnType<typeof vi.fn>
    start: ReturnType<typeof vi.fn>
    stop: ReturnType<typeof vi.fn>
  }

  let mockResume: ReturnType<typeof vi.fn>
  let mockClose: ReturnType<typeof vi.fn>
  let mockState: string

  beforeEach(() => {
    mockState = 'running'
    mockResume = vi.fn().mockResolvedValue(undefined)
    mockClose = vi.fn().mockResolvedValue(undefined)

    mockGainNode = {
      gain: {
        setValueAtTime: vi.fn(),
        exponentialRampToValueAtTime: vi.fn(),
      },
      connect: vi.fn(),
    }

    mockOscillatorNode = {
      type: 'square',
      frequency: {
        setValueAtTime: vi.fn(),
      },
      connect: vi.fn(),
      start: vi.fn(),
      stop: vi.fn(),
    }

    class FakeAudioContext {
      get state() {
        return mockState
      }
      currentTime = 0
      destination = {}
      resume = mockResume
      createGain = vi.fn().mockReturnValue(mockGainNode)
      createOscillator = vi.fn().mockReturnValue(mockOscillatorNode)
      close = mockClose
    }

    vi.stubGlobal('AudioContext', FakeAudioContext)
  })

  afterEach(() => {
    vi.restoreAllMocks()
    vi.unstubAllGlobals()
  })

  it('should attach click listener to document on mount and remove on unmount', () => {
    const addEventListenerSpy = vi.spyOn(document, 'addEventListener')
    const removeEventListenerSpy = vi.spyOn(document, 'removeEventListener')

    const { unmount } = renderHook(() => useButtonSound())

    expect(addEventListenerSpy).toHaveBeenCalledWith('click', expect.any(Function))

    unmount()

    expect(removeEventListenerSpy).toHaveBeenCalledWith('click', expect.any(Function))
  })

  it('should play sound when an active button is clicked', () => {
    renderHook(() => useButtonSound())

    const button = document.createElement('button')
    document.body.appendChild(button)

    button.click()

    expect(mockGainNode.connect).toHaveBeenCalled()
    expect(mockOscillatorNode.start).toHaveBeenCalledWith(0)
    expect(mockOscillatorNode.stop).toHaveBeenCalledWith(0.095)

    document.body.removeChild(button)
  })

  it('should play sound when a child element inside an active button is clicked', () => {
    renderHook(() => useButtonSound())

    const button = document.createElement('button')
    const span = document.createElement('span')
    button.appendChild(span)
    document.body.appendChild(button)

    span.click()

    expect(mockOscillatorNode.start).toHaveBeenCalled()

    document.body.removeChild(button)
  })

  it('should NOT play sound when a non-button element is clicked', () => {
    renderHook(() => useButtonSound())

    const div = document.createElement('div')
    document.body.appendChild(div)

    div.click()

    expect(mockGainNode.connect).not.toHaveBeenCalled()

    document.body.removeChild(div)
  })

  it('should NOT play sound when a disabled button is clicked', () => {
    renderHook(() => useButtonSound())

    const button = document.createElement('button')
    button.setAttribute('disabled', 'true')
    document.body.appendChild(button)

    button.click()

    expect(mockGainNode.connect).not.toHaveBeenCalled()

    document.body.removeChild(button)
  })

  it('should resume suspended AudioContext when button is clicked', () => {
    mockState = 'suspended'

    renderHook(() => useButtonSound())

    const button = document.createElement('button')
    document.body.appendChild(button)

    button.click()

    expect(mockResume).toHaveBeenCalled()

    document.body.removeChild(button)
  })

  it('should close AudioContext on unmount if it was instantiated', () => {
    const { unmount } = renderHook(() => useButtonSound())

    const button = document.createElement('button')
    document.body.appendChild(button)
    button.click()

    unmount()

    expect(mockClose).toHaveBeenCalled()

    document.body.removeChild(button)
  })
})
