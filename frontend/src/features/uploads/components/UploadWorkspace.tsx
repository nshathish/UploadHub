import { useState } from 'react'

import { FilePicker } from '@/features/uploads/components/FilePicker'
import { ChunkedUploadPage } from '@/features/uploads/components/ChunkedUploadPage'
import { UploadStatus } from '@/features/uploads/components/UploadStatus'
import { useDirectUpload } from '@/features/uploads/useDirectUpload'
import { requiresChunkedUpload, validateFile } from '@/features/uploads/validation'

export function UploadWorkspace() {
  const [file, setFile] = useState<File | null>(null)
  const [validationError, setValidationError] = useState<string | null>(null)
  const upload = useDirectUpload()
  const isBusy = upload.state.phase === 'initiating' || upload.state.phase === 'uploading'

  function selectFile(nextFile: File) {
    const error = validateFile(nextFile)
    setValidationError(error)
    if (error) {
      setFile(null)
      return
    }

    setFile(nextFile)
    upload.reset()
  }

  function reset() {
    upload.reset()
    setFile(null)
    setValidationError(null)
  }

  if (file && requiresChunkedUpload(file)) {
    return <ChunkedUploadPage file={file} onBack={reset} />
  }

  return (
    <main className="min-h-screen bg-slate-950 px-4 py-10 text-slate-900 sm:px-6 lg:px-8">
      <div className="mx-auto max-w-5xl">
        <header className="flex items-center gap-3 text-white"><div className="flex size-10 items-center justify-center rounded-xl bg-indigo-500 text-lg font-black">U</div><div><p className="text-lg font-bold tracking-tight">UploadHub</p><p className="text-sm text-slate-400">Secure document intake</p></div></header>
        <section className="mt-10 grid overflow-hidden rounded-3xl bg-white shadow-2xl shadow-slate-950/30 lg:grid-cols-[1.05fr_0.95fr]">
          <div className="p-7 sm:p-10">
            <p className="text-sm font-semibold uppercase tracking-[0.18em] text-indigo-700">Document upload</p>
            <h1 className="mt-3 text-3xl font-bold tracking-tight text-slate-950 sm:text-4xl">Send your verification document securely.</h1>
            <p className="mt-4 max-w-xl text-base leading-7 text-slate-600">Your file is uploaded directly to protected cloud storage. UploadHub never routes document bytes through this application server.</p>
            <div className="mt-8">{!file ? <FilePicker disabled={isBusy} onFileSelected={selectFile} /> : <UploadStatus file={file} state={upload.state} onCancel={upload.cancel} onReset={reset} />}</div>
            {validationError && <p className="mt-4 rounded-lg bg-rose-50 px-4 py-3 text-sm text-rose-900">{validationError}</p>}
            {file && upload.state.phase === 'idle' && <button className="mt-5 w-full rounded-lg bg-indigo-600 px-5 py-3 text-sm font-semibold text-white transition hover:bg-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2" type="button" onClick={() => upload.start(file)}>Start secure upload</button>}
          </div>
          <aside className="bg-indigo-700 p-7 text-white sm:p-10">
            <p className="text-sm font-semibold uppercase tracking-[0.18em] text-indigo-200">How it works</p>
            <ol className="mt-8 space-y-7"><Step number="01" title="Select your document" description="We validate file type and size before the transfer begins." /><Step number="02" title="Receive a secure upload link" description="The server authorizes one short-lived, file-specific storage URL." /><Step number="03" title="Upload and verify" description="Your browser transfers directly to storage, then the server verifies receipt." /></ol>
            <div className="mt-10 border-t border-white/20 pt-6 text-sm leading-6 text-indigo-100">Supported files: PDF, JPEG, and PNG. Files above 25 MiB automatically use chunked transfer.</div>
          </aside>
        </section>
      </div>
    </main>
  )
}

type StepProps = { number: string; title: string; description: string }

function Step({ number, title, description }: StepProps) {
  return <li className="flex gap-4"><span className="flex size-8 shrink-0 items-center justify-center rounded-full border border-indigo-300/60 text-xs font-bold text-indigo-100">{number}</span><div><p className="font-semibold">{title}</p><p className="mt-1 text-sm leading-6 text-indigo-100">{description}</p></div></li>
}
