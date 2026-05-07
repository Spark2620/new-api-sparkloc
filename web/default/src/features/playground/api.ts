import { api } from '@/lib/api'
import { getChannels } from '@/features/channels/api'
import { getChannelTypeLabel } from '@/features/channels/lib'
import { API_ENDPOINTS } from './constants'
import type {
  ChatCompletionRequest,
  ChatCompletionResponse,
  ModelOption,
  GroupOption,
} from './types'

/**
 * Send chat completion request (non-streaming)
 */
export async function sendChatCompletion(
  payload: ChatCompletionRequest
): Promise<ChatCompletionResponse> {
  const res = await api.post(API_ENDPOINTS.CHAT_COMPLETIONS, payload, {
    skipErrorHandler: true,
  } as Record<string, unknown>)
  return res.data
}

/**
 * Get user available models
 */
export async function getUserModels(): Promise<ModelOption[]> {
  const res = await api.get(API_ENDPOINTS.USER_MODELS)
  const { data } = res

  if (!data.success || !Array.isArray(data.data)) {
    return []
  }

  return data.data.map((model: string) => ({
    label: model,
    value: model,
  }))
}

/**
 * Get user groups
 */
export async function getUserGroups(
  t?: (key: string) => string
): Promise<GroupOption[]> {
  const response = await getChannels({ status: 'enabled', page_size: 100 })
  const translate = t || ((key: string) => key)
  const autoOption: GroupOption = {
    label: translate('Auto'),
    value: 'auto',
    ratio: 1,
    desc: translate('Automatically select an available channel'),
  }

  if (!response.success || !response.data?.items) {
    return [autoOption]
  }

  const channelOptions = response.data.items.map((channel) => {
    const creator =
      channel.owner_username ||
      (channel.owner_user_id > 0
        ? `#${channel.owner_user_id}`
        : translate('Unknown'))
    const modelsPreview = (channel.models || '')
      .split(',')
      .map((model) => model.trim())
      .filter(Boolean)
      .slice(0, 3)
      .join(', ')
    const details = [
      `${translate('Channel Creator')}: ${creator}`,
      translate(getChannelTypeLabel(channel.type)),
      channel.tag || undefined,
      modelsPreview || undefined,
    ].filter(Boolean)

    return {
      label: channel.name,
      value: `channel-${channel.id}`,
      ratio: channel.supply_ratio || 1,
      creator,
      desc: details.join(' - '),
    }
  })

  return [autoOption, ...channelOptions]
}
