import { useMutation, useQueryClient } from '@tanstack/react-query'
import type { TagResponse } from '~types/tag'
import { createVideo, deleteVideo, addTagsToVideo, removeTagFromVideo } from '~/api/modApi'

interface UseVideoMutationsCallbacks {
  onCreateSuccess?: () => void
  onDeleteSuccess?: () => void
  onAddTagSuccess?: (updatedTags: TagResponse[]) => void
  onRemoveTagSuccess?: () => void
}

export const useVideoMutations = (callbacks?: UseVideoMutationsCallbacks) => {
  const queryClient = useQueryClient()

  const createMutation = useMutation({
    mutationFn: ({ youtubeId, tagIds }: { youtubeId: string; tagIds?: number[] }) =>
      createVideo(youtubeId, tagIds),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mod-videos'] })
      callbacks?.onCreateSuccess?.()
    },
  })

  const deleteMutation = useMutation({
    mutationFn: deleteVideo,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mod-videos'] })
      callbacks?.onDeleteSuccess?.()
    },
  })

  const addTagMutation = useMutation({
    mutationFn: ({ videoId, tagIds }: { videoId: string; tagIds: string[] }) =>
      addTagsToVideo(videoId, tagIds),
    onSuccess: (updatedTags) => {
      queryClient.invalidateQueries({ queryKey: ['mod-videos'] })
      callbacks?.onAddTagSuccess?.(updatedTags)
    },
  })

  const removeTagMutation = useMutation({
    mutationFn: ({ videoId, tagId }: { videoId: string; tagId: string }) =>
      removeTagFromVideo(videoId, tagId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mod-videos'] })
      callbacks?.onRemoveTagSuccess?.()
    },
  })

  return {
    createMutation,
    deleteMutation,
    addTagMutation,
    removeTagMutation,
  }
}
