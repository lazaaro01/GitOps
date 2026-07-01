import { useState } from 'react'

export default function RollbackModal({
  currentVersion,
  versions,
  onConfirm,
  onClose,
}: {
  currentVersion: string
  versions: string[]
  onConfirm: (targetVersion: string, reason: string) => void
  onClose: () => void
}) {
  const [target, setTarget] = useState('')
  const [reason, setReason] = useState('')

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60">
      <div className="bg-gray-900 border border-gray-700 rounded-xl w-full max-w-lg mx-4 p-6 space-y-4">
        <h2 className="text-lg font-semibold text-gray-100">Confirmar Rollback</h2>

        <div className="text-sm text-gray-400 space-y-1">
          <p>Deploy atual: <span className="text-gray-200">{currentVersion}</span></p>
          <p className="text-yellow-400 text-xs">
            O state do Terraform será restaurado para a versão selecionada
          </p>
        </div>

        <div>
          <label className="block text-xs text-gray-400 mb-1">Rollback para</label>
          <select
            value={target}
            onChange={(e) => setTarget(e.target.value)}
            className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-gray-200"
          >
            <option value="">Selecione uma versão...</option>
            {versions.map((v) => (
              <option key={v} value={v}>{v}</option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-xs text-gray-400 mb-1">Motivo</label>
          <textarea
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-gray-200 h-20 resize-none"
            placeholder="O novo deploy quebrou a API..."
          />
        </div>

        <div className="flex justify-end gap-3 pt-2">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm text-gray-400 hover:text-white transition-colors"
          >
            Cancelar
          </button>
          <button
            onClick={() => target && onConfirm(target, reason)}
            disabled={!target}
            className="px-4 py-2 text-sm bg-red-600 hover:bg-red-500 disabled:bg-gray-700 disabled:text-gray-500 text-white rounded-lg transition-colors"
          >
            Confirmar Rollback
          </button>
        </div>
      </div>
    </div>
  )
}
