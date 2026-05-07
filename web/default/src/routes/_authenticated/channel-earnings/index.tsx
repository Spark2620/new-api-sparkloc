import z from 'zod'
import { createFileRoute } from '@tanstack/react-router'
import { ChannelEarnings } from '@/features/channel-earnings'

const channelEarningsSearchSchema = z.object({
  keyword: z.string().optional().catch(''),
  page: z.coerce.number().int().min(1).optional().catch(1),
  pageSize: z.coerce.number().int().min(1).optional().catch(20),
})

export const Route = createFileRoute('/_authenticated/channel-earnings/')({
  validateSearch: channelEarningsSearchSchema,
  component: ChannelEarnings,
})
