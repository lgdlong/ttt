"""
Logging utilities for LLM providers
"""

import json
import os
from datetime import datetime
from pathlib import Path
from typing import Optional, Any


class APILogger:
    """Logger for API calls and responses"""
    
    def __init__(self, log_file: str = "logs/gemini_logs.txt"):
        self.log_file = log_file
        self.log_dir = Path(log_file).parent
        
        # Create logs directory if not exists
        self.log_dir.mkdir(parents=True, exist_ok=True)
    
    def log_api_call(
        self,
        provider: str,
        model: str,
        status: str,
        input_tokens: Optional[int] = None,
        output_tokens: Optional[int] = None,
        total_tokens: Optional[int] = None,
        duration_ms: Optional[float] = None,
        error: Optional[str] = None,
        api_key_index: Optional[int] = None,
        prompt_length: Optional[int] = None,
        response_preview: Optional[str] = None,
        extra_info: Optional[dict] = None,
    ):
        """
        Log an API call with details
        
        Args:
            provider: Provider name (e.g., 'gemini', 'openai')
            model: Model name
            status: Status ('success', 'error', 'rate_limited', etc.)
            input_tokens: Number of input tokens
            output_tokens: Number of output tokens
            total_tokens: Total tokens used
            duration_ms: Request duration in milliseconds
            error: Error message if failed
            api_key_index: Which API key was used (for rotation tracking)
            prompt_length: Original prompt length in characters
            response_preview: First 100 chars of response
            extra_info: Additional metadata
        """
        
        timestamp = datetime.now()
        
        log_entry = {
            "timestamp": timestamp.isoformat(),
            "timestamp_readable": timestamp.strftime("%Y-%m-%d %H:%M:%S.%f")[:-3],
            "provider": provider,
            "model": model,
            "status": status,
            "duration_ms": duration_ms,
            "tokens": {
                "input": input_tokens,
                "output": output_tokens,
                "total": total_tokens,
            },
            "request": {
                "prompt_length": prompt_length,
            },
            "response": {
                "preview": response_preview[:100] if response_preview else None,
                "length": len(response_preview) if response_preview else None,
            },
            "api_key": {
                "index": api_key_index,
            },
            "error": error,
        }
        
        # Add extra info if provided
        if extra_info:
            log_entry["extra"] = extra_info
        
        # Write to log file
        try:
            with open(self.log_file, "a", encoding="utf-8") as f:
                f.write(json.dumps(log_entry, ensure_ascii=False, indent=2) + "\n")
                f.write("-" * 80 + "\n")
        except Exception as e:
            print(f"Failed to write log: {e}")
    
    def log_summary(self, provider: str):
        """Print a summary of API usage"""
        try:
            if not Path(self.log_file).exists():
                print(f"No logs found at {self.log_file}")
                return
            
            with open(self.log_file, "r", encoding="utf-8") as f:
                lines = f.readlines()
            
            # Parse JSON entries
            entries = []
            current_json = ""
            
            for line in lines:
                if line.startswith("-"):
                    if current_json.strip():
                        try:
                            entry = json.loads(current_json)
                            if entry.get("provider") == provider:
                                entries.append(entry)
                        except:
                            pass
                    current_json = ""
                else:
                    current_json += line
            
            if not entries:
                print(f"No entries found for {provider}")
                return
            
            # Calculate stats
            success_count = len([e for e in entries if e.get("status") == "success"])
            error_count = len([e for e in entries if e.get("status") == "error"])
            total_input_tokens = sum([e.get("tokens", {}).get("input") or 0 for e in entries])
            total_output_tokens = sum([e.get("tokens", {}).get("output") or 0 for e in entries])
            total_duration = sum([e.get("duration_ms") or 0 for e in entries])
            
            print(f"\n{'='*60}")
            print(f"API Usage Summary - {provider.upper()}")
            print(f"{'='*60}")
            print(f"Total Calls:        {len(entries)}")
            print(f"  Success:          {success_count}")
            print(f"  Errors:           {error_count}")
            print(f"Total Input Tokens: {total_input_tokens:,}")
            print(f"Total Output Tokens:{total_output_tokens:,}")
            print(f"Total Duration:     {total_duration:.2f}ms ({total_duration/1000:.2f}s)")
            
            if success_count > 0:
                avg_duration = total_duration / success_count
                print(f"Avg Duration:       {avg_duration:.2f}ms")
            
            print(f"{'='*60}\n")
            
        except Exception as e:
            print(f"Error reading logs: {e}")


# Global logger instance
_logger = APILogger()


def get_logger() -> APILogger:
    """Get the global logger instance"""
    return _logger
