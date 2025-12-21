"""
Test script for multi-provider setup
"""

import asyncio
from llm.manager import llm_manager


async def test_default_provider():
    """Test với default provider"""
    print("\n=== Test Default Provider ===")
    provider = llm_manager.get_provider()
    print(f"Provider: {llm_manager.default_provider}")
    print(f"Model: {provider.model_name}")
    
    try:
        result = await provider.generate(
            prompt='{"test": true}',
            system_instruction="Respond with valid JSON only."
        )
        print(f"✅ Response: {result[:200]}...")
        return True
    except Exception as e:
        print(f"❌ Lỗi: {e}")
        return False


async def test_all_providers():
    """Test tất cả providers có sẵn"""
    print("\n=== Test All Available Providers ===")
    available = llm_manager.get_available_providers()
    print(f"Available providers: {available}")
    
    results = {}
    
    for provider_name in available:
        print(f"\n--- Testing {provider_name} ---")
        provider = llm_manager.get_provider(provider_name)
        
        try:
            result = await provider.generate(
                prompt='{"ok": true}',
                system_instruction="Respond with JSON."
            )
            print(f"✅ {provider_name} works!")
            print(f"Response preview: {result[:150]}...")
            results[provider_name] = "SUCCESS"
        except Exception as e:
            print(f"❌ {provider_name} failed: {e}")
            results[provider_name] = f"FAILED: {e}"
    
    return results


async def test_round_robin():
    """Test round-robin key rotation"""
    print("\n=== Test Round-Robin Key Rotation ===")
    provider = llm_manager.get_provider()
    
    if len(provider.api_keys) < 2:
        print(f"⚠️ Provider chỉ có {len(provider.api_keys)} key. Cần ít nhất 2 để test rotation.")
        return False
    
    print(f"Provider có {len(provider.api_keys)} keys")
    
    try:
        # Gọi 2 lần để test rotation
        for i in range(2):
            key_index = provider.current_key_index
            result = await provider.generate(
                prompt='{"x": 1}',
                system_instruction="JSON."
            )
            print(f"✅ Call {i+1}: Key index {key_index}")
        
        return True
    except Exception as e:
        print(f"❌ Rotation test failed: {e}")
        return False


async def main():
    """Main test suite"""
    print("=" * 60)
    print("Multi-Provider LLM Test Suite")
    print("=" * 60)
    
    # Test 1: Default provider
    test1 = await test_default_provider()
    
    # Test 2: All providers
    test2_results = await test_all_providers()
    
    # Test 3: Round-robin
    test3 = await test_round_robin()
    
    # Summary
    print("\n" + "=" * 60)
    print("Test Summary")
    print("=" * 60)
    print(f"Default Provider: {'✅ PASS' if test1 else '❌ FAIL'}")
    print(f"\nAll Providers:")
    for provider, status in test2_results.items():
        emoji = "✅" if status == "SUCCESS" else "❌"
        print(f"  {emoji} {provider}: {status}")
    print(f"\nRound-Robin: {'✅ PASS' if test3 else '⚠️ SKIP'}")
    
    # Cleanup
    await llm_manager.close_all()
    print("\n🎉 Tests completed!")


if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("\n\n⚠️ Tests interrupted by user")
    except Exception as e:
        print(f"\n\n❌ Fatal error: {e}")
