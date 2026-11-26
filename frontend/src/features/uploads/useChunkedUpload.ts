import { useRef, useState } from 'react';

import { uploadBlocks } from '@/api/chunkedBlobUpload';
import { completeUpload, initiateChunkedUpload, resumeChunkedUpload } from '@/api/uploads';
import type { ChunkState } from '@/features/uploads/components/ChunkProgressMap';
import type { UploadState } from '@/features/uploads/types';

export function useChunkedUpload() {
  const [state, setState] = useState<UploadState>({ phase: 'idle' });
  const [chunks, setChunks] = useState<ChunkState[]>([]);
  const abortController = useRef<AbortController | null>(null);

  async function start(file: File) {
    await upload(file);
  }

  async function retry(file: File) {
    if (state.phase !== 'error' || !state.uploadId) return;
    await upload(file, state.uploadId);
  }

  async function resume(file: File, uploadId: string) {
    await upload(file, uploadId);
  }

  async function upload(file: File, retryUploadId?: string) {
    const controller = new AbortController();
    let uploadID = retryUploadId;
    abortController.current = controller;
    setState({ phase: 'initiating' });

    try {
      const upload = retryUploadId
        ? await resumeChunkedUpload(retryUploadId)
        : await initiateChunkedUpload({ fileName: file.name, size: file.size }, crypto.randomUUID());
      uploadID = upload.id;

      if (('uploadMode' in upload && upload.uploadMode !== 'chunked') || !upload.blockSize) {
        throw new Error('The server did not prepare a chunked upload session.');
      }

      const completedChunkIndexes = 'completedChunkIndexes' in upload ? upload.completedChunkIndexes : [];
      setChunks(
        Array.from(
          { length: Math.ceil(file.size / upload.blockSize) },
          (_, index) => completedChunkIndexes.includes(index) ? 'complete' : 'queued',
        ),
      );

      setState({ phase: 'uploading', progress: 0, uploadedBytes: 0 });
      await uploadBlocks(upload.uploadUrl, upload.id, file, upload.blockSize, completedChunkIndexes, {
        signal: controller.signal,
        onProgress: (uploadedBytes, totalBytes) =>
          setState({
            phase: 'uploading',
            progress: Math.round((uploadedBytes / totalBytes) * 100),
            uploadedBytes,
          }),
        onChunkStart: (index) =>
          setChunks((current) => updateChunk(current, index, 'uploading')),
        onChunkComplete: (index) =>
          setChunks((current) => updateChunk(current, index, 'complete')),
        onChunkFailed: (index) =>
          setChunks((current) => updateChunk(current, index, 'failed')),
      });

      setState({ phase: 'verifying', uploadId: upload.id });
      await completeUpload(upload.id);
      setState({ phase: 'uploaded', uploadId: upload.id });
    } catch (error) {
      if (error instanceof DOMException && error.name === 'AbortError') {
        setState({ phase: 'canceled' });
      } else {
        setState({
          phase: 'error',
          message:
            error instanceof Error
              ? error.message
              : 'Unable to upload the document.',
          uploadId: uploadID,
        });
      }
    } finally {
      abortController.current = null;
    }
  }

  function cancel() {
    abortController.current?.abort();
  }

  function reset() {
    abortController.current?.abort();
    abortController.current = null;
    setChunks([]);
    setState({ phase: 'idle' });
  }

  return { cancel, chunks, reset, resume, retry, start, state };
}

function updateChunk(
  chunks: ChunkState[],
  index: number,
  state: ChunkState,
): ChunkState[] {
  return chunks.map((chunk, chunkIndex) =>
    chunkIndex === index ? state : chunk,
  );
}
