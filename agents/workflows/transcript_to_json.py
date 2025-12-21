"""
Transcript to JSON Workflow
Chuyển đổi transcript thành JSON structure
"""

import asyncio
import os
import json
import time
from pathlib import Path
from typing import Optional
from llm.manager import llm_manager
from utils import ContextOptimizer, get_logger


# ========== WORKFLOW CONFIGURATION ==========
# Base path là workflow directory
WORKFLOW_DIR = os.path.dirname(os.path.abspath(__file__))
AGENTS_DIR = os.path.dirname(WORKFLOW_DIR)

# Input/Output directories
INPUT_DIR = os.path.join(AGENTS_DIR, "resources", "transcript_to_json", "input")
OUTPUT_DIR = os.path.join(AGENTS_DIR, "resources", "transcript_to_json", "output")

INPUT_FILE_EXT = ".f.txt"
OUTPUT_FILE_EXT = ".json"
ERROR_FILE_SUFFIX = ".error.txt"
MAX_CONCURRENT_WORKERS = 3
ENCODING = "utf-8"
JSON_INDENT = 2

# LLM Configuration
DEFAULT_TEMPERATURE = 0.2
MAX_OUTPUT_TOKENS = int(os.getenv("MAX_OUTPUT_TOKENS", "16384"))  # Load from env or default
# ============================================


# Create output directory if not exists
Path(OUTPUT_DIR).mkdir(parents=True, exist_ok=True)

# Initialize logger
logger = get_logger()


# Load system prompt từ file
def load_system_prompt() -> str:
    """Load system prompt từ file instruction"""
    prompt_file = os.path.join(
        AGENTS_DIR,
        "prompts",
        "transcript_to_json_instruction.md"
    )
    
    if os.path.exists(prompt_file):
        with open(prompt_file, 'r', encoding='utf-8') as f:
            return f.read()
    
    # Fallback prompt
    return """
You are a Transcript Analysis Expert.
Task: Convert transcript content into structured JSON.
Mandatory requirements:
1. Keep the 'content' section verbatim (100% exact from input).
2. Segment content logically (5-15 sentences per segment).
3. Provide concise titles for each segment.
4. Create a detailed summary (300+ words).
Output: Valid JSON object only.
"""


SYSTEM_PROMPT = load_system_prompt()


async def process_file(
    file_path: str,
    provider: Optional[str] = None
) -> dict:
    """
    Xử lý một file transcript
    
    Args:
        file_path: Đường dẫn file input
        provider: Provider name (optional, dùng default nếu None)
        
    Returns:
        Dict với status và message
    """
    file_name = os.path.basename(file_path)
    print(f"🚀 Processing: {file_name}")
    
    start_time = time.time()
    
    try:
        # 1. Đọc file
        with open(file_path, 'r', encoding=ENCODING) as f:
            raw_content = f.read()
        
        print(f"  📄 Input size: {len(raw_content)} chars")
        
        # 2. Optimize content
        optimized_content = ContextOptimizer.compact_history(raw_content)
        print(f"  ✂️ Optimized size: {len(optimized_content)} chars")
        
        # 3. Get LLM provider
        llm = llm_manager.get_provider(provider)
        
        # 4. Generate JSON
        print(f"  🤖 Calling {llm.model_name}...")
        json_str = await llm.generate(
            prompt=optimized_content,
            system_instruction=SYSTEM_PROMPT,
            temperature=DEFAULT_TEMPERATURE,
            max_tokens=MAX_OUTPUT_TOKENS
        )
        
        # 5. Validate JSON
        data = ContextOptimizer.parse_json_safely(json_str)
        if not data:
            print(f"  ⚠️ Failed to parse JSON, trying direct parse...")
            try:
                data = json.loads(json_str)
                print(f"  ✅ Direct parse successful")
            except json.JSONDecodeError as e:
                print(f"  ❌ JSON decode error: {e}")
                
                # Save error for debugging
                error_path = os.path.join(OUTPUT_DIR, file_name + ERROR_FILE_SUFFIX)
                with open(error_path, 'w', encoding=ENCODING) as f:
                    f.write(json_str)
                
                return {
                    "status": "error",
                    "file": file_name,
                    "message": f"Invalid JSON output: {e}",
                    "duration": time.time() - start_time
                }
        
        # 6. Save result
        output_path = os.path.join(
            OUTPUT_DIR,
            file_name.replace(INPUT_FILE_EXT, OUTPUT_FILE_EXT)
        )
        with open(output_path, 'w', encoding=ENCODING) as f:
            json.dump(data, f, ensure_ascii=False, indent=JSON_INDENT)
        
        duration = time.time() - start_time
        print(f"✅ Completed: {file_name} ({duration:.2f}s)")
        
        return {
            "status": "success",
            "file": file_name,
            "output": output_path,
            "duration": duration
        }
    
    except Exception as e:
        duration = time.time() - start_time
        print(f"❌ Error processing {file_name}: {e}")
        
        return {
            "status": "error",
            "file": file_name,
            "message": str(e),
            "duration": duration
        }


