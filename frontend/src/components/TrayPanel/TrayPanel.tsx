import { useState } from 'react'
import type { CSSProperties } from 'react'
import { StatusCard } from './StatusCard'
import { QuickActions } from './QuickActions'
import { useEngineStore } from '../../store/engineStore'
import { TextAnalyzer } from '../Analyzer/TextAnalyzer'
import * as App from '../../../wailsjs/go/ipc/App'

type View = 'main' | 'analyze'

interface TrayPanelProps {
  onSettings: () => void
}

export function TrayPanel({ onSettings }: TrayPanelProps) {
  const { mapping, setMapping } = useEngineStore()
  const [view, setView] = useState<View>('main')
  const [fixResult, setFixResult] = useState<string | null>(null)

  const handleToggle = () => App.ToggleEngine()

  const handleFixClipboard = async () => {
    const result = await App.FixClipboard()
    if (result.fixed) {
      setFixResult(`Fixed: ${result.original?.slice(0, 20)}… → ${result.repaired?.slice(0, 20)}…`)
    } else {
      setFixResult(result.message || result.error || 'No issues found')
    }
    setTimeout(() => setFixResult(null), 4000)
  }

  if (view === 'analyze') {
    return <TextAnalyzer onBack={() => setView('main')} />
  }

  return (
    <div className="flex flex-col h-screen bg-[#0F0F12] text-[#F0F0F5] p-4 gap-4 select-none">
      {/* Drag bar */}
      <div className="flex items-center justify-between" style={{ WebkitAppRegion: 'drag' } as CSSProperties}>
        <div className="flex items-center gap-2">
          <span className="text-[#4F8EF7] text-sm">◉</span>
          <span className="text-sm font-semibold">Sinhala Assistant</span>
        </div>
        <button
          onClick={() => App.Quit()}
          style={{ WebkitAppRegion: 'no-drag' } as CSSProperties}
          className="text-[#8888AA] hover:text-[#F0F0F5] text-xs w-6 h-6 flex items-center justify-center rounded"
        >
          ✕
        </button>
      </div>

      <StatusCard onToggle={handleToggle} />

      {/* Layout selector */}
      <div className="flex items-center gap-3">
        <span className="text-xs text-[#8888AA]">Layout:</span>
        <select
          value={mapping}
          onChange={(e) => setMapping(e.target.value)}
          className="flex-1 bg-[#1A1A20] border border-[#2A2A38] rounded-lg px-3 py-1.5
                     text-xs text-[#F0F0F5] focus:outline-none focus:border-[#4F8EF7]"
        >
          <option value="phonetic">Phonetic</option>
          <option value="wijesekera">Wijesekera</option>
        </select>
      </div>

      <QuickActions onFixClipboard={handleFixClipboard} onAnalyze={() => setView('analyze')} />

      {fixResult && (
        <div className="text-xs text-[#3ECF8E] bg-[#1A1A20] border border-[#3ECF8E]/30 rounded-lg px-3 py-2">
          {fixResult}
        </div>
      )}

      <div className="border-t border-[#2A2A38]" />

      {/* Footer */}
      <div className="flex justify-between items-center mt-auto">
        <button
          onClick={onSettings}
          className="text-xs text-[#8888AA] hover:text-[#F0F0F5] transition-colors"
        >
          ⚙ Settings
        </button>
        <button
          onClick={() => App.Quit()}
          className="text-xs text-[#8888AA] hover:text-[#E5534B] transition-colors"
        >
          Quit
        </button>
      </div>
    </div>
  )
}
