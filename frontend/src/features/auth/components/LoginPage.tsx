import { useState, type FormEvent } from 'react'

import { useAuth } from '@/features/auth/AuthContext'

export function LoginPage() {
  const { signIn } = useAuth()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError(null)
    setIsSubmitting(true)
    try {
      await signIn(username, password)
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Unable to sign in.')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <main className="min-h-screen bg-slate-950 px-4 py-10 text-slate-900 sm:px-6 lg:px-8">
      <div className="mx-auto flex min-h-[calc(100vh-5rem)] max-w-md items-center">
        <section className="w-full overflow-hidden rounded-3xl bg-white shadow-2xl shadow-slate-950/30">
          <div className="bg-indigo-700 p-8 text-white sm:p-10">
            <div className="flex size-11 items-center justify-center rounded-xl bg-indigo-400 text-xl font-black">U</div>
            <p className="mt-7 text-sm font-semibold uppercase tracking-[0.18em] text-indigo-200">UploadHub</p>
            <h1 className="mt-3 text-3xl font-bold tracking-tight">Sign in to upload securely.</h1>
            <p className="mt-4 text-base leading-7 text-indigo-100">Your account authorizes document uploads and keeps each upload associated with its owner.</p>
          </div>
          <form className="p-8 sm:p-10" onSubmit={submit}>
            <label className="block text-sm font-semibold text-slate-800" htmlFor="username">Username</label>
            <input className="mt-2 w-full rounded-lg border border-slate-300 px-3 py-2.5 text-base outline-none transition focus:border-indigo-500 focus:ring-2 focus:ring-indigo-200" id="username" name="username" autoComplete="username" value={username} onChange={(event) => setUsername(event.target.value)} required />
            <label className="mt-5 block text-sm font-semibold text-slate-800" htmlFor="password">Password</label>
            <input className="mt-2 w-full rounded-lg border border-slate-300 px-3 py-2.5 text-base outline-none transition focus:border-indigo-500 focus:ring-2 focus:ring-indigo-200" id="password" name="password" type="password" autoComplete="current-password" value={password} onChange={(event) => setPassword(event.target.value)} required />
            {error && <p className="mt-5 rounded-lg bg-rose-50 px-4 py-3 text-sm text-rose-900">{error}</p>}
            <button className="mt-6 w-full rounded-lg bg-indigo-600 px-5 py-3 text-sm font-semibold text-white transition hover:bg-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 disabled:cursor-not-allowed disabled:bg-slate-400" type="submit" disabled={isSubmitting}>{isSubmitting ? 'Signing in…' : 'Sign in'}</button>
          </form>
        </section>
      </div>
    </main>
  )
}
