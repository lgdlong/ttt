# 📘 TTT Archive - UI/UX Documentation (Final Specs)

Version: 2.0

Tech Stack: React, Material UI (MUI) v6/v7

Style: Clean Utility (Tập trung nội dung, Học thuật)

Font: Inter

----------

## 1. Design System (Hệ thống thiết kế)

### Màu sắc (Color Palette)

-   **Primary (Màu chính):** `#008080` (Teal/Cyan đậm). Dùng cho nút bấm chính, liên kết, icon active.
    
-   **Background (Nền):** `#FFFFFF` (Nền chính), `#F8FAFC` (Nền phụ/Nền trang chủ).
    
-   **Text (Chữ):** `#1E293B` (Đen xám - Tiêu đề), `#475569` (Xám vừa - Nội dung), `#94A3B8` (Xám nhạt - Metadata).
    
-   **Highlight (Khi đọc script):** `#E0F2F1` (Xanh Teal rất nhạt).
    
-   **Badge Uy tín:** `#10B981` (Xanh lá Emerald).
    

### Typography (Font chữ)

-   **Font Family:** `Inter`, sans-serif.
    
-   **Tiêu đề (H1/H2):** Weight 600 hoặc 700.
    
-   **Nội dung (Body):** Weight 400, Line-height 1.6 (để dễ đọc đoạn văn dài).
    

----------

## 2. Layout Chung (App Shell)

Web sử dụng layout tối giản, bỏ qua Sidebar bên trái để tập trung không gian hiển thị Grid video.

### Header (Thanh điều hướng)

-   **Vị trí:** Sticky (Dính chặt trên cùng).
    
-   **Chiều cao:** 64px.
    
-   **Màu nền:** Trắng (có border-bottom mỏng).
    
-   **Phần bên Trái:** Logo "TTT Archive" (Text đậm hoặc Icon).
    
-   **Phần ở Giữa:** Thanh tìm kiếm (Search Bar) bo tròn, rộng, có nút Search icon bên phải.
    
-   **Phần bên Phải:** Nút chuyển đổi ngôn ngữ (nếu có) + Avatar User (Dropdown).
    

----------

## 3. Trang Chủ (Homepage)

Giao diện dạng lưới (Grid) sạch sẽ, hiển thị nhiều video nhất có thể nhưng không rối mắt.

### Khu vực 1: Filter Bar (Bộ lọc nhanh)

Nằm ngay dưới Header, cố định hoặc trôi theo khi cuộn.

-   **Thành phần:** Một hàng ngang các Chips (thẻ từ khóa).
    
-   **Nội dung:** "Tất cả", "Kỷ luật", "Tài chính", "Tâm lý học", "Coding"...
    
-   **Trạng thái:** Chip đang chọn sẽ có màu nền Teal, Chip chưa chọn màu xám nhạt.
    

### Khu vực 2: Video Grid (Danh sách Video)

-   **Layout:** Grid responsive (Grid2).
    
    -   Desktop lớn: 4 cột.
        
    -   Laptop: 3 cột.
        
    -   Tablet: 2 cột.
        
    -   Mobile: 1 cột.
        

### Chi tiết UI của 1 Video Card (Thẻ Video)

-   **Thumbnail:** Tỷ lệ 16:9. Bo góc nhẹ (8px).
    
-   **Badge Uy tín:** Nếu script đã được duyệt, hiển thị icon "Verified" (Tích xanh) nhỏ ở góc thumbnail hoặc ngay cạnh tiêu đề.
    
-   **Tiêu đề:** Tối đa 2 dòng, font Inter semi-bold.
    
-   **Metadata:** Hiển thị "Ngày đăng" • "Lượt xem". Màu chữ xám nhạt.
    
-   **Tags:** Hiển thị tối đa 2 tag quan trọng nhất dưới dạng text nhỏ màu Teal.
    

----------

## 4. Trang Chi tiết Video (Video Detail Page)

Layout chia đôi màn hình (Split View) dành cho Desktop.

### Cấu trúc Grid (Desktop)

-   **Cột Trái (Main Content):** Chiếm 65-70% chiều rộng. Chứa Video Player và thông tin.
    
-   **Cột Phải (Transcript Sidebar):** Chiếm 30-35% chiều rộng. Chứa nội dung bài nói.
    

### Chi tiết Cột Trái (Video & Info)

1.  **Video Player:** Full chiều rộng cột trái. Tỷ lệ 16:9.
    
2.  **Tiêu đề Video:** Font size 24px, Bold. Nằm ngay dưới video.
    
