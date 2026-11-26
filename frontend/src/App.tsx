import { useState } from 'react'

import { useAuth } from '@/features/auth/AuthContext'
import { AuthProvider } from '@/features/auth/AuthProvider'
import { LoginPage } from '@/features/auth/components/LoginPage'
import { UploadWorkspace } from '@/features/uploads/components/UploadWorkspace'
import { UploadsPage } from '@/features/uploads/components/UploadsPage'
import { ChunkedUploadPage } from '@/features/uploads/components/ChunkedUploadPage'
import type { UploadSession } from '@/features/uploads/types'

export default function App() {
  return <AuthProvider><Application /></AuthProvider>
}

function Application() {
  const { session } = useAuth()
  if (!session) {
    return <LoginPage />
  }

  return <AuthenticatedApplication key={session.token} />
}

function AuthenticatedApplication() {
  const [page, setPage] = useState<'uploads' | 'new-upload' | 'resume'>('uploads')
  const [resume, setResume] = useState<{ upload: UploadSession; file: File } | null>(null)

  if (page === 'uploads') return <UploadsPage onNewUpload={() => setPage('new-upload')} onResumeUpload={(upload, file) => { setResume({ upload, file }); setPage('resume') }} />
  if (page === 'resume' && resume) return <ChunkedUploadPage file={resume.file} resumeUploadId={resume.upload.id} onBack={() => setPage('uploads')} onBackToUploads={() => setPage('uploads')} />
  return <UploadWorkspace onBackToUploads={() => setPage('uploads')} />
}

