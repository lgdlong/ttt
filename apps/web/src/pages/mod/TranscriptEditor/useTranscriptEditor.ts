import { useState, useCallback } from 'react'
import type { YouTubePlayer } from 'react-youtube'
import type { SegmentResponse } from '~/types/video'

interface UseTranscriptEditorOptions {
  playerRef: React.RefObject<YouTubePlayer | null>
  segments: SegmentResponse[]
  setActiveIndex: (index: number) => void
}

/**
 * Simplified editor logic hook for keyboard shortcuts and playback control.
 *
 * Note: Text editing and saving is now handled by:
 * - TranscriptRow (local state)
 * - useUpdateSegment (API mutations)
 */
export const useTranscriptEditor = ({
  playerRef,
  segments,
  setActiveIndex,
}: UseTranscriptEditorOptions) => {
  const [playing, setPlaying] = useState(false)

  // Handle keyboard shortcuts for navigation and playback
  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent, index: number) => {
      // Enter: Next line + Play
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault()
        const nextIndex = Math.min(index + 1, segments.length - 1)
        setActiveIndex(nextIndex)
        setPlaying(true)
      }

      // Shift + Enter: Previous line + Seek
      if (e.key === 'Enter' && e.shiftKey) {
        e.preventDefault()
        const prevIndex = Math.max(index - 1, 0)
        setActiveIndex(prevIndex)
        if (playerRef.current && segments[prevIndex]) {
          playerRef.current.seekTo(segments[prevIndex].start_time / 1000, true)
        }
      }

      // Ctrl + Space: Toggle Play/Pause
      if (e.ctrlKey && e.key === ' ') {
        e.preventDefault()
        setPlaying((prev) => !prev)
      }

      // Ctrl + R: Replay current segment
      if (e.ctrlKey && e.key === 'r') {
        e.preventDefault()
        if (playerRef.current && segments[index]) {
          playerRef.current.seekTo(segments[index].start_time / 1000, true)
          setPlaying(true)
        }
      }
    },
    [segments, setActiveIndex, playerRef]
  )

  return {
    playing,
    setPlaying,
    handleKeyDown,
  }
}
