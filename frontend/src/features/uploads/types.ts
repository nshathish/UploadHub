export type InitiateUploadRequest = {
  fileName: string
  size: number
}

export type InitiateUploadResponse = {
  id: string
  blobName: string
  uploadUrl: string
  expiresAt: string
  uploadMode: 'single' | 'chunked'
  blockSize?: number
}

export type UploadSession = {
  id: string
  fileName: string
  size: number
  status: 'initiated' | 'uploading' | 'uploaded' | 'validating' | 'accepted' | 'rejected' | 'canceled' | 'expired' | 'failed'
  expiresAt: string
}

export type UploadState =
  | { phase: 'idle' }
  | { phase: 'initiating' }
  | { phase: 'uploading'; progress: number; uploadedBytes: number }
  | { phase: 'verifying'; uploadId: string }
  | { phase: 'uploaded'; uploadId: string }
  | { phase: 'canceled' }
  | { phase: 'error'; message: string }
