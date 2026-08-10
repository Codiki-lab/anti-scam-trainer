import type { Topic } from '../model/types'

export function getCompletedLevelCount(topic: Topic) {
  return topic.levels.filter((level) => level.stars > 0).length
}
