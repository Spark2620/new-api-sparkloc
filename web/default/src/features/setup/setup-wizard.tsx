import { useEffect, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from '@tanstack/react-router'
import { ShieldCheck } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import { useSystemConfig } from '@/hooks/use-system-config'
import { useStatus } from '@/hooks/use-status'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { ErrorState } from '@/components/error-state'
import { LanguageSwitcher } from '@/components/language-switcher'
import { LoadingState } from '@/components/loading-state'
import { LegalConsent } from '@/features/auth/components/legal-consent'
import { OAuthProviders } from '@/features/auth/components/oauth-providers'
import { getSetupStatus } from './api'
import { DatabaseStep } from './components/database-step'

export function SetupWizard() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { systemName, logo, loading: systemConfigLoading } = useSystemConfig()
  const { status } = useStatus()
  const [agreedToLegal, setAgreedToLegal] = useState(false)

  const {
    data: statusResponse,
    isLoading,
    isError,
    refetch,
  } = useQuery({
    queryKey: ['setup-status'],
    queryFn: getSetupStatus,
    retry: false,
  })

  useEffect(() => {
    if (statusResponse?.success && statusResponse.data?.status) {
      navigate({ to: '/' })
    }
  }, [statusResponse, navigate])

  const hasUserAgreement = Boolean(status?.user_agreement_enabled)
  const hasPrivacyPolicy = Boolean(status?.privacy_policy_enabled)
  const requiresLegalConsent = hasUserAgreement || hasPrivacyPolicy

  useEffect(() => {
    setAgreedToLegal(!requiresLegalConsent)
  }, [requiresLegalConsent])

  const setupStatus = statusResponse?.data
  const sparklocReady = Boolean(status?.sparkloc_oauth)

  return (
    <div className='bg-muted/40 relative min-h-svh py-10'>
      <div className='absolute top-4 right-4 sm:top-6 sm:right-6'>
        <LanguageSwitcher />
      </div>
      <div className='container mx-auto flex max-w-4xl flex-col gap-8 px-4 sm:px-6'>
        <div className='flex flex-col items-center gap-3'>
          <div className='relative h-12 w-12'>
            {systemConfigLoading ? (
              <Skeleton className='absolute inset-0 rounded-full' />
            ) : (
              <img
                src={logo}
                alt={t('System logo')}
                className='h-12 w-12 rounded-full object-cover shadow-sm'
              />
            )}
          </div>
          {systemConfigLoading ? (
            <Skeleton className='h-7 w-40' />
          ) : (
            <h1 className='text-2xl font-semibold tracking-tight'>
              {t('Initialize')} {systemName}
            </h1>
          )}
          <p className='text-muted-foreground max-w-xl text-center text-sm sm:text-base'>
            {t(
              'Use your Sparkloc community account to initialize the root administrator.'
            )}
          </p>
        </div>

        <Card className='shadow-lg'>
          <CardHeader className='space-y-2'>
            <CardTitle className='flex items-center gap-2 text-xl font-semibold'>
              <ShieldCheck className='h-5 w-5' />
              {t('Sparkloc administrator initialization')}
            </CardTitle>
            <CardDescription>
              {t(
                'The first successful Sparkloc login creates the root administrator. Later Sparkloc logins create or access regular accounts.'
              )}
            </CardDescription>
          </CardHeader>

          <CardContent className='space-y-6'>
            {isLoading ? (
              <LoadingState message={t('Loading setup status…')} />
            ) : isError ? (
              <ErrorState
                title={t('We could not load the setup status.')}
                onRetry={() => refetch()}
              />
            ) : (
              <DatabaseStep status={setupStatus} />
            )}

            <div className='bg-card rounded-lg border p-4'>
              <div className='space-y-4'>
                <LegalConsent
                  status={status}
                  checked={agreedToLegal}
                  onCheckedChange={setAgreedToLegal}
                />
                <OAuthProviders
                  status={status}
                  disabled={requiresLegalConsent && !agreedToLegal}
                />
                {!sparklocReady && (
                  <p className='text-muted-foreground text-sm'>
                    {t(
                      'Sparkloc OAuth is not configured. Please set SPARKLOC_CLIENT_ID and SPARKLOC_CLIENT_SECRET in .env.'
                    )}
                  </p>
                )}
              </div>
            </div>

            <div
              className={cn(
                'text-muted-foreground rounded-lg border px-4 py-3 text-sm',
                setupStatus?.root_init && 'border-amber-300 text-amber-700'
              )}
            >
              {setupStatus?.root_init
                ? t(
                    'A root user already exists, but setup is not marked complete. Sign in with the bound Sparkloc account or check the database setup record.'
                  )
                : t(
                    'No root administrator exists yet. The next successful Sparkloc login will become the root administrator.'
                  )}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
