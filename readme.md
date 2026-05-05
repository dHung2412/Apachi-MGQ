# Apachi-MGQ: Lightweight Data Catalog & Governance

Dự án phát triển hệ thống quản lý Metadata, Data Lineage và Data Governance sử dụng ngôn ngữ Go. Hệ thống được thiết kế với mục tiêu cung cấp các tính năng quản lý dữ liệu cốt lõi tương đương DataHub nhưng với cấu trúc tinh gọn, tiêu thụ ít tài nguyên và đạt hiệu suất xử lý cao.

## 1. Mục tiêu hệ thống

- **Quản lý Metadata & Traceability:** Ghi nhận và theo dõi cấu trúc dữ liệu, sự dịch chuyển của dữ liệu từ hệ thống nguồn đến các hệ thống tiêu thụ (chi tiết đến cấp độ cột - Column-level Lineage).
- **Data Governance tích hợp:** Triển khai cơ chế phân quyền (RBAC), kiểm soát truy cập, định nghĩa quyền sở hữu đa tầng (Multiple Owners) và quản lý phân loại dữ liệu (Classification).
- **Thực thi chính sách (Policy Enforcement):** Tự động chặn các thao tác vi phạm quy định bảo mật hoặc quy trình dữ liệu thông qua Policy Engine.
- **Kiểm toán (Audit/Compliance):** Lưu vết toàn bộ các thao tác (đọc, ghi, xóa, từ chối quyền) dưới dạng log bất biến (immutable) phục vụ giám sát và bảo mật.
- **Hiệu suất xử lý:** Tận dụng khả năng xử lý đồng thời (concurrency) của Go kết hợp cơ chế hàng đợi (channels, worker pool) để xử lý dữ liệu lớn từ các tác vụ ETL/Airflow.

## 2. Kiến trúc hệ thống

Hệ thống được tổ chức thành 3 lớp chức năng chính:

1. **Ingestion Layer:**
   - Tiếp nhận Metadata thông qua REST API (Echo/Gin framework).
   - Hỗ trợ đa dạng nguồn cấp liệu: Airflow DAGs/Tasks, SQL Query Logs, định nghĩa công việc ETL, API thủ công.
   - Sử dụng Go Channels để đưa dữ liệu vào hàng đợi và xử lý bất đồng bộ.

2. **Processing & Governance Layer:**
   - **Lineage Engine:** Tính toán thay đổi (diff) của schema, ánh xạ quan hệ cột.
   - **Policy Engine:** Đánh giá các điều kiện chính sách (dựa trên CEL) trước khi cho phép thực thi API hoặc thay đổi Metadata.
   - **Access Control:** Xác thực (JWT) và phân quyền dựa trên vai trò (Role-Based Access Control).

3. **Storage Layer:**
   - **PostgreSQL:** Lưu trữ Metadata của Dataset, cấu hình User/Role, Policies, Tags và thông tin lịch sử.
   - **Memgraph (Neo4j driver/Bolt):** Lưu trữ đồ thị liên kết dữ liệu (Nodes & Edges) để truy vấn Data Lineage.
   - **External Storage (S3 Object Lock):** Lưu trữ Audit Logs đảm bảo tính bất biến, tách biệt khỏi cơ sở dữ liệu nghiệp vụ.

## 3. Công nghệ sử dụng

| Thành phần                             | Công nghệ / Công cụ                                  |
| :------------------------------------- | :--------------------------------------------------- |
| **Backend API**                        | Golang (Echo framework)                              |
| **Cơ sở dữ liệu quan hệ (Metadata)**   | PostgreSQL (GORM)                                    |
| **Cơ sở dữ liệu đồ thị (Lineage)**     | Memgraph (tương thích Neo4j / Bolt protocol, Cypher) |
| **Policy Engine**                      | Common Expression Language (CEL)                     |
| **Bộ nhớ đệm (Cache)**                 | Redis                                                |
| **Điều phối dữ liệu (Nguồn Metadata)** | Apache Airflow 2.x                                   |
| **Bảo mật / Xác thực**                 | JWT (echo-jwt/v5)                                    |
| **Lưu trữ Audit Logs**                 | S3-compatible storage (hỗ trợ Object Lock)           |

## 4. Mô hình dữ liệu & Quản trị (Governance Model)

### 4.1. Thực thể cơ sở

- **Dataset:** Thông tin bảng, schema dữ liệu, định dạng cột.
- **Job/Task:** Thông tin tiến trình xử lý và định danh công việc từ Airflow.
- **Edge/Link:** Định nghĩa quan hệ dịch chuyển hoặc biến đổi dữ liệu giữa các Dataset.

### 4.2. Phân quyền và Sở hữu (RBAC & Ownership)

Hệ thống hỗ trợ 5 vai trò (Roles) có các đặc quyền khác nhau trên tài nguyên (Dataset, Tag, Policy, User, Lineage):

- **Admin:** Quyền quản trị cao nhất trên toàn hệ thống.
- **DataOwner:** Có quyền đọc, ghi, xóa và gán quyền sở hữu cho Dataset mà họ quản lý.
- **DataConsumer:** Quyền đọc và truy vấn Dataset.
- **Viewer:** Chỉ có quyền xem thông tin (không thực thi truy vấn nội dung dữ liệu).
- **Customer:** Người dùng bên ngoài hệ thống, giới hạn truy cập trên một số Dataset nhất định.

