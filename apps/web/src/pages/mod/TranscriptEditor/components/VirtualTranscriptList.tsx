import React, { useRef, useEffect, useCallback } from 'react'
import { useVirtualizer } from '@tanstack/react-virtual'
import { useQuery } from '@tanstack/react-query'
import type { YouTubePlayer } from 'react-youtube'
import axiosInstance from '~/lib/axios'
import type { SegmentResponse } from '~/types/video'
import { TranscriptRow, type TranscriptSegment } from '../TranscriptRow'
import { useUpdateSegment } from '../hooks/useSegmentMutation'

interface TranscriptData {
  video_id: string
  segments: SegmentResponse[]
}

interface VirtualTranscriptListProps {
  videoId: string
  playerRef: React.RefObject<YouTubePlayer | null>
  activeIndex: number
  shouldScrollToActive: boolean
  onActiveIndexChange: (index: number) => void
  onEditStart: () => void
  onKeyDown: (e: React.KeyboardEvent, index: number) => void
}

/**
 * High-performance virtualized transcript list.
 *
 * Performance optimizations:
 * - Only renders visible rows (viewport + overscan)
 * - Reduces DOM nodes from 600+ to ~20
 * - Uses TanStack Virtual for efficient scrolling
 * - Atomic updates via PATCH /transcript-segments/:id
 * - Optimistic UI for instant feedback
 */
export const VirtualTranscriptList: React.FC<VirtualTranscriptListProps> = ({
  videoId,
  activeIndex,
  shouldScrollToActive,
  onActiveIndexChange,
  onEditStart,
  onKeyDown,
}) => {
  const parentRef = useRef<HTMLDivElement>(null)

  // Fetch transcript data using TanStack Query
  const { data, isLoading, error } = useQuery({
    queryKey: ['transcript', videoId],
    queryFn: async () => {
      const response = await axiosInstance.get<TranscriptData>(`/videos/${videoId}/transcript`)
      return response.data
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 10 * 60 * 1000, // 10 minutes
  })

  // Mutation for updating segments
  const { mutate: updateSegment } = useUpdateSegment(videoId)

  // Setup virtualizer
  const rowVirtualizer = useVirtualizer({
    count: data?.segments.length || 0,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 60, // Estimated row height
    overscan: 10, // Render 10 extra rows above/below viewport
  })

  // Auto-scroll to active segment when it changes
  useEffect(() => {
    if (shouldScrollToActive && activeIndex >= 0 && activeIndex < (data?.segments.length || 0)) {
      rowVirtualizer.scrollToIndex(activeIndex, {
        align: 'center',
        behavior: 'smooth',
      })
    }
  }, [shouldScrollToActive, activeIndex, data?.segments.length, rowVirtualizer])

  const handleUpdate = useCallback(
    (id: number, text: string) => {
      updateSegment({ id, text })
    },
    [updateSegment]
  )

  const handleSeek = useCallback(
    (index: number) => {
      onActiveIndexChange(index)
    },
    [onActiveIndexChange]
  )

  if (isLoading) {
    return (
      <div
        style={{
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          height: '100%',
          color: '#666',
        }}
      >
        Loading transcript...
      </div>
    )
  }

  if (error) {
    return (
      <div
        style={{
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          height: '100%',
          color: '#d32f2f',
        }}
      >
        Failed to load transcript. Please try again.
      </div>
    )
  }

  if (!data || data.segments.length === 0) {
    return (
      <div
        style={{
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          height: '100%',
          color: '#666',
        }}
      >
        No transcript available.
      </div>
    )
  }

  const segments: TranscriptSegment[] = data.segments.map((seg) => ({
    ...seg,
    edited: false, // Initial state, will be updated by mutation
  }))

  return (
    <div
      ref={parentRef}
      style={{
        height: '100%',
        overflowY: 'auto',
        padding: '16px',
      }}
    >
      {/* Header */}
      <div
        style={{
          marginBottom: '16px',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}
      >
        <h3 style={{ margin: 0, fontSize: '18px', fontWeight: 600, color: '#333' }}>
          Transcript ({segments.length} segments)
        </h3>
      </div>

      {/* Virtual list container */}
      <div
        style={{
          height: `${rowVirtualizer.getTotalSize()}px`,
          width: '100%',
          position: 'relative',
        }}
      >
        {rowVirtualizer.getVirtualItems().map((virtualItem) => {
          const segment = segments[virtualItem.index]
          return (
            <TranscriptRow
              key={virtualItem.key}
              segment={segment}
              index={virtualItem.index}
              isActive={virtualItem.index === activeIndex}
              onUpdate={handleUpdate}
              onSeek={handleSeek}
              onKeyDown={onKeyDown}
              onEditStart={onEditStart}
              measureRef={rowVirtualizer.measureElement}
              style={{
                position: 'absolute',
                top: 0,
                left: 0,
                width: '100%',
                transform: `translateY(${virtualItem.start}px)`,
                boxSizing: 'border-box',
              }}
            />
          )
        })}
      </div>
    </div>
  )
}

export default VirtualTranscriptList
