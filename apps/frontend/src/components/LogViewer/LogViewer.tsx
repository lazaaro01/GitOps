import { useEffect, useRef } from 'react'
import type { DeploymentLog } from '../../types'

export default function LogViewer({
  logs,
  loading,
}: {
  logs: DeploymentLog[]
  loading: boolean
}) {
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [logs.length])

  return (
    <div className="bg-gray-950 rounded-lg border border-gray-800 font-mono text-xs">
      <div className="h-64 overflow-y-auto p-3 space-y-0.5">
        {logs.length === 0 && !loading && (
          <div className="text-gray-500 italic">Nenhum log disponível</div>
        )}
        {logs.map((log, i) => {
          const time = log.created_at
            ? new Date(log.created_at).toLocaleTimeString('pt-BR')
            : ''
          return (
            <div
              key={`${log.sequence}-${i}`}
              className={`${
                log.level === 'error' ? 'text-red-400' :
                log.level === 'warn' ? 'text-yellow-400' :
                log.level === 'debug' ? 'text-gray-500' :
                'text-gray-300'
              }`}
            >
              <span className="text-gray-600">{time}</span>{' '}
              <span className="uppercase text-[10px] text-gray-500">[{log.level}]</span>{' '}
              {log.message}
            </div>
          )
        })}
        {loading && (
          <div className="text-yellow-400 animate-pulse">Aguardando logs...</div>
        )}
        <div ref={bottomRef} />
      </div>
    </div>
  )
}
