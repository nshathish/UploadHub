import { useEffect, useRef, useState } from 'react'

import { listUploads } from '@/api/uploads'
import { useAuth } from '@/features/auth/AuthContext'
import type { UploadSession } from '@/features/uploads/types'

type UploadsPageProps = {
  onNewUpload: () => void
  onResumeUpload: (upload: UploadSession, file: File) => void
}

export function UploadsPage({ onNewUpload, onResumeUpload }: UploadsPageProps) {
  const { session, signOut } = useAuth()
  const [uploads, setUploads] = useState<UploadSession[]>([])
  const [error, setError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [resumeUpload, setResumeUpload] = useState<UploadSession | null>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    let isCurrent = true
    void listUploads()
      .then((result) => {
        if (isCurrent) {
          setUploads(result)
        }
      })
      .catch((error) => {
        if (isCurrent) {
          setError(error instanceof Error ? error.message : 'Unable to load uploads.')
        }
      })
      .finally(() => {
        if (isCurrent) {
          setIsLoading(false)
        }
      })
    return () => { isCurrent = false }
  }, [])

  function selectResumeFile(file: File) {
    if (!resumeUpload) return
    if (file.name !== resumeUpload.fileName || file.size !== resumeUpload.size) {
      setError('Choose the original file with the same name and size to resume this upload.')
      return
    }
    onResumeUpload(resumeUpload, file)
  }

  return (
    <main className="min-h-screen bg-slate-950 px-4 py-10 text-slate-900 sm:px-6 lg:px-8">
      <div className="mx-auto max-w-5xl">
        <header className="flex items-center justify-between gap-4 text-white">
          <div className="flex items-center gap-3"><div className="flex size-10 items-center justify-center rounded-xl bg-indigo-500 text-lg font-black">U</div><div><p className="text-lg font-bold tracking-tight">UploadHub</p><p className="text-sm text-slate-400">Secure document intake</p></div></div>
          <div className="flex items-center gap-3"><span className="hidden text-sm text-slate-300 sm:inline">Signed in as {session?.user.username}</span><button className="rounded-lg border border-slate-600 px-3 py-2 text-sm font-semibold transition hover:border-slate-400 hover:bg-slate-800" type="button" onClick={() => void signOut()}>Sign out</button></div>
        </header>
        <section className="mt-10 rounded-3xl bg-white p-7 shadow-2xl shadow-slate-950/30 sm:p-10">
          <div className="flex flex-wrap items-start justify-between gap-5"><div><p className="text-sm font-semibold uppercase tracking-[0.18em] text-indigo-700">Your documents</p><h1 className="mt-3 text-3xl font-bold tracking-tight text-slate-950 sm:text-4xl">Uploads</h1><p className="mt-3 text-base leading-7 text-slate-600">Review the upload sessions associated with your account.</p></div><button className="rounded-lg bg-indigo-600 px-5 py-3 text-sm font-semibold text-white transition hover:bg-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2" type="button" onClick={onNewUpload}>New upload</button></div>
          {isLoading && <p className="mt-10 text-sm text-slate-600">Loading uploads…</p>}
          {error && <p className="mt-8 rounded-lg bg-rose-50 px-4 py-3 text-sm text-rose-900">{error}</p>}
          {!isLoading && !error && uploads.length === 0 && <div className="mt-10 rounded-2xl border border-dashed border-slate-300 bg-slate-50 p-8 text-center"><p className="font-semibold text-slate-900">No uploads yet</p><p className="mt-2 text-sm leading-6 text-slate-600">Start a new upload to send your first verification document.</p></div>}
          <input ref={inputRef} className="sr-only" type="file" accept="application/pdf,image/jpeg,image/png" onChange={(event) => { const [file] = Array.from(event.target.files ?? []); if (file) selectResumeFile(file); event.target.value = '' }} />
          {!isLoading && uploads.length > 0 && <ul className="mt-8 divide-y divide-slate-200 rounded-2xl border border-slate-200">{uploads.map((upload) => <li key={upload.id} className="flex flex-wrap items-center justify-between gap-4 p-5"><div className="min-w-0"><p className="truncate font-semibold text-slate-900">{upload.fileName}</p><p className="mt-1 text-sm text-slate-500">{formatFileSize(upload.size)} · {formatDate(upload.createdAt)} · {upload.uploadMode} upload</p></div><div className="flex items-center gap-3"><StatusBadge status={upload.status} />{upload.status === 'initiated' && upload.uploadMode === 'chunked' && <button className="rounded-lg border border-cyan-700 px-3 py-2 text-sm font-semibold text-cyan-800 transition hover:bg-cyan-50" type="button" onClick={() => { setResumeUpload(upload); inputRef.current?.click() }}>Resume upload</button>}</div></li>)}</ul>}
        </section>
      </div>
    </main>
  )
}

function StatusBadge({ status }: { status: UploadSession['status'] }) {
  const styles: Record<UploadSession['status'], string> = {
    initiated: 'bg-slate-100 text-slate-700', uploading: 'bg-blue-50 text-blue-800', uploaded: 'bg-emerald-50 text-emerald-800', validating: 'bg-amber-50 text-amber-900', accepted: 'bg-emerald-50 text-emerald-800', rejected: 'bg-rose-50 text-rose-800', canceled: 'bg-slate-100 text-slate-700', expired: 'bg-slate-100 text-slate-700', failed: 'bg-rose-50 text-rose-800',
  }
  return <span className={`rounded-full px-3 py-1 text-sm font-semibold ${styles[status]}`}>{status}</span>
}

function formatFileSize(bytes: number): string {
  return bytes < 1024 * 1024 ? `${Math.ceil(bytes / 1024)} KiB` : `${(bytes / (1024 * 1024)).toFixed(1)} MiB`
}

function formatDate(value: string): string {
  return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value))
}
