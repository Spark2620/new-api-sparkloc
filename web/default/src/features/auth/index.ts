// ============================================================================
// API Functions
// ============================================================================

export {
  logout,
  getOAuthState,
} from './api'

// ============================================================================
// Types
// ============================================================================

export type {
  ApiResponse,
  SystemStatus,
  OAuthProvider,
  AuthFormProps,
} from './types'

// ============================================================================
// Constants & Schemas
// ============================================================================

export {
  OAUTH_BIND_STORAGE_KEY,
} from './constants'

// ============================================================================
// Utilities
// ============================================================================

export {
  buildSparklocOAuthUrl,
  getAvailableOAuthProviders,
  hasOAuthProviders,
} from './lib/oauth'

export {
  saveUserId,
  getUserId,
  removeUserId,
  getAffiliateCode,
  saveAffiliateCode,
} from './lib/storage'

export {
  isValidOTP,
  isValidBackupCode,
  formatBackupCode,
  cleanBackupCode,
  isValidEmail,
} from './lib/validation'

// ============================================================================
// Hooks
// ============================================================================

export { useTurnstile } from './hooks/use-turnstile'
export { useOAuthLogin } from './hooks/use-oauth-login'
export { useAuthRedirect } from './hooks/use-auth-redirect'

// ============================================================================
// Components
// ============================================================================

export { AuthLayout } from './auth-layout'
export { OAuthProviders } from './components/oauth-providers'
export { TermsFooter } from './components/terms-footer'
export { LegalConsent } from './components/legal-consent'
export { SignIn } from './sign-in'
