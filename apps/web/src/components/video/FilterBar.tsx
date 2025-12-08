import React from 'react'
import { Stack, Chip } from '@mui/material'

interface FilterBarProps {
  categories: string[]
  selectedCategory: string
  onCategoryChange: (category: string) => void
}

/**
 * FilterBar Component
 * Displays category chips for filtering videos
 * Following UI_Spec.md specifications
 */
const FilterBar: React.FC<FilterBarProps> = ({
  categories,
  selectedCategory,
  onCategoryChange,
}) => {
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
      {categories.map((category) => (
        <Chip
          key={category}
          label={category}
          clickable
          color={selectedCategory === category ? 'primary' : 'default'}
          onClick={() => onCategoryChange(category)}
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
