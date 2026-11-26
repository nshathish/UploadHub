import { useMemo } from 'react'

import { UploadStatus } from '@/features/uploads/components/UploadStatus'
import { ChunkProgressMap } from '@/features/uploads/components/ChunkProgressMap'
import { chunkCountForFile } from '@/features/uploads/chunkSizing'
import { useChunkedUpload } from '@/features/uploads/useChunkedUpload'
import { singleUploadLimit } from '@/features/uploads/validation'

type ChunkedUploadPageProps = {
  file: File
  onBack: () => void
  onBackToUploads: () => void
  resumeUploadId?: string
}

export function ChunkedUploadPage({ file, onBack, onBackToUploads, resumeUploadId }: ChunkedUploadPageProps) {
  const upload = useChunkedUpload()
  const isIdle = upload.state.phase === 'idle'
  const isBusy = upload.state.phase === 'initiating' || upload.state.phase === 'uploading' || upload.state.phase === 'verifying'
  const queuedChunks = useMemo(() => Array.from({ length: chunkCountForFile(file.size) }, () => 'queued' as const), [file.size])
  const chunks = upload.chunks.length > 0 ? upload.chunks : queuedChunks

  return (
    <main className="min-h-screen bg-slate-950 px-4 py-10 text-slate-900 sm:px-6 lg:px-8">
      <div className="mx-auto max-w-3xl">
        <header className="flex items-center justify-between gap-4 text-white"><div className="flex items-center gap-3"><div className="flex size-10 items-center justify-center rounded-xl bg-cyan-500 text-lg font-black">U</div><div><p className="text-lg font-bold tracking-tight">UploadHub</p><p className="text-sm text-slate-400">Large document transfer</p></div></div><button className="rounded-lg border border-slate-600 px-3 py-2 text-sm font-semibold transition hover:border-slate-400 hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-50" type="button" onClick={onBackToUploads} disabled={isBusy}>My uploads</button></header>
        <section className="mt-10 overflow-hidden rounded-3xl bg-white shadow-2xl shadow-slate-950/30">
          <div className="border-b border-slate-200 bg-cyan-50 p-7 sm:p-10">
            <p className="text-sm font-semibold uppercase tracking-[0.18em] text-cyan-800">Chunked upload</p>
            <h1 className="mt-3 text-3xl font-bold tracking-tight text-slate-950 sm:text-4xl">Large file upload</h1>
            <p className="mt-4 max-w-2xl text-base leading-7 text-slate-600">This document is larger than {formatFileSize(singleUploadLimit)}. Your browser will stage it in smaller blocks directly to secure storage, then commit the complete file.</p>
          </div>
          <div className="p-7 sm:p-10">
            <UploadStatus file={file} state={upload.state} onCancel={upload.cancel} onReset={onBack} />
            <ChunkProgressMap chunks={chunks} />
            {isIdle && !resumeUploadId && <button className="mt-5 w-full rounded-lg bg-cyan-700 px-5 py-3 text-sm font-semibold text-white transition hover:bg-cyan-600 focus:outline-none focus:ring-2 focus:ring-cyan-600 focus:ring-offset-2" type="button" onClick={() => upload.start(file)}>Start chunked upload</button>}
            {isIdle && resumeUploadId && <button className="mt-5 w-full rounded-lg bg-cyan-700 px-5 py-3 text-sm font-semibold text-white transition hover:bg-cyan-600 focus:outline-none focus:ring-2 focus:ring-cyan-600 focus:ring-offset-2" type="button" onClick={() => upload.resume(file, resumeUploadId)}>Resume remaining chunks</button>}
            {upload.state.phase === 'error' && upload.state.uploadId && <button className="mt-5 w-full rounded-lg bg-cyan-700 px-5 py-3 text-sm font-semibold text-white transition hover:bg-cyan-600 focus:outline-none focus:ring-2 focus:ring-cyan-600 focus:ring-offset-2" type="button" onClick={() => upload.retry(file)}>Retry remaining chunks</button>}
            {isIdle && <button className="mt-3 w-full rounded-lg px-5 py-3 text-sm font-semibold text-slate-700 transition hover:bg-slate-100" type="button" onClick={onBack}>Choose a different file</button>}
          </div>
        </section>
      </div>
    </main>
  )
}

function formatFileSize(bytes: number): string {
  return `${(bytes / (1024 * 1024)).toFixed(0)} MiB`
}
