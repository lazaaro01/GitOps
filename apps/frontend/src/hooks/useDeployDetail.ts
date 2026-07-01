import { useState, useEffect, useCallback } from 'react'
import { fetchDeployment } from '../services/api'
import type { Deployment, DeploymentLog } from '../types'

export function useDeployDetail(id: string) {
  const [deployment, setDeployment] = useState<Deployment | null>(null)
  const [logs, setLogs] = useState<DeploymentLog[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    if (!id) return
    setLoading(true)
    setError(null)
    try {
      const res = await fetchDeployment(id)
      setDeployment(res.deployment)
      setLogs(res.logs)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load')
    } finally {
      setLoading(false)
    }
  }, [id])

  useEffect(() => { load() }, [load])

  return { deployment, logs, setLogs, loading, error, reload: load }
}
