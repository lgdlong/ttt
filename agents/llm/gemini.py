"""
Google Gemini Provider
Sử dụng official google-genai SDK
"""

# ========== PROVIDER CONFIGURATION ==========
import os
DEFAULT_TEMPERATURE = 0.2
MAX_OUTPUT_TOKENS = int(os.getenv("MAX_OUTPUT_TOKENS", "30000"))
MAX_RETRIES = int(os.getenv("GEMINI_MAX_RETRIES", "5"))
RETRY_BASE_DELAY = float(os.getenv("GEMINI_RETRY_DELAY", "1.0"))  # seconds
RETRY_JITTER_MAX = float(os.getenv("GEMINI_RETRY_JITTER", "0.5"))  # seconds
# ============================================

import asyncio
import random
import time
from typing import Optional
from google import genai
from base import BaseLLMProvider
from utils.logger import get_logger


class GeminiProvider(BaseLLMProvider):
    """Google Gemini API Provider using official SDK"""
    
    def __init__(self, api_keys: list[str], model_name: str = "gemini-2.5-flash"):
        super().__init__(api_keys, model_name)
        self.max_retries = MAX_RETRIES
        self.retry_base_delay = RETRY_BASE_DELAY
        self.jitter_max = RETRY_JITTER_MAX
        self.logger = get_logger()
    
    def _get_client(self, api_key: str):
        """Tạo Gemini client với API key"""
        return genai.Client(api_key=api_key)
    
    async def generate(
        self,
        prompt: str,
        system_instruction: Optional[str] = None,
        temperature: float = DEFAULT_TEMPERATURE,
        max_tokens: int = MAX_OUTPUT_TOKENS
    ) -> str:
        """
        Generate text using Gemini API
        
        Args:
            prompt: User prompt
            system_instruction: System instruction
            temperature: Temperature (0.0-1.0)
            max_tokens: Max output tokens
            
        Returns:
            Generated JSON string
        """
        
        start_time = time.time()
        
        for attempt in range(1, self.max_retries + 1):
            api_key = self.get_next_key()
            client = self._get_client(api_key)
            
            try:
                # Configure generation config cục bộ
                generation_config = genai.types.GenerateContentConfig(
                    temperature=temperature,
                    max_output_tokens=max_tokens,
                    response_mime_type="application/json"
                )
                
                # Add system instruction nếu có
                if system_instruction:
                    full_prompt = f"{system_instruction}\n\n{prompt}"
                else:
                    full_prompt = prompt
                
                # Generate using new API
                response = await asyncio.to_thread(
                    client.models.generate_content,
                    model=self.model_name,
                    contents=full_prompt,
                    config=generation_config
                )
                
                # Extract text from response
                if response.text:
                    # Calculate metrics
                    duration_ms = (time.time() - start_time) * 1000
                    input_tokens = response.usage_metadata.prompt_token_count if hasattr(response, 'usage_metadata') else None
                    output_tokens = response.usage_metadata.candidates_token_count if hasattr(response, 'usage_metadata') else None
                    total_tokens = response.usage_metadata.total_token_count if hasattr(response, 'usage_metadata') else None
                    
                    # Log successful API call
                    self.logger.log_api_call(
                        provider="gemini",
                        model=self.model_name,
                        status="success",
                        input_tokens=input_tokens,
                        output_tokens=output_tokens,
                        total_tokens=total_tokens,
                        duration_ms=duration_ms,
                        api_key_index=self.current_key_index,
                        prompt_length=len(full_prompt),
                        response_preview=response.text,
                    )
                    
                    return response.text
                else:
                    raise ValueError("Gemini trả về response trống hoặc bị block (Safety)")
            
            except Exception as e:
                error_msg = str(e).lower()
                duration_ms = (time.time() - start_time) * 1000
                
                # Retry nếu là rate limit hoặc server error
                if any(keyword in error_msg for keyword in ['rate', 'quota', '429', '500', '503']):
                    print(f"Lỗi (Attempt {attempt}/{self.max_retries}): {e}")
                    print(f"Đổi key và retry...")
                    
                    # Log rate limit error
                    self.logger.log_api_call(
                        provider="gemini",
                        model=self.model_name,
                        status="rate_limited",
                        error=str(e),
                        duration_ms=duration_ms,
                        api_key_index=self.current_key_index,
                        prompt_length=len(full_prompt) if 'full_prompt' in locals() else None,
                    )
                    
                    # Exponential backoff + jitter
                    wait_time = (self.retry_base_delay * (2 ** (attempt - 1))) + random.uniform(0, self.jitter_max)
                    await asyncio.sleep(wait_time)
                    continue
                else:
                    # Lỗi khác không retry
                    duration_ms = (time.time() - start_time) * 1000
                    self.logger.log_api_call(
                        provider="gemini",
                        model=self.model_name,
                        status="error",
                        error=str(e),
                        duration_ms=duration_ms,
                        api_key_index=self.current_key_index,
                        prompt_length=len(full_prompt) if 'full_prompt' in locals() else None,
                    )
                    raise Exception(f"Gemini API Error: {e}")
        
        # Log final failure after all retries
        duration_ms = (time.time() - start_time) * 1000
        self.logger.log_api_call(
            provider="gemini",
            model=self.model_name,
            status="failed_after_retries",
            error=f"Failed after {self.max_retries} retries",
            duration_ms=duration_ms,
            api_key_index=self.current_key_index,
            prompt_length=len(full_prompt) if 'full_prompt' in locals() else None,
        )
        raise Exception(f"Gemini: Đã thử {self.max_retries} lần nhưng thất bại")
    
    async def close(self):
        """Cleanup"""
        pass
