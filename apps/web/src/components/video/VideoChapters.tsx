import React from 'react'
import { Box, Typography, Accordion, AccordionSummary, AccordionDetails } from '@mui/material'
import ExpandMoreIcon from '@mui/icons-material/ExpandMore'
import PlayArrowIcon from '@mui/icons-material/PlayArrow'
import type { ChapterResponse } from '~/types/video'

interface VideoChaptersProps {
  chapters?: ChapterResponse[]
  onSeek?: (time: number) => void
}

export const VideoChapters: React.FC<VideoChaptersProps> = ({ chapters, onSeek }) => {
  if (!chapters || chapters.length === 0) return null

  /**
   * Format milliseconds to MM:SS or HH:MM:SS
   */
  const formatTime = (milliseconds: number): string => {
    const totalSeconds = Math.floor(milliseconds / 1000)
    const hours = Math.floor(totalSeconds / 3600)
    const minutes = Math.floor((totalSeconds % 3600) / 60)
    const secs = totalSeconds % 60

    if (hours > 0) {
      return `${hours}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
    }
    return `${minutes}:${secs.toString().padStart(2, '0')}`
  }

  return (
    <Box sx={{ mb: 4 }}>
      <Typography variant="h6" gutterBottom fontWeight={700} fontFamily="Inter" mb={2}>
        Mục lục chi tiết
      </Typography>
      <Box>
        {chapters.map((chapter, index) => (
          <Accordion 
            key={chapter.id} 
            disableGutters 
            elevation={0} 
            sx={{ 
              '&:before': { display: 'none' }, 
              borderBottom: '1px solid',
              borderColor: 'divider',
              bgcolor: 'transparent'
            }}
          >
            <AccordionSummary expandIcon={<ExpandMoreIcon />} sx={{ px: 0 }}>
              <Box 
                onClick={(e) => {
                  if (onSeek && chapter.start_time > 0) {
                    e.stopPropagation();
                    onSeek(chapter.start_time);
                  }
                }}
                sx={{ 
                  display: 'flex', 
                  alignItems: 'center', 
                  gap: 1,
                  cursor: chapter.start_time > 0 ? 'pointer' : 'default',
                  '&:hover': chapter.start_time > 0 ? { color: 'primary.main' } : {}
                }}
              >
                {chapter.start_time > 0 && <PlayArrowIcon sx={{ fontSize: 18 }} />}
                <Typography fontWeight={600} fontFamily="Inter">
                  {index + 1}. {chapter.title}
                </Typography>
                {chapter.start_time > 0 && (
                  <Typography variant="caption" color="text.secondary">
                    ({formatTime(chapter.start_time)})
                  </Typography>
                )}
              </Box>
            </AccordionSummary>
            <AccordionDetails sx={{ px: 0, pt: 0, pb: 2 }}>
               <Typography variant="body2" color="text.secondary" sx={{ whiteSpace: 'pre-line', lineHeight: 1.8 }}>
                 {chapter.content}
               </Typography>
            </AccordionDetails>
          </Accordion>
        ))}
      </Box>
    </Box>
  )
}
