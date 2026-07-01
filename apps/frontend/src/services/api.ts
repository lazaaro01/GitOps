import axios from 'axios'
import type { APIResponse, DeployListData, Deployment, DeploymentLog } from '../types'

const http = axios.create({
  baseURL: '/api',
  headers: { 'Content-Type': 'application/json' },
})

export async function fetchDeployments(
  page = 1,
  limit = 20,
  status?: string,
): Promise<DeployListData> {
  const params: Record<string, string | number> = { page, limit }
  if (status) params.status = status

  const { data } = await http.get<APIResponse<DeployListData>>('/deployments', { params })
  if (!data.success) throw new Error(data.error ?? 'Unknown error')
  return data.data!
}

export async function fetchDeployment(id: string): Promise<{
  deployment: Deployment
  logs: DeploymentLog[]
}> {
  const { data } = await http.get<APIResponse<{ deployment: Deployment; logs: DeploymentLog[] }>>(
    `/deployments/${id}`,
  )
  if (!data.success) throw new Error(data.error ?? 'Unknown error')
  return data.data!
}

export async function fetchLogs(
  id: string,
  params?: { step?: string; offset?: number },
): Promise<{ lines: DeploymentLog[]; total_lines: number }> {
  const { data } = await http.get<
    APIResponse<{ lines: DeploymentLog[]; total_lines: number }>
  >(`/deployments/${id}/logs`, { params })
  if (!data.success) throw new Error(data.error ?? 'Unknown error')
  return data.data!
}

export async function createDeploy(appName: string, imageTag: string, envVars?: Record<string, string>) {
  const { data } = await http.post<APIResponse<Deployment>>('/deploy', {
    app_name: appName,
    image_tag: imageTag,
    env_vars: envVars,
  })
  if (!data.success) throw new Error(data.error ?? 'Unknown error')
  return data.data!
}

export async function cancelDeploy(id: string) {
  const { data } = await http.put<APIResponse<{ message: string }>>(`/deployments/${id}/cancel`)
  if (!data.success) throw new Error(data.error ?? 'Unknown error')
  return data.data!
}

export async function rollbackDeploy(id: string, targetVersion: string) {
  const { data } = await http.post<APIResponse<{ rollback_job_id: string; status: string; message: string }>>(
    `/deployments/${id}/rollback`,
    { target_version: targetVersion },
  )
  if (!data.success) throw new Error(data.error ?? 'Unknown error')
  return data.data!
}

export async function retryDeploy(id: string) {
  const { data } = await http.post<APIResponse<{ retry_job_id: string; status: string; message: string }>>(
    `/deployments/${id}/retry`,
  )
  if (!data.success) throw new Error(data.error ?? 'Unknown error')
  return data.data!
}

export async function downloadLogs(id: string): Promise<Blob> {
  const { data } = await http.get<Blob>(`/deployments/${id}/logs/download`, {
    responseType: 'blob',
  })
  return data
}

export default http
