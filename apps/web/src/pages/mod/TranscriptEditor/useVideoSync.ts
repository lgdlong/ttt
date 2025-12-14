import { useState, useEffect, useCallback, useRef } from 'react'
import type { YouTubePlayer } from 'react-youtube'
import type { SegmentResponse } from '~/types/video'

interface UseVideoSyncOptions {
  playerRef: React.RefObject<YouTubePlayer | null>
  segments: SegmentResponse[]
  playing: boolean
}

/**
 * Hook for syncing video playback with transcript highlighting.
 *
 * Optimizations:
 * - Works with virtualized lists (no DOM element access)
 * - Efficient polling at 200ms intervals
 * - Only updates when segment actually changes
 * - Auto-scroll can be disabled by user interaction
 */
export const useVideoSync = ({ playerRef, segments, playing }: UseVideoSyncOptions) => {
  const [activeIndex, setActiveIndex] = useState(0)
  const syncIntervalRef = useRef<number | null>(null)

  // Update active segment based on video time
  const updateActiveSegment = useCallback(async () => {
    if (!playerRef.current || segments.length === 0) return

    try {
      const player = playerRef.current
      if (!player || !player.getCurrentTime) return

      const currentTimeSec = await player.getCurrentTime()
      const currentTimeMs = currentTimeSec * 1000

      // Find currently playing segment
      const newActiveIndex = segments.findIndex(
        (seg) => seg.start_time <= currentTimeMs && currentTimeMs < seg.end_time
      )

      // Only update if index changed (prevents unnecessary re-renders)
      if (newActiveIndex !== -1 && newActiveIndex !== activeIndex) {
        setActiveIndex(newActiveIndex)
      }
    } catch (err) {
      if (import.meta.env.DEV) {
        console.warn('updateActiveSegment error:', err)
      }
    }
  }, [segments, activeIndex, playerRef])

  // Setup sync interval
  useEffect(() => {
    if (playing && segments.length > 0) {
      syncIntervalRef.current = window.setInterval(updateActiveSegment, 200)
    } else {
      if (syncIntervalRef.current) {
        clearInterval(syncIntervalRef.current)
        syncIntervalRef.current = null
      }
    }

    return () => {
      if (syncIntervalRef.current) {
        clearInterval(syncIntervalRef.current)
      }
    }
  }, [playing, segments, updateActiveSegment])

  return {
    activeIndex,
    setActiveIndex,
  }
}
