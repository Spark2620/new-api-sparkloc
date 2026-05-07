export interface ChannelEarningsItem {
  created_at: number
  channel_id: number
  channel_name: string
  consumer: string
  self_use: boolean
  model_name: string
  quota: number
  prompt_tokens: number
  completion_tokens: number
  total_tokens: number
  payout_amount: number
  payout_daily_amount: number
  payout_earned_amount: number
}

export interface ChannelEarningsSummary {
  owned_channels: number
  matched_requests: number
  matched_consumption: number
  total_earnings: number
}

export interface ChannelEarningsResponse {
  success: boolean
  message?: string
  data?: {
    page: number
    page_size: number
    total: number
    keyword?: string
    items: ChannelEarningsItem[]
    summary?: ChannelEarningsSummary
  }
}

export interface ChannelEarningsParams {
  keyword?: string
  p?: number
  page_size?: number
}
