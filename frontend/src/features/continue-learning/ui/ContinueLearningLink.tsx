import { Link } from 'react-router-dom'
import type { ContinueAction } from '@/entities/learning-path'
import { uiStyles } from '@/shared/ui-kit'
import { getContinueLearningLabel, getContinueLearningPath } from '../model/getContinueLearningPath'

interface ContinueLearningLinkProps {
  action: ContinueAction | null
  basePath?: string
  className?: string
  label?: string
}

export function ContinueLearningLink({
  action,
  basePath = '',
  className = uiStyles.primaryButton,
  label,
}: ContinueLearningLinkProps) {
  return (
    <Link className={className} to={getContinueLearningPath(action, basePath)}>
      {label ?? getContinueLearningLabel(action)}
    </Link>
  )
}
