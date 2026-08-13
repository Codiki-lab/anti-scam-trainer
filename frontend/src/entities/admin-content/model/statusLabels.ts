import type { ContentStatus } from './types'

export const statusLabels: Record<ContentStatus, string> = {
  draft: 'Черновик',
  published: 'Опубликовано',
  archived: 'В архиве',
}
