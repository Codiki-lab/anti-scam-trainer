import type {
  ButtonHTMLAttributes,
  InputHTMLAttributes,
  ReactNode,
  SelectHTMLAttributes,
  TextareaHTMLAttributes,
} from 'react'
import styles from './AdminPage.module.scss'

interface FieldMetaProps {
  label: string
  hint?: string
}

interface FieldShellProps extends FieldMetaProps {
  children: ReactNode
}

function FieldShell({ label, hint, children }: FieldShellProps) {
  return (
    <label className={styles.field}>
      <span>{label}</span>
      {children}
      {hint && <small>{hint}</small>}
    </label>
  )
}

export function TextField({
  label,
  hint,
  ...props
}: FieldMetaProps & InputHTMLAttributes<HTMLInputElement>) {
  return (
    <FieldShell label={label} hint={hint}>
      <input {...props} />
    </FieldShell>
  )
}

export function TextAreaField({
  label,
  hint,
  ...props
}: FieldMetaProps & TextareaHTMLAttributes<HTMLTextAreaElement>) {
  return (
    <FieldShell label={label} hint={hint}>
      <textarea rows={4} {...props} />
    </FieldShell>
  )
}

export function SelectField({
  label,
  hint,
  children,
  ...props
}: FieldShellProps & SelectHTMLAttributes<HTMLSelectElement>) {
  return (
    <FieldShell label={label} hint={hint}>
      <select {...props}>{children}</select>
    </FieldShell>
  )
}

export function EditorActions({ children }: { children: ReactNode }) {
  return <div className={styles.editorActions}>{children}</div>
}

export function PrimaryAction(props: ButtonHTMLAttributes<HTMLButtonElement>) {
  return <button {...props} className={styles.primaryAction} type="button" />
}

export function SecondaryAction(props: ButtonHTMLAttributes<HTMLButtonElement>) {
  return <button {...props} className={styles.secondaryAction} type="button" />
}

export function DangerAction(props: ButtonHTMLAttributes<HTMLButtonElement>) {
  return <button {...props} className={styles.dangerAction} type="button" />
}
