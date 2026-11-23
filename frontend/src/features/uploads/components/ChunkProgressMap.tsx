export type ChunkState = 'queued' | 'uploading' | 'complete' | 'failed'

type ChunkProgressMapProps = {
  chunks: ChunkState[]
}

const stateClassNames: Record<ChunkState, string> = {
  queued: 'bg-slate-300',
  uploading: 'bg-blue-500 ring-2 ring-blue-200',
  complete: 'bg-emerald-500',
  failed: 'bg-rose-500',
}

export function ChunkProgressMap({ chunks }: ChunkProgressMapProps) {
  const completed = chunks.filter((chunk) => chunk === 'complete').length
  const active = chunks.findIndex((chunk) => chunk === 'uploading')

  return (
    <section className="mt-6 rounded-2xl border border-slate-200 bg-slate-50 p-5" aria-label="Chunk upload progress">
      <div className="flex flex-wrap items-baseline justify-between gap-2">
        <div><p className="font-semibold text-slate-900">Upload chunks</p><p className="mt-1 text-sm text-slate-600">{completed} of {chunks.length} chunks completed{active >= 0 ? ` · uploading chunk ${active + 1}` : ''}</p></div>
        <div className="flex flex-wrap gap-x-3 gap-y-1 text-xs font-medium text-slate-600"><Legend color="bg-slate-300" label="Queued" /><Legend color="bg-blue-500" label="Uploading" /><Legend color="bg-emerald-500" label="Complete" /><Legend color="bg-rose-500" label="Failed" /></div>
      </div>
      <div className="mt-5 grid grid-cols-10 gap-1.5 sm:grid-cols-16 md:grid-cols-20" role="list">
        {chunks.map((state, index) => <span key={index} className={`h-3 rounded-sm ${stateClassNames[state]}`} role="listitem" title={`Chunk ${index + 1}: ${state}`} aria-label={`Chunk ${index + 1}: ${state}`} />)}
      </div>
    </section>
  )
}

type LegendProps = { color: string; label: string }

function Legend({ color, label }: LegendProps) {
  return <span className="flex items-center gap-1.5"><span className={`size-2.5 rounded-sm ${color}`} />{label}</span>
}
