import type { SystemStatus, OAuthProvider } from '../types'

export {
  buildGitHubOAuthUrl,
  buildDiscordOAuthUrl,
  buildOIDCOAuthUrl,
  buildLinuxDOOAuthUrl,
  buildSparklocOAuthUrl,
} from '@/lib/oauth'

// ============================================================================
// OAuth Providers Utilities
// ============================================================================

/**
 * Get available OAuth providers from system status
 */
export function getAvailableOAuthProviders(
  status: SystemStatus | null
): OAuthProvider[] {
  if (!status) return []

  const providers: OAuthProvider[] = []

  if (status.sparkloc_oauth) {
    providers.push({
      name: 'Sparkloc',
      type: 'sparkloc',
      enabled: true,
      clientId: status.sparkloc_client_id,
      authEndpoint: status.sparkloc_authorize_endpoint,
    })
  }

  return providers
}

/**
 * Check if any OAuth provider is available
 */
export function hasOAuthProviders(status: SystemStatus | null): boolean {
  if (!status) return false
  return !!status.sparkloc_oauth
}
