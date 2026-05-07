import { useMemo } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { SectionPageLayout } from '@/components/layout'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { getRouteApi } from '@tanstack/react-router'
import { LeaderboardTable } from './components/leaderboard-table'
import type { LeaderboardType } from './types'

const route = getRouteApi('/_authenticated/leaderboard/')

export function Leaderboard() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const search = route.useSearch()
  const currentType = (search.type ?? 'consumption') as LeaderboardType

  const title = useMemo(
    () =>
      currentType === 'contribution'
        ? t('Contribution Ranking')
        : t('Consumption Ranking'),
    [currentType, t]
  )

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>{t('Leaderboard')}</SectionPageLayout.Title>
      <SectionPageLayout.Description>
        {t(
          'Track the most active users by total consumption and the top contributors by cumulative channel earnings.'
        )}
      </SectionPageLayout.Description>
      <SectionPageLayout.Actions>
        <Tabs
          value={currentType}
          onValueChange={(value) => {
            void navigate({
              to: '/leaderboard',
              search: (prev) => ({
                ...prev,
                type: value as LeaderboardType,
                page: undefined,
                pageSize: prev.pageSize,
              }),
            })
          }}
        >
          <TabsList className='grid w-full grid-cols-2 sm:w-auto'>
            <TabsTrigger value='consumption'>{t('Consumption Ranking')}</TabsTrigger>
            <TabsTrigger value='contribution'>{t('Contribution Ranking')}</TabsTrigger>
          </TabsList>
        </Tabs>
      </SectionPageLayout.Actions>
      <SectionPageLayout.Content>
        <div className='space-y-3 sm:space-y-4'>
          <div className='text-muted-foreground text-sm'>{title}</div>
          <LeaderboardTable type={currentType} />
        </div>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
