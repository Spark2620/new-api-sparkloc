import { api } from '@/lib/api'
import type {
  ChannelEarningsParams,
  ChannelEarningsResponse,
} from './types'

export async function getChannelEarnings(
  params: ChannelEarningsParams = {}
): Promise<ChannelEarningsResponse> {
  const search = new URLSearchParams()
  if (params.keyword) search.set('keyword', params.keyword)
  if (params.p) search.set('p', String(params.p))
  if (params.page_size) search.set('page_size', String(params.page_size))
  const res = await api.get(`/api/user/channel-earnings?${search.toString()}`)
  return res.data
}
