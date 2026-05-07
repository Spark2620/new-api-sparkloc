import { useEffect, useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { getRouteApi, useNavigate } from '@tanstack/react-router'
import {
  type ColumnDef,
  flexRender,
  getCoreRowModel,
  useReactTable,
} from '@tanstack/react-table'
import { Coins, Search } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { PageFooterPortal, SectionPageLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  DataTablePagination,
  MobileCardList,
  TableEmpty,
  TableSkeleton,
} from '@/components/data-table'
import { useMediaQuery } from '@/hooks'
import { useTableUrlState } from '@/hooks/use-table-url-state'
import {
  formatCompactNumber,
  formatLogQuota,
  formatQuota,
  formatTimestampToDate,
} from '@/lib/format'
import { cn } from '@/lib/utils'
import { getChannelEarnings } from './api'
import type { ChannelEarningsItem } from './types'

const route = getRouteApi('/_authenticated/channel-earnings/')

function SummaryCard(props: {
  label: string
  value: string
  hint?: string
}) {
  return (
    <div className='rounded-md border bg-card/60 px-4 py-3'>
      <div className='text-muted-foreground text-xs'>{props.label}</div>
      <div className='mt-1 text-lg font-semibold tabular-nums'>
        {props.value}
      </div>
      {props.hint ? (
        <div className='text-muted-foreground mt-1 text-xs'>{props.hint}</div>
      ) : null}
    </div>
  )
}

export function ChannelEarnings() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const search = route.useSearch()
  const keyword = search.keyword ?? ''
  const [inputValue, setInputValue] = useState(keyword)
  const isMobile = useMediaQuery('(max-width: 640px)')

  useEffect(() => {
    setInputValue(keyword)
  }, [keyword])

  const { pagination, onPaginationChange, ensurePageInRange } = useTableUrlState(
    {
      search: route.useSearch(),
      navigate: route.useNavigate(),
      pagination: { defaultPage: 1, defaultPageSize: isMobile ? 10 : 20 },
    }
  )

  const currentPage = pagination.pageIndex + 1

  const { data, isLoading, isFetching } = useQuery({
    queryKey: [
      'channel-earnings',
      keyword,
      currentPage,
      pagination.pageSize,
    ],
    queryFn: async () => {
      const result = await getChannelEarnings({
        keyword,
        p: currentPage,
        page_size: pagination.pageSize,
      })
      if (!result.success) {
        return {
          items: [],
          total: 0,
          summary: {
            owned_channels: 0,
            matched_requests: 0,
            matched_consumption: 0,
            total_earnings: 0,
          },
        }
      }
      return {
        items: result.data?.items || [],
        total: result.data?.total || 0,
        summary: result.data?.summary || {
          owned_channels: 0,
          matched_requests: 0,
          matched_consumption: 0,
          total_earnings: 0,
        },
      }
    },
    placeholderData: (previousData) => previousData,
  })

  const columns = useMemo<ColumnDef<ChannelEarningsItem>[]>(
    () => [
      {
        accessorKey: 'created_at',
        header: t('Time'),
        meta: { label: t('Time'), mobileBadge: true },
        cell: ({ row }) => (
          <span className='font-mono text-xs tabular-nums'>
            {formatTimestampToDate(row.original.created_at)}
          </span>
        ),
      },
      {
        id: 'channel',
        header: t('Channel'),
        meta: { label: t('Channel'), mobileTitle: true },
        cell: ({ row }) => (
          <div className='min-w-0'>
            <div className='truncate text-sm font-medium'>
              {row.original.channel_name || `#${row.original.channel_id}`}
            </div>
            <div className='text-muted-foreground mt-0.5 text-xs'>
              #{row.original.channel_id}
            </div>
          </div>
        ),
      },
      {
        accessorKey: 'consumer',
        header: t('User'),
        meta: { label: t('User') },
        cell: ({ row }) => (
          <div className='min-w-0'>
            <div className='truncate text-sm'>{row.original.consumer}</div>
            <div className='text-muted-foreground mt-0.5 text-xs'>
              {row.original.self_use ? t('Self use') : t('Community user')}
            </div>
          </div>
        ),
      },
      {
        accessorKey: 'model_name',
        header: t('Model'),
        meta: { label: t('Model') },
        cell: ({ row }) => (
          <span className='font-mono text-xs'>{row.original.model_name || '-'}</span>
        ),
      },
      {
        accessorKey: 'total_tokens',
        header: 'Tokens',
        meta: { label: 'Tokens', mobileHidden: true },
        cell: ({ row }) => (
          <div className='text-sm'>
            <div className='font-mono tabular-nums'>
              {formatCompactNumber(row.original.total_tokens)}
            </div>
            <div className='text-muted-foreground mt-0.5 text-xs'>
              {formatCompactNumber(row.original.prompt_tokens)} /{' '}
              {formatCompactNumber(row.original.completion_tokens)}
            </div>
          </div>
        ),
      },
      {
        accessorKey: 'quota',
        header: t('Consumption'),
        meta: { label: t('Consumption') },
        cell: ({ row }) => (
          <span className='font-mono text-sm font-medium tabular-nums'>
            {formatLogQuota(row.original.quota)}
          </span>
        ),
      },
      {
        accessorKey: 'payout_amount',
        header: t('Channel Earnings'),
        meta: { label: t('Channel Earnings') },
        cell: ({ row }) => (
          <div className='text-sm'>
            <div className='font-mono font-semibold tabular-nums text-emerald-600 dark:text-emerald-400'>
              {formatQuota(row.original.payout_amount)}
            </div>
            <div className='text-muted-foreground mt-0.5 text-xs'>
              {t('Daily')}: {formatQuota(row.original.payout_daily_amount)} /{' '}
              {t('Earned')}: {formatQuota(row.original.payout_earned_amount)}
            </div>
          </div>
        ),
      },
    ],
    [t]
  )

  const items = data?.items || []
  const table = useReactTable({
    data: items,
    columns,
    state: {
      pagination,
    },
    getCoreRowModel: getCoreRowModel(),
    onPaginationChange,
    manualPagination: true,
    pageCount: Math.max(1, Math.ceil((data?.total || 0) / pagination.pageSize)),
  })

  const pageCount = table.getPageCount()
  useEffect(() => {
    ensurePageInRange(pageCount)
  }, [pageCount, ensurePageInRange])

  const handleSearchSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    const trimmed = inputValue.trim()
    void navigate({
      to: '/channel-earnings',
      search: (prev) => ({
        ...prev,
        keyword: trimmed || undefined,
        page: undefined,
        pageSize: prev.pageSize,
      }),
    })
  }

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>{t('Channel Earnings')}</SectionPageLayout.Title>
      <SectionPageLayout.Description>
        {t(
          'View how community members use your channels and how much balance those requests transfer to you. This page hides request content and only keeps the operational data needed to manage supply.'
        )}
      </SectionPageLayout.Description>
      <SectionPageLayout.Actions>
        <form onSubmit={handleSearchSubmit} className='flex w-full gap-2 sm:w-auto'>
          <div className='relative min-w-0 flex-1 sm:w-80'>
            <Search className='text-muted-foreground absolute top-1/2 left-3 size-4 -translate-y-1/2' />
            <Input
              value={inputValue}
              onChange={(event) => setInputValue(event.target.value)}
              placeholder={t('Search by channel, model, user, request ID, or channel ID')}
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
          <div className='grid gap-3 md:grid-cols-4'>
            <SummaryCard
              label={t('Owned Channels')}
              value={String(data?.summary?.owned_channels || 0)}
            />
            <SummaryCard
              label={t('Matched Requests')}
              value={String(data?.summary?.matched_requests || 0)}
            />
            <SummaryCard
              label={t('Matched Consumption')}
              value={formatQuota(data?.summary?.matched_consumption || 0)}
            />
            <SummaryCard
              label={t('Total Channel Earnings')}
              value={formatQuota(data?.summary?.total_earnings || 0)}
              hint={t('Includes all historical channel payout credits')}
            />
          </div>

          <div className='bg-muted/30 flex items-center gap-2 rounded-md border px-4 py-3 text-sm'>
            <Coins className='text-primary size-4' />
            <span>
              {t(
                'Only your own channels are included. Request contents, API keys, IPs, and raw identifiers are not shown here.'
              )}
            </span>
          </div>

          {isMobile ? (
            <MobileCardList
              table={table}
              isLoading={isLoading}
              emptyTitle={t('No channel earnings data')}
              emptyDescription={t(
                'No matched requests were found for your channels yet.'
              )}
            />
          ) : (
            <div
              className={cn(
                'overflow-hidden rounded-md border transition-opacity duration-150',
                isFetching && !isLoading && 'pointer-events-none opacity-60'
              )}
            >
              <Table>
                <TableHeader>
                  {table.getHeaderGroups().map((headerGroup) => (
                    <TableRow key={headerGroup.id}>
                      {headerGroup.headers.map((header) => (
                        <TableHead key={header.id}>
                          {header.isPlaceholder
                            ? null
                            : flexRender(
                                header.column.columnDef.header,
                                header.getContext()
                              )}
                        </TableHead>
                      ))}
                    </TableRow>
                  ))}
                </TableHeader>
                <TableBody>
                  {isLoading ? (
                    <TableSkeleton table={table} keyPrefix='channel-earnings-skeleton' />
                  ) : table.getRowModel().rows.length === 0 ? (
                    <TableEmpty
                      colSpan={columns.length}
                      title={t('No channel earnings data')}
                      description={t(
                        'No matched requests were found for your channels yet.'
                      )}
                    />
                  ) : (
                    table.getRowModel().rows.map((row) => (
                      <TableRow key={row.id}>
                        {row.getVisibleCells().map((cell) => (
                          <TableCell key={cell.id}>
                            {flexRender(
                              cell.column.columnDef.cell,
                              cell.getContext()
                            )}
                          </TableCell>
                        ))}
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </div>
          )}
        </div>
      </SectionPageLayout.Content>
      <PageFooterPortal>
        <DataTablePagination table={table} />
      </PageFooterPortal>
    </SectionPageLayout>
  )
}
