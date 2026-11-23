export type ChunkedBlobUploadOptions = {
  signal: AbortSignal
  onProgress: (uploadedBytes: number, totalBytes: number) => void
}

export async function uploadBlocks(
  uploadUrl: string,
  uploadId: string,
  file: File,
  blockSize: number,
  options: ChunkedBlobUploadOptions,
): Promise<void> {
  const blockIds: string[] = []
  let uploadedBytes = 0

  for (let index = 0; uploadedBytes < file.size; index += 1) {
    const end = Math.min(uploadedBytes + blockSize, file.size)
    const blockId = createBlockId(uploadId, index)
    await stageBlock(uploadUrl, blockId, file.slice(uploadedBytes, end), uploadedBytes, file.size, options)
    blockIds.push(blockId)
    uploadedBytes = end
    options.onProgress(uploadedBytes, file.size)
  }

  await commitBlocks(uploadUrl, blockIds, file.type, options.signal)
}

function createBlockId(uploadId: string, index: number): string {
  return btoa(`${uploadId}:${String(index).padStart(8, '0')}`)
}

function storageUrl(uploadUrl: string, parameters: Record<string, string>): string {
  const url = new URL(uploadUrl)
  for (const [name, value] of Object.entries(parameters)) {
    url.searchParams.set(name, value)
  }
  return url.toString()
}

function stageBlock(uploadUrl: string, blockId: string, chunk: Blob, completedBytes: number, totalBytes: number, options: ChunkedBlobUploadOptions): Promise<void> {
  return sendRequest(
    storageUrl(uploadUrl, { comp: 'block', blockid: blockId }),
    chunk,
    {},
    options.signal,
    (loaded) => options.onProgress(completedBytes + loaded, totalBytes),
  )
}

function commitBlocks(uploadUrl: string, blockIds: string[], contentType: string, signal: AbortSignal): Promise<void> {
  const body = `<?xml version="1.0" encoding="utf-8"?><BlockList>${blockIds.map((blockId) => `<Latest>${blockId}</Latest>`).join('')}</BlockList>`
  const headers: Record<string, string> = { 'Content-Type': 'application/xml' }
  if (contentType) {
    headers['x-ms-blob-content-type'] = contentType
  }
  return sendRequest(storageUrl(uploadUrl, { comp: 'blocklist' }), body, headers, signal)
}

function sendRequest(
  url: string,
  body: Document | XMLHttpRequestBodyInit,
  headers: Record<string, string>,
  signal: AbortSignal,
  onProgress?: (loaded: number) => void,
): Promise<void> {
  return new Promise((resolve, reject) => {
    const request = new XMLHttpRequest()
    const abort = () => request.abort()
    if (signal.aborted) {
      abort()
      reject(new DOMException('Upload canceled.', 'AbortError'))
      return
    }
    signal.addEventListener('abort', abort, { once: true })
    request.open('PUT', url)
    for (const [name, value] of Object.entries(headers)) {
      request.setRequestHeader(name, value)
    }
    request.upload.addEventListener('progress', (event) => {
      if (event.lengthComputable) {
        onProgress?.(event.loaded)
      }
    })
    request.addEventListener('load', () => {
      signal.removeEventListener('abort', abort)
      if (request.status >= 200 && request.status < 300) {
        resolve()
        return
      }
      reject(new Error(`Storage chunk upload failed with HTTP ${request.status}.`))
    })
    request.addEventListener('error', () => {
      signal.removeEventListener('abort', abort)
      reject(new Error('Storage chunk upload failed due to a network error.'))
    })
    request.addEventListener('abort', () => {
      signal.removeEventListener('abort', abort)
      reject(new DOMException('Upload canceled.', 'AbortError'))
    })
    request.send(body)
  })
}
