import { describe, expect, it } from 'vitest'
import type { Attempt } from '../../../shared/api'
import {
  buildEvents,
  buildMessages,
  riskGradeOf,
  riskMarkerPercent,
  toneOf,
} from './missionPlayHelpers'

function attemptWith(overrides: Partial<Attempt> = {}): Attempt {
  return {
    attemptId: 'att-1',
    scenario: {
      id: 'buyer-fake-delivery',
      version: 1,
      slug: 'buyer-fake-delivery',
      role: 'buyer',
      title: 'Ссылка на доставку',
    },
    status: 'in_progress',
    score: 100,
    currentNodeId: 'channel-decision',
    revealedNodes: [],
    decisions: [],
    startedAt: '2026-08-08T10:00:00Z',
    updatedAt: '2026-08-08T10:00:00Z',
    ...overrides,
  }
}

describe('buildMessages', () => {
  it('ставит в ленту закреплённую реплику игрока, а не подпись кнопки', () => {
    const attempt = attemptWith({
      revealedNodes: [
        { id: 'greeting', type: 'message', sender: 'seller', text: 'Товар ещё доступен.' },
        { id: 'channel-decision', type: 'decision', prompt: 'Что вы сделаете?' },
        { id: 'pressure', type: 'message', sender: 'seller', text: 'Оплатите по ссылке.' },
      ],
      decisions: [
        {
          nodeId: 'channel-decision',
          choiceId: 'move-to-messenger',
          label: 'Перейти в мессенджер',
          playerReply: 'Хорошо, напишите мне в мессенджер.',
          consequence: {
            severity: 'dangerous',
            title: 'Защита площадки потеряна',
            explanation: 'Вне площадки переписка не поможет вернуть деньги.',
          },
        },
      ],
    })

    const messages = buildMessages(attempt)

    expect(messages.map((message) => message.text)).toEqual([
      'Товар ещё доступен.',
      'Хорошо, напишите мне в мессенджер.',
      'Оплатите по ссылке.',
    ])
    expect(messages[1].own).toBe(true)
    expect(messages[1].role).toBe('buyer')
    expect(messages[1].tone).toBe('risky')
  })

  it('не показывает реплику за узел решения, пока выбор не сделан', () => {
    const attempt = attemptWith({
      revealedNodes: [
        { id: 'greeting', type: 'message', sender: 'seller', text: 'Здравствуйте.' },
        { id: 'channel-decision', type: 'decision', prompt: 'Что вы сделаете?' },
      ],
    })

    expect(buildMessages(attempt)).toHaveLength(1)
  })

  it('не пускает системные факты и финал в переписку', () => {
    const attempt = attemptWith({
      status: 'completed',
      currentNodeId: 'safe-ending',
      revealedNodes: [
        { id: 'offer', type: 'message', sender: 'seller', text: 'Оплатите по ссылке.' },
        {
          id: 'sms',
          type: 'message',
          sender: 'system',
          text: 'Приходит SMS: «Код 481926 для входа в интернет-банк».',
        },
        {
          id: 'safe-ending',
          type: 'terminal',
          outcome: {
            type: 'safe',
            title: 'Сделка прошла безопасно',
            explanation: 'Вы остались в защищённом канале.',
          },
        },
      ],
    })

    const messages = buildMessages(attempt)

    expect(messages).toHaveLength(1)
    expect(messages[0].text).toBe('Оплатите по ссылке.')
  })

  it('в сценарии продавца реплика игрока идёт от продавца и на его стороне', () => {
    const attempt = attemptWith({
      scenario: {
        id: 'seller-gpu-return-swap',
        version: 1,
        slug: 'seller-gpu-return-swap',
        role: 'seller',
        title: 'Подмена видеокарты',
      },
      revealedNodes: [
        { id: 'claim', type: 'message', sender: 'buyer', text: 'Видеокарта не работает.' },
        { id: 'return-decision', type: 'decision', prompt: 'Что вы сделаете?' },
      ],
      decisions: [
        {
          nodeId: 'return-decision',
          choiceId: 'request-video',
          label: 'Запросить видео распаковки',
          playerReply: 'Пришлите видео неисправности и фото серийного номера.',
          consequence: {
            severity: 'safe',
            title: 'Доказательство зафиксировано',
            explanation: 'Видео распаковки подтверждает комплектность.',
          },
        },
      ],
    })

    const [opponent, player] = buildMessages(attempt)

    expect(opponent.role).toBe('buyer')
    expect(opponent.own).toBe(false)
    expect(player.role).toBe('seller')
    expect(player.own).toBe(true)
  })
})

describe('buildEvents', () => {
  it('собирает факты обстановки в отдельную хронику', () => {
    const attempt = attemptWith({
      revealedNodes: [
        { id: 'offer', type: 'message', sender: 'seller', text: 'Оплатите по ссылке.' },
        { id: 'no-order', type: 'message', sender: 'system', text: 'В приложении заказа нет.' },
      ],
    })

    expect(buildEvents(attempt)).toEqual([
      { id: 'no-order', text: 'В приложении заказа нет.' },
    ])
  })
})

describe('шкала риска', () => {
  it('переводит результат в уровень риска', () => {
    expect(riskGradeOf(100)).toBe('low')
    expect(riskGradeOf(80)).toBe('low')
    expect(riskGradeOf(60)).toBe('medium')
    expect(riskGradeOf(20)).toBe('high')
  })

  it('переводит результат в положение стрелки в процентах', () => {
    expect(riskMarkerPercent(100)).toBe(0)
    expect(riskMarkerPercent(50)).toBe(50)
    // Стрелка не должна выходить за правый край шкалы.
    expect(riskMarkerPercent(0)).toBe(96)
    expect(riskMarkerPercent(-40)).toBe(96)
  })

  it('переводит степень опасности в тон реплики', () => {
    expect(toneOf('safe')).toBe('safe')
    expect(toneOf('warning')).toBe('neutral')
    expect(toneOf('dangerous')).toBe('risky')
  })
})
