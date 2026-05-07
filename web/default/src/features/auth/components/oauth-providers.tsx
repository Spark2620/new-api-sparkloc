import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { useOAuthLogin } from '../hooks/use-oauth-login'
import type { SystemStatus } from '../types'

type OAuthProvidersProps = {
  status: SystemStatus | null
  disabled?: boolean
  className?: string
}

export function OAuthProviders({
  status,
  disabled = false,
  className,
}: OAuthProvidersProps) {
  const { t } = useTranslation()
  const {
    isLoading,
    sparklocButtonText,
    sparklocButtonDisabled,
    handleSparklocLogin,
  } = useOAuthLogin(status)

  if (!status?.sparkloc_oauth) return null

  return (
    <div className={cn('space-y-3', className)}>
      <Button
        variant='default'
        type='button'
        disabled={disabled || isLoading || sparklocButtonDisabled}
        onClick={handleSparklocLogin}
        className='h-12 w-full justify-center gap-3 rounded-lg'
      >
        <img
          src='/logo.png'
          alt=''
          className='h-5 w-5 rounded-sm object-contain'
          aria-hidden='true'
        />
        {sparklocButtonText || t('Continue with Sparkloc')}
      </Button>
    </div>
  )
}
