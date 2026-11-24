import { createContext, useContext } from 'react'

import type { AuthenticatedSession } from '@/features/auth/types'

export type AuthContextValue = {
  session: AuthenticatedSession | null
  signIn: (username: string, password: string) => Promise<void>
  signOut: () => Promise<void>
}

export const AuthContext = createContext<AuthContextValue | null>(null)

export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used inside AuthProvider.')
  }
  return context
}
