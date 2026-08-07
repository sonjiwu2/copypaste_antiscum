import { PixelIcon } from '../../../shared/ui/PixelIcon'

export function HeroBanner({ onAction }: { onAction: () => void }) {
  return (
    <section className='hero-card pixel-card' aria-labelledby='hero-title'>
      <div className='hero-copy'>
        <h1 id='hero-title'>
          Прокачай навыки.
          <br />
          Защити себя.
        </h1>
        <p>
          Учись распознавать мошенников, защищай себя
          <br className='desktop-break' /> и совершай сделки с уверенностью.
        </p>
        <button className='button button--primary hero-button' type='button' onClick={onAction}>
          <span>Начать миссию</span>
          <PixelIcon name='chevron' size={20} />
        </button>
      </div>
      <div className='hero-art' aria-hidden='true'>
        <img
          src='/assets/anti-scam/hero-frames/frame-01.png'
          alt=''
          className='hero-frame-static pixel-art'
          draggable={false}
        />
      </div>
    </section>
  )
}
