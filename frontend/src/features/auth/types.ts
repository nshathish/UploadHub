export type AuthenticatedUser = {
  id: string
  username: string
}

export type AuthenticatedSession = {
  token: string
  expiresAt: string
  user: AuthenticatedUser
}
