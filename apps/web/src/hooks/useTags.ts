import { useQuery, useSuspenseQuery } from '@tanstack/react-query'
import { tagApi } from '~/api/tagApi'
import type { TagSearchRequest } from '~/types/video'

/**
 * Query keys for tag-related queries
 */
export const tagKeys = {
  all: ['tags'] as const,
  lists: () => [...tagKeys.all, 'list'] as const,
  list: (params?: { page?: number; limit?: number }) => [...tagKeys.lists(), params] as const,
  details: () => [...tagKeys.all, 'detail'] as const,
  detail: (id: string) => [...tagKeys.details(), id] as const,
  search: (params: TagSearchRequest) => [...tagKeys.all, 'search', params] as const,
}

/**
 * Hook to fetch all approved tags
 * Uses regular query (not suspense) as it's less critical data
 */
export function useTags() {
  return useQuery({
    queryKey: tagKeys.list(),
    queryFn: () => tagApi.fetchAllTagsApproved(),
    staleTime: Infinity, // Tags rarely change
  })
}

/**
 * Hook to fetch single tag by ID
 * Uses suspense for better UX with Suspense boundaries
 */
export function useTagDetail(id: string) {
  return useSuspenseQuery({
    queryKey: tagKeys.detail(id),
    queryFn: () => tagApi.fetchTagById(id),
  })
}

/**
 * Hook to search tags by semantic similarity
 */
export function useTagSearch(params: TagSearchRequest) {
  return useQuery({
    queryKey: tagKeys.search(params),
    queryFn: () => tagApi.searchTags(params),
    enabled: params.q.length >= 2,
  })
}
