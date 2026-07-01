import { useState, useEffect, useCallback } from 'react'
import { fetchDeployments } from '../services/api'
import type { Deployment } from '../types'

export function useDeployments(page = 1, limit = 20, status?: string) {
  const [deployments, setDeployments] = useState<Deployment[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const res = await fetchDeployments(page, limit, status)
      setDeployments(res.deployments)
      setTotal(res.total)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load')
    } finally {
      setLoading(false)
    }
  }, [page, limit, status])

  useEffect(() => { load() }, [load])

  return { deployments, total, loading, error, reload: load }
}
