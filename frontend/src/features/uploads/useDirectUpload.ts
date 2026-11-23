import { useRef, useState } from 'react'

import { uploadBlob } from '@/api/blobUpload'
import { completeUpload, initiateUpload } from '@/api/uploads'
import type { UploadState } from '@/features/uploads/types'

export function useDirectUpload() {
  const [state, setState] = useState<UploadState>({ phase: 'idle' })
  const abortController = useRef<AbortController | null>(null)

  async function start(file: File) {
    const controller = new AbortController()
    abortController.current = controller
    setState({ phase: 'initiating' })

    try {
      const upload = await initiateUpload(
        { fileName: file.name, size: file.size },
        crypto.randomUUID(),
      )

      setState({ phase: 'uploading', progress: 0, uploadedBytes: 0 })
      await uploadBlob(upload.uploadUrl, file, {
        signal: controller.signal,
        onProgress: (uploadedBytes, totalBytes) => {
          setState({
            phase: 'uploading',
            progress: Math.round((uploadedBytes / totalBytes) * 100),
            uploadedBytes,
          })
        },
      })

      setState({ phase: 'verifying', uploadId: upload.id })
      await completeUpload(upload.id)
      setState({ phase: 'uploaded', uploadId: upload.id })
    } catch (error) {
      if (error instanceof DOMException && error.name === 'AbortError') {
        setState({ phase: 'canceled' })
      } else {
        setState({
          phase: 'error',
          message: error instanceof Error ? error.message : 'Unable to upload the document.',
        })
      }
    } finally {
      abortController.current = null
    }
  }

  function cancel() {
    abortController.current?.abort()
  }

  function reset() {
    abortController.current?.abort()
    abortController.current = null
    setState({ phase: 'idle' })
  }

  return { cancel, reset, start, state }
}
