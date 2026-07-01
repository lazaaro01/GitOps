import { useState, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useDeployDetail } from '../hooks/useDeployDetail'
import { useSSE } from '../hooks/useSSE'
import StatusBadge from '../components/StatusBadge/StatusBadge'
import DeployTimeline from '../components/DeployTimeline/DeployTimeline'
import LogViewer from '../components/LogViewer/LogViewer'
import RollbackModal from '../components/RollbackModal/RollbackModal'
import { rollbackDeploy, retryDeploy, downloadLogs, cancelDeploy } from '../services/api'
import type { DeploymentLog } from '../types'

export default function DeployDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { deployment, logs, setLogs, loading, reload } = useDeployDetail(id ?? '')
  const [showRollback, setShowRollback] = useState(false)
  const [toast, setToast] = useState<string | null>(null)

  const showToast = useCallback((msg: string) => {
    setToast(msg)
    setTimeout(() => setToast(null), 4000)
  }, [])

  const handleSSELog = useCallback((log: DeploymentLog) => {
    setLogs((prev) => [...prev, log])
  }, [setLogs])

  const handleSSEComplete = useCallback(() => {
    reload()
  }, [reload])

  const isActive = deployment?.status === 'in_progress' || deployment?.status === 'queued'
  useSSE(isActive ? id : undefined, handleSSELog, handleSSEComplete)

  const handleRollback = useCallback(async (targetVersion: string) => {
    if (!id) return
    try {
      await rollbackDeploy(id, targetVersion)
      showToast(`Rollback para ${targetVersion} enfileirado!`)
      setShowRollback(false)
    } catch (err) {
      showToast(`Erro: ${err instanceof Error ? err.message : 'Falha no rollback'}`)
    }
  }, [id, showToast])

  const handleRetry = useCallback(async () => {
    if (!id) return
    try {
      await retryDeploy(id)
      showToast('Reexecução enfileirada!')
      reload()
    } catch (err) {
      showToast(`Erro: ${err instanceof Error ? err.message : 'Falha ao reexecutar'}`)
    }
  }, [id, showToast, reload])

  const handleCancel = useCallback(async () => {
    if (!id) return
    try {
      await cancelDeploy(id)
      showToast('Deploy cancelado!')
      reload()
    } catch (err) {
      showToast(`Erro: ${err instanceof Error ? err.message : 'Falha ao cancelar'}`)
    }
  }, [id, showToast, reload])

  const handleDownload = useCallback(async () => {
    if (!id) return
    try {
      const blob = await downloadLogs(id)
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `deploy-${id.slice(0, 8)}.log`
      a.click()
      URL.revokeObjectURL(url)
    } catch (err) {
      showToast(`Erro ao baixar logs: ${err instanceof Error ? err.message : 'Erro'}`)
    }
  }, [id, showToast])

  const previousVersions = deployment
    ? [`${deployment.app_name}-prev-1`, `${deployment.app_name}-prev-2`]
    : []

  if (loading) {
    return <div className="text-center py-12 text-gray-500 animate-pulse">Carregando...</div>
  }

  if (!deployment) {
    return (
      <div className="text-center py-12 text-gray-500">
        Deploy não encontrado
        <button onClick={() => navigate('/')} className="block mx-auto mt-4 text-emerald-400 hover:underline text-sm">
          Voltar ao Dashboard
        </button>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {toast && (
        <div className="fixed top-4 right-4 z-50 bg-gray-800 border border-gray-700 text-gray-100 px-4 py-2 rounded-lg shadow-lg text-sm">
          {toast}
        </div>
      )}

      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <button onClick={() => navigate('/')} className="text-gray-400 hover:text-white text-sm">
            &larr; Voltar
          </button>
          <h1 className="text-xl font-semibold">{deployment.app_name}</h1>
          <span className="text-gray-500 font-mono text-xs">{deployment.image_tag}</span>
          <StatusBadge status={deployment.status} />
        </div>

        <div className="flex gap-2">
          {deployment.status === 'success' && (
            <button
              onClick={() => setShowRollback(true)}
              className="px-3 py-1.5 text-xs bg-red-600 hover:bg-red-500 text-white rounded-lg transition-colors"
            >
              Rollback
            </button>
          )}
          {(deployment.status === 'failed' || deployment.status === 'cancelled') && (
            <button
              onClick={handleRetry}
              className="px-3 py-1.5 text-xs bg-yellow-600 hover:bg-yellow-500 text-white rounded-lg transition-colors"
            >
              Retry
            </button>
          )}
          {(deployment.status === 'pending' || deployment.status === 'queued') && (
            <button
              onClick={handleCancel}
              className="px-3 py-1.5 text-xs bg-gray-600 hover:bg-gray-500 text-white rounded-lg transition-colors"
            >
              Cancelar
            </button>
          )}
        </div>
      </div>

      {deployment.error_message && (
        <div className="bg-red-900/30 border border-red-800 rounded-lg px-4 py-3 text-sm text-red-300">
          {deployment.error_message}
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-1">
          <div className="bg-gray-900 border border-gray-800 rounded-lg p-4 space-y-3">
            <h2 className="text-sm font-medium text-gray-400">Timeline</h2>
            <DeployTimeline logs={logs} deploymentStatus={deployment.status} />
          </div>

          <div className="bg-gray-900 border border-gray-800 rounded-lg p-4 mt-4 space-y-2">
            <h2 className="text-sm font-medium text-gray-400">Informações</h2>
            <div className="text-xs space-y-1 text-gray-400">
              <p><span className="text-gray-500">ID:</span> {deployment.id.slice(0, 8)}...</p>
              <p><span className="text-gray-500">Criado:</span> {new Date(deployment.created_at).toLocaleString('pt-BR')}</p>
              {deployment.finished_at && (
                <p><span className="text-gray-500">Finalizado:</span> {new Date(deployment.finished_at).toLocaleString('pt-BR')}</p>
              )}
            </div>
          </div>
        </div>

        <div className="lg:col-span-2 space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-sm font-medium text-gray-400">
              Logs ({logs.length} linhas)
            </h2>
            <button
              onClick={handleDownload}
              className="text-xs text-gray-400 hover:text-white transition-colors"
            >
              Download
            </button>
          </div>
          <LogViewer logs={logs} loading={isActive && logs.length === 0} />
        </div>
      </div>

      {showRollback && (
        <RollbackModal
          currentVersion={deployment.image_tag}
          versions={previousVersions}
          onConfirm={handleRollback}
          onClose={() => setShowRollback(false)}
        />
      )}
    </div>
  )
}
