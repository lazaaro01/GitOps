export type DeploymentStatus =
  | 'pending'
  | 'queued'
  | 'in_progress'
  | 'success'
  | 'failed'
  | 'cancelled'

export interface Deployment {
  id: string
  app_name: string
  image_tag: string
  status: DeploymentStatus
  error_message?: string
  created_at: string
  updated_at: string
  finished_at?: string
}

export interface DeploymentLog {
  id: string
  deployment_id: string
  step: string
  level: 'info' | 'warn' | 'error' | 'debug'
  message: string
  sequence: number
  created_at: string
}

export interface APIResponse<T = unknown> {
  success: boolean
  data?: T
  error?: string
}

export interface DeployListData {
  deployments: Deployment[]
  total: number
  limit: number
  offset: number
}

export interface SSEEvent {
  type: 'deploy_update' | 'deploy_log' | 'deploy_completed'
  deploy_id: string
  status?: string
  step?: string
  level?: string
  message?: string
}
