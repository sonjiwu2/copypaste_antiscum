import { useEffect, useRef } from 'react'
import { ROLE_LABELS, type MissionRole } from '../../../entities/mission'
import { ART_ROOT } from '../../../shared/config/assets'
import type { ChatMessage } from '../lib/missionPlayHelpers'

/** Буквы вариантов ответа. Нейтральные метки не намекают на правильный выбор. */
const CHOICE_LETTERS = ['А', 'Б', 'В', 'Г', 'Д']

/** «Диалог с …»: русскому заголовку нужен творительный падеж. */
const OPPONENT_LABELS: Record<'buyer' | 'seller', string> = {
  buyer: 'покупателем',
  seller: 'продавцом',
}

export interface ConversationChoice {
  id: string
  text: string
}

export interface ConversationPanelProps {
  playerRole: MissionRole
  missionId: string
  messages: ChatMessage[]
  choices: ConversationChoice[]
  disabled: boolean
  onChoose: (index: number) => void
}

export function ConversationPanel({
  playerRole,
  missionId,
  messages,
  choices,
  disabled,
  onChoose,
}: ConversationPanelProps) {
  const opponentRole = playerRole === 'seller' ? 'buyer' : 'seller'
  const conversationEndRef = useRef<HTMLDivElement>(null)

  // После выбора сервер раскрывает сразу несколько новых узлов. Прокручиваем
  // ленту к ним, иначе на высоком диалоге пользователь видит прежний экран и
  // замечает только изменения в правой колонке.
  useEffect(() => {
    conversationEndRef.current?.scrollIntoView({ behavior: 'smooth', block: 'end' })
  }, [messages.length])

  return (
    <section className='conversation-panel' aria-label='Переписка по сценарию'>
      <div className='conversation-header'>
        <strong>ВЫ — {ROLE_LABELS[playerRole].toUpperCase()}</strong>
        <b>Диалог с {OPPONENT_LABELS[opponentRole]}</b>
        <small>#{missionId.slice(0, 12)}</small>
      </div>

      {/* В ленте только слова сторон сделки: факты обстановки живут в хронике. */}
      <div className='conversation-scroll' aria-live='polite'>
        {messages.map((message) => (
          <ChatRow key={message.id} message={message} />
        ))}
        <div ref={conversationEndRef} aria-hidden='true' />
      </div>

      {choices.length > 0 && (
        <div className='response-area'>
          <div className='response-heading'>
            <span />
            <h2>ВЫБЕРИТЕ ОТВЕТ</h2>
            <span />
          </div>
          {/*
            Все варианты оформлены одинаково. Подсветка безопасного ответа до
            выбора лишила бы сценарий смысла, поэтому отличаются только буквы.
          */}
          <div className='response-grid'>
            {choices.map((choice, index) => (
              <button
                className='response-choice'
                type='button'
                key={choice.id}
                onClick={() => onChoose(index)}
                disabled={disabled}
              >
                <span className='response-choice__letter' aria-hidden='true'>
                  {CHOICE_LETTERS[index] ?? index + 1}
                </span>
                <strong>{choice.text}</strong>
              </button>
            ))}
          </div>
        </div>
      )}
    </section>
  )
}

/**
 * Реплика одной из сторон сделки.
 *
 * Сторона ленты выбирается по тому, чья это реплика, а не по роли: иначе в
 * сценарии продавца слова самого игрока оказались бы на стороне собеседника.
 * Аватар при этом остаётся привязан к роли говорящего.
 */
function ChatRow({ message }: { message: ChatMessage }) {
  return (
    <div className={`chat-row chat-row--${message.own ? 'own' : 'opponent'}`}>
      <span className='chat-avatar'>
        <img src={`${ART_ROOT}/${message.role}-avatar.png`} alt='' className='pixel-art' />
      </span>
      <div className={`chat-bubble${message.tone ? ` chat-bubble--${message.tone}` : ''}`}>
        <p>{message.text}</p>
        <small>{message.time}</small>
      </div>
    </div>
  )
}
