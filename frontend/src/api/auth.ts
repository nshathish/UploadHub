import { clearStoredSession, getSessionToken } from '@/features/auth/session'
import type { AuthenticatedSession } from '@/features/auth/types'

export class AuthApiError extends Error {
  readonly status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = 'AuthApiError'
    this.status = status
  }
}

export async function login(username: string, password: string): Promise<AuthenticatedSession> {
  const response = await fetch('/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
  })
  if (!response.ok) {
    const message = await response.text()
    throw new AuthApiError(message || 'Unable to sign in.', response.status)
  }
  return response.json() as Promise<AuthenticatedSession>
}

export async function logout(): Promise<void> {
  const token = getSessionToken()
  if (!token) {
    clearStoredSession()
    return
  }
  try {
    await fetch('/api/auth/logout', { method: 'POST', headers: { Authorization: `Bearer ${token}` } })
  } finally {
    clearStoredSession()
  }
}
