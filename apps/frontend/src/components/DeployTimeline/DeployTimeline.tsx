import type { DeploymentLog } from '../../types'

const steps = [
  { key: 'received',       label: 'Recebido' },
  { key: 'terraform_init', label: 'Terraform Init' },
  { key: 'terraform_plan', label: 'Terraform Plan' },
  { key: 'terraform_apply',label: 'Terraform Apply' },
  { key: 'health_check',   label: 'Health Check' },
  { key: 'completed',      label: 'Concluído' },
]

function statusForStep(step: string, logs: DeploymentLog[], deploymentStatus: string): 'pending' | 'running' | 'done' | 'failed' {
  const stepLogs = logs.filter((l) => l.step === step)
  if (stepLogs.length === 0) {
    if (deploymentStatus === 'success' || deploymentStatus === 'failed') return 'done'
    return 'pending'
  }
  const hasError = stepLogs.some((l) => l.level === 'error')
  if (hasError) return 'failed'

  if (step === 'completed') {
    const hasComplete = stepLogs.some((l) => l.message.includes('completed'))
    return hasComplete ? 'done' : 'pending'
  }

  return 'done'
}

export default function DeployTimeline({
  logs,
  deploymentStatus,
}: {
  logs: DeploymentLog[]
  deploymentStatus: string
}) {
  return (
    <div className="space-y-2">
      {steps.map((step) => {
        const st = statusForStep(step.key, logs, deploymentStatus)
        return (
          <div key={step.key} className="flex items-center gap-3">
            <div className="flex-shrink-0 w-6 flex justify-center">
              {st === 'done' && <span className="text-emerald-400 text-sm">✓</span>}
              {st === 'failed' && <span className="text-red-400 text-sm">✕</span>}
              {st === 'running' && (
                <span className="w-2 h-2 rounded-full bg-yellow-400 animate-pulse" />
              )}
              {st === 'pending' && <span className="text-gray-600 text-sm">○</span>}
            </div>
            <span className={`text-sm ${
              st === 'done' ? 'text-emerald-300' :
              st === 'failed' ? 'text-red-300' :
              st === 'running' ? 'text-yellow-300' : 'text-gray-500'
            }`}>
              {step.label}
            </span>
          </div>
        )
      })}
    </div>
  )
}
