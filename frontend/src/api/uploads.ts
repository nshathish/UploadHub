import type { InitiateUploadRequest, InitiateUploadResponse, UploadSession } from '@/features/uploads/types'
import { getSessionToken } from '@/features/auth/session'

export class UploadApiError extends Error {
  readonly status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = 'UploadApiError'
    this.status = status
  }
}

export async function initiateUpload(
  request: InitiateUploadRequest,
  idempotencyKey: string,
): Promise<InitiateUploadResponse> {
  return initiate(request, idempotencyKey, '/api/uploads/initiate')
}

export async function initiateChunkedUpload(
  request: InitiateUploadRequest,
  idempotencyKey: string,
): Promise<InitiateUploadResponse> {
  return initiate(request, idempotencyKey, '/api/uploads/initiate-chunked')
}

async function initiate(
  request: InitiateUploadRequest,
  idempotencyKey: string,
  endpoint: string,
): Promise<InitiateUploadResponse> {
  const token = getSessionToken()
  if (!token) {
    throw new UploadApiError('Your session has ended. Sign in again.', 401)
  }

  const response = await fetch(endpoint, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
      'Idempotency-Key': idempotencyKey,
    },
    body: JSON.stringify(request),
  })

  if (!response.ok) {
    const message = await response.text()
    throw new UploadApiError(message || 'Unable to initiate the upload.', response.status)
  }

  return response.json() as Promise<InitiateUploadResponse>
}

export async function completeUpload(uploadId: string): Promise<UploadSession> {
  const token = getSessionToken()
  if (!token) {
    throw new UploadApiError('Your session has ended. Sign in again.', 401)
  }

  const response = await fetch(`/api/uploads/${uploadId}/complete`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`,
    },
  })

  if (!response.ok) {
    const message = await response.text()
    throw new UploadApiError(message || 'Unable to verify the upload.', response.status)
  }

  return response.json() as Promise<UploadSession>
}

export async function listUploads(): Promise<UploadSession[]> {
  const token = getSessionToken()
  if (!token) {
    throw new UploadApiError('Your session has ended. Sign in again.', 401)
  }

  const response = await fetch('/api/uploads', { headers: { Authorization: `Bearer ${token}` } })
  if (!response.ok) {
    const message = await response.text()
    throw new UploadApiError(message || 'Unable to load uploads.', response.status)
  }

  return response.json() as Promise<UploadSession[]>
}

export type ChunkResume = { id: string; uploadUrl: string; blockSize: number; completedChunkIndexes: number[] }

export async function resumeChunkedUpload(uploadId: string): Promise<ChunkResume> {
  const token = getSessionToken()
  if (!token) throw new UploadApiError('Your session has ended. Sign in again.', 401)
  const response = await fetch(`/api/uploads/${uploadId}/chunks`, { headers: { Authorization: `Bearer ${token}` } })
  if (!response.ok) throw new UploadApiError((await response.text()) || 'Unable to resume the upload.', response.status)
  return response.json() as Promise<ChunkResume>
}
