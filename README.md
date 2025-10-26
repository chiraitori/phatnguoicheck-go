# License Plate Checker - API Tra Cứu Phương Tiện Vi Phạm

API server để tra cứu thông tin vi phạm giao thông từ website CSGT Việt Nam với tính năng tự động giải captcha.

## Tính Năng

- ✅ Tự động giải captcha bằng Tesseract OCR (primary) + OCR.space API (fallback)
- ✅ Tra cứu thông tin vi phạm giao thông
- ✅ Parse chi tiết các vi phạm (biển số, loại xe, thời gian, địa điểm, hành vi, trạng thái, đơn vị phát hiện, nơi giải quyết)
- ✅ Tự động retry khi captcha sai (tối đa 9 lần)
- ✅ Đếm số lượng vi phạm
- ✅ Config qua file .env

## Yêu Cầu

- Go 1.18 trở lên
- Tesseract OCR (optional, cho tính năng OCR local)
  - Windows: `choco install tesseract` hoặc download từ [UB Mannheim](https://github.com/UB-Mannheim/tesseract/wiki)
  - Linux: `sudo apt install tesseract-ocr`
  - Mac: `brew install tesseract`

## Cài Đặt

1. Clone repository:
```bash
git clone https://github.com/chiraitori/phatnguoicheck-go.git
cd phatnguoicheck-go
```

2. Cài đặt dependencies:
```bash
go mod download
```

3. Tạo file `.env` từ template:
```bash
cp .env.example .env
```

4. Cập nhật API key trong file `.env`:
```env
OCR_API_KEY=your_api_key_here
PORT=8080
```

Lấy API key miễn phí tại: https://ocr.space/ocrapi

## Chạy Server

```bash
go run .
```

Hoặc build và chạy:
```bash
go build -o LicensePlatecheck
./LicensePlatecheck
```

Server sẽ khởi động trên port được config trong `.env` (mặc định: 8080)

## Sử Dụng API

### Endpoint: POST `/check-license-plate`

**Request:**
```json
{
  "license_plate": "98B378578",
  "vehicle_type": "2"
}
```

**Vehicle Types:**
- `1`: Ô tô
- `2`: Xe máy

**Response (Thành công):**
```json
{
  "success": true,
  "href": "https://www.csgt.vn/tra-cuu-phuong-tien-vi-pham.html?...",
  "error": "",
  "attempts": 2,
  "violation_count": 2,
  "details": {
    "violations": [
      {
        "license_plate": "98B3-785.78",
        "plate_color": "Nền mầu trắng, chữ và số màu đen",
        "vehicle_type": "Xe máy",
        "violation_time": "08:44, 16/10/2025",
        "location": "Km 95+900m, QL1A, Xã Kép, Bắc Ninh",
        "behavior": "16824.7.2.b.01.Điều khiển xe chạy quá tốc độ quy định từ 05 km/h đến dưới 10 km/h",
        "status": "Chưa xử phạt",
        "detecting_unit": "Đội Cảnh sát giao thông đường bộ số 4 - Phòng Cảnh sát giao thông - Công an Tỉnh Bắc Ninh",
        "resolution_point": "Đường Xương Giang, phường Bắc Giang, tỉnh Bắc Ninh"
      }
    ]
  }
}
```

**Response (Không có vi phạm):**
```json
{
  "success": true,
  "href": "...",
  "error": "",
  "attempts": 1,
  "violation_count": 0,
  "details": {
    "message": "Không tìm thấy thông tin vi phạm"
  }
}
```

### Ví Dụ Với cURL

**Windows (PowerShell):**
```powershell
curl.exe -X POST "http://localhost:8080/check-license-plate" `
  -H "Content-Type: application/json" `
  -d '{"license_plate":"98B378578","vehicle_type":"2"}'
```

**Linux/Mac:**
```bash
curl -X POST "http://localhost:8080/check-license-plate" \
  -H "Content-Type: application/json" \
  -d '{"license_plate":"98B378578","vehicle_type":"2"}'
```

## Cách Hoạt Động

1. **Tải captcha** từ website CSGT
2. **Xử lý ảnh**: Chuyển sang grayscale, tăng contrast
3. **Giải captcha**:
   - Thử Tesseract OCR (local) trước
   - Nếu fail, dùng OCR.space API
4. **Gửi request** tra cứu với captcha đã giải
5. **Parse kết quả** từ HTML response
6. **Retry** nếu captcha sai (tối đa 9 lần)

## Cấu Trúc Project

```
.
├── main.go           # Source code chính
├── go.mod            # Go modules
├── go.sum            # Dependencies checksums
├── .env              # Config (không commit)
├── .env.example      # Template config
├── .gitignore        # Git ignore rules
└── README.md         # Documentation này
```

## Cấu Hình

File `.env` hỗ trợ các biến sau:

```env
# API key cho OCR.space (fallback khi Tesseract fail)
OCR_API_KEY=your_api_key_here

# Port server (mặc định: 8080)
PORT=8080
```

## Lưu Ý

- **Rate limiting**: Website CSGT có thể giới hạn số request
- **Tesseract**: Không bắt buộc, nếu không có sẽ dùng API
- **API key**: Miễn phí nhưng có giới hạn calls/tháng
- **Retry logic**: Tự động retry khi captcha sai, tối đa 9 lần

## Xử Lý Lỗi

- **404**: Captcha sai → Tự động retry
- **Timeout**: Request quá lâu → Tăng timeout trong code
- **No API key**: Server vẫn chạy nhưng chỉ dùng Tesseract

## License

MIT

## Đóng Góp

Pull requests are welcome! 

1. Fork repo
2. Tạo feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to branch (`git push origin feature/AmazingFeature`)
5. Mở Pull Request

## Liên Hệ

Nếu có vấn đề hoặc câu hỏi, vui lòng mở issue trên GitHub.
