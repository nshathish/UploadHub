const maxFileSize = 25 * 1024 * 1024

const allowedFileTypes = new Map([
  ['application/pdf', '.pdf'],
  ['image/jpeg', '.jpg'],
  ['image/png', '.png'],
])

export function validateFile(file: File): string | null {
  if (file.size === 0) {
    return 'Choose a file that is not empty.'
  }

  if (file.size > maxFileSize) {
    return 'Choose a file smaller than 25 MiB.'
  }

  const extension = file.name.slice(file.name.lastIndexOf('.')).toLowerCase()
  const expectedExtension = allowedFileTypes.get(file.type)
  if (!expectedExtension || (extension !== expectedExtension && !(file.type === 'image/jpeg' && extension === '.jpeg'))) {
    return 'Choose a PDF, JPEG, or PNG document.'
  }

  return null
}
