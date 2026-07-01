import { DeployBarChart, DeployPieChart } from '../components/HistoryChart/HistoryChart'
import { useDeployments } from '../hooks/useDeployments'

export default function History() {
  const { deployments, loading } = useDeployments(1, 100)

  if (loading) {
    return <div className="text-center py-12 text-gray-500 animate-pulse">Carregando...</div>
  }

  const successRate = deployments.length > 0
    ? Math.round((deployments.filter((d) => d.status === 'success').length / deployments.length) * 100)
    : 0

  const failRate = deployments.length > 0
    ? Math.round((deployments.filter((d) => d.status === 'failed').length / deployments.length) * 100)
    : 0

  const cancelRate = deployments.length > 0
    ? Math.round((deployments.filter((d) => d.status === 'cancelled').length / deployments.length) * 100)
    : 0

  return (
    <div className="space-y-6">
      <h1 className="text-xl font-semibold">Histórico Visual</h1>

      <div className="grid grid-cols-3 gap-4">
        <div className="bg-gray-900 border border-gray-800 rounded-lg p-4 text-center">
          <div className="text-2xl font-bold text-emerald-400">{successRate}%</div>
          <div className="text-xs text-gray-400 mt-1">Sucesso</div>
        </div>
        <div className="bg-gray-900 border border-gray-800 rounded-lg p-4 text-center">
          <div className="text-2xl font-bold text-red-400">{failRate}%</div>
          <div className="text-xs text-gray-400 mt-1">Falha</div>
        </div>
        <div className="bg-gray-900 border border-gray-800 rounded-lg p-4 text-center">
          <div className="text-2xl font-bold text-gray-400">{cancelRate}%</div>
          <div className="text-xs text-gray-400 mt-1">Cancelado</div>
        </div>
      </div>

      <div className="bg-gray-900 border border-gray-800 rounded-lg p-4">
        <h2 className="text-sm font-medium text-gray-400 mb-4">Deploys por Dia</h2>
        <DeployBarChart deployments={deployments} />
      </div>

      <div className="bg-gray-900 border border-gray-800 rounded-lg p-4">
        <h2 className="text-sm font-medium text-gray-400 mb-4">Distribuição por Status</h2>
        <DeployPieChart deployments={deployments} />
      </div>
    </div>
  )
}
