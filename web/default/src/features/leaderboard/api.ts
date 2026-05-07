import { api } from '@/lib/api'
import type { LeaderboardParams, LeaderboardResponse } from './types'

export async function getLeaderboard(
  params: LeaderboardParams = {}
): Promise<LeaderboardResponse> {
  const { type = 'consumption', p = 1, page_size = 20 } = params
  const res = await api.get(
    `/api/user/leaderboard?type=${type}&p=${p}&page_size=${page_size}`
  )
  return res.data
}
