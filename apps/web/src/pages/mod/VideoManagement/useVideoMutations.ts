import { useMutation, useQueryClient } from '@tanstack/react-query'
import type { Video } from '~types/video'
import { createVideo, deleteVideo, addTagsToVideo, removeTagFromVideo } from '~/api/modApi'

interface UseVideoMutationsCallbacks {
  onCreateSuccess?: () => void
  onDeleteSuccess?: () => void
  onAddTagSuccess?: (updatedVideo: Video) => void
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
    mutationFn: ({ videoId, tagIds }: { videoId: number; tagIds: number[] }) =>
      addTagsToVideo(videoId, tagIds),
    onSuccess: (updatedVideo) => {
      queryClient.invalidateQueries({ queryKey: ['mod-videos'] })
      callbacks?.onAddTagSuccess?.(updatedVideo)
    },
  })

  const removeTagMutation = useMutation({
    mutationFn: ({ videoId, tagId }: { videoId: number; tagId: string }) =>
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