async def worker(
    semaphore: asyncio.Semaphore,
    file_path: str,
    provider: Optional[str] = None
) -> dict:
    """Worker với semaphore để giới hạn concurrency"""
    async with semaphore:
        return await process_file(file_path, provider)


async def run_workflow(
    input_dir: Optional[str] = None,
    provider: Optional[str] = None,
    max_workers: int = MAX_CONCURRENT_WORKERS
):
    """
    Chạy workflow transcript to JSON
    
    Args:
        input_dir: Thư mục chứa file input (optional, dùng default nếu None)
        provider: Provider name (optional)
        max_workers: Số worker chạy đồng thời
    """
    if input_dir is None:
        input_dir = INPUT_DIR
    
    print("\n" + "=" * 70)
    print("🎬 TRANSCRIPT TO JSON WORKFLOW")
    print("=" * 70)
    
    # Check if input directory exists
    if not os.path.exists(input_dir):
        print(f"❌ Input directory not found: {input_dir}")
        return
    
    # Scan files and filter out already processed ones
    files = []
    skipped_count = 0
    for f in os.listdir(input_dir):
        if f.endswith(INPUT_FILE_EXT):
            input_path = os.path.join(input_dir, f)
            output_file_name = f.replace(INPUT_FILE_EXT, OUTPUT_FILE_EXT)
            output_path = os.path.join(OUTPUT_DIR, output_file_name)
            
            if os.path.exists(output_path):
                skipped_count += 1
                continue
            
            files.append(input_path)
    
    if not files and skipped_count == 0:
        print(f"⚠️ Không tìm thấy file {INPUT_FILE_EXT} nào trong {input_dir}")
        return
    
    if not files and skipped_count > 0:
        print(f"✅ Tất cả {skipped_count} file đã được xử lý xong. Không có file mới.")
        return

    print(f"📁 Input directory:  {input_dir}")
    print(f"📤 Output directory: {OUTPUT_DIR}")
    print(f"📊 Found {len(files) + skipped_count} file(s), processing {len(files)} new file(s) (skipped {skipped_count})")
    print(f"🤖 Provider: {provider or llm_manager.default_provider}")
    print(f"⚙️ Max workers: {max_workers}")
    print("=" * 70 + "\n")
    
    # Create semaphore
    semaphore = asyncio.Semaphore(max_workers)
    
    # Create tasks
    tasks = [
        worker(semaphore, file_path, provider)
        for file_path in files
    ]
    
    # Run pipeline
    start_time = time.time()
    results = await asyncio.gather(*tasks)
    
    # Cleanup
    await llm_manager.close_all()
    
    # Summary
    duration = time.time() - start_time
    success_count = sum(1 for r in results if r["status"] == "success")
    error_count = len(results) - success_count
    total_duration = sum(r.get("duration", 0) for r in results)
    
    print("\n" + "=" * 70)
    print("📊 WORKFLOW SUMMARY")
    print("=" * 70)
    print(f"✅ Success: {success_count}/{len(results)}")
    print(f"❌ Error: {error_count}/{len(results)}")
    print(f"⏱️ Total duration: {duration:.2f}s")
    print(f"⏱️ LLM API time: {total_duration:.2f}s")
    
    if success_count > 0:
        print(f"\n✨ Output files saved to: {OUTPUT_DIR}")
    
    print("=" * 70 + "\n")
    
    return results


if __name__ == "__main__":
    # Thiết lập event loop policy cho Windows
    if os.name == 'nt':
        asyncio.set_event_loop_policy(asyncio.WindowsSelectorEventLoopPolicy())
    
    # Run workflow
    asyncio.run(run_workflow())
