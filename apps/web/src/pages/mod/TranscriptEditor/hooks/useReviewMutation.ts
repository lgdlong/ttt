import { useMutation, useQueryClient } from '@tanstack/react-query'
import axiosInstance from '~/lib/axios'
import type { SubmitReviewRequest, VideoTranscriptReviewResponse } from '~/types/video'

interface UseSubmitReviewParams {
  videoId: string
}

/**
 * Hook for submitting transcript review
 *
 * Usage:
 * ```ts
 * const { mutate: submitReview, isPending } = useSubmitReview({ videoId })
 *
 * submitReview(
 *   { notes: 'Optional review notes' },
 *   {
 *     onSuccess: (data) => {
 *       toast.success(data.message)
 *     },
 *     onError: (error) => {
 *       toast.error(error.response?.data?.message || 'Failed to submit review')
 *     }
 *   }
 * )
 * ```
 */
export function useSubmitReview({ videoId }: UseSubmitReviewParams) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (
      request: SubmitReviewRequest = {}
    ): Promise<VideoTranscriptReviewResponse> => {
      const response = await axiosInstance.post<VideoTranscriptReviewResponse>(
        `/videos/${videoId}/reviews`,
        request
      )
      return response.data
    },
    onSuccess: () => {
      // Invalidate review stats and user review status queries
      queryClient.invalidateQueries({
        queryKey: ['video-review-stats', videoId],
      })
      queryClient.invalidateQueries({
        queryKey: ['user-review-status', videoId],
      })
      // Optionally invalidate video list to update status
      queryClient.invalidateQueries({
        queryKey: ['videos'],
      })
    },
  })
}

/**
 * Hook for checking if user has reviewed a video
 *
 * Usage:
 * ```ts
 * const { data: reviewStatus } = useUserReviewStatus({ videoId })
 * if (reviewStatus?.has_reviewed) {
 *   // Show "Already reviewed" message
 * }
 * ```
 */
export function useUserReviewStatus({ videoId }: UseSubmitReviewParams) {
  return {
    queryKey: ['user-review-status', videoId],
    queryFn: async () => {
      const response = await axiosInstance.get<{
        video_id: string
        has_reviewed: boolean
      }>(`/videos/${videoId}/reviews/status`)
      return response.data
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 10 * 60 * 1000, // 10 minutes
  }
}

/**
 * Hook for getting video review statistics
 *
 * Usage:
 * ```ts
 * const { data: reviewStats } = useVideoReviewStats({ videoId })
 * console.log(`${reviewStats?.review_count} reviews`)
 * ```
 */
export function useVideoReviewStats({ videoId }: UseSubmitReviewParams) {
  return {
    queryKey: ['video-review-stats', videoId],
    queryFn: async () => {
      const response = await axiosInstance.get<{
        video_id: string
        review_count: number
      }>(`/videos/${videoId}/reviews/stats`)
      return response.data
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 10 * 60 * 1000, // 10 minutes
  }
}
