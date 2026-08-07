import { PixelIcon } from '../../../shared/ui/PixelIcon'

export function Toast({ message, onDismiss }: { message: string; onDismiss: () => void }) {
  return (
    <div className='toast' role='status' aria-live='polite'>
      <PixelIcon name='shield' size={23} />
      <span>{message}</span>
      <button type='button' onClick={onDismiss} aria-label='Dismiss notification'>
        ×
      </button>
    </div>
  )
}
