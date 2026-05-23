import { create } from 'zustand'

interface EngineState {
  active: boolean
  mapping: string
  error: string | null
  setActive: (v: boolean) => void
  setMapping: (v: string) => void
  setError: (v: string | null) => void
}

export const useEngineStore = create<EngineState>((set) => ({
  active: false,
  mapping: 'phonetic',
  error: null,
  setActive: (active) => set({ active, error: null }),
  setMapping: (mapping) => set({ mapping }),
  setError: (error) => set({ error }),
}))
