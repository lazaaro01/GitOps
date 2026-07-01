import { useState, useCallback } from 'react'
import { useDeployments } from '../hooks/useDeployments'
import DeployTable from '../components/DeployTable/DeployTable'
import { DeployBarChart, DeployPieChart } from '../components/HistoryChart/HistoryChart'
import { createDeploy, retryDeploy } from '../services/api'

const statusFilters = [
  { value: '', label: 'Todos' },
  { value: 'success', label: 'Sucesso' },
  { value: 'failed', label: 'Falha' },
  { value: 'in_progress', label: 'Em Andamento' },
]

export default function Dashboard() {
  const [page, setPage] = useState(1)
  const [statusFilter, setStatusFilter] = useState('')
  const [showNew, setShowNew] = useState(false)
  const [appName, setAppName] = useState('')
  const [imageTag, setImageTag] = useState('')
  const [creating, setCreating] = useState(false)
  const [toast, setToast] = useState<string | null>(null)
  const limit = 20

  const { deployments, total, loading, reload } = useDeployments(page, limit, statusFilter || undefined)

  const showToast = useCallback((msg: string) => {
    setToast(msg)
    setTimeout(() => setToast(null), 4000)
  }, [])

  const handleCreate = useCallback(async () => {
    if (!appName || !imageTag) return
    setCreating(true)
    try {
      const deploy = await createDeploy(appName, imageTag)
      showToast(`Deploy ${deploy.app_name} criado com sucesso!`)
      setShowNew(false)
      setAppName('')
      setImageTag('')
      reload()
    } catch (err) {
      showToast(`Erro: ${err instanceof Error ? err.message : 'Falha ao criar deploy'}`)
    } finally {
      setCreating(false)
    }
  }, [appName, imageTag, showToast, reload])

  const handleRetry = useCallback(async (id: string) => {
    try {
      await retryDeploy(id)
      showToast('Deploy reexecutado com sucesso!')
      reload()
    } catch (err) {
      showToast(`Erro: ${err instanceof Error ? err.message : 'Falha ao reexecutar'}`)
    }
  }, [showToast, reload])

  const totalPages = Math.max(1, Math.ceil(total / limit))

  return (
    <div className="space-y-6">
      {toast && (
        <div className="fixed top-4 right-4 z-50 bg-gray-800 border border-gray-700 text-gray-100 px-4 py-2 rounded-lg shadow-lg text-sm">
          {toast}
        </div>
      )}

      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">Dashboard</h1>
        <button
          onClick={() => setShowNew(!showNew)}
          className="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white text-sm rounded-lg transition-colors"
        >
          {showNew ? 'Cancelar' : 'Novo Deploy'}
        </button>
      </div>

      {showNew && (
        <div className="bg-gray-900 border border-gray-800 rounded-lg p-4 space-y-3">
          <input
            value={appName}
            onChange={(e) => setAppName(e.target.value)}
            placeholder="app_name (ex: my-app)"
            className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm"
          />
          <input
            value={imageTag}
            onChange={(e) => setImageTag(e.target.value)}
            placeholder="image_tag (ex: nginx:latest)"
            className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm"
          />
          <button
            onClick={handleCreate}
            disabled={creating || !appName || !imageTag}
            className="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 disabled:bg-gray-700 disabled:text-gray-500 text-white text-sm rounded-lg transition-colors"
          >
            {creating ? 'Criando...' : 'Criar Deploy'}
          </button>
        </div>
      )}

      <div className="flex gap-2">
        {statusFilters.map((f) => (
          <button
            key={f.value}
            onClick={() => { setStatusFilter(f.value); setPage(1) }}
            className={`px-3 py-1.5 text-xs rounded-lg transition-colors ${
              statusFilter === f.value
                ? 'bg-emerald-600 text-white'
                : 'bg-gray-800 text-gray-400 hover:text-gray-200'
            }`}
          >
            {f.label}
          </button>
        ))}
      </div>

      {loading ? (
        <div className="text-center py-12 text-gray-500 animate-pulse">Carregando...</div>
      ) : (
        <>
          <DeployTable deployments={deployments} onRetry={handleRetry} />

          {totalPages > 1 && (
            <div className="flex justify-center items-center gap-2 text-sm">
              <button
                onClick={() => setPage(Math.max(1, page - 1))}
                disabled={page <= 1}
                className="px-3 py-1 bg-gray-800 rounded disabled:opacity-40 hover:bg-gray-700 transition-colors"
              >
                &lt;
              </button>
              <span className="text-gray-400">
                {page} de {totalPages}
              </span>
              <button
                onClick={() => setPage(Math.min(totalPages, page + 1))}
                disabled={page >= totalPages}
                className="px-3 py-1 bg-gray-800 rounded disabled:opacity-40 hover:bg-gray-700 transition-colors"
              >
                &gt;
              </button>
            </div>
          )}
        </>
      )}

      {deployments.length > 0 && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 pt-4">
          <div className="bg-gray-900 border border-gray-800 rounded-lg p-4">
            <h2 className="text-sm font-medium text-gray-400 mb-3">Deploys por Dia</h2>
            <DeployBarChart deployments={deployments} />
          </div>
          <div className="bg-gray-900 border border-gray-800 rounded-lg p-4">
            <h2 className="text-sm font-medium text-gray-400 mb-3">Distribuição por Status</h2>
            <DeployPieChart deployments={deployments} />
          </div>
        </div>
      )}
    </div>
  )
}
