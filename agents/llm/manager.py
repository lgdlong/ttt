"""
LLM Manager - Factory cho các LLM providers
"""

import os
from typing import Optional, Dict
from dotenv import load_dotenv
from llm import GeminiProvider, OpenAIProvider
from base import BaseLLMProvider

# Load environment variables
load_dotenv()


class LLMManager:
    """
    Quản lý và tạo các LLM providers
    Factory pattern với support cho nhiều providers
    """
    
    def __init__(self):
        self.providers: Dict[str, BaseLLMProvider] = {}
        self._load_providers()
    
    def _load_providers(self):
        """Load providers từ environment variables"""
        
        # Load Gemini
        gemini_keys = os.getenv("GEMINI_API_KEYS", "")
        if gemini_keys:
            keys = [k.strip() for k in gemini_keys.split(",") if k.strip()]
            if keys:
                model = os.getenv("GEMINI_MODEL", "gemini-2.0-flash-exp")
                self.providers["gemini"] = GeminiProvider(keys, model)
                print(f"✅ Loaded Gemini provider with {len(keys)} key(s), model: {model}")
        
        # Load OpenAI
        openai_keys = os.getenv("OPENAI_API_KEYS", "")
        if openai_keys:
            keys = [k.strip() for k in openai_keys.split(",") if k.strip()]
            if keys:
                model = os.getenv("OPENAI_MODEL", "gpt-4o-mini")
                base_url = os.getenv("OPENAI_BASE_URL")  # Custom base URL for third-party services
                self.providers["openai"] = OpenAIProvider(keys, model, base_url)
                
                if base_url:
                    print(f"✅ Loaded OpenAI provider with {len(keys)} key(s), model: {model}, base_url: {base_url}")
                else:
                    print(f"✅ Loaded OpenAI provider with {len(keys)} key(s), model: {model}")
        
        if not self.providers:
            raise ValueError(
                "Không tìm thấy API key cho bất kỳ provider nào!\n"
                "Hãy set ít nhất một trong các biến môi trường:\n"
                "  - GEMINI_API_KEYS=key1,key2\n"
                "  - OPENAI_API_KEYS=key1,key2"
            )
        
        # Set default provider
        default = os.getenv("DEFAULT_AI_PROVIDER", "gemini").lower()
        if default not in self.providers:
            default = list(self.providers.keys())[0]
        
        self.default_provider = default
        print(f"🎯 Default provider: {self.default_provider}")
    
    def get_provider(self, provider: Optional[str] = None) -> BaseLLMProvider:
        """
        Lấy provider instance
        
        Args:
            provider: Tên provider ("gemini", "openai"). Nếu None, dùng default
            
        Returns:
            Provider instance
        """
        provider_name = provider or self.default_provider
        
        if provider_name not in self.providers:
            available = ", ".join(self.providers.keys())
            raise ValueError(
                f"Provider '{provider_name}' không có sẵn.\n"
                f"Available: {available}"
            )
        
        return self.providers[provider_name]
    
    def get_available_providers(self) -> list[str]:
        """Lấy danh sách providers có sẵn"""
        return list(self.providers.keys())
    
    async def close_all(self):
        """Đóng tất cả providers"""
        for provider in self.providers.values():
            await provider.close()


# Singleton instance
llm_manager = LLMManager()
