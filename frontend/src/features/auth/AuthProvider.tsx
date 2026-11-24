import { useState, type ReactNode } from 'react'

import { login, logout } from '@/api/auth'
import { AuthContext } from '@/features/auth/AuthContext'
import { clearStoredSession, getStoredSession, storeSession } from '@/features/auth/session'
import type { AuthenticatedSession } from '@/features/auth/types'

export function AuthProvider({ children }: { children: ReactNode }) {
  const [session, setSession] = useState<AuthenticatedSession | null>(() => getStoredSession())

  async function signIn(username: string, password: string) {
    const nextSession = await login(username, password)
    storeSession(nextSession)
    setSession(nextSession)
  }

  async function signOut() {
    try {
      await logout()
    } finally {
      clearStoredSession()
      setSession(null)
    }
  }

  return <AuthContext.Provider value={{ session, signIn, signOut }}>{children}</AuthContext.Provider>
}
