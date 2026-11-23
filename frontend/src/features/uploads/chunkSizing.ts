const defaultBlockSize = 4 * 1024 * 1024
const maximumBlockCount = 50000

export function chunkCountForFile(size: number): number {
  const blockSize = chunkBlockSizeForFile(size)
  return Math.ceil(size / blockSize)
}

function chunkBlockSizeForFile(size: number): number {
  const minimumBlockSize = Math.ceil(size / maximumBlockCount)
  if (minimumBlockSize <= defaultBlockSize) {
    return defaultBlockSize
  }
  return Math.ceil(minimumBlockSize / defaultBlockSize) * defaultBlockSize
}
