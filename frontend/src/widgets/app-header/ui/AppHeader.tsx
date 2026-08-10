import { useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { useLogoutMutation, type Account } from '@/entities/user'
import { Brand } from '@/shared/brand'
import { useIsPreview } from '@/shared/runtime-mode'
import styles from './AppHeader.module.scss'

const navItems = [
  ['/dashboard', 'Главная'],
  ['/lessons', 'Обучение'],
  ['/chats', 'Тренировки'],
  ['/progress', 'Прогресс'],
  ['/achievements', 'Достижения'],
] as const

interface AppHeaderProps {
  account: Account
  basePath?: string
}

export function AppHeader({ account, basePath = '' }: AppHeaderProps) {
  const location = useLocation()
  const navigate = useNavigate()
  const isPreview = useIsPreview()
  const [menuOpen, setMenuOpen] = useState(false)
  const [logout] = useLogoutMutation()

  const signOut = async () => {
    if (isPreview) {
      navigate(`${basePath}/login`)
      return
    }

    try {
      await logout().unwrap()
    } finally {
      navigate('/login')
    }
  }

  const isActive = (path: string) =>
    location.pathname === path ||
    (path.endsWith('/lessons') && location.pathname.startsWith(`${path}/`))

  return (
    <header className={styles.topbar}>
      <Brand />
      <nav className={styles.nav}>
        {account.accessRole === 'admin' && !basePath && (
          <Link className={isActive('/admin') ? styles.active : undefined} to="/admin">
            Админ-панель
          </Link>
        )}
        {navItems.map(([to, label]) => {
          const href = `${basePath}${to}`

          return (
            <Link key={to} className={isActive(href) ? styles.active : undefined} to={href}>
              {label}
            </Link>
          )
        })}
      </nav>
      <div className={styles.actions}>
        <span className={styles.streak} aria-label={`Серия ${account.streak.current} дня`}>
          🔥 <b>{account.streak.current}</b>
        </span>
        <button
          className={styles.avatar}
          type="button"
          onClick={() => setMenuOpen((open) => !open)}
        >
          {account.username.slice(0, 1).toUpperCase()}
        </button>
        {menuOpen && (
          <div className={styles.profileMenu}>
            <b>{account.username}</b>
            <span>{account.trainingRole === 'buyer' ? 'Покупатель' : 'Продавец'}</span>
            <button type="button" onClick={signOut}>
              Выйти
            </button>
          </div>
        )}
      </div>
    </header>
  )
}
