import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { cn } from '@/lib/utils'
import { useStatus } from '@/hooks/use-status'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { LegalConsent } from '@/features/auth/components/legal-consent'
import { OAuthProviders } from '@/features/auth/components/oauth-providers'
import type { AuthFormProps } from '@/features/auth/types'

export function UserAuthForm({
  className,
  redirectTo: _redirectTo,
  ...props
}: AuthFormProps) {
  const { t } = useTranslation()
  const [agreedToLegal, setAgreedToLegal] = useState(false)
  const { status } = useStatus()

  const hasUserAgreement = Boolean(status?.user_agreement_enabled)
  const hasPrivacyPolicy = Boolean(status?.privacy_policy_enabled)
  const requiresLegalConsent = hasUserAgreement || hasPrivacyPolicy
  const sparklocReady = Boolean(status?.sparkloc_oauth)

  useEffect(() => {
    setAgreedToLegal(!requiresLegalConsent)
  }, [requiresLegalConsent])

  const disabled = requiresLegalConsent && !agreedToLegal

  return (
    <div className={cn('grid gap-4', className)} {...props}>
      <LegalConsent
        status={status}
        checked={agreedToLegal}
        onCheckedChange={setAgreedToLegal}
      />

      <div
        onClick={() => {
          if (disabled) {
            toast.error(t('Please agree to the legal terms first'))
          }
        }}
      >
        <OAuthProviders status={status} disabled={disabled} />
      </div>

      {!sparklocReady && (
        <Alert>
          <AlertDescription>
            {t(
              'Sparkloc OAuth is not configured. Please set SPARKLOC_CLIENT_ID and SPARKLOC_CLIENT_SECRET in .env.'
            )}
          </AlertDescription>
        </Alert>
      )}
    </div>
  )
}
