make run-workflow WORKFLOW=transcript_to_json PROVIDER=openai
make run-workflow WORKFLOW=transcript_to_json PROVIDER=gemini
make run-workflow WORKFLOW=other_workflow PROVIDER=openai

make run-workflow  # Chạy transcript_to_json với gemini (mặc định)

make test PROVIDER=openai
make test PROVIDER=gemini
make test-all