import { useMutation, useQueryClient } from '@tanstack/react-query'
import { updateTranscriptSegment } from '~/api/videoApi'
import type { SegmentResponse } from '~/types/video'

interface UpdateSegmentVariables {
  id: number
  text: string
}

interface TranscriptData {
  video_id: string
  segments: SegmentResponse[]
}

/**
 * Hook for updating a single transcript segment with optimistic updates.
 *
 * Features:
 * - Optimistic UI: Updates UI immediately before server responds
 * - Automatic rollback on error
 * - Invalidates query on success to sync with server
 */
export const useUpdateSegment = (videoId: string) => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ id, text }: UpdateSegmentVariables) => {
      return updateTranscriptSegment(id, { text_content: text })
    },

    // CRITICAL: Optimistic update - Update UI immediately
    onMutate: async ({ id, text }) => {
      // Cancel any outgoing refetches to avoid overwriting optimistic update
      await queryClient.cancelQueries({ queryKey: ['transcript', videoId] })

      // Snapshot previous value for rollback
      const previousData = queryClient.getQueryData<TranscriptData>(['transcript', videoId])

      // Optimistically update to the new value
      queryClient.setQueryData<TranscriptData>(['transcript', videoId], (old) => {
        if (!old) return old

        return {
          ...old,
          segments: old.segments.map((seg) =>
            seg.id === id ? { ...seg, text, edited: true } : seg
          ),
        }
      })

      // Return context with previous data for rollback
      return { previousData }
    },

    // Rollback on error
    onError: (err, _variables, context) => {
      if (context?.previousData) {
        queryClient.setQueryData(['transcript', videoId], context.previousData)
      }
      console.error('Failed to update segment:', err)
    },

    // Refetch on success (optional - can be removed if you trust optimistic update)
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['transcript', videoId] })
    },
  })
}
