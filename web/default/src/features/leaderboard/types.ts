import { z } from 'zod'

export const leaderboardTypeSchema = z.enum(['consumption', 'contribution'])
export type LeaderboardType = z.infer<typeof leaderboardTypeSchema>

export const leaderboardEntrySchema = z.object({
  rank: z.number(),
  user_id: z.number(),
  username: z.string(),
  display_name: z.string().optional(),
  role: z.number(),
  group: z.string(),
  value: z.number(),
  value_label: z.string().optional(),
  is_current_user: z.boolean().optional(),
})
export type LeaderboardEntry = z.infer<typeof leaderboardEntrySchema>

export interface LeaderboardParams {
  type?: LeaderboardType
  p?: number
  page_size?: number
}

export interface LeaderboardResponse {
  success: boolean
  message?: string
  data?: {
    type: LeaderboardType
    page: number
    page_size: number
    total: number
    items: LeaderboardEntry[]
  }
}
