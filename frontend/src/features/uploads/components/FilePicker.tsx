import { useRef } from 'react'

type FilePickerProps = {
  disabled: boolean
  onFileSelected: (file: File) => void
}

export function FilePicker({ disabled, onFileSelected }: FilePickerProps) {
  const inputRef = useRef<HTMLInputElement>(null)

  return (
    <div className="rounded-2xl border border-dashed border-slate-300 bg-slate-50 p-8 text-center transition hover:border-indigo-400 hover:bg-indigo-50/50">
      <input
        ref={inputRef}
        className="sr-only"
        type="file"
        accept="application/pdf,image/jpeg,image/png"
        disabled={disabled}
        onChange={(event) => {
          const [file] = Array.from(event.target.files ?? [])
          if (file) {
            onFileSelected(file)
          }
          event.target.value = ''
        }}
      />
      <div className="mx-auto mb-4 flex size-12 items-center justify-center rounded-full bg-indigo-100 text-2xl text-indigo-700">↑</div>
      <p className="text-base font-semibold text-slate-900">Select a document to upload</p>
      <p className="mt-2 text-sm leading-6 text-slate-600">PDF, JPEG, or PNG. Files above 25 MiB use chunked transfer automatically.</p>
      <button
        className="mt-5 rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white transition hover:bg-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 disabled:cursor-not-allowed disabled:bg-slate-400"
        type="button"
        disabled={disabled}
        onClick={() => inputRef.current?.click()}
      >
        Choose file
      </button>
    </div>
  )
}
