export type InitiateUploadRequest = {
  fileName: string
  size: number
}

export type InitiateUploadResponse = {
  id: string
  blobName: string
  uploadUrl: string
  expiresAt: string
}

export type UploadState =
  | { phase: 'idle' }
  | { phase: 'initiating' }
  | { phase: 'uploading'; progress: number; uploadedBytes: number }
  | { phase: 'transferred'; uploadId: string }
  | { phase: 'canceled' }
  | { phase: 'error'; message: string }
