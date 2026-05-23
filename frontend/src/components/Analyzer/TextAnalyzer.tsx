import { useState } from 'react'
import * as App from '../../../wailsjs/go/ipc/App'
import { normalization } from '../../../wailsjs/go/models'

const issueLabels: Record<number, string> = {
  1: 'Kombuwa before base consonant',
  2: 'Vowel sign before base consonant',
  3: 'Duplicate virama (al-lakuna)',
  4: 'Orphaned virama',
  5: 'Wrong cluster order',
  6: 'Legacy font encoding detected',
}

interface Props {
  onBack: () => void
}

export function TextAnalyzer({ onBack }: Props) {
  const [input, setInput] = useState('')
  const [issues, setIssues] = useState<normalization.Detection[]>([])
  const [repaired, setRepaired] = useState('')
  const [analyzed, setAnalyzed] = useState(false)

  const handleAnalyze = async () => {
    const detected = await App.AnalyzeText(input)
    const fixed = await App.RepairText(input)
    setIssues(detected || [])
    setRepaired(fixed)
    setAnalyzed(true)
  }

  const handleCopyRepaired = () => {
    navigator.clipboard.writeText(repaired)
  }

  const handleReplaceClipboard = () => App.FixClipboard()

  return (
    <div className="flex flex-col h-screen bg-[#0F0F12] text-[#F0F0F5] p-4 gap-4">
      <div className="flex items-center gap-3">
        <button onClick={onBack} className="text-[#8888AA] hover:text-[#F0F0F5] text-sm">← Back</button>
        <span className="text-sm font-semibold">Text Analyzer</span>
      </div>

      <div>
        <label className="text-xs text-[#8888AA] block mb-1">Paste Sinhala text:</label>
        <textarea
          value={input}
          onChange={(e) => { setInput(e.target.value); setAnalyzed(false) }}
          className="w-full h-24 bg-[#1A1A20] border border-[#2A2A38] rounded-lg p-3 text-sm
                     text-[#F0F0F5] resize-none focus:outline-none focus:border-[#4F8EF7]"
          placeholder="ෙකොළඹ…"
          dir="auto"
        />
        <button
          onClick={handleAnalyze}
          disabled={!input.trim()}
          className="mt-2 w-full py-2 text-xs rounded-lg bg-[#4F8EF7] hover:bg-[#4F8EF7]/80
                     disabled:opacity-40 disabled:cursor-not-allowed text-white font-semibold transition-colors"
        >
          Analyze ▶
        </button>
      </div>

      {analyzed && (
        <>
          <div>
            <div className="text-xs text-[#8888AA] mb-2">
              Issues Found:{' '}
              <span className={issues.length > 0 ? 'text-[#E5534B]' : 'text-[#3ECF8E]'}>
                {issues.length}
              </span>
            </div>
            {issues.length === 0 ? (
              <div className="text-xs text-[#3ECF8E] bg-[#1A1A20] rounded-lg p-3 border border-[#3ECF8E]/30">
                ✓ No issues found — text is well-formed
              </div>
            ) : (
              <div className="bg-[#1A1A20] border border-[#2A2A38] rounded-lg divide-y divide-[#2A2A38]">
                {issues.map((issue, i) => (
                  <div key={i} className="p-3">
                    <div className="text-xs text-[#E5534B]">
                      ⚠ Position {issue.StartIndex}: {issueLabels[issue.Issue] ?? `Issue #${issue.Issue}`}
                    </div>
                    {issue.Context && (
                      <div className="text-xs text-[#8888AA] mt-0.5">Context: "{issue.Context}"</div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>

          {repaired && repaired !== input && (
            <div>
              <div className="text-xs text-[#8888AA] mb-1">Repaired:</div>
              <div
                className="bg-[#1A1A20] border border-[#3ECF8E]/30 rounded-lg p-3 text-sm text-[#F0F0F5]"
                dir="auto"
              >
                {repaired} ✓
              </div>
              <div className="flex gap-2 mt-2">
                <button
                  onClick={handleCopyRepaired}
                  className="flex-1 py-1.5 text-xs rounded-lg bg-[#1A1A20] border border-[#2A2A38]
                             text-[#F0F0F5] hover:border-[#4F8EF7] transition-colors"
                >
                  Copy Repaired
                </button>
                <button
                  onClick={handleReplaceClipboard}
                  className="flex-1 py-1.5 text-xs rounded-lg bg-[#4F8EF7] text-white
                             hover:bg-[#4F8EF7]/80 transition-colors"
                >
                  Replace Clipboard
                </button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  )
}
