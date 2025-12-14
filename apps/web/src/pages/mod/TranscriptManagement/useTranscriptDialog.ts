import { useState, useCallback } from 'react'
import axiosInstance from '~/lib/axios'
import type { Video } from '~/types/video'
import type { TranscriptResponse } from './TranscriptViewDialog'

export const useTranscriptDialog = () => {
  const [openViewDialog, setOpenViewDialog] = useState(false)
  const [selectedVideo, setSelectedVideo] = useState<Video | null>(null)
  const [transcript, setTranscript] = useState<TranscriptResponse | null>(null)
  const [loadingTranscript, setLoadingTranscript] = useState(false)

  const fetchTranscript = async (videoId: string): Promise<TranscriptResponse | null> => {
    try {
      const response = await axiosInstance.get(`/videos/${videoId}/transcript`)
      return response.data
    } catch {
      return null
    }
  }

  const handleViewTranscript = useCallback(async (video: Video) => {
    setSelectedVideo(video)
    setOpenViewDialog(true)
    setLoadingTranscript(true)
    try {
      const data = await fetchTranscript(video.id)
      setTranscript(data)
    } finally {
      setLoadingTranscript(false)
    }
  }, [])

  const handleCloseViewDialog = useCallback(() => {
    setOpenViewDialog(false)
    setSelectedVideo(null)
    setTranscript(null)
  }, [])

  return {
    openViewDialog,
    selectedVideo,
    transcript,
    loadingTranscript,
    handleViewTranscript,
    handleCloseViewDialog,
  }
}
