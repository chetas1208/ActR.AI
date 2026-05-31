import { defineStore } from 'pinia'
import type { WorkflowJob } from '~/types'

export const useWorkflowStore = defineStore('workflow', () => {
  const recentJobs = ref<WorkflowJob[]>([])

  function addJob(job: WorkflowJob) {
    const idx = recentJobs.value.findIndex(j => j.id === job.id)
    if (idx >= 0) {
      recentJobs.value[idx] = job
    } else {
      recentJobs.value.unshift(job)
      if (recentJobs.value.length > 10) {
        recentJobs.value = recentJobs.value.slice(0, 10)
      }
    }
  }

  function updateJob(id: string, updates: Partial<WorkflowJob>) {
    const idx = recentJobs.value.findIndex(j => j.id === id)
    if (idx >= 0) {
      recentJobs.value[idx] = { ...recentJobs.value[idx], ...updates }
    }
  }

  return { recentJobs, addJob, updateJob }
}, {
  persist: {
    storage: typeof window !== 'undefined' ? localStorage : undefined,
  },
})
