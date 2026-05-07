// ============================================================================
// API Responses
// ============================================================================

export interface ApiResponse {
  success: boolean
  message: string
  data?: unknown
}

// ============================================================================
// System Status
// ============================================================================

export interface SystemStatus {
  success?: boolean
  message?: string
  data?: {
    version?: string
    system_name?: string
    logo?: string
    sparkloc_oauth?: boolean
    sparkloc_authorize_endpoint?: string
    sparkloc_client_id?: string
    sparkloc_scopes?: string
    passkey_login?: boolean
    turnstile_check?: boolean
    turnstile_site_key?: string
    self_use_mode_enabled?: boolean
    display_in_currency?: boolean
    display_token_stat_enabled?: boolean
    quota_per_unit?: number
    quota_display_type?: string
    usd_exchange_rate?: number
    custom_currency_symbol?: string
    custom_currency_exchange_rate?: number
    demo_site_enabled?: boolean
    user_agreement_enabled?: boolean
    privacy_policy_enabled?: boolean
    [key: string]: unknown
  }
  // Allow direct access to common properties
  version?: string
  system_name?: string
  logo?: string
  sparkloc_oauth?: boolean
  sparkloc_authorize_endpoint?: string
  sparkloc_client_id?: string
  sparkloc_scopes?: string
  passkey_login?: boolean
  turnstile_check?: boolean
  turnstile_site_key?: string
  self_use_mode_enabled?: boolean
  display_in_currency?: boolean
  display_token_stat_enabled?: boolean
  quota_per_unit?: number
  quota_display_type?: string
  usd_exchange_rate?: number
  custom_currency_symbol?: string
  custom_currency_exchange_rate?: number
  demo_site_enabled?: boolean
  user_agreement_enabled?: boolean
  privacy_policy_enabled?: boolean
  [key: string]: unknown
}

// ============================================================================
// OAuth
// ============================================================================

export interface OAuthProvider {
  name: string
  type: 'sparkloc'
  enabled: boolean
  clientId?: string
  authEndpoint?: string
}

// ============================================================================
// Form Props
// ============================================================================

export interface AuthFormProps extends React.HTMLAttributes<HTMLDivElement> {
  redirectTo?: string
}
