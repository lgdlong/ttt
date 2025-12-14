import { useMutation, useQueryClient } from '@tanstack/react-query'
import { createTag, mergeTags, updateTagApproval } from './api'

interface UseTagMutationsCallbacks {
  onCreateSuccess?: () => void
  onMergeSuccess?: () => void
  onApprovalSuccess?: () => void
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

  const mergeMutation = useMutation({
    mutationFn: mergeTags,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mod-tags'] })
      callbacks?.onMergeSuccess?.()
    },
  })

  const approvalMutation = useMutation({
    mutationFn: updateTagApproval,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mod-tags'] })
      callbacks?.onApprovalSuccess?.()
    },
  })

  return {
    createMutation,
    mergeMutation,
    approvalMutation,
  }
}
