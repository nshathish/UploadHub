export const singleUploadLimit = 25 * 1024 * 1024

const allowedFileTypes = new Map([
  ['application/pdf', '.pdf'],
  ['image/jpeg', '.jpg'],
  ['image/png', '.png'],
])

export function validateFile(file: File): string | null {
  if (file.size === 0) {
    return 'Choose a file that is not empty.'
  }

  const extension = file.name.slice(file.name.lastIndexOf('.')).toLowerCase()
  const expectedExtension = allowedFileTypes.get(file.type)
  if (!expectedExtension || (extension !== expectedExtension && !(file.type === 'image/jpeg' && extension === '.jpeg'))) {
    return 'Choose a PDF, JPEG, or PNG document.'
  }

  return null
}

export function requiresChunkedUpload(file: File): boolean {
  return file.size > singleUploadLimit
}
