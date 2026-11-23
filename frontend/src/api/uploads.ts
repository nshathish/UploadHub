import type { InitiateUploadRequest, InitiateUploadResponse, UploadSession } from '@/features/uploads/types'

const developmentToken = import.meta.env.VITE_DEV_AUTH_TOKEN

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
  if (!developmentToken) {
    throw new UploadApiError('Missing VITE_DEV_AUTH_TOKEN.', 0)
  }

  const response = await fetch('/api/uploads/initiate', {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${developmentToken}`,
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
  if (!developmentToken) {
    throw new UploadApiError('Missing VITE_DEV_AUTH_TOKEN.', 0)
  }

  const response = await fetch(`/api/uploads/${uploadId}/complete`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${developmentToken}`,
    },
  })

  if (!response.ok) {
    const message = await response.text()
    throw new UploadApiError(message || 'Unable to verify the upload.', response.status)
  }

  return response.json() as Promise<UploadSession>
}
