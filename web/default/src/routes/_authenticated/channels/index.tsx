import z from 'zod'
import { createFileRoute } from '@tanstack/react-router'
import { Channels } from '@/features/channels'

const channelsSearchSchema = z.object({
  page: z.number().optional().catch(1),
  pageSize: z.number().optional().catch(10),
  filter: z.string().optional().catch(''),
  status: z.array(z.string()).optional().catch([]),
  type: z.array(z.string()).optional().catch([]),
  group: z.array(z.string()).optional().catch([]),
  model: z.string().optional().catch(''),
})

export const Route = createFileRoute('/_authenticated/channels/')({
  validateSearch: channelsSearchSchema,
  component: Channels,
})
