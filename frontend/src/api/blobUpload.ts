export type BlobUploadOptions = {
  signal: AbortSignal
  onProgress: (uploadedBytes: number, totalBytes: number) => void
}

export function uploadBlob(
  uploadUrl: string,
  file: File,
  options: BlobUploadOptions,
): Promise<void> {
  return new Promise((resolve, reject) => {
    const request = new XMLHttpRequest()
    const abort = () => request.abort()

    options.signal.addEventListener('abort', abort, { once: true })
    request.open('PUT', uploadUrl)
    request.setRequestHeader('x-ms-blob-type', 'BlockBlob')

    request.upload.addEventListener('progress', (event) => {
      if (event.lengthComputable) {
        options.onProgress(event.loaded, event.total)
      }
    })

    request.addEventListener('load', () => {
      options.signal.removeEventListener('abort', abort)
      if (request.status >= 200 && request.status < 300) {
        resolve()
        return
      }
      reject(new Error(`Storage upload failed with HTTP ${request.status}.`))
    })

    request.addEventListener('error', () => {
      options.signal.removeEventListener('abort', abort)
      reject(new Error('Storage upload failed due to a network error.'))
    })

    request.addEventListener('abort', () => {
      options.signal.removeEventListener('abort', abort)
      reject(new DOMException('Upload canceled.', 'AbortError'))
    })

    request.send(file)
  })
}
