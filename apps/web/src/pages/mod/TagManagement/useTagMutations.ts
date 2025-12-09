import { useMutation, useQueryClient } from '@tanstack/react-query'
import { createTag, updateTag, deleteTag } from './api'

interface UseTagMutationsCallbacks {
  onCreateSuccess?: () => void
  onUpdateSuccess?: () => void
  onDeleteSuccess?: () => void
}

export const useTagMutations = (callbacks?: UseTagMutationsCallbacks) => {
  const queryClient = useQueryClient()

  const createMutation = useMutation({
    mutationFn: createTag,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mod-tags'] })
      callbacks?.onCreateSuccess?.()
    },
  })

  const updateMutation = useMutation({
    mutationFn: updateTag,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mod-tags'] })
      callbacks?.onUpdateSuccess?.()
    },
  })

  const deleteMutation = useMutation({
    mutationFn: deleteTag,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mod-tags'] })
      callbacks?.onDeleteSuccess?.()
    },
  })

  return {
    createMutation,
    updateMutation,
    deleteMutation,
  }
}
