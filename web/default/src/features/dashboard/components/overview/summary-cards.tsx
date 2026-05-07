import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { getSelf } from '@/lib/api'
import { getCurrencyLabel, isCurrencyDisplayEnabled } from '@/lib/currency'
import { formatNumber, formatQuota } from '@/lib/format'
import { useStatus } from '@/hooks/use-status'
import { StaggerContainer, StaggerItem } from '@/components/page-transition'
import { useSummaryCardsConfig } from '@/features/dashboard/hooks/use-dashboard-config'
import { StatCard } from '../ui/stat-card'

export function SummaryCards() {
  const { status, loading } = useStatus()
  const { data: selfResponse, isLoading: selfLoading } = useQuery({
    queryKey: ['dashboard-self-summary'],
    queryFn: getSelf,
    staleTime: 30000,
  })
  const user = selfResponse?.success ? selfResponse.data : null

  const summaryValues = useMemo(() => {
    const remainQuota = Number(user?.available_quota ?? user?.quota ?? 0)
    const usedQuota = Number(user?.used_quota ?? 0)
    const requestCount = Number(user?.request_count ?? 0)

    return {
      remainDisplay: formatQuota(remainQuota),
      usedDisplay: formatQuota(usedQuota),
      requestCountDisplay: formatNumber(requestCount),
    }
  }, [user])

  const currencyEnabledFromStore = isCurrencyDisplayEnabled()
  const statusCurrencyFlag =
    typeof status?.display_in_currency === 'boolean'
      ? Boolean(status.display_in_currency)
      : undefined
  const currencyEnabled =
    statusCurrencyFlag !== undefined
      ? statusCurrencyFlag
      : currencyEnabledFromStore
  const currencyLabel = currencyEnabled ? getCurrencyLabel() : 'Tokens'

  const items = useSummaryCardsConfig({
    ...summaryValues,
    currencyEnabled,
    currencyLabel,
  }).map((config) => ({
    title: config.title,
    value: config.value,
    desc: config.description,
    icon: config.icon,
  }))

  return (
    <div className='overflow-hidden rounded-lg border'>
      <StaggerContainer className='divide-border/60 grid grid-cols-3 divide-x'>
        {items.map((it) => (
          <StaggerItem key={it.title} className='px-3 py-3 sm:px-5 sm:py-4'>
            <StatCard
              title={it.title}
              value={it.value}
              description={it.desc}
              icon={it.icon}
              loading={loading || selfLoading}
            />
          </StaggerItem>
        ))}
      </StaggerContainer>
    </div>
  )
}
