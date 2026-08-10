import { Glyph } from '../../../shared/ui/Glyph'
import { SCENE_ROOT } from '../../../shared/config/assets'

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
          <Glyph name='chevron' size={20} />
        </button>
      </div>
      <div className='hero-art' aria-hidden='true'>
        <img
          src={`${SCENE_ROOT}/hero-character.png`}
          alt=''
          className='hero-frame-static pixel-art'
          draggable={false}
        />
      </div>
    </section>
  )
}
