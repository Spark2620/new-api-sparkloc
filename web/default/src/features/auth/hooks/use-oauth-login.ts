import { useState, useRef, useEffect } from 'react'
import type { AxiosRequestConfig } from 'axios'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { useAuthStore } from '@/stores/auth-store'
import { api } from '@/lib/api'
import { getOAuthState } from '../api'
import { buildSparklocOAuthUrl } from '../lib/oauth'
import type { SystemStatus } from '../types'

type LogoutRequestConfig = AxiosRequestConfig & {
  skipErrorHandler?: boolean
}

/**
 * Hook for managing OAuth login
 */
export function useOAuthLogin(status: SystemStatus | null) {
  const { t } = useTranslation()
  const [isLoading, setIsLoading] = useState(false)
  const [sparklocButtonText, setSparklocButtonText] = useState('')
  const [sparklocButtonDisabled, setSparklocButtonDisabled] = useState(false)
  const sparklocTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const { auth } = useAuthStore()

  useEffect(() => {
    setSparklocButtonText(t('Continue with Sparkloc'))

    return () => {
      if (sparklocTimeoutRef.current) {
        clearTimeout(sparklocTimeoutRef.current)
      }
    }
  }, [t])

  const resetSession = async () => {
    try {
      auth.reset()
    } catch (_error) {
      // ignore store reset errors
    }
    try {
      await api.get('/api/user/logout', {
        skipErrorHandler: true,
      } as LogoutRequestConfig)
    } catch (_error) {
      // ignore logout errors
    }
  }

  const handleSparklocLogin = async () => {
    if (!status?.sparkloc_client_id || !status?.sparkloc_authorize_endpoint) {
      toast.error(t('Sparkloc OAuth is not configured'))
      return
    }
    if (sparklocButtonDisabled) return

    setIsLoading(true)
    setSparklocButtonDisabled(true)
    setSparklocButtonText(t('Redirecting to Sparkloc...'))

    if (sparklocTimeoutRef.current) {
      clearTimeout(sparklocTimeoutRef.current)
    }

    sparklocTimeoutRef.current = setTimeout(() => {
      setIsLoading(false)
      setSparklocButtonText(
        t('Request timed out, please refresh and restart Sparkloc login')
      )
      setSparklocButtonDisabled(true)
    }, 20000)

    try {
      await resetSession()
      const state = await getOAuthState()
      if (!state) {
        toast.error(t('Failed to initialize OAuth'))
        if (sparklocTimeoutRef.current) {
          clearTimeout(sparklocTimeoutRef.current)
        }
        setIsLoading(false)
        setSparklocButtonText(t('Continue with Sparkloc'))
        setSparklocButtonDisabled(false)
        return
      }

      const url = buildSparklocOAuthUrl(
        status.sparkloc_authorize_endpoint,
        status.sparkloc_client_id,
        state,
        status.sparkloc_scopes
      )
      window.open(url, '_self')
    } catch (_error) {
      toast.error(t('Failed to start Sparkloc login'))
      if (sparklocTimeoutRef.current) {
        clearTimeout(sparklocTimeoutRef.current)
      }
      setIsLoading(false)
      setSparklocButtonText(t('Continue with Sparkloc'))
      setSparklocButtonDisabled(false)
    }
  }

  return {
    isLoading,
    sparklocButtonText,
    sparklocButtonDisabled,
    handleSparklocLogin,
  }
}
