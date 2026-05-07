import z from 'zod'
import { createFileRoute } from '@tanstack/react-router'
import { Leaderboard } from '@/features/leaderboard'

const leaderboardSearchSchema = z.object({
  type: z.enum(['consumption', 'contribution']).optional().catch('consumption'),
  page: z.number().optional().catch(1),
  pageSize: z.number().optional().catch(20),
})

export const Route = createFileRoute('/_authenticated/leaderboard/')({
  validateSearch: leaderboardSearchSchema,
  component: Leaderboard,
})
