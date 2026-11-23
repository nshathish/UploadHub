import type { UploadState } from '@/features/uploads/types'

type UploadStatusProps = {
  file: File
  state: UploadState
  onCancel: () => void
  onReset: () => void
}

export function UploadStatus({ file, state, onCancel, onReset }: UploadStatusProps) {
  const isActive = state.phase === 'initiating' || state.phase === 'uploading'
  const isVerifying = state.phase === 'verifying'
  const canReset = state.phase === 'uploaded' || state.phase === 'canceled' || state.phase === 'error'

  return (
    <section className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0">
          <p className="truncate font-semibold text-slate-900">{file.name}</p>
          <p className="mt-1 text-sm text-slate-500">{formatFileSize(file.size)}</p>
        </div>
        {isActive ? (
          <button className="text-sm font-semibold text-slate-600 hover:text-slate-900" type="button" onClick={onCancel}>Cancel</button>
        ) : isVerifying ? (
          <span className="text-sm font-semibold text-slate-500">Verifying</span>
        ) : canReset ? (
          <button className="text-sm font-semibold text-indigo-700 hover:text-indigo-600" type="button" onClick={onReset}>Upload another</button>
        ) : null}
      </div>

      {state.phase === 'initiating' && <p className="mt-5 text-sm text-slate-600">Preparing a secure upload…</p>}

      {state.phase === 'uploading' && (
        <div className="mt-5">
          <div className="mb-2 flex justify-between text-sm text-slate-600"><span>Uploading to secure storage</span><span>{state.progress}%</span></div>
          <div className="h-2 overflow-hidden rounded-full bg-slate-100"><div className="h-full rounded-full bg-indigo-600 transition-all" style={{ width: `${state.progress}%` }} /></div>
          <p className="mt-2 text-xs text-slate-500">{formatFileSize(state.uploadedBytes)} transferred</p>
        </div>
      )}

      {state.phase === 'verifying' && (
        <div className="mt-5 rounded-xl bg-amber-50 p-4 text-sm text-amber-950">
          <p className="font-semibold">Transfer complete</p>
          <p className="mt-1 leading-6">Verifying your document with secure storage…</p>
          <p className="mt-2 font-mono text-xs text-amber-800">Upload ID: {state.uploadId}</p>
        </div>
      )}

      {state.phase === 'uploaded' && (
        <div className="mt-5 rounded-xl bg-emerald-50 p-4 text-sm text-emerald-950">
          <p className="font-semibold">Upload confirmed</p>
          <p className="mt-1 leading-6">Your document was received and verified by UploadHub.</p>
          <p className="mt-2 font-mono text-xs text-emerald-800">Upload ID: {state.uploadId}</p>
        </div>
      )}

      {state.phase === 'canceled' && <p className="mt-5 text-sm text-slate-600">Upload canceled. You can choose the file again.</p>}

      {state.phase === 'error' && (
        <div className="mt-5 rounded-xl bg-rose-50 p-4 text-sm text-rose-900"><p className="font-semibold">Upload could not be completed</p><p className="mt-1 leading-6">{state.message}</p></div>
      )}
    </section>
  )
}

function formatFileSize(bytes: number) {
  if (bytes < 1024 * 1024) {
    return `${Math.ceil(bytes / 1024)} KiB`
  }

  return `${(bytes / (1024 * 1024)).toFixed(1)} MiB`
}
