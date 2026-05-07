import { useEffect, useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { getRouteApi, useNavigate } from '@tanstack/react-router'
import { Activity, Search } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { SectionPageLayout } from '@/components/layout'
import { StatusBadge } from '@/components/status-badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { cn } from '@/lib/utils'
import { getChannelAvailability } from './api'
import type {
  ChannelAvailabilityItem,
  ChannelAvailabilityMinute,
} from './types'

const route = getRouteApi('/_authenticated/channel-availability/')
const defaultPageSize = 20

function formatPercent(value: number) {
  return `${(value * 100).toFixed(1)}%`
}

function formatLatency(value: number) {
  if (!Number.isFinite(value) || value <= 0) {
    return '--'
  }
  return `${Math.round(value)} ms`
}

function formatMinuteLabel(ts: number) {
  return new Intl.DateTimeFormat(undefined, {
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(ts * 1000))
}

function formatSecondLabel(ts: number) {
  return new Intl.DateTimeFormat(undefined, {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  }).format(new Date(ts * 1000))
}

function minuteBarClassName(minute: ChannelAvailabilityMinute) {
  if (!minute.has_traffic) {
    return 'bg-muted/70'
  }
  switch (minute.status) {
    case 'healthy':
      return 'bg-emerald-500/90'
    case 'degraded':
      return 'bg-amber-500/90'
    case 'failed':
      return 'bg-rose-500/90'
    default:
      return 'bg-muted/70'
  }
}

function ChannelAvailabilityRow({
  item,
}: {
  item: ChannelAvailabilityItem
}) {
  const { t } = useTranslation()
  const totalRequests = item.success_count + item.failure_count
  const trafficRange = useMemo(() => {
    const first = item.minutes[0]
    const last = item.minutes[item.minutes.length - 1]
    if (!first || !last) {
      return null
    }
    return {
      start: formatMinuteLabel(first.bucket_start),
      end: formatMinuteLabel(last.bucket_start),
    }
  }, [item.minutes])

  return (
    <div className='grid gap-4 border-b px-4 py-4 last:border-b-0 lg:grid-cols-[120px_260px_minmax(0,1fr)] lg:gap-6'>
      <div className='space-y-1.5'>
        <div className='text-foreground text-2xl font-semibold tabular-nums'>
          {item.has_traffic ? formatPercent(item.overall_availability) : '--'}
        </div>
        <div className='text-muted-foreground text-xs'>
          {item.has_traffic
            ? t('{{count}} requests', { count: totalRequests })
            : t('No traffic yet')}
        </div>
        <div className='text-muted-foreground text-xs tabular-nums'>
          {t('Avg latency')}: {formatLatency(item.avg_latency_ms)}
        </div>
      </div>

      <div className='min-w-0 space-y-2'>
        <div className='flex min-w-0 items-center gap-2'>
          <span className='truncate text-sm font-medium'>{item.name}</span>
          <StatusBadge
            label={item.type_name}
            autoColor={item.type_name}
            size='sm'
            copyable={false}
          />
        </div>
        <div className='text-muted-foreground flex flex-wrap items-center gap-x-3 gap-y-1 text-xs'>
          <span>
            {t('Channel Creator')}: @{item.creator_username || '-'}
          </span>
          <span>
            {t('Channel Ratio')}: {item.supply_ratio.toFixed(2)}x
          </span>
          <span>
            {t('Success')}: {item.success_count}
          </span>
          <span>
            {t('Failed')}: {item.failure_count}
          </span>
        </div>
        <div className='text-muted-foreground text-xs'>
          {item.last_traffic_at > 0
            ? t('Last traffic: {{time}}', {
                time: formatSecondLabel(item.last_traffic_at),
              })
            : t('No traffic in the last hour')}
        </div>
      </div>

      <div className='space-y-2'>
        <div className='flex items-center justify-between gap-3'>
          <div className='text-muted-foreground flex items-center gap-2 text-xs'>
            <StatusBadge
              label={t('Healthy')}
              variant='success'
              showDot
              copyable={false}
            />
            <StatusBadge
              label={t('Degraded')}
              variant='warning'
              showDot
              copyable={false}
            />
            <StatusBadge
              label={t('Failed')}
              variant='danger'
              showDot
              copyable={false}
            />
            <StatusBadge
              label={t('Idle')}
              variant='neutral'
              showDot
              copyable={false}
            />
          </div>
          {trafficRange ? (
            <div className='text-muted-foreground hidden text-[11px] sm:block'>
              {trafficRange.start} - {trafficRange.end}
            </div>
          ) : null}
        </div>
        <div className='overflow-x-auto pb-1'>
          <div className='flex min-w-[720px] gap-1'>
            {item.minutes.map((minute) => (
              <Tooltip key={`${item.channel_id}-${minute.bucket_start}`}>
                <TooltipTrigger asChild>
                  <div
                    className={cn(
                      'h-10 flex-1 rounded-[4px] transition-opacity hover:opacity-85',
                      minuteBarClassName(minute)
                    )}
                  />
                </TooltipTrigger>
                <TooltipContent side='top' sideOffset={6} className='space-y-1'>
                  <div className='font-medium'>
                    {formatMinuteLabel(minute.bucket_start)}
                  </div>
                  <div>{t('Requests')}: {minute.total_requests}</div>
                  <div>{t('Success')}: {minute.success_count}</div>
                  <div>{t('Failed')}: {minute.failure_count}</div>
                  <div>
                    {t('Availability')}: {formatPercent(minute.availability)}
                  </div>
                  <div>{t('Avg latency')}: {formatLatency(minute.avg_latency_ms)}</div>
                </TooltipContent>
              </Tooltip>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}

function ChannelAvailabilitySkeleton() {
  return (
    <div className='divide-y rounded-md border'>
      {Array.from({ length: 5 }).map((_, index) => (
        <div
          key={`channel-availability-skeleton-${index}`}
          className='grid gap-4 px-4 py-4 lg:grid-cols-[120px_260px_minmax(0,1fr)] lg:gap-6'
        >
          <div className='space-y-2'>
            <Skeleton className='h-8 w-20' />
            <Skeleton className='h-4 w-24' />
            <Skeleton className='h-4 w-18' />
          </div>
          <div className='space-y-2'>
            <Skeleton className='h-5 w-32' />
            <Skeleton className='h-4 w-44' />
            <Skeleton className='h-4 w-40' />
          </div>
          <div className='space-y-2'>
            <Skeleton className='h-4 w-48' />
            <div className='flex gap-1'>
              {Array.from({ length: 20 }).map((__, minuteIndex) => (
                <Skeleton
                  key={`channel-availability-skeleton-bar-${index}-${minuteIndex}`}
                  className='h-10 flex-1 rounded-[4px]'
                />
              ))}
            </div>
          </div>
        </div>
      ))}
    </div>
  )
}

export function ChannelAvailability() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const search = route.useSearch()
  const page = search.page ?? 1
  const pageSize = search.pageSize ?? defaultPageSize
  const keyword = search.keyword ?? ''
  const [inputValue, setInputValue] = useState(keyword)

  useEffect(() => {
    setInputValue(keyword)
  }, [keyword])

  const { data, isLoading, isFetching } = useQuery({
    queryKey: ['channel-availability', keyword, page, pageSize],
    queryFn: async () => {
      const result = await getChannelAvailability({
        keyword,
        p: page,
        page_size: pageSize,
      })
      if (!result.success) {
        return {
          items: [],
          total: 0,
          windowMinutes: 60,
        }
      }
      return {
        items: result.data?.items || [],
        total: result.data?.total || 0,
        windowMinutes: result.data?.window?.minutes || 60,
      }
    },
    placeholderData: (previousData) => previousData,
  })

  const items = data?.items || []
  const total = data?.total || 0
  const totalPages = Math.max(1, Math.ceil(total / pageSize))

  const handleSearchSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    const trimmed = inputValue.trim()
    void navigate({
      to: '/channel-availability',
      search: (prev) => ({
        ...prev,
        keyword: trimmed || undefined,
        page: undefined,
        pageSize: prev.pageSize,
      }),
    })
  }

  const handlePageChange = (nextPage: number) => {
    void navigate({
      to: '/channel-availability',
      search: (prev) => ({
        ...prev,
        keyword: prev.keyword,
        page: nextPage <= 1 ? undefined : nextPage,
        pageSize: prev.pageSize,
      }),
    })
  }

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>
        {t('Channel Availability')}
      </SectionPageLayout.Title>
      <SectionPageLayout.Description>
        {t(
          'Observe the last hour of real channel traffic. Only enabled channels are shown, disabled channels are excluded, and the system never sends active probe requests here.'
        )}
      </SectionPageLayout.Description>
      <SectionPageLayout.Actions>
        <form onSubmit={handleSearchSubmit} className='flex w-full gap-2 sm:w-auto'>
          <div className='relative min-w-0 flex-1 sm:w-72'>
            <Search className='text-muted-foreground absolute top-1/2 left-3 size-4 -translate-y-1/2' />
            <Input
              value={inputValue}
              onChange={(event) => setInputValue(event.target.value)}
              placeholder={t('Search by channel, creator, or ID')}
              className='pl-9'
            />
          </div>
          <Button type='submit' variant='outline' disabled={isFetching}>
            {t('Search')}
          </Button>
        </form>
      </SectionPageLayout.Actions>
      <SectionPageLayout.Content>
        <div className='space-y-4'>
          <div className='bg-muted/30 flex flex-wrap items-center justify-between gap-3 rounded-md border px-4 py-3 text-sm'>
            <div className='flex items-center gap-2'>
              <Activity className='text-primary size-4' />
              <span>{t('{{count}} enabled channels', { count: total })}</span>
            </div>
            <div className='text-muted-foreground text-xs sm:text-sm'>
              {t('Window')}: {data?.windowMinutes || 60} {t('minutes')}
            </div>
          </div>

          {isLoading ? (
            <ChannelAvailabilitySkeleton />
          ) : items.length === 0 ? (
            <div className='rounded-md border px-6 py-12 text-center'>
              <div className='text-base font-medium'>
                {t('No channel availability data')}
              </div>
              <div className='text-muted-foreground mt-2 text-sm'>
                {t(
                  'No enabled channels matched the current search, or no eligible channels are available to display.'
                )}
              </div>
            </div>
          ) : (
            <div
              className={cn(
                'overflow-hidden rounded-md border transition-opacity',
                isFetching && 'opacity-70'
              )}
            >
              {items.map((item) => (
                <ChannelAvailabilityRow key={item.channel_id} item={item} />
              ))}
            </div>
          )}

          <div className='flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between'>
            <div className='text-muted-foreground text-sm'>
              {t('Page {{page}} of {{totalPages}}', {
                page,
                totalPages,
              })}
            </div>
            <div className='flex items-center gap-2'>
              <Button
                variant='outline'
                size='sm'
                disabled={page <= 1 || isFetching}
                onClick={() => handlePageChange(page - 1)}
              >
                {t('Previous page')}
              </Button>
              <Button
                variant='outline'
                size='sm'
                disabled={page >= totalPages || isFetching}
                onClick={() => handlePageChange(page + 1)}
              >
                {t('Next page')}
              </Button>
            </div>
          </div>
        </div>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
