import json
import os
import glob

# --- Constants ---
# Thư mục mặc định chứa output của workflow transcript_to_json
DEFAULT_TARGET_DIR = os.path.join("agents", "resources", "transcript_to_json", "output")
ALTERNATE_TARGET_DIR = "resources/transcript_to_json/output"

# Quy tắc tính start_time (ms)
START_TIME_INCREMENT = 1000

# Tên các trường trong JSON
TRANSCRIPT_FIELD = 'transcript'
START_TIME_FIELD = 'start_time'

# Cấu hình định dạng JSON
JSON_INDENT = 2
FILE_ENCODING = 'utf-8'

def update_json_files(directory):
    """
    Duyệt qua tất cả các tệp .json trong thư mục chỉ định và cập nhật trường 'start_time' 
    theo quy tắc: phần tử đầu tiên = 0, các phần tử sau tăng dần 1000ms.
    
    Args:
        directory (str): Đường dẫn đến thư mục chứa các tệp JSON cần xử lý.
    """
    # Tạo pattern để tìm tất cả các file .json
    pattern = os.path.join(directory, "*.json")
    files = glob.glob(pattern)
    
    if not files:
        print(f"⚠️ Không tìm thấy tệp JSON nào trong {directory}")
        return

    for file_path in files:
        print(f"⏳ Đang xử lý {file_path}...")
        try:
            # Đọc nội dung file JSON
            with open(file_path, 'r', encoding=FILE_ENCODING) as f:
                data = json.load(f)
            
            # Kiểm tra xem có trường TRANSCRIPT_FIELD và nó có phải là list không
            if TRANSCRIPT_FIELD in data and isinstance(data[TRANSCRIPT_FIELD], list):
                # Cập nhật start_time theo thứ tự tăng dần
                for index, item in enumerate(data[TRANSCRIPT_FIELD]):
                    item[START_TIME_FIELD] = index * START_TIME_INCREMENT
                
                # Ghi đè lại nội dung đã cập nhật vào file (giữ nguyên encoding và định dạng thụt lề)
                with open(file_path, 'w', encoding=FILE_ENCODING) as f:
                    json.dump(data, f, ensure_ascii=False, indent=JSON_INDENT)
                print(f"✅ Đã cập nhật thành công: {file_path}")
            else:
                print(f"❓ Không tìm thấy mảng '{TRANSCRIPT_FIELD}' trong {file_path}, bỏ qua...")
        except Exception as e:
            print(f"❌ Lỗi khi xử lý {file_path}: {e}")

if __name__ == "__main__":
    # Kiểm tra sự tồn tại của thư mục mục tiêu
    if os.path.exists(DEFAULT_TARGET_DIR):
        update_json_files(DEFAULT_TARGET_DIR)
    elif os.path.exists(ALTERNATE_TARGET_DIR):
        update_json_files(ALTERNATE_TARGET_DIR)
    else:
        print(f"❌ Thư mục không tồn tại: {DEFAULT_TARGET_DIR}")