import { useEngineStore } from '../../store/engineStore'

// Sinhala character shown on each key for the phonetic layout
const PHONETIC_LABELS: Record<string, string> = {
  q: '', w: '', e: 'එ', r: 'ර', t: 'ත', y: 'ය', u: 'උ', i: 'ඉ', o: 'ඔ', p: 'ප',
  a: 'අ', s: 'ස', d: 'ද', f: 'ෆ', g: 'ග', h: 'හ', j: 'ජ', k: 'ක', l: 'ල',
  z: '', x: '', c: 'ච', v: 'ව', b: 'බ', n: 'න', m: 'ම',
}

const KEYBOARD_ROWS = [
  ['q','w','e','r','t','y','u','i','o','p'],
  ['a','s','d','f','g','h','j','k','l'],
  ['z','x','c','v','b','n','m'],
]

export function KeyLayout() {
  const { mapping } = useEngineStore()

  return (
    <div className="p-4 bg-[#1A1A20] rounded-xl border border-[#2A2A38]">
      <div className="text-xs text-[#8888AA] mb-3">
        Keyboard Layout — {mapping}
      </div>
      {KEYBOARD_ROWS.map((row, i) => (
        <div key={i} className="flex gap-1 mb-1 justify-center">
          {row.map((key) => {
            const sinhala = PHONETIC_LABELS[key]
            return (
              <div
                key={key}
                className={`w-9 h-9 rounded-md border flex flex-col items-center justify-center ${
                  sinhala
                    ? 'bg-[#242430] border-[#4F8EF7]/40 text-white'
                    : 'bg-[#1A1A20] border-[#2A2A38] text-[#4A4A5A]'
                }`}
              >
                {sinhala && <span className="text-[10px] text-[#4F8EF7] leading-none">{sinhala}</span>}
                <span className="text-[9px] text-[#8888AA] leading-none">{key}</span>
              </div>
            )
          })}
        </div>
      ))}
    </div>
  )
}
