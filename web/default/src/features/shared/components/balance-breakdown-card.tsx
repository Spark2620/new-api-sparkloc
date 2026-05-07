import { Info, WalletCards } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { formatQuota, formatTimestampToDate } from '@/lib/format'
import { Button } from '@/components/ui/button'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'

export interface BalanceBreakdownData {
  quota?: number
  available_quota?: number
  daily_credit?: number
  earned_credit?: number
  legacy_quota?: number
  daily_credit_expires_at?: number
}

interface BalanceBreakdownCardProps {
  data: BalanceBreakdownData | null | undefined
  compact?: boolean
  nested?: boolean
}

export function BalanceBreakdownCard({
  data,
  compact = false,
  nested = false,
}: BalanceBreakdownCardProps) {
  const { t } = useTranslation()

  const totalBalance = data?.available_quota ?? data?.quota ?? 0
  const dailyCredit = data?.daily_credit ?? 0
  const earnedCredit = data?.earned_credit ?? 0
  const legacyQuota = data?.legacy_quota ?? 0
  const dailyResetAt = data?.daily_credit_expires_at

  const items = [
    {
      label: t('Current daily credit'),
      value: formatQuota(dailyCredit),
      description:
        dailyResetAt && dailyResetAt > 0
          ? t('Daily check-in credit resets at {{time}}', {
              time: formatTimestampToDate(dailyResetAt),
            })
          : t('Daily check-in credit resets at the configured daily reset time'),
    },
    {
      label: t('Channel earnings balance'),
      value: formatQuota(earnedCredit),
      description: t(
        'Balance earned when other users consume channels provided by you'
      ),
    },
    {
      label: t('Legacy quota'),
      value: formatQuota(legacyQuota),
      description: t(
        'Existing quota that does not expire daily and remains available until consumed'
      ),
    },
  ]

  return (
    <div
      className={
        nested
          ? 'rounded-lg p-0'
          : 'bg-card rounded-lg border p-4 sm:p-5'
      }
    >
      <div className='flex items-start justify-between gap-3'>
        <div className='min-w-0'>
          <div className='flex items-center gap-2'>
            <div className='text-muted-foreground text-xs font-medium tracking-wider uppercase'>
              {t('Available balance')}
            </div>
            <Popover>
              <PopoverTrigger asChild>
                <Button
                  type='button'
                  variant='ghost'
                  size='icon'
                  className='text-muted-foreground hover:text-foreground h-5 w-5 rounded-full'
                >
                  <Info className='size-3.5' />
                </Button>
              </PopoverTrigger>
              <PopoverContent align='end' className='w-80 space-y-2 text-sm'>
                <div className='font-medium'>{t('Balance details')}</div>
                <p className='text-muted-foreground leading-6'>
                  {t(
                    'Total available balance is the sum of current daily credit, channel earnings balance, and legacy quota'
                  )}
                </p>
                <p className='text-muted-foreground leading-6'>
                  {t(
                    'Daily reward is based on community trust level and leaderboard score'
                  )}
                </p>
              </PopoverContent>
            </Popover>
          </div>
          <div className='text-foreground mt-2 font-mono text-2xl font-bold tracking-tight tabular-nums sm:text-3xl'>
            {formatQuota(totalBalance)}
          </div>
          <div className='text-muted-foreground mt-1 text-sm'>
            {t('Remaining quota')}
          </div>
        </div>
        <div className='bg-primary/10 text-primary flex h-10 w-10 shrink-0 items-center justify-center rounded-xl'>
          <WalletCards className='h-5 w-5' />
        </div>
      </div>

      <div className={compact ? 'mt-4 grid gap-3' : 'mt-4 grid gap-3 md:grid-cols-3'}>
        {items.map((item) => (
          <div
            key={item.label}
            className='bg-muted/35 rounded-lg border px-3 py-3 sm:px-4'
          >
            <div className='text-muted-foreground text-xs font-medium tracking-wider uppercase'>
              {item.label}
            </div>
            <div className='text-foreground mt-1.5 font-mono text-base font-semibold tabular-nums sm:text-lg'>
              {item.value}
            </div>
            <div className='text-muted-foreground mt-1 text-xs leading-5'>
              {item.description}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
