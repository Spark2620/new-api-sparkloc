import { api } from '@/lib/api'
import type {
  ChannelAvailabilityParams,
  ChannelAvailabilityResponse,
} from './types'

export async function getChannelAvailability(
  params: ChannelAvailabilityParams = {}
): Promise<ChannelAvailabilityResponse> {
  const { keyword = '', p = 1, page_size = 20 } = params
  const search = new URLSearchParams({
    keyword,
    p: String(p),
    page_size: String(page_size),
  })
  const res = await api.get(`/api/user/channel-availability?${search.toString()}`)
  return res.data
}
