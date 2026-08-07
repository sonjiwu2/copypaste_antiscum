export type ChoiceTone = 'safe' | 'neutral' | 'risky'

export type Choice = {
  text: string
  tone: ChoiceTone
  risk: number
  xp: number
}

export type DemoStep = {
  seller: string
  objective: string
  signal: string
  choices: Choice[]
}

export type SavedProgress = {
  step: number
  risk: number
  xp: number
  choices: Array<number | string>
  startedAt: number
}

export type ChatMessage = {
  id: string
  side: 'seller' | 'buyer'
  text: string
  time: string
  tone?: ChoiceTone
}

export const PROVIDED = '/assets/anti-scam/provided'
export const TOTAL_STEPS = 5

export const DEMO_STEPS: DemoStep[] = [
  {
    seller: 'Здравствуйте! Товар в наличии, готов отправлять прямо сегодня.',
    objective: 'Уточните детали и проверьте, не пытается ли продавец торопить вас.',
    signal: 'Продавец создаёт искусственную торопливость в диалоге.',
    choices: [
      { text: 'Мне интересно, но сначала я хочу всё проверить.', tone: 'safe', risk: -1, xp: 8 },
      { text: 'Подтвердите свежие фото товара.', tone: 'neutral', risk: 0, xp: 5 },
      { text: 'Отлично! Покупаю прямо сейчас.', tone: 'risky', risk: 1, xp: 2 },
    ],
  },
  {
    seller: 'Товар абсолютно новый. Переведите оплату по ссылке, чтобы не платить комиссию.',
    objective: 'Сохраняйте оплату только внутри официальной площадки.',
    signal: 'Запрос на оплату или уход по сторонней ссылке.',
    choices: [
      { text: 'Я оплачиваю исключительно через платформу.', tone: 'safe', risk: -1, xp: 10 },
      { text: 'Почему вы не хотите использовать встроенную доставку?', tone: 'neutral', risk: 0, xp: 6 },
      { text: 'Хорошо, скидывайте ссылку.', tone: 'risky', risk: 2, xp: 1 },
    ],
  },
  {
    seller: 'Такая низкая цена только сегодня. У меня ещё два покупателя, определяйтесь за 5 минут.',
    objective: 'Не поддавайтесь давлению и проверьте реальную рыночную цену.',
    signal: 'Спешка и подозрительно низкая цена.',
    choices: [
      { text: 'Я не люблю спешку. Сначала сравню цены.', tone: 'safe', risk: -1, xp: 10 },
      { text: 'Какова средняя рыночная цена этого товара?', tone: 'neutral', risk: 0, xp: 6 },
      { text: 'Забронируйте за мной — я переведу задаток.', tone: 'risky', risk: 2, xp: 1 },
    ],
  },
  {
    seller: 'Чтобы подтвердить, что вы настоящий покупатель, диктуйте СМС-код с телефона.',
    objective: 'Защищайте коды из СМС и личные данные своего профиля.',
    signal: 'Запрос конфиденциального кода или пароля.',
    choices: [
      {
        text: 'СМС-коды передавать нельзя. Отправляю жалобу в поддержку.',
        tone: 'safe',
        risk: -2,
        xp: 12,
      },
      { text: 'Зачем вам код от моего аккаунта?', tone: 'neutral', risk: 0, xp: 5 },
      { text: 'Код из СМС: 482913.', tone: 'risky', risk: 3, xp: 0 },
    ],
  },
  {
    seller: 'Последний шанс. Оплачивайте сейчас или я отдаю товар другому.',
    objective: 'Примите безопасное итоговое решение.',
    signal: 'Повторные угрозы и попытки манипуляции.',
    choices: [
      {
        text: 'Прекратить диалог, заблокировать продавца и пожаловаться.',
        tone: 'safe',
        risk: -2,
        xp: 15,
      },
      { text: 'Выйти из диалога и всё хорошо обдумать.', tone: 'neutral', risk: -1, xp: 8 },
      { text: 'Срочно перевести деньги, чтобы не упустить сделку.', tone: 'risky', risk: 3, xp: 0 },
    ],
  },
]

export const PRACTICAL_TIPS = [
  ['Проверка продавца', 'Проверяйте дату регистрации аккаунта и реальные отзывы.'],
  ['Безопасные платежи', 'Используйте только встроенную безопасную сделку площадки.'],
  ['Анализ цены', 'Сравнивайте предложение с рыночными ценами.'],
  ['Жалоба и защита', 'Блокируйте профиль при подозрительных просьбах.'],
]

export function loadProgress(missionId: string): SavedProgress {
  try {
    const raw = localStorage.getItem(`antiscam.demo.${missionId}`)
    if (raw) {
      const parsed = JSON.parse(raw) as SavedProgress
      if (parsed && typeof parsed.step === 'number' && Array.isArray(parsed.choices)) {
        return parsed
      }
    }
  } catch {
    /* demo storage is optional */
  }
  return { step: 0, risk: 2, xp: 0, choices: [], startedAt: Date.now() }
}

export function clockTime(offset: number): string {
  const minutes = 32 + offset
  return `10:${String(minutes).padStart(2, '0')}`
}

export function buildMessages(progress: SavedProgress): ChatMessage[] {
  const messages: ChatMessage[] = []
  const maxStep = Math.max(0, Math.min(progress.step, DEMO_STEPS.length - 1))
  for (let index = 0; index <= maxStep; index += 1) {
    const stepData = DEMO_STEPS[index] ?? DEMO_STEPS[0]
    messages.push({
      id: `seller-${index}`,
      side: 'seller',
      text: stepData.seller,
      time: clockTime(index * 2),
    })
    const choiceIndex = progress.choices[index]
    if (choiceIndex !== undefined) {
      const choice = typeof choiceIndex === 'number'
        ? stepData.choices[choiceIndex]
        : stepData.choices.find((c) => c.text === String(choiceIndex))
      if (choice) {
        messages.push({
          id: `buyer-${index}`,
          side: 'buyer',
          text: choice.text,
          time: clockTime(index * 2 + 1),
          tone: choice.tone,
        })
      }
    }
  }
  return messages
}
