"""
Base Provider Interface
Định nghĩa interface chung cho tất cả LLM providers
"""

from abc import ABC, abstractmethod
from typing import Optional


class BaseLLMProvider(ABC):
    """Abstract base class cho tất cả LLM providers"""
    
    def __init__(self, api_keys: list[str], model_name: str):
        """
        Khởi tạo provider
        
        Args:
            api_keys: Danh sách API keys (support round-robin)
            model_name: Tên model
        """
        self.api_keys = api_keys
        self.model_name = model_name
        self.current_key_index = 0
    
    def get_next_key(self) -> str:
        """Lấy API key tiếp theo (round-robin)"""
        key = self.api_keys[self.current_key_index]
        self.current_key_index = (self.current_key_index + 1) % len(self.api_keys)
        return key
    
    @abstractmethod
    async def generate(
        self, 
        prompt: str, 
        system_instruction: Optional[str] = None,
        temperature: float = 0.2,
        max_tokens: int = 8192
    ) -> str:
        """
        Generate text từ prompt
        
        Args:
            prompt: User prompt
            system_instruction: System instruction (optional)
            temperature: Temperature (0.0-1.0)
            max_tokens: Max output tokens
            
        Returns:
            Generated text
        """
        pass
    
    @abstractmethod
    async def close(self):
        """Cleanup resources"""
        pass
