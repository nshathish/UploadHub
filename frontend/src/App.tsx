import { useState } from 'react'

import { useAuth } from '@/features/auth/AuthContext'
import { AuthProvider } from '@/features/auth/AuthProvider'
import { LoginPage } from '@/features/auth/components/LoginPage'
import { UploadWorkspace } from '@/features/uploads/components/UploadWorkspace'
import { UploadsPage } from '@/features/uploads/components/UploadsPage'

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
  const [page, setPage] = useState<'uploads' | 'new-upload'>('uploads')

  return page === 'uploads' ? <UploadsPage onNewUpload={() => setPage('new-upload')} /> : <UploadWorkspace onBackToUploads={() => setPage('uploads')} />
}

