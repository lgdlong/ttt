"""
Main Entry Point
Chạy workflow transcript to JSON
"""

import asyncio
import sys
from workflows import run_transcript_to_json


if __name__ == "__main__":
    # Parse command line arguments (optional)
    provider = None
    if len(sys.argv) > 1:
        provider = sys.argv[1]
    
    # Run workflow
    asyncio.run(run_transcript_to_json(provider=provider))