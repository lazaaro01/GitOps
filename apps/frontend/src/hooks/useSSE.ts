import { useEffect, useRef, useCallback } from 'react'
import { connectSSE } from '../services/sse'
import type { SSEEvent, DeploymentLog } from '../types'

export function useSSE(
  deployId: string | undefined,
  onLog: (log: DeploymentLog) => void,
  onComplete: () => void,
) {
  const onLogRef = useRef(onLog)
  const onCompleteRef = useRef(onComplete)
  onLogRef.current = onLog
  onCompleteRef.current = onComplete

  const cleanup = useRef<(() => void) | null>(null)

  const handler = useCallback((event: SSEEvent) => {
    if (event.type === 'deploy_log' && event.message) {
      onLogRef.current({
        id: '',
        deployment_id: event.deploy_id,
        step: event.step ?? '',
        level: (event.level as DeploymentLog['level']) ?? 'info',
        message: event.message,
        sequence: 0,
        created_at: new Date().toISOString(),
      })
    }
    if (event.type === 'deploy_completed') {
      onCompleteRef.current()
    }
  }, [])

  useEffect(() => {
    if (!deployId) return
    cleanup.current = connectSSE(deployId, handler)
    return () => cleanup.current?.()
  }, [deployId, handler])
}
