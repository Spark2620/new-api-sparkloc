import { useEffect, useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  type ColumnDef,
  flexRender,
  getCoreRowModel,
  useReactTable,
} from '@tanstack/react-table'
import { getRouteApi } from '@tanstack/react-router'
import { Trophy } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
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
import { PageFooterPortal } from '@/components/layout'
import { useMediaQuery } from '@/hooks'
import { useTableUrlState } from '@/hooks/use-table-url-state'
import { formatQuota } from '@/lib/format'
import { cn } from '@/lib/utils'
import { getLeaderboard } from '../api'
import type { LeaderboardEntry, LeaderboardType } from '../types'

const route = getRouteApi('/_authenticated/leaderboard/')

function getInitials(entry: LeaderboardEntry) {
  const source = entry.display_name?.trim() || entry.username.trim()
  if (!source) return 'U'
  return source.slice(0, 2).toUpperCase()
}

function getAvatarStyle(seed: string) {
  let hash = 0
  for (let i = 0; i < seed.length; i++) {
    hash = (hash * 31 + seed.charCodeAt(i)) >>> 0
  }
  const hue = hash % 360
  return {
    backgroundColor: `hsl(${hue} 62% 56% / 0.86)`,
    color: 'white',
  }
}

export function LeaderboardTable(props: { type: LeaderboardType }) {
  const { t } = useTranslation()
  const isMobile = useMediaQuery('(max-width: 640px)')
  const [columnVisibility, setColumnVisibility] = useState({})

  const {
    pagination,
    onPaginationChange,
    ensurePageInRange,
  } = useTableUrlState({
    search: route.useSearch(),
    navigate: route.useNavigate(),
    pagination: { defaultPage: 1, defaultPageSize: isMobile ? 10 : 20 },
  })

  const { data, isLoading, isFetching } = useQuery({
    queryKey: [
      'leaderboard',
      props.type,
      pagination.pageIndex + 1,
      pagination.pageSize,
    ],
    queryFn: async () => {
      const result = await getLeaderboard({
        type: props.type,
        p: pagination.pageIndex + 1,
        page_size: pagination.pageSize,
      })
      if (!result.success) {
        return { items: [], total: 0 }
      }
      return {
        items: result.data?.items || [],
        total: result.data?.total || 0,
      }
    },
    placeholderData: (previousData) => previousData,
  })

  const columns = useMemo<ColumnDef<LeaderboardEntry>[]>(
    () => [
      {
        accessorKey: 'rank',
        header: t('Rank'),
        meta: { label: t('Rank'), mobileBadge: true },
        cell: ({ row }) => {
          const rank = row.original.rank
          const accent =
            rank === 1
              ? 'text-amber-600 dark:text-amber-300'
              : rank === 2
                ? 'text-slate-600 dark:text-slate-300'
                : rank === 3
                  ? 'text-orange-700 dark:text-orange-300'
                  : 'text-muted-foreground'

          return (
            <div
              className={cn(
                'inline-flex items-center gap-1.5 font-mono text-xs font-semibold',
                accent
              )}
            >
              {rank <= 3 && <Trophy className='size-3.5' />}
              <span>#{rank}</span>
            </div>
          )
        },
      },
      {
        id: 'user',
        header: t('User'),
        meta: { label: t('User'), mobileTitle: true },
        cell: ({ row }) => {
          const entry = row.original
          const displayName = entry.display_name?.trim() || entry.username
          return (
            <div className='flex min-w-0 items-center gap-2.5'>
              <Avatar className='size-8 shrink-0 ring-1 ring-border/60'>
                <AvatarFallback
                  className='text-[11px] font-semibold'
                  style={getAvatarStyle(entry.username)}
                >
                  {getInitials(entry)}
                </AvatarFallback>
              </Avatar>
              <div className='min-w-0'>
                <div className='flex min-w-0 items-center gap-2'>
                  <span className='truncate text-sm font-medium'>
                    {displayName}
                  </span>
                  {entry.is_current_user && (
                    <span className='bg-primary/10 text-primary rounded px-1.5 py-0.5 text-[10px] font-medium'>
                      {t('You')}
                    </span>
                  )}
                </div>
                <div className='text-muted-foreground mt-0.5 truncate text-xs'>
                  @{entry.username}
                </div>
              </div>
            </div>
          )
        },
      },
      {
        accessorKey: 'value',
        header:
          props.type === 'contribution'
            ? t('Total Contribution')
            : t('Total Consumption'),
        meta: {
          label:
            props.type === 'contribution'
              ? t('Total Contribution')
              : t('Total Consumption'),
        },
        cell: ({ row }) => (
          <span className='font-mono text-sm font-semibold tabular-nums'>
            {formatQuota(row.original.value)}
          </span>
        ),
      },
    ],
    [props.type, t]
  )

  const items = data?.items || []
  const table = useReactTable({
    data: items,
    columns,
    state: {
      pagination,
      columnVisibility,
    },
    onColumnVisibilityChange: setColumnVisibility,
    getCoreRowModel: getCoreRowModel(),
    onPaginationChange,
    manualPagination: true,
    pageCount: Math.max(1, Math.ceil((data?.total || 0) / pagination.pageSize)),
  })

  const pageCount = table.getPageCount()
  useEffect(() => {
    ensurePageInRange(pageCount)
  }, [pageCount, ensurePageInRange])

  return (
    <>
      <div className='space-y-3 sm:space-y-4'>
        {isMobile ? (
          <MobileCardList
            table={table}
            isLoading={isLoading}
            emptyTitle={t('No leaderboard data')}
            emptyDescription={t(
              'No ranking data is available yet. Once users start consuming or contributing channels, rankings will appear here.'
            )}
            getRowClassName={(row) =>
              row.original.is_current_user
                ? 'bg-primary/5 border-l-4 border-l-primary'
                : undefined
            }
          />
        ) : (
          <div
            className={cn(
              'overflow-hidden rounded-md border transition-opacity duration-150',
              isFetching && !isLoading && 'pointer-events-none opacity-50'
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
                  <TableSkeleton
                    table={table}
                    keyPrefix={`leaderboard-${props.type}-skeleton`}
                  />
                ) : table.getRowModel().rows.length === 0 ? (
                  <TableEmpty
                    colSpan={columns.length}
                    title={t('No leaderboard data')}
                    description={t(
                      'No ranking data is available yet. Once users start consuming or contributing channels, rankings will appear here.'
                    )}
                  />
                ) : (
                  table.getRowModel().rows.map((row) => (
                    <TableRow
                      key={row.id}
                      className={cn(
                        row.original.is_current_user && 'bg-primary/5'
                      )}
                    >
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
      <PageFooterPortal>
        <DataTablePagination table={table} />
      </PageFooterPortal>
    </>
  )
}
