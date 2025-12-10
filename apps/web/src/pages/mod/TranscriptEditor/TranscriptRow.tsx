import React, { useRef, useEffect, useCallback, useState } from 'react'
import type { SegmentResponse } from '~/types/video'

interface TranscriptSegment extends SegmentResponse {
  id: number
  edited?: boolean
}

interface TranscriptRowProps {
  segment: TranscriptSegment
  index: number
  isActive: boolean
  onUpdate: (id: number, text: string) => void
  onSeek?: (index: number) => void
  onKeyDown: (e: React.KeyboardEvent, index: number) => void
  onEditStart: () => void
  style?: React.CSSProperties // For virtualization positioning
  measureRef?: (node: Element | null) => void // For dynamic height measurement
}

const formatTime = (ms: number) => {
  const seconds = Math.floor(ms / 1000)
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  return `${m}:${s.toString().padStart(2, '0')}`
}

/**
 * Ultra-lightweight TranscriptRow component with:
 * - Local state management (zero-lag typing)
 * - Native HTML textarea (no MUI overhead)
 * - Minimalist CSS (performance-first)
 * - Virtualization-ready (absolute positioning support)
 */
export const TranscriptRow = React.memo<TranscriptRowProps>(
  ({ segment, index, isActive, onUpdate, onSeek, onKeyDown, onEditStart, style, measureRef }) => {
    const textareaRef = useRef<HTMLTextAreaElement>(null)

    // CRITICAL: Local state - never triggers parent re-render while typing
    const [localText, setLocalText] = useState(segment.text)

    // Sync with parent only when segment changes (after successful save)
    useEffect(() => {
      setLocalText(segment.text)
    }, [segment.text])

    // Auto-focus when active
    useEffect(() => {
      if (isActive && textareaRef.current) {
        textareaRef.current.focus()
      }
    }, [isActive])

    // Auto-resize textarea based on content
    useEffect(() => {
      if (textareaRef.current) {
        textareaRef.current.style.height = 'auto'
        textareaRef.current.style.height = `${textareaRef.current.scrollHeight}px`
      }
    }, [localText])

    const handleChange = useCallback(
      (e: React.ChangeEvent<HTMLTextAreaElement>) => {
        setLocalText(e.target.value)
        onEditStart() // Pause video when editing starts
      },
      [onEditStart]
    )

    const handleBlur = useCallback(() => {
      if (localText !== segment.text) {
        onUpdate(segment.id, localText)
      }
    }, [segment.id, segment.text, localText, onUpdate])

    const handleFocus = useCallback(() => {
      onSeek?.(index)
    }, [index, onSeek])

    const handleKeyDown = useCallback(
      (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
        // Save on Enter, then delegate to parent for navigation
        if (e.key === 'Enter' && !e.shiftKey && localText !== segment.text) {
          onUpdate(segment.id, localText)
        }
        onKeyDown(e, index)
      },
      [segment.id, segment.text, localText, index, onUpdate, onKeyDown]
    )

    return (
      <div
        ref={measureRef}
        data-index={index}
        style={{
          ...style,
          display: 'flex',
          gap: '12px',
          padding: '12px',
          marginBottom: '4px',
          borderRadius: '4px',
          transition: 'all 0.15s ease',
          backgroundColor: isActive ? '#f0f7ff' : '#fafafa',
          opacity: isActive ? 1 : 0.7,
          border: isActive ? '2px solid #1976d2' : '1px solid #e0e0e0',
          cursor: 'text',
          boxSizing: 'border-box',
        }}
        onClick={() => textareaRef.current?.focus()}
      >
        {/* Timestamp */}
        <span
          style={{
            minWidth: '60px',
            paddingTop: '4px',
            fontFamily: 'monospace',
            fontSize: '12px',
            fontWeight: isActive ? 600 : 400,
            color: isActive ? '#1976d2' : '#666',
            userSelect: 'none',
          }}
        >
          {formatTime(segment.start_time)}
        </span>

        {/* Textarea */}
        <textarea
          ref={textareaRef}
          value={localText}
          onChange={handleChange}
          onBlur={handleBlur}
          onFocus={handleFocus}
          onKeyDown={handleKeyDown}
          style={{
            flex: 1,
            border: 'none',
            outline: 'none',
            backgroundColor: 'transparent',
            resize: 'none',
            fontFamily: 'Inter, system-ui, sans-serif',
            fontSize: isActive ? '15px' : '14px',
            fontWeight: isActive ? 500 : 400,
            color: isActive ? '#000' : '#333',
            lineHeight: '1.6',
            minHeight: '24px',
            overflow: 'hidden',
          }}
        />

        {/* Edited indicator */}
        {segment.edited && (
          <span
            style={{
              alignSelf: 'flex-start',
              padding: '2px 8px',
              borderRadius: '12px',
              fontSize: '11px',
              fontWeight: 600,
              color: '#ed6c02',
              backgroundColor: '#fff4e5',
              userSelect: 'none',
            }}
          >
            Edited
          </span>
        )}
      </div>
    )
  }
)

TranscriptRow.displayName = 'TranscriptRow'

export type { TranscriptSegment, TranscriptRowProps }
