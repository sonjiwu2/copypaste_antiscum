import { PixelIcon } from '../../../shared/ui/PixelIcon'

/**
 * Состояния каталога до появления данных.
 *
 * Оба занимают то же место, что и сам каталог, поэтому при загрузке страница
 * не «прыгает».
 */

export function CatalogLoading() {
  return (
    <main className='missions-layout missions-layout--state'>
      <div className='catalog-state' role='status'>
        <PixelIcon name='in-progress' size={44} className='catalog-state__spinner' />
        <strong>Загружаем миссии</strong>
        <p>Ещё пара секунд — подбираем сценарии.</p>
      </div>
    </main>
  )
}

export function CatalogError({ onRetry }: { onRetry: () => void }) {
  return (
    <main className='missions-layout missions-layout--state'>
      <div className='catalog-state catalog-state--error'>
        <PixelIcon name='shield' size={44} />
        <strong>Не удалось загрузить миссии</strong>
        <p>Сервер не ответил. Проверьте соединение и попробуйте ещё раз.</p>
        <button type='button' onClick={onRetry}>
          Повторить
        </button>
      </div>
    </main>
  )
}
