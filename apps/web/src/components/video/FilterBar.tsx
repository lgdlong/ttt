import React from 'react'
import { Stack, Chip } from '@mui/material'

interface FilterBarProps {
  selectedFilter: string
  onFilterChange: (filter: string) => void
}

const FILTERS = [
  { value: 'all', label: 'Tất cả' },
  { value: 'has_transcript', label: 'Có phụ đề' },
  { value: 'no_transcript', label: 'Chưa có phụ đề' },
  { value: 'is_reviewed', label: 'Đã duyệt' },
]

/**
 * FilterBar Component
 * Displays filter chips for filtering videos by transcript status
 */
const FilterBar: React.FC<FilterBarProps> = ({ selectedFilter, onFilterChange }) => {
  return (
    <Stack
      direction="row"
      spacing={1}
      sx={{
        mb: 4,
        overflowX: 'auto',
        pb: 1,
        // Hide scrollbar but keep functionality
        '&::-webkit-scrollbar': {
          height: 4,
        },
        '&::-webkit-scrollbar-thumb': {
          backgroundColor: '#CBD5E1',
          borderRadius: 4,
        },
      }}
    >
      {FILTERS.map((filter) => (
        <Chip
          key={filter.value}
          label={filter.label}
          clickable
          color={selectedFilter === filter.value ? 'primary' : 'default'}
          onClick={() => onFilterChange(filter.value)}
          sx={{
            fontWeight: 500,
            flexShrink: 0,
          }}
        />
      ))}
    </Stack>
  )
}

export default FilterBar
