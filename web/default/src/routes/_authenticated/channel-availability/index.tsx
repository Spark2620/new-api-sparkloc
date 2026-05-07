import z from 'zod'
import { createFileRoute } from '@tanstack/react-router'
import { ChannelAvailability } from '@/features/channel-availability'

const channelAvailabilitySearchSchema = z.object({
  keyword: z.string().optional().catch(''),
  page: z.number().optional().catch(1),
  pageSize: z.number().optional().catch(20),
})

export const Route = createFileRoute('/_authenticated/channel-availability/')({
  validateSearch: channelAvailabilitySearchSchema,
  component: ChannelAvailability,
})
