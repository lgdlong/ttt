"""
OpenAI Provider
Sử dụng official openai SDK
"""

# ========== PROVIDER CONFIGURATION ==========
import os
DEFAULT_TEMPERATURE = 0.2
MAX_OUTPUT_TOKENS = int(os.getenv("MAX_OUTPUT_TOKENS", "30000"))
MAX_RETRIES = int(os.getenv("OPENAI_MAX_RETRIES", "5"))
RETRY_BASE_DELAY = float(os.getenv("OPENAI_RETRY_DELAY", "1.0"))  # seconds
RETRY_JITTER_MAX = float(os.getenv("OPENAI_RETRY_JITTER", "0.5"))  # seconds
# ============================================

import asyncio
import random
import time
from typing import Optional
from openai import AsyncOpenAI
from base import BaseLLMProvider
from utils.logger import APILogger


class OpenAIProvider(BaseLLMProvider):
    """OpenAI API Provider using official SDK"""
    
    def __init__(self, api_keys: list[str], model_name: str = "gpt-4o-mini", base_url: Optional[str] = None):
        super().__init__(api_keys, model_name)
        self.base_url = base_url  # Custom base URL for third-party services
        self.max_retries = MAX_RETRIES
        self.retry_base_delay = RETRY_BASE_DELAY
        self.jitter_max = RETRY_JITTER_MAX
        self.logger = APILogger(log_file="logs/openai_logs.txt")  # Logger for OpenAI
    
    def _get_client(self, api_key: str) -> AsyncOpenAI:
        """Tạo OpenAI client với API key và custom base_url"""
        if self.base_url:
            return AsyncOpenAI(api_key=api_key, base_url=self.base_url)
        return AsyncOpenAI(api_key=api_key)
    
    async def generate(
        self,
        prompt: str,
        system_instruction: Optional[str] = None,
        temperature: float = DEFAULT_TEMPERATURE,
        max_tokens: int = MAX_OUTPUT_TOKENS
    ) -> str:
        """
        Generate text using OpenAI API
        
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
                # Build messages
                messages = []
                full_prompt = ""
                if system_instruction:
                    messages.append({"role": "system", "content": system_instruction})
                    full_prompt = f"{system_instruction}\n\n{prompt}"
                else:
                    full_prompt = prompt
                messages.append({"role": "user", "content": prompt})
                
                # Call OpenAI API
                response = await client.chat.completions.create(
                    model=self.model_name,
                    messages=messages,
                    temperature=temperature,
                    max_tokens=max_tokens,
                    response_format={"type": "json_object"}  # JSON mode
                )
                
                # Extract content
                if response.choices and response.choices[0].message.content:
                    content = response.choices[0].message.content
                    
                    # Calculate metrics
                    duration_ms = (time.time() - start_time) * 1000
                    input_tokens = response.usage.prompt_tokens
                    output_tokens = response.usage.completion_tokens
                    total_tokens = response.usage.total_tokens
                    
                    # Check if response was truncated
                    if response.choices[0].finish_reason == 'length':
                        print(f"  ⚠️ WARNING: Response truncated at {output_tokens} tokens (max_tokens={max_tokens})")
                        print(f"  💡 Consider increasing max_tokens or splitting the input")
                    
                    # Log successful API call
                    self.logger.log_api_call(
                        provider="openai",
                        model=self.model_name,
                        status="success",
                        input_tokens=input_tokens,
                        output_tokens=output_tokens,
                        total_tokens=total_tokens,
                        duration_ms=duration_ms,
                        api_key_index=self.current_key_index,
                        prompt_length=len(full_prompt),
                        response_preview=content,
                    )
                    
                    return content
                else:
                    raise ValueError("OpenAI trả về response trống")
            
            except Exception as e:
                error_msg = str(e).lower()
                duration_ms = (time.time() - start_time) * 1000
                
                # Retry nếu là rate limit hoặc server error
                if any(keyword in error_msg for keyword in ['rate', 'limit', '429', '500', '503']):
                    print(f"Lỗi (Attempt {attempt}/{self.max_retries}): {e}")
                    print(f"Đổi key và retry...")
                    
                    # Log rate limit error
                    self.logger.log_api_call(
                        provider="openai",
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
                    self.logger.log_api_call(
                        provider="openai",
                        model=self.model_name,
                        status="error",
                        error=str(e),
                        duration_ms=duration_ms,
                        api_key_index=self.current_key_index,
                        prompt_length=len(full_prompt) if 'full_prompt' in locals() else None,
                    )
                    raise Exception(f"OpenAI API Error: {e}")
            finally:
                await client.close()
        
        # Log final failure after all retries
        duration_ms = (time.time() - start_time) * 1000
        self.logger.log_api_call(
            provider="openai",
            model=self.model_name,
            status="failed_after_retries",
            error=f"Failed after {self.max_retries} attempts",
            duration_ms=duration_ms,
            api_key_index=self.current_key_index,
            prompt_length=len(full_prompt) if 'full_prompt' in locals() else None,
        )
        
        raise Exception(f"OpenAI: Đã thử {self.max_retries} lần nhưng thất bại")
    
    async def close(self):
        """Cleanup"""
        pass
