import type { DeploymentStatus } from '../../types'

const map: Record<DeploymentStatus, { label: string; style: string }> = {
  pending:     { label: 'Pendente',     style: 'bg-gray-700 text-gray-300' },
  queued:      { label: 'Na Fila',      style: 'bg-blue-900 text-blue-300' },
  in_progress: { label: 'Em Andamento', style: 'bg-yellow-900 text-yellow-300 animate-pulse' },
  success:     { label: 'Sucesso',      style: 'bg-emerald-900 text-emerald-300' },
  failed:      { label: 'Falha',        style: 'bg-red-900 text-red-300' },
  cancelled:   { label: 'Cancelado',    style: 'bg-gray-700 text-gray-400' },
}

export default function StatusBadge({ status }: { status: DeploymentStatus }) {
  const s = map[status] ?? { label: status, style: 'bg-gray-700 text-gray-300' }
  return (
    <span className={`inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium ${s.style}`}>
      {status === 'in_progress' && (
        <span className="w-1.5 h-1.5 rounded-full bg-yellow-300 animate-ping" />
      )}
      {s.label}
    </span>
  )
}
