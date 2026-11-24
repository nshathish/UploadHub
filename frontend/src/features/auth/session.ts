import type { AuthenticatedSession } from '@/features/auth/types'

const storageKey = 'uploadhub.auth-session'

export function getStoredSession(): AuthenticatedSession | null {
  const value = sessionStorage.getItem(storageKey)
  if (!value) {
    return null
  }
  try {
    return JSON.parse(value) as AuthenticatedSession
  } catch {
    sessionStorage.removeItem(storageKey)
    return null
  }
}

export function getSessionToken(): string | null {
  return getStoredSession()?.token ?? null
}

export function storeSession(session: AuthenticatedSession): void {
  sessionStorage.setItem(storageKey, JSON.stringify(session))
}

export function clearStoredSession(): void {
  sessionStorage.removeItem(storageKey)
}
