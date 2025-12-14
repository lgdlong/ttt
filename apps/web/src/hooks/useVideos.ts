import { useQuery, useSuspenseQuery } from '@tanstack/react-query'
import { videoApi } from '~/api/videoApi'
import type { ListVideoRequest, TranscriptSearchRequest, TagSearchRequest } from '~/types/video'

/**
 * Query keys for video-related queries
 */
export const videoKeys = {
  all: ['videos'] as const,
  lists: () => [...videoKeys.all, 'list'] as const,
  list: (params?: ListVideoRequest) => [...videoKeys.lists(), params] as const,
  details: () => [...videoKeys.all, 'detail'] as const,
  detail: (id: string) => [...videoKeys.details(), id] as const,
  transcript: (videoId: string) => [...videoKeys.all, 'transcript', videoId] as const,
}

export const searchKeys = {
  all: ['search'] as const,
  transcripts: (params: TranscriptSearchRequest) =>
    [...searchKeys.all, 'transcripts', params] as const,
  tags: (params: TagSearchRequest) => [...searchKeys.all, 'tags', params] as const,
}

/**
 * Hook to fetch videos list with pagination
 * Uses useSuspenseQuery for better UX with Suspense boundaries
 */
export function useVideos(params?: ListVideoRequest) {
  return useSuspenseQuery({
    queryKey: videoKeys.list(params),
    queryFn: () => videoApi.fetchVideos(params),
  })
}

/**
 * Hook to fetch single video detail
 */
export function useVideoDetail(id: string) {
  return useSuspenseQuery({
    queryKey: videoKeys.detail(id),
    queryFn: () => videoApi.fetchVideoById(id),
  })
}

/**
 * Hook to fetch video transcript
 */
export function useVideoTranscript(videoId: string) {
  return useSuspenseQuery({
    queryKey: videoKeys.transcript(videoId),
    queryFn: () => videoApi.fetchVideoTranscript(videoId),
  })
}

/**
 * Hook to search transcripts
 */
export function useTranscriptSearch(params: TranscriptSearchRequest) {
  return useQuery({
    queryKey: searchKeys.transcripts(params),
    queryFn: () => videoApi.searchTranscripts(params),
    enabled: params.q.length >= 2, // Only search when query is at least 2 chars
  })
}
