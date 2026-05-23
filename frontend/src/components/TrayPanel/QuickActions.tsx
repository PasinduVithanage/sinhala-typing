interface QuickActionsProps {
  onFixClipboard: () => void
  onAnalyze: () => void
  onGuide: () => void
}

export function QuickActions({ onFixClipboard, onAnalyze, onGuide }: QuickActionsProps) {
  return (
    <div className="flex gap-2">
      <button
        onClick={onFixClipboard}
        className="flex-1 py-2 px-3 text-xs rounded-lg bg-[#1A1A20] border border-[#2A2A38]
                   text-[#F0F0F5] hover:bg-[#242430] hover:border-[#4F8EF7] transition-colors"
      >
        Fix Clipboard
      </button>
      <button
        onClick={onAnalyze}
        className="flex-1 py-2 px-3 text-xs rounded-lg bg-[#1A1A20] border border-[#2A2A38]
                   text-[#F0F0F5] hover:bg-[#242430] hover:border-[#4F8EF7] transition-colors"
      >
        Analyze Text
      </button>
      <button
        onClick={onGuide}
        className="py-2 px-3 text-xs rounded-lg bg-[#1A1A20] border border-[#2A2A38]
                   text-[#4F8EF7] hover:bg-[#242430] hover:border-[#4F8EF7] transition-colors"
        title="Phonetic typing guide"
      >
        ?
      </button>
    </div>
  )
}
