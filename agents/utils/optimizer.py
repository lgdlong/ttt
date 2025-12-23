"""
Context Optimizer Utilities
"""

import json
import re


class ContextOptimizer:
    """Tối ưu context và parse JSON"""
    
    @staticmethod
    def estimate_tokens(text: str) -> int:
        """
        Ước lượng số token
        Quy tắc: 1 token ~= 4 ký tự tiếng Anh, 2-3 ký tự tiếng Việt
        """
        if not text:
            return 0
        return len(text) // 3
    
    @staticmethod
    def compact_history(content: str, max_chars: int = 100000) -> str:
        """
        Cắt bớt content nếu quá dài
        
        Args:
            content: Nội dung cần cắt
            max_chars: Độ dài tối đa
            
        Returns:
            Content đã được compact
        """
        if len(content) <= max_chars:
            return content
        
        print(f"⚠️ Content quá dài ({len(content)} chars). Cắt xuống {max_chars}...")
        return content[:max_chars] + "\n...[TRUNCATED]..."
    
    @staticmethod
    def parse_json_safely(text: str) -> dict | None:
        """
        Parse JSON an toàn từ text có thể chứa markdown hoặc text thừa
        
        Args:
            text: Text chứa JSON
            
        Returns:
            Dict nếu parse thành công, None nếu thất bại
        """
        try:
            # Cách 1: Thử parse trực tiếp
            return json.loads(text)
        except json.JSONDecodeError:
            pass
        
        try:
            # Cách 2: Remove markdown code blocks
            clean_text = text.replace("```json", "").replace("```", "").strip()
            return json.loads(clean_text)
        except json.JSONDecodeError:
            pass
        
        try:
            # Cách 3: Tìm JSON object bằng regex
            match = re.search(r'\{.*\}', text, re.DOTALL)
            if match:
                return json.loads(match.group())
        except json.JSONDecodeError:
            pass
        
        return None
