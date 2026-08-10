/**
 * Корни статики. Файлы лежат в `public`, поэтому пути абсолютные и не зависят
 * от того, из какой папки собирается модуль.
 *
 * Картинки готовит `npm run process-assets` из папки `ui reference`.
 */

/** Иконки интерфейса: 128 px, показываются в размере 16–58 px. */
export const ICON_ROOT = '/assets/icons'

/** Иллюстрации карточек и аватары переписки: 192 px. */
export const ART_ROOT = '/assets/art'

/** Крупные картинки: логотип, персонажи ролей, герой главной страницы. */
export const SCENE_ROOT = '/assets/scene'

/** Иллюстрации отдельной страницы правил. */
export const RULES_ROOT = '/assets/rules'

/** Музыка и другие длинные звуковые дорожки. */
export const AUDIO_ROOT = '/assets/audio'
