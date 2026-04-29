# Lightweight Data Catalog (Go-based)

Dự án xây dựng hệ thống quản lý Metadata và Data Lineage tinh gọn, hiệu suất cao bằng ngôn ngữ Go. Hệ thống được thiết kế để thay thế các tính năng cốt lõi của DataHub cho mục đích cá nhân và tối ưu hóa tài nguyên.

## 🚀 Mục tiêu hệ thống
- **Traceability:** Theo dõi dòng chảy dữ liệu từ nguồn đến nơi tiêu thụ (Source to Consumption).
- **Column-level Lineage:** Chi tiết thay đổi đến từng cấp độ cột.
- **Performance:** Tận dụng concurrency của Go để xử lý hàng ngàn event từ Airflow mà không gây trễ.
- **Simplicity:** Loại bỏ các tính năng thừa, tập trung vào Metadata, Schema, và Ownership.

## 🏗 Kiến trúc hệ thống (Architecture)
Hệ thống bao gồm 3 lớp chính:
1. **Ingestion Layer (Go):** Nhận Metadata qua REST API (Echo framework). Sử dụng Go Channels để xử lý bất đồng bộ (Async).
2. **Processing Layer:** Logic so sánh schema (diff), ánh xạ quan hệ các cột và quản lý phiên bản.
3. **Storage Layer:**
   - **PostgreSQL:** Lưu trữ Metadata thực thể, Tags, Policies và History.
   - **Memgraph:** Lưu trữ đồ thị Lineage (Nodes & Edges) qua Bolt protocol.

## 🛠 Tech Stack đề xuất
| Thành phần | Công nghệ |
| :--- | :--- |
| **Backend** | Golang (Echo framework) |
| **Metadata DB** | PostgreSQL (GORM) |
| **Graph DB** | Memgraph (Neo4j Go driver / Bolt) |
| **Data Orchestrator** | Apache Airflow 2.x (Triggering events) |
| **Communication** | REST API |
| **Authentication** | JWT (echo-jwt/v5) |

## 📊 Mô hình dữ liệu cốt lõi
- **Dataset:** Thông tin bảng, schema, định dạng cột.
- **Job/Task:** Định danh tiến trình xử lý từ Airflow.
- **User/Owner:** Người chịu trách nhiệm cho dữ liệu.
- **Edge (Link):** Mô tả sự dịch chuyển/biến đổi dữ liệu giữa các Dataset.

## 📝 Lộ trình triển khai
1. **Giai đoạn 1:** Xây dựng Core API và Schema Registry bằng Go.
2. **Giai đoạn 2:** Tích hợp Memgraph để xử lý quan hệ đồ thị (Lineage).
3. **Giai đoạn 3:** CLI + API hoàn chỉnh và tính năng so sánh Schema (Schema Diff).
4. **Giai đoạn 4:** Tối ưu hóa hiệu suất với Worker Pool và Redis Cache.

Cấu trúc thư mục dự án:
go-data-catalog/
├── cmd/
│   └── server/
│       └── main.go           # Điểm khởi đầu của ứng dụng
├── internal/
│   ├── api/
│   │   ├── handlers/         # Xử lý các request HTTP từ Airflow
│   │   └── middleware/       # Auth, logging...
│   ├── model/
│   │   ├── dataset.go        # Định nghĩa Schema bảng, cột
│   │   └── lineage.go        # Định nghĩa cấu trúc Graph (Link)
│   ├── repository/
│   │   ├── postgres/         # Tương tác với PostgreSQL (Metadata)
│   │   └── memgraph/         # Tương tác với Memgraph (Lineage)
│   └── service/
│       ├── ingestion.go      # Logic xử lý bất đồng bộ qua Channels
│       └── catalog.go        # Logic nghiệp vụ quản lý Metadata
├── pkg/
│   ├── graph/                # Tiện ích bổ trợ cho Graph DB
│   └── config/               # Cấu hình hệ thống (Viper/Env)
├── airflow-provider/
│   ├── hooks/                # Custom Hook gửi metadata đến API
│   └── callbacks/            # on_success_callback cho DAGs
├── scripts/                  # SQL migrations, Neo4j constraints
├── .env                      # Biến môi trường
├── go.mod
└── README.md

---
*Dự án được thiết kế cho các kỹ sư dữ liệu cần một giải pháp quản lý metadata linh hoạt và mạnh mẽ.*