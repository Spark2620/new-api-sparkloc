import { z } from 'zod'

export const channelAvailabilityMinuteSchema = z.object({
  bucket_start: z.number(),
  success_count: z.number(),
  failure_count: z.number(),
  avg_latency_ms: z.number(),
  availability: z.number(),
  status: z.enum(['idle', 'healthy', 'degraded', 'failed']),
  has_traffic: z.boolean(),
  total_requests: z.number(),
})

export const channelAvailabilityItemSchema = z.object({
  channel_id: z.number(),
  name: z.string(),
  type: z.number(),
  type_name: z.string(),
  creator_user_id: z.number(),
  creator_username: z.string(),
  supply_ratio: z.number(),
  status: z.number(),
  overall_availability: z.number(),
  success_count: z.number(),
  failure_count: z.number(),
  avg_latency_ms: z.number(),
  last_traffic_at: z.number(),
  has_traffic: z.boolean(),
  minutes: z.array(channelAvailabilityMinuteSchema),
})

export type ChannelAvailabilityMinute = z.infer<
  typeof channelAvailabilityMinuteSchema
>
export type ChannelAvailabilityItem = z.infer<
  typeof channelAvailabilityItemSchema
>

export interface ChannelAvailabilityParams {
  keyword?: string
  p?: number
  page_size?: number
}

export interface ChannelAvailabilityResponse {
  success: boolean
  message?: string
  data?: {
    page: number
    page_size: number
    total: number
    keyword: string
    items: ChannelAvailabilityItem[]
    window: {
      minutes: number
    }
  }
}
