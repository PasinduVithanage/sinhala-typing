import type { CSSProperties } from 'react'

interface PhoneticGuideProps {
  onBack: () => void
}

const EXAMPLES = [
  { roman: 'pa·si·du',   sinhala: 'පසිදු' },
  { roman: 'si·n·ha·la', sinhala: 'සිංහල' },
  { roman: 'a·yu·boo·wa·n', sinhala: 'ආයුබෝවන්' },
  { roman: 'ka·ma·ra',   sinhala: 'කාමර' },
  { roman: 'ge·da·ra',   sinhala: 'ගෙදර' },
]

// Using ක as the base consonant to show all vowel combinations
const VOWEL_COMBOS = [
  { keys: 'k',   sinhala: 'ක',  note: 'inherent a' },
  { keys: 'ka',  sinhala: 'කා', note: 'ā (long a)' },
  { keys: 'ki',  sinhala: 'කි', note: 'i' },
  { keys: 'kI',  sinhala: 'කී', note: 'ī (long i)' },
  { keys: 'ku',  sinhala: 'කු', note: 'u' },
  { keys: 'kU',  sinhala: 'කූ', note: 'ū (long u)' },
  { keys: 'kE',  sinhala: 'කෙ', note: 'e' },
  { keys: 'kee', sinhala: 'කේ', note: 'ē (long e)' },
  { keys: 'ko',  sinhala: 'කො', note: 'o' },
  { keys: 'koo', sinhala: 'කෝ', note: 'ō (long o)' },
  { keys: 'kH',  sinhala: 'ක්', note: 'hal (no vowel)' },
]

const CONSONANTS = [
  ['k','ක'],['g','ග'],['ng','ඞ'],['c','ච'],['j','ජ'],['t','ත'],
  ['d','ද'],['n','න'],['p','ප'],['b','බ'],['m','ම'],['y','ය'],
  ['r','ර'],['l','ල'],['v','ව'],['s','ස'],['h','හ'],['sh','ශ'],
  ['T','ට'],['D','ඩ'],['N','ණ'],['L','ළ'],['f','ෆ'],
]

const SPECIALS = [
  { keys: 'M',  sinhala: 'ං', label: 'anusvara' },
  { keys: 'kHr', sinhala: 'ක්‍ර', label: 'rakar cluster' },
  { keys: 'kHy', sinhala: 'ක්‍ය', label: 'yansaya cluster' },
  { keys: 'rHk', sinhala: 'ර්‍ක', label: 'repaya cluster' },
]

export function PhoneticGuide({ onBack }: PhoneticGuideProps) {
  return (
    <div className="flex flex-col h-screen bg-[#0F0F12] text-[#F0F0F5] select-none">
      {/* Header */}
      <div
        className="flex items-center gap-2 px-4 py-3 border-b border-[#2A2A38]"
        style={{ WebkitAppRegion: 'drag' } as CSSProperties}
      >
        <button
          onClick={onBack}
          style={{ WebkitAppRegion: 'no-drag' } as CSSProperties}
          className="text-[#8888AA] hover:text-[#F0F0F5] text-xs w-6 h-6 flex items-center justify-center rounded"
        >
          ←
        </button>
        <span className="text-sm font-semibold">Phonetic Typing Guide</span>
      </div>

      {/* Scrollable content */}
      <div className="flex-1 overflow-y-auto px-4 py-3 space-y-4">

        {/* Concept intro */}
        <div className="bg-[#1A1A20] rounded-xl p-3 border border-[#4F8EF7]/30">
          <div className="text-[10px] text-[#8888AA] uppercase tracking-wider mb-2">How it works</div>
          <p className="text-xs text-[#C0C0D0] leading-relaxed">
            Type English letters that <em className="text-[#4F8EF7] not-italic">sound like</em> the Sinhala syllable.
            Consonant + vowel keys combine automatically.
          </p>
        </div>

        {/* Word examples */}
        <div>
          <div className="text-[10px] text-[#8888AA] uppercase tracking-wider mb-2">Word examples</div>
          <div className="space-y-1.5">
            {EXAMPLES.map(({ roman, sinhala }) => (
              <div
                key={roman}
                className="flex items-center justify-between bg-[#1A1A20] rounded-lg px-3 py-2
                           border border-[#2A2A38]"
              >
                <span className="text-xs font-mono text-[#4F8EF7]">{roman}</span>
                <span className="text-base text-[#F0F0F5]">{sinhala}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Vowel combinations */}
        <div>
          <div className="text-[10px] text-[#8888AA] uppercase tracking-wider mb-2">
            Vowel signs <span className="normal-case">(using ක as base)</span>
          </div>
          <div className="grid grid-cols-2 gap-1">
            {VOWEL_COMBOS.map(({ keys, sinhala, note }) => (
              <div
                key={keys}
                className="flex items-center gap-2 bg-[#1A1A20] rounded-lg px-2 py-1.5
                           border border-[#2A2A38]"
              >
                <span className="text-[11px] font-mono text-[#4F8EF7] w-8 shrink-0">{keys}</span>
                <span className="text-base text-[#F0F0F5] w-6 shrink-0">{sinhala}</span>
                <span className="text-[9px] text-[#6666AA] truncate">{note}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Consonants reference */}
        <div>
          <div className="text-[10px] text-[#8888AA] uppercase tracking-wider mb-2">Consonants</div>
          <div className="grid grid-cols-4 gap-1">
            {CONSONANTS.map(([roman, sinhala]) => (
              <div
                key={roman}
                className="flex items-center justify-between bg-[#1A1A20] rounded-lg px-2 py-1.5
                           border border-[#2A2A38]"
              >
                <span className="text-[11px] font-mono text-[#4F8EF7]">{roman}</span>
                <span className="text-sm text-[#F0F0F5]">{sinhala}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Special / clusters */}
        <div>
          <div className="text-[10px] text-[#8888AA] uppercase tracking-wider mb-2">Special & clusters</div>
          <div className="space-y-1">
            {SPECIALS.map(({ keys, sinhala, label }) => (
              <div
                key={keys}
                className="flex items-center gap-3 bg-[#1A1A20] rounded-lg px-3 py-1.5
                           border border-[#2A2A38]"
              >
                <span className="text-[11px] font-mono text-[#4F8EF7] w-10 shrink-0">{keys}</span>
                <span className="text-base text-[#F0F0F5] w-8 shrink-0">{sinhala}</span>
                <span className="text-[10px] text-[#6666AA]">{label}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Tip */}
        <div className="bg-[#1A1A20] rounded-xl p-3 border border-[#3ECF8E]/20">
          <div className="text-[10px] text-[#3ECF8E] uppercase tracking-wider mb-1">Tip</div>
          <p className="text-[10px] text-[#8888AA] leading-relaxed">
            Uppercase <span className="text-[#4F8EF7] font-mono">T D N L</span> = retroflex sounds.
            Use <span className="text-[#4F8EF7] font-mono">H</span> for hal-akura (্).
            Space commits the current syllable.
          </p>
        </div>

      </div>
    </div>
  )
}