**Mô hình đa sở hữu (Multiple Owners):**
Mỗi Dataset có thể được gán nhiều Owner phân theo loại:

- `Technical Owner`: Chịu trách nhiệm về cấu trúc schema và hệ thống nền tảng.
- `Business Owner`: Chịu trách nhiệm về ý nghĩa dữ liệu và logic nghiệp vụ.

### 4.3. Quản lý Chính sách và Kiểm toán

- **Policy Engine:** Cho phép định nghĩa các quy tắc như chặn công khai Dataset nếu chưa có Owner, yêu cầu lý do hợp lệ khi thao tác với dữ liệu nhạy cảm (PII). Chính sách áp dụng tại các điểm kiểm tra: tạo mới dataset, gán tag, thực thi API, hoặc từ hệ thống Airflow.
- **Audit Logging:** Ghi lại mọi sự kiện (READ, WRITE, DELETE, ACCESS_ATTEMPT, POLICY_VIOLATION). Dữ liệu log mang cấu trúc JSON, lưu trữ dạng append-only tại External Collector (S3) để đảm bảo không thể sửa đổi hoặc xóa bỏ.

## 5. Lộ trình phát triển (Roadmap)

Dự án được chia thành 5 giai đoạn triển khai:

**Giai đoạn 1: Core API & Storage (Cơ sở hạ tầng API & Lưu trữ)**

- Khởi tạo kiến trúc Go module (cmd, internal, pkg).
- Xây dựng Ingestion API nhận Metadata.
- Tích hợp PostgreSQL và định nghĩa cấu trúc bảng (Datasets, Jobs, Users).
- Thiết lập kết nối Memgraph (Neo4j driver) và tạo nodes/edges cơ bản cho Lineage.
- Triển khai logic xác thực cấu trúc (Schema Validation) cho Metadata đầu vào.

**Giai đoạn 2: Lineage Logic & Airflow Integration (Đường dẫn dữ liệu & Tích hợp)**

- Viết logic phân tích ánh xạ cấp độ cột (Column-level Mapping).
- Phát triển Airflow Provider/Hook đẩy sự kiện (Task runs, DAG status, input/output) thông qua API (Push method).
- Tự động hóa liên kết Lineage từ metadata thu thập được.

**Giai đoạn 3: Core Governance (Quản trị dữ liệu cốt lõi)**

- Triển khai mô hình RBAC và định nghĩa các vai trò.
- Cập nhật cấu trúc Ownership (hỗ trợ Technical/Business Owner).
- Xây dựng Policy Engine cơ bản sử dụng CEL.

**Giai đoạn 4: Enforcement & Audit (Thực thi chính sách & Kiểm toán)**

- Tích hợp Policy Engine vào API Handlers và Job processing để chặn thao tác không hợp lệ.
- Viết Interface cho Audit Logger và kết nối ghi log ra hệ thống ngoài (S3).
- Xây dựng giao diện/API tra cứu Audit Event cho phân quyền Admin/DataOwner.

**Giai đoạn 5: Optimization & UI (Tối ưu hóa & Giao diện)**

- Tối ưu hiệu năng bằng Go Channels (Async Worker Pool) để xử lý lượng Ingestion request lớn.
- Thêm Redis Cache cho các dữ liệu Schema/Lineage có tần suất truy cập cao.
- Phát triển Metadata Browser UI liệt kê Dataset/Owner.
- Xây dựng Lineage Visualization UI hiển thị đồ thị dữ liệu sử dụng thư viện trực quan hóa (D3.js / React Flow).

## 6. Cấu trúc thư mục dự án

```text
go-data-catalog/
├── cmd/
│   └── server/
│       └── main.go           # Entry point của ứng dụng
├── internal/
│   ├── api/
│   │   ├── handlers/         # Controller nhận request HTTP
│   │   └── middleware/       # Auth (JWT), Logging, Policy Enforcement check
│   ├── model/
│   │   ├── dataset.go        # Schema Dataset, Column, Tag
│   │   ├── lineage.go        # Định nghĩa Graph Node/Edge
│   │   ├── user.go           # Định nghĩa User, Role
│   │   ├── owner.go          # Quan hệ Multiple Owners
│   │   ├── policy.go         # Cấu trúc Policy và Rule (CEL)
│   │   └── audit.go          # Schema sự kiện Audit
│   ├── repository/
│   │   ├── postgres/         # Tương tác PostgreSQL (Metadata, Governance)
│   │   ├── memgraph/         # Tương tác Memgraph (Lineage)
│   │   ├── policy_repo.go    # Thao tác đọc/ghi rules
│   │   └── audit_repo.go     # Collector hook gửi log ra storage
│   └── service/
│       ├── ingestion.go      # Xử lý async qua Channels
│       ├── catalog.go        # Quản lý vòng đời Metadata
│       └── governance.go     # Xử lý RBAC, thực thi Policy Engine
├── pkg/
│   ├── graph/                # Tiện ích giao tiếp Graph DB
│   └── config/               # Cấu hình môi trường (Viper)
├── airflow-provider/
│   ├── hooks/                # Custom Hook tích hợp hệ thống Airflow
│   └── callbacks/            # Callback sự kiện gửi trạng thái DAG/Task
├── scripts/                  # Chứa script migration SQL, thiết lập Cypher/Bolt
├── .env                      # Tập tin cấu hình biến môi trường
├── go.mod                    # Dependency module Golang
└── README.md                 # Tài liệu hệ thống
```
