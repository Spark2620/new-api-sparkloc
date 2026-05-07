import { Activity, BarChart3 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { formatQuota } from '@/lib/format'
import { Skeleton } from '@/components/ui/skeleton'
import { BalanceBreakdownCard } from '@/features/shared/components/balance-breakdown-card'
import type { UserWalletData } from '../types'

interface WalletStatsCardProps {
  user: UserWalletData | null
  loading?: boolean
}

export function WalletStatsCard(props: WalletStatsCardProps) {
  const { t } = useTranslation()
  if (props.loading) {
    return (
      <div className='overflow-hidden rounded-lg border'>
        <div className='p-4 sm:p-5'>
          <Skeleton className='h-4 w-24' />
          <Skeleton className='mt-3 h-10 w-40' />
          <div className='mt-4 grid gap-3 md:grid-cols-3'>
            {Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className='rounded-lg border px-3 py-3 sm:px-4'>
                <Skeleton className='h-3.5 w-24' />
                <Skeleton className='mt-2 h-6 w-24' />
                <Skeleton className='mt-2 h-10 w-full' />
              </div>
            ))}
          </div>
        </div>
        <div className='border-t'>
          <div className='divide-border/60 grid grid-cols-2 divide-x'>
            {Array.from({ length: 2 }).map((_, i) => (
              <div key={i} className='px-3 py-3 sm:px-5 sm:py-4'>
                <Skeleton className='h-3.5 w-20' />
                <Skeleton className='mt-2 h-7 w-28' />
                <Skeleton className='mt-1.5 h-3.5 w-24' />
              </div>
            ))}
          </div>
        </div>
      </div>
    )
  }

  const stats = [
    {
      label: t('Total Usage'),
      value: formatQuota(props.user?.used_quota ?? 0),
      description: t('Total consumed quota'),
      icon: BarChart3,
    },
    {
      label: t('API Requests'),
      value: (props.user?.request_count ?? 0).toLocaleString(),
      description: t('Total requests made'),
      icon: Activity,
    },
  ]

  return (
    <div className='overflow-hidden rounded-lg border'>
      <div className='p-4 sm:p-5'>
        <BalanceBreakdownCard data={props.user} nested />
      </div>
      <div className='border-t'>
        <div className='divide-border/60 grid grid-cols-2 divide-x'>
          {stats.map((item) => (
            <div key={item.label} className='px-3 py-3 sm:px-5 sm:py-4'>
              <div className='flex items-center gap-2'>
                <item.icon className='text-muted-foreground/60 size-3.5 shrink-0' />
                <div className='text-muted-foreground truncate text-xs font-medium tracking-wider uppercase'>
                  {item.label}
                </div>
              </div>

              <div className='text-foreground mt-1.5 font-mono text-base font-bold tracking-tight break-all tabular-nums sm:mt-2 sm:text-2xl'>
                {item.value}
              </div>
              <div className='text-muted-foreground/60 mt-1 hidden text-xs md:block'>
                {item.description}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
