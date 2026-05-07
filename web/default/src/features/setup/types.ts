export interface SetupStatus {
  status: boolean
  root_init: boolean
  database_type: string
}

export interface SetupResponse {
  success: boolean
  message?: string
  data?: SetupStatus
}
