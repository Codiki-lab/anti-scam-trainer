import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { Link, useSearchParams } from 'react-router-dom'
import { z } from 'zod'
import { RoleSelector, type UserRole } from '@/entities/user'
import { useIsPreview } from '@/shared/runtime-mode'
import { uiStyles } from '@/shared/ui-kit'
import { useAuthFlow } from '../model/useAuthFlow'
import styles from './AuthForm.module.scss'

type AuthValues = { username: string; password: string; trainingRole: UserRole }
const credentialsSchema = z.object({
  username: z.string().trim().min(3, 'Введите минимум 3 символа.'),
  password: z.string().min(6, 'Введите минимум 6 символов.'),
  trainingRole: z.enum(['buyer', 'seller']),
})

export function AuthForm({ mode }: { mode: 'login' | 'register' }) {
  const isRegister = mode === 'register'
  const isPreview = useIsPreview()
  const [searchParams] = useSearchParams()
  const {
    register,
    handleSubmit,
    setValue,
    watch,
    formState: { errors },
  } = useForm<AuthValues>({
    resolver: zodResolver(credentialsSchema),
    defaultValues: { trainingRole: 'buyer' },
  })
  const trainingRole = watch('trainingRole')
  const { submit, error, isSubmitting } = useAuthFlow(mode, trainingRole)
  const authBasePath = isPreview ? '/preview' : ''

  return (
    <form className={styles.card} onSubmit={handleSubmit(submit)}>
      <p className={uiStyles.eyebrow}>Тренажёр безопасности</p>
      <h2>{isRegister ? 'Создайте аккаунт' : 'Войдите в аккаунт'}</h2>
      <p className={`${uiStyles.muted} ${styles.description}`}>
        {isRegister
          ? 'Выберите роль — её можно будет изменить позднее.'
          : 'Продолжайте обучение с того места, где остановились.'}
      </p>
      {!isRegister && searchParams.get('registered') === '1' && (
        <p className={uiStyles.formSuccess} role="status">
          Аккаунт создан. Теперь войдите с указанными данными.
        </p>
      )}
      {isRegister && (
        <RoleSelector
          value={trainingRole}
          onChange={(role) => setValue('trainingRole', role, { shouldValidate: true })}
        />
      )}
      <label>
        Логин
        <input autoComplete="username" placeholder="Например, alex" {...register('username')} />
      </label>
      {errors.username && <p className={uiStyles.fieldError}>{errors.username.message}</p>}
      <label>
        Пароль
        <input
          type="password"
          autoComplete={isRegister ? 'new-password' : 'current-password'}
          placeholder="Минимум 6 символов"
          {...register('password')}
        />
      </label>
      {errors.password && <p className={uiStyles.fieldError}>{errors.password.message}</p>}
      {error && <p className={uiStyles.formError}>{error}</p>}
      <button
        className={`${uiStyles.primaryButton} ${uiStyles.full} ${styles.submit}`}
        disabled={isSubmitting}
      >
        {isSubmitting ? 'Проверяем…' : isRegister ? 'Создать аккаунт' : 'Войти'}
      </button>
      <p className={styles.switch}>
        {isRegister ? 'Уже есть аккаунт?' : 'Впервые в тренажёре?'}{' '}
        <Link to={`${authBasePath}/${isRegister ? 'login' : 'register'}`}>
          {isRegister ? 'Войти' : 'Зарегистрироваться'}
        </Link>
      </p>
    </form>
  )
}
