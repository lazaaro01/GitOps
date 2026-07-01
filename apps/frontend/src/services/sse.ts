import type { SSEEvent } from '../types'

type EventCallback = (event: SSEEvent) => void

export function connectSSE(deployId: string, onEvent: EventCallback): () => void {
  const url = `/api/events?deploy_id=${deployId}`
  const source = new EventSource(url)

  source.addEventListener('deploy_update', (e) => {
    try {
      onEvent({ type: 'deploy_update', ...JSON.parse(e.data) })
    } catch { /* ignore parse errors */ }
  })

  source.addEventListener('deploy_log', (e) => {
    try {
      onEvent({ type: 'deploy_log', ...JSON.parse(e.data) })
    } catch { /* ignore parse errors */ }
  })

  source.addEventListener('deploy_completed', (e) => {
    try {
      onEvent({ type: 'deploy_completed', ...JSON.parse(e.data) })
      source.close()
    } catch { /* ignore parse errors */ }
  })

  source.onerror = () => {
    source.close()
  }

  return () => source.close()
}
