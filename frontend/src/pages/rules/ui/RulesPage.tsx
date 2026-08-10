import { ICON_ROOT, RULES_ROOT, SCENE_ROOT } from '../../../shared/config/assets'
import './rules.css'

interface RuleStep {
  icon: string
  title: string
  text: string
}

const STEPS: RuleStep[] = [
  {
    icon: 'rules-missions-map.png',
    title: '1. МИССИИ',
    text: 'Выбирайте сценарии покупателя или продавца и проходите реальные ситуации общения.',
  },
  {
    icon: 'rules-dialogues.png',
    title: '2. ДИАЛОГИ',
    text: 'Изучайте сообщения собеседника и замечайте подозрительные формулировки.',
  },
  {
    icon: 'rules-choice-signpost.png',
    title: '3. ВЫБОР',
    text: 'Принимайте решение, как действовать дальше, не раскрывая личные данные.',
  },
  {
    icon: 'rules-review.png',
    title: '4. РАЗБОР',
    text: 'После завершения миссии получайте объяснение ошибок, советы и награды.',
  },
]

const SAFE_DEAL_RULES = [
  'Общайтесь и оплачивайте только внутри платформы.',
  'Не переходите по подозрительным ссылкам из сообщений.',
  'Не сообщайте коды, пароли и данные карты.',
  'Проверяйте профиль, рейтинг и историю продавца или покупателя.',
  'Если собеседник торопит или уводит в другой мессенджер — это повод насторожиться.',
]

export function RulesPage({ onNavigate }: { onNavigate: (command: string) => void }) {
  return (
    <main className='rules-page'>
      <div className='rules-page__inner'>
        <header className='rules-intro'>
          <div className='rules-intro__title'>
            <img src={`${RULES_ROOT}/rules-clipboard.png`} alt='' className='pixel-art' />
            <h1>Правила сервиса</h1>
          </div>
          <p>Как пользоваться тренажёром и совершать безопасные сделки на платформе.</p>
        </header>

        <section className='rules-card rules-flow' aria-labelledby='rules-flow-title'>
          <h2 id='rules-flow-title'>КАК РАБОТАЕТ СЕРВИС</h2>
          <div className='rules-flow__grid'>
            {STEPS.map((step) => (
              <article className='rules-step' key={step.title}>
                <img
                  src={`${RULES_ROOT}/${step.icon}`}
                  alt=''
                  className='rules-step__image pixel-art'
                />
                <h3>{step.title}</h3>
                <p>{step.text}</p>
              </article>
            ))}
          </div>
        </section>

        <div className='rules-details'>
          <section className='rules-card rules-purpose' aria-labelledby='rules-purpose-title'>
            <h2 id='rules-purpose-title'>ЗАЧЕМ ЭТО НУЖНО</h2>
            <div className='rules-purpose__copy'>
              <img
                src={`${RULES_ROOT}/rules-lightbulb.png`}
                alt=''
                className='rules-purpose__bulb pixel-art'
              />
              <p>
                Тренажёр помогает распознавать схемы обмана, безопасно общаться на платформе и
                принимать более осторожные решения при покупке и продаже.
              </p>
            </div>
            <p className='rules-purpose__tip'>
              <img src={`${SCENE_ROOT}/brand-shield.png`} alt='' className='pixel-art' />
              Чем больше практики, тем легче замечать рискованные ситуации.
            </p>
          </section>

          <section className='rules-card rules-safety' aria-labelledby='rules-safety-title'>
            <h2 id='rules-safety-title'>ПРАВИЛА БЕЗОПАСНОЙ СДЕЛКИ</h2>
            <ul>
              {SAFE_DEAL_RULES.map((rule) => (
                <li key={rule}>{rule}</li>
              ))}
            </ul>
            <div className='rules-safety__helper' aria-hidden='true'>
              <img
                src={`${RULES_ROOT}/rules-alert.png`}
                alt=''
                className='rules-safety__alert pixel-art'
              />
              <img
                src={`${ICON_ROOT}/profile-avatar.png`}
                alt=''
                className='rules-safety__avatar pixel-art'
              />
            </div>
          </section>
        </div>

        <button className='rules-cta' type='button' onClick={() => onNavigate('MISSIONS')}>
          К СПИСКУ МИССИЙ
          <span aria-hidden='true'>→</span>
        </button>
      </div>
    </main>
  )
}
