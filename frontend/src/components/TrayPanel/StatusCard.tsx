import { useEngineStore } from '../../store/engineStore'

interface StatusCardProps {
  onToggle: () => void
}

export function StatusCard({ onToggle }: StatusCardProps) {
  const { active, error } = useEngineStore()

  return (
    <div className="bg-[#1A1A20] rounded-xl p-5 border border-[#2A2A38]">
      <div className="flex items-center justify-between">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <span className={`w-2.5 h-2.5 rounded-full ${active ? 'bg-[#3ECF8E]' : 'bg-[#8888AA]'}`} />
            <span className="text-sm font-semibold text-[#F0F0F5]">
              SINHALA TYPING
            </span>
          </div>
          <div className={`text-xs ml-4 ${active ? 'text-[#3ECF8E]' : 'text-[#8888AA]'}`}>
            {active ? 'ACTIVE' : 'INACTIVE'}
          </div>
          {error && <div className="text-xs text-[#E5534B] mt-1 ml-4">{error}</div>}
        </div>
        <button
          onClick={onToggle}
          className={`relative w-12 h-6 rounded-full transition-colors duration-200 ${
            active ? 'bg-[#3ECF8E]' : 'bg-[#2A2A38]'
          }`}
        >
          <span
            className={`absolute top-0.5 w-5 h-5 bg-white rounded-full shadow transition-transform duration-200 ${
              active ? 'translate-x-6' : 'translate-x-0.5'
            }`}
          />
        </button>
      </div>
    </div>
  )
}
