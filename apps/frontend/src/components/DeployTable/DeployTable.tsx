import { Link } from 'react-router-dom'
import StatusBadge from '../StatusBadge/StatusBadge'
import type { Deployment } from '../../types'

export default function DeployTable({
  deployments,
  onRetry,
}: {
  deployments: Deployment[]
  onRetry: (id: string) => void
}) {
  if (deployments.length === 0) {
    return (
      <div className="text-center py-12 text-gray-500">
        Nenhum deploy encontrado
      </div>
    )
  }

  return (
    <div className="overflow-x-auto rounded-lg border border-gray-800">
      <table className="w-full text-sm">
        <thead>
          <tr className="bg-gray-900 text-gray-400 uppercase text-xs tracking-wider">
            <th className="text-left px-4 py-3">App</th>
            <th className="text-left px-4 py-3">Imagem</th>
            <th className="text-left px-4 py-3">Status</th>
            <th className="text-left px-4 py-3">Data</th>
            <th className="text-right px-4 py-3">Ações</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-800">
          {deployments.map((d) => (
            <tr key={d.id} className="hover:bg-gray-900/50 transition-colors">
              <td className="px-4 py-3">
                <Link to={`/deploy/${d.id}`} className="text-emerald-400 hover:underline font-medium">
                  {d.app_name}
                </Link>
              </td>
              <td className="px-4 py-3 text-gray-400 font-mono text-xs">{d.image_tag}</td>
              <td className="px-4 py-3"><StatusBadge status={d.status} /></td>
              <td className="px-4 py-3 text-gray-400 text-xs">
                {new Date(d.created_at).toLocaleString('pt-BR')}
              </td>
              <td className="px-4 py-3 text-right">
                <div className="flex justify-end gap-2">
                  <Link
                    to={`/deploy/${d.id}`}
                    className="text-xs text-gray-400 hover:text-white transition-colors"
                  >
                    Detalhes
                  </Link>
                  {d.status === 'failed' && (
                    <button
                      onClick={() => onRetry(d.id)}
                      className="text-xs text-yellow-400 hover:text-yellow-300 transition-colors"
                    >
                      Retry
                    </button>
                  )}
                </div>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
