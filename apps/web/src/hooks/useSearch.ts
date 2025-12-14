import { useQuery } from '@tanstack/react-query'
import { videoApi } from '~/api/videoApi'
import type { ListVideoRequest } from '~/types/video'

/**
 * Query keys for search
 */
export const searchQueryKeys = {
  all: ['globalSearch'] as const,
  videos: (query: string) => [...searchQueryKeys.all, 'videos', query] as const,
}

/**
 * Hook for searching videos with debounce
 * Returns search results with title and tags
 */
export function useVideoSearch(query: string, enabled: boolean = true) {
  const trimmedQuery = query.trim()

  return useQuery({
    queryKey: searchQueryKeys.videos(trimmedQuery),
    queryFn: async () => {
      const params: ListVideoRequest = {
        q: trimmedQuery,
        page: 1,
        limit: 10,
        has_transcript: true,
      }
      return videoApi.fetchVideos(params)
    },
    enabled: enabled && trimmedQuery.length >= 2,
    staleTime: 30000, // 30 seconds
    gcTime: 60000, // 1 minute
  })
}
