import { useEffect } from 'react'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import { useEngineStore } from '../store/engineStore'

export function useEngine() {
  const { setActive, setError } = useEngineStore()

  useEffect(() => {
    EventsOn('engine:started', () => setActive(true))
    EventsOn('engine:stopped', () => setActive(false))
    EventsOn('engine:error', (msg: string) => setError(msg))

    return () => {
      EventsOff('engine:started')
      EventsOff('engine:stopped')
      EventsOff('engine:error')
    }
  }, [setActive, setError])
}
