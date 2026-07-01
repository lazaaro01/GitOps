import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
  Legend,
} from 'recharts'
import type { Deployment } from '../../types'

const COLORS = {
  success: '#34d399',
  failed: '#f87171',
  cancelled: '#6b7280',
  in_progress: '#fbbf24',
  pending: '#9ca3af',
  queued: '#60a5fa',
}

function statusColor(status: string): string {
  return COLORS[status as keyof typeof COLORS] ?? '#6b7280'
}

export function DeployPieChart({ deployments }: { deployments: Deployment[] }) {
  const counts: Record<string, number> = {}
  for (const d of deployments) {
    counts[d.status] = (counts[d.status] ?? 0) + 1
  }

  const data = Object.entries(counts).map(([name, value]) => ({
    name,
    value,
    color: statusColor(name),
  }))

  if (data.length === 0) return null

  return (
    <ResponsiveContainer width="100%" height={200}>
      <PieChart>
        <Pie
          data={data}
          dataKey="value"
          nameKey="name"
          cx="50%"
          cy="50%"
          outerRadius={70}
          innerRadius={40}
        >
          {data.map((entry, i) => (
            <Cell key={i} fill={entry.color} />
          ))}
        </Pie>
        <Tooltip />
        <Legend
          formatter={(value: string) => (
            <span className="text-xs text-gray-400 capitalize">{value}</span>
          )}
        />
      </PieChart>
    </ResponsiveContainer>
  )
}

export function DeployBarChart({ deployments }: { deployments: Deployment[] }) {
  const daily: Record<string, { success: number; failed: number }> = {}

  for (const d of deployments) {
    const day = new Date(d.created_at).toLocaleDateString('pt-BR', {
      day: '2-digit', month: '2-digit',
    })
    if (!daily[day]) daily[day] = { success: 0, failed: 0 }
    if (d.status === 'success') daily[day].success++
    else if (d.status === 'failed') daily[day].failed++
  }

  const data = Object.entries(daily).map(([day, counts]) => ({
    day,
    ...counts,
  }))

  if (data.length === 0) return null

  return (
    <ResponsiveContainer width="100%" height={200}>
      <BarChart data={data}>
        <XAxis dataKey="day" tick={{ fontSize: 11, fill: '#9ca3af' }} />
        <YAxis tick={{ fontSize: 11, fill: '#9ca3af' }} />
        <Tooltip
          contentStyle={{ backgroundColor: '#1f2937', border: '1px solid #374151', borderRadius: 8 }}
          labelStyle={{ color: '#e5e7eb' }}
        />
        <Bar dataKey="success" name="Sucesso" fill={COLORS.success} radius={[2, 2, 0, 0]} />
        <Bar dataKey="failed" name="Falha" fill={COLORS.failed} radius={[2, 2, 0, 0]} />
      </BarChart>
    </ResponsiveContainer>
  )
}
