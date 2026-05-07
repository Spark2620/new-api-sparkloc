import { UserRound } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { StatusBadge } from '@/components/status-badge'
import type { UserProfile } from '../../types'

interface AccountBindingsTabProps {
  profile: UserProfile | null
  onUpdate: () => void
}

export function AccountBindingsTab({
  profile,
  onUpdate: _onUpdate,
}: AccountBindingsTabProps) {
  const { t } = useTranslation()
  const sparklocId = profile?.sparkloc_id || ''

  if (!profile) return null

  return (
    <div className='grid grid-cols-1 gap-2.5 sm:gap-3'>
      <div className='flex items-center justify-between gap-2.5 rounded-lg border p-2.5 sm:gap-3 sm:p-3'>
        <div className='flex min-w-0 items-center gap-2.5 sm:gap-3'>
          <div className='bg-muted shrink-0 rounded-md p-1.5 sm:p-2'>
            <UserRound className='h-4 w-4' />
          </div>
          <div className='min-w-0'>
            <div className='flex items-center gap-1.5'>
              <p className='text-sm font-medium'>{t('Sparkloc')}</p>
              {sparklocId ? (
                <StatusBadge
                  label={t('Bound')}
                  variant='success'
                  copyable={false}
                />
              ) : (
                <StatusBadge
                  label={t('Not bound')}
                  variant='neutral'
                  copyable={false}
                />
              )}
            </div>
            <p className='text-muted-foreground truncate text-xs'>
              {sparklocId || t('Sign in with Sparkloc to bind this account')}
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}
