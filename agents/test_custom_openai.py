"""
Test OpenAI Third-Party Service
"""

import asyncio
import os
from dotenv import load_dotenv

load_dotenv()

# Test với custom base_url
from llm.openai import OpenAIProvider


async def test_custom_openai():
    """Test OpenAI provider with custom base_url"""
    
    # Get config from .env
    api_keys = os.getenv("OPENAI_API_KEYS", "").split(",")
    model = os.getenv("OPENAI_MODEL", "gpt-4o-mini")
    base_url = os.getenv("OPENAI_BASE_URL")
    
    if not api_keys or not api_keys[0]:
        print("❌ OPENAI_API_KEYS not set in .env")
        return
    
    print("=" * 60)
    print("Test OpenAI Third-Party Service")
    print("=" * 60)
    print(f"API Keys: {len(api_keys)} key(s)")
    print(f"Model: {model}")
    print(f"Base URL: {base_url or 'Default (https://api.openai.com/v1)'}")
    print("=" * 60)
    
    # Create provider
    provider = OpenAIProvider(api_keys, model, base_url)
    
    # Test 1: Simple JSON response
    print("\n--- Test 1: Simple JSON ---")
    try:
        result = await provider.generate(
            prompt='{"test": "Hello from custom OpenAI service"}',
            system_instruction="You are a helpful assistant. Respond with valid JSON only.",
            temperature=0.0,
            max_tokens=100
        )
        print(f"✅ Success!")
        print(f"Response: {result[:200]}...")
    except Exception as e:
        print(f"❌ Error: {e}")
    
    # Test 2: Vietnamese translation
    print("\n--- Test 2: Vietnamese Translation ---")
    try:
        result = await provider.generate(
            prompt='{"term": "Chủ nghĩa thực dụng"}',
            system_instruction="""You are an expert terminologist. 
Translate the Vietnamese input into its most standard, professional English equivalent.
Return ONLY valid JSON with format: {"translation": "..."}""",
            temperature=0.0,
            max_tokens=50
        )
        print(f"✅ Success!")
        print(f"Response: {result}")
    except Exception as e:
        print(f"❌ Error: {e}")
    
    # Cleanup
    await provider.close()
    
    print("\n" + "=" * 60)
    print("Test completed!")
    print("=" * 60)


if __name__ == "__main__":
    asyncio.run(test_custom_openai())
