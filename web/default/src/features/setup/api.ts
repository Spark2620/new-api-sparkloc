import { api } from '@/lib/api'
import type { SetupResponse } from './types'

export async function getSetupStatus(): Promise<SetupResponse> {
  const res = await api.get('/api/setup', {
    // We want fresh status on every visit.
    params: {
      t: Date.now(),
    },
  })
  return res.data
}
