import type {
  WorkflowDetails,
  PresignUploadResponse,
  YouTubeStartResponse,
  ProviderStatus,
} from '~/types'

export function useApi() {
  const config = useRuntimeConfig()
  const baseURL = config.public.apiBaseUrl

  async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const url = `${baseURL}${path}`
    const response = await fetch(url, {
      headers: { 'Content-Type': 'application/json', ...options.headers },
      ...options,
    })
    const data = await response.json()
    if (!response.ok) {
      throw Object.assign(new Error(data.error || 'Request failed'), { status: response.status, data })
    }
    return data as T
  }

  return {
    health: () => request<{ ok: boolean; service: string }>('/api/health'),
    ready:  () => request<{ ok: boolean; checks: Record<string, unknown> }>('/api/ready'),
    providerStatus: () => request<ProviderStatus>('/api/providers/status'),

    presignUpload: (p: { filename: string; contentType: string; fileSize: number; title?: string }) =>
      request<PresignUploadResponse>('/api/uploads/presign', { method: 'POST', body: JSON.stringify(p) }),

    presignTranscript: (p: { jobId: string; filename: string; contentType: string }) =>
      request<{ uploadUrl: string; objectKey: string }>('/api/uploads/transcript/presign', { method: 'POST', body: JSON.stringify(p) }),

    startYouTubeWorkflow: (youtubeUrl: string) =>
      request<YouTubeStartResponse>('/api/workflows/youtube/start', { method: 'POST', body: JSON.stringify({ youtubeUrl }) }),

    startUploadWorkflow: (p: { jobId: string; objectKey: string; title?: string }) =>
      request<{ jobId: string; status: string }>('/api/workflows/upload/start', { method: 'POST', body: JSON.stringify(p) }),

    startDirectFileWorkflow: (p: { fileUrl: string; title?: string; sourceRightsConfirmed: boolean }) =>
      request<{ jobId: string; status: string }>('/api/workflows/direct-file/start', { method: 'POST', body: JSON.stringify(p) }),

    continueWorkflow: (jobId: string) =>
      request<{ jobId: string; status: string; progress?: number }>(`/api/workflows/${jobId}/continue`, { method: 'POST', body: '{}' }),

    getWorkflow: (jobId: string) =>
      request<WorkflowDetails>(`/api/workflows/${jobId}`),

    runAction: (actionCardId: string) =>
      request<{ runId: string; status: string }>(`/api/actions/${actionCardId}/run`, { method: 'POST', body: JSON.stringify({ mode: 'daytona' }) }),

    runRtrvr: (jobId: string, task: string, targetUrls?: string[]) =>
      request<{ browserRunId: string; status: string; outputKey?: string }>('/api/rtrvr/run', {
        method: 'POST', body: JSON.stringify({ jobId, task, targetUrls }),
      }),
  }
}

// Direct Tigris upload helper with progress
export async function uploadToTigris(uploadUrl: string, file: File, onProgress?: (pct: number) => void): Promise<void> {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest()
    xhr.open('PUT', uploadUrl)
    xhr.setRequestHeader('Content-Type', file.type)
    if (onProgress) {
      xhr.upload.onprogress = e => {
        if (e.lengthComputable) onProgress(Math.round((e.loaded / e.total) * 100))
      }
    }
    xhr.onload = () => xhr.status >= 200 && xhr.status < 300 ? resolve() : reject(new Error(`Upload failed: ${xhr.status}`))
    xhr.onerror = () => reject(new Error('Upload network error'))
    xhr.send(file)
  })
}