3.  **Hàng Actions:**
    
    -   Nút "Like" (Icon ngón tay cái).
        
    -   Nút "Lưu xem sau" (Icon Bookmark).
        
    -   Nút "Share" (Icon chia sẻ).
        
    -   _Style:_ Button dạng Text hoặc Outlined nhẹ nhàng, màu Teal.
        
4.  **Danh sách Tags:** Các chip nhỏ (pill shape) nằm ngang. Click vào sẽ nhảy sang trang tìm kiếm tag đó.
    
5.  **Thông tin Author:** Avatar tròn + Tên Youtuber + Số sub.
    

### Chi tiết Cột Phải (Transcript - Interactive)

-   **Container:** Chiều cao cố định (bằng chiều cao Video + Info), có thanh cuộn riêng (`overflow-y: auto`).
    
-   **Logic hiển thị:**
    
    -   Chia script thành các **Đoạn văn** (Paragraph).
        
    -   Mỗi đoạn văn gồm **9 câu** script ghép lại.
        
-   **Giao diện từng đoạn:**
    
    -   **Time Point:** Thời gian bắt đầu của đoạn (ví dụ `04:20`) hiển thị nhỏ, màu Teal đậm, có thể click để tua.
        
    -   **Text Body:** Các câu nối tiếp nhau.
        
    -   **Hiệu ứng:** Khi video chạy đến câu nào, câu đó sáng nền màu `#E0F2F1`. Hover chuột vào câu bất kỳ sẽ tô đậm nhẹ để người dùng biết có thể click.
        

----------

## 5. Hướng dẫn Code (MUI v7 Syntax)

Dưới đây là cấu trúc code React sử dụng `Grid2` (Cú pháp mới nhất thay thế cho Grid cũ) để bạn copy.

### A. Code Trang Chủ (Homepage)

JavaScript

```
import React from 'react';
import { 
  Box, Container, Typography, Card, CardMedia, CardContent, 
  Chip, Stack, Avatar, IconButton 
} from '@mui/material';
import Grid from '@mui/material/Grid2'; // MUI v6/v7 sử dụng Grid2
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import SearchIcon from '@mui/icons-material/Search';

const Homepage = () => {
  // Mock data
  const videos = Array.from({ length: 8 }); 

  return (
    <Box sx={{ bgcolor: '#F8FAFC', minHeight: '100vh' }}>
      
      {/* 1. Header Minimal */}
      <Box component="header" sx={{ 
        position: 'sticky', top: 0, zIndex: 10, 
        bgcolor: 'white', borderBottom: '1px solid #e2e8f0', 
        height: 64, display: 'flex', alignItems: 'center', px: 3, gap: 2 
      }}>
        <Typography variant="h6" fontWeight="800" color="primary.main">TTT ARCHIVE</Typography>
        
        {/* Search Bar */}
        <Box sx={{ 
          flex: 1, maxWidth: 600, mx: 'auto', 
          bgcolor: '#f1f5f9', borderRadius: 99, 
          display: 'flex', alignItems: 'center', px: 2, py: 0.5 
        }}>
          <input 
            placeholder="Tìm kiếm bài học..." 
            style={{ border: 'none', background: 'transparent', width: '100%', outline: 'none', padding: '8px' }} 
          />
          <IconButton><SearchIcon /></IconButton>
        </Box>
        
        <Avatar sx={{ width: 32, height: 32 }} />
      </Box>

      <Container maxWidth="xl" sx={{ py: 3 }}>
        
        {/* 2. Filter Bar */}
        <Stack direction="row" spacing={1} sx={{ mb: 4, overflowX: 'auto', pb: 1 }}>
          {['Tất cả', 'Kỷ luật', 'Tài chính', 'Mindset', 'Sức khỏe'].map((tag, i) => (
            <Chip 
              key={tag} label={tag} clickable 
              color={i === 0 ? 'primary' : 'default'} // Active cái đầu tiên
              sx={{ fontWeight: 500 }}
            />
          ))}
        </Stack>

        {/* 3. Video Grid (Grid2 Syntax) */}
        <Grid container spacing={3}>
          {videos.map((_, index) => (
            <Grid size={{ xs: 12, sm: 6, md: 4, lg: 3 }} key={index}> 
              <Card sx={{ 
                height: '100%', borderRadius: 2, boxShadow: 'none', 
                border: '1px solid #e2e8f0', cursor: 'pointer',
                transition: 'transform 0.2s', '&:hover': { transform: 'translateY(-4px)' }
              }}>
                {/* Thumbnail */}
                <Box sx={{ position: 'relative' }}>
                  <CardMedia component="img" height="180" image="https://placehold.co/600x400" />
                  <Box sx={{ 
                    position: 'absolute', bottom: 8, right: 8, 
                    bgcolor: 'rgba(0,0,0,0.8)', color: 'white', 
                    fontSize: 12, px: 0.5, borderRadius: 1 
                  }}>
                    12:05
                  </Box>
                </Box>

                {/* Content */}
                <CardContent sx={{ pb: '16px !important' }}>
                  <Typography variant="subtitle1" fontWeight="600" lineHeight={1.3} mb={1}>
                    Làm sao để giữ kỷ luật bản thân mỗi ngày?
                  </Typography>
                  
                  {/* Badge Uy tín + Metadata */}
                  <Stack direction="row" alignItems="center" spacing={0.5} mb={1}>
                    <CheckCircleIcon sx={{ fontSize: 14, color: '#10B981' }} />
                    <Typography variant="caption" color="#10B981" fontWeight="600">Script Verified</Typography>
                  </Stack>

                  <Typography variant="caption" color="text.secondary">
                    2 năm trước • 1.5M views
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
          ))}
        </Grid>

      </Container>
    </Box>
  );
};

```

