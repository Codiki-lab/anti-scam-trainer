import { CaretDown, ChartLineUp, Medal, SignOut } from '@phosphor-icons/react'
import { useEffect, useRef, useState } from 'react'
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
  const menuRef = useRef<HTMLDivElement>(null)
  const [logout] = useLogoutMutation()

  useEffect(() => {
    if (!menuOpen) return

    const closeOnOutsideClick = (event: PointerEvent) => {
      if (!menuRef.current?.contains(event.target as Node)) setMenuOpen(false)
    }
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setMenuOpen(false)
    }

    document.addEventListener('pointerdown', closeOnOutsideClick)
    document.addEventListener('keydown', closeOnEscape)
    return () => {
      document.removeEventListener('pointerdown', closeOnOutsideClick)
      document.removeEventListener('keydown', closeOnEscape)
    }
  }, [menuOpen])

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
      <div className={styles.actions} ref={menuRef}>
        <span className={styles.streak} aria-label={`Серия ${account.streak.current} дня`}>
          🔥 <b>{account.streak.current}</b>
        </span>
        <button
          className={styles.avatar}
          type="button"
          aria-expanded={menuOpen}
          aria-haspopup="menu"
          aria-label="Открыть меню профиля"
          onClick={() => setMenuOpen((open) => !open)}
        >
          <span>{account.username.slice(0, 1).toUpperCase()}</span>
          <CaretDown size={14} weight="bold" aria-hidden="true" />
        </button>
        {menuOpen && (
          <div className={styles.profileMenu} role="menu">
            <div className={styles.profileMenuHeader}>
              <span>{account.username.slice(0, 1).toUpperCase()}</span>
              <div>
                <b>{account.username}</b>
                <small>{account.trainingRole === 'buyer' ? 'Покупатель' : 'Продавец'}</small>
              </div>
            </div>
            <div className={styles.profileMenuLinks}>
              <Link
                role="menuitem"
                to={`${basePath}/achievements`}
                onClick={() => setMenuOpen(false)}
              >
                <Medal size={19} weight="duotone" aria-hidden="true" />
                <span>
                  <b>Достижения</b>
                  <small>Награды и следующие цели</small>
                </span>
              </Link>
              <Link role="menuitem" to={`${basePath}/progress`} onClick={() => setMenuOpen(false)}>
                <ChartLineUp size={19} weight="duotone" aria-hidden="true" />
                <span>
                  <b>Прогресс</b>
                  <small>Темы, Баллы и Звёзды</small>
                </span>
              </Link>
            </div>
            <button className={styles.logoutButton} role="menuitem" type="button" onClick={signOut}>
              <SignOut size={19} aria-hidden="true" />
              Выйти из аккаунта
            </button>
          </div>
        )}
      </div>
    </header>
  )
}