### B. Code Trang Chi Tiết (Detail Page)

JavaScript

```
import React from 'react';
import { Box, Container, Typography, Chip, Stack, Button, Avatar } from '@mui/material';
import Grid from '@mui/material/Grid2'; // Import Grid2
import BookmarkBorderIcon from '@mui/icons-material/BookmarkBorder';
import ThumbUpOffAltIcon from '@mui/icons-material/ThumbUpOffAlt';

// Component hiển thị transcript (Đã tối ưu)
import TranscriptParagraphs from './TranscriptParagraphs'; 

const VideoDetail = () => {
  return (
    <Container maxWidth="xl" sx={{ mt: 3, mb: 5 }}>
      {/* Layout Split View: Left (Main) 8 - Right (Side) 4 */}
      <Grid container spacing={4}>
        
        {/* --- CỘT TRÁI: VIDEO PLAYER & INFO --- */}
        <Grid size={{ xs: 12, md: 8 }}>
          
          {/* Video Player */}
          <Box sx={{ 
            width: '100%', aspectRatio: '16/9', bgcolor: 'black', 
            borderRadius: 3, mb: 2, overflow: 'hidden' 
          }}>
            {/* Embed Iframe Youtube Here */}
          </Box>

          {/* Title */}
          <Typography variant="h5" fontWeight="700" fontFamily="Inter" gutterBottom>
            Wicked's costume designer on how to tell stories with clothes
          </Typography>

          {/* Actions & Meta */}
          <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 2 }}>
             <Typography variant="body2" color="text.secondary">
               243,296 lượt xem • 15 thg 11, 2025
             </Typography>
             
             <Stack direction="row" spacing={1}>
               <Button startIcon={<ThumbUpOffAltIcon />} color="inherit">Thích</Button>
               <Button startIcon={<BookmarkBorderIcon />} color="inherit">Lưu</Button>
             </Stack>
          </Stack>

          {/* Tags */}
          <Stack direction="row" spacing={1} mb={3}>
            {['Design', 'Creativity', 'Art', 'Fashion'].map(tag => (
              <Chip key={tag} label={tag} size="small" onClick={() => {}} />
            ))}
          </Stack>

          {/* Author Info (Divider Top) */}
          <Box sx={{ display: 'flex', gap: 2, pt: 3, borderTop: '1px solid #eee' }}>
            <Avatar sx={{ width: 48, height: 48 }} />
            <Box>
              <Typography variant="subtitle1" fontWeight="600">TED Talks</Typography>
              <Typography variant="caption" color="text.secondary">20M subscribers</Typography>
            </Box>
          </Box>
        </Grid>

        {/* --- CỘT PHẢI: TRANSCRIPT (SCROLLABLE) --- */}
        <Grid size={{ xs: 12, md: 4 }}>
          <Box sx={{ 
            height: 'calc(100vh - 100px)', // Full chiều cao trừ header
            position: 'sticky', top: 80,
            display: 'flex', flexDirection: 'column'
          }}>
            <Typography variant="h6" fontWeight="600" mb={2}>Transcript</Typography>
            
            {/* Vùng cuộn nội dung */}
            <Box sx={{ 
              flex: 1, 
              overflowY: 'auto', 
              pr: 1,
              // Custom Scrollbar cho đẹp
              '&::-webkit-scrollbar': { width: '6px' },
              '&::-webkit-scrollbar-thumb': { backgroundColor: '#cbd5e1', borderRadius: '4px' }
            }}>
              {/* Component Paragraph Logic 9 câu/đoạn */}
              <TranscriptParagraphs /> 
            </Box>
          </Box>
        </Grid>

      </Grid>
    </Container>
  );
};

```