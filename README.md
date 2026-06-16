# PrintSec War Room - Vue + Go + MSSQL

列印資安專案群組工作分享系統，技術棧為：

- Frontend：Vue 3 + Vite
- Backend：Go 1.22 HTTP API
- Database：Microsoft SQL Server
- Uploads：後端集中保存檔案，MSSQL 保存檔案 metadata

## 專案結構

```text
frontend/             Vue 3 前端
backend/              Go API 後端
backend/migrations/   MSSQL table schema 與種子資料
database/             Docker MSSQL 建庫初始化
deploy/               docker compose 部署
```

## 功能

- 共用行事曆：新增、編輯、刪除、查看同仁編輯的預定行程
- 專案流程進度表：專案名稱、規格內容、報價內容、報價日期、客製化需求、客製化人天、POC 日期、POC 結果、預計安裝日期、完成日期、備注說明
- 報價日期與客製化日期：可填說明並上傳檔案
- 檔案集中歸檔：上傳到 Go 後端 `uploads/`，metadata 寫入 MSSQL
- 結案歸檔：勾選結案後從進行中移除
- 管理區域：已完成結案瀏覽、未成案瀏覽
- 多人同步：所有資料集中於 MSSQL

## 快速啟動 Docker 版本

請先安裝 Docker Desktop。

```bash
cd deploy
docker compose up --build
```

開啟：

```text
http://localhost:8088
```

API：

```text
http://localhost:8080/api/health
```

MSSQL：

```text
localhost:1433
Database: PrintSecWarRoom
User: sa
Password: YourStrong!Passw0rd
```

正式部署前請修改 `.env.example` 內的密碼，並複製成 `.env`。

## 手動開發啟動

### 1. 建立 MSSQL

先建立資料庫：

```sql
CREATE DATABASE PrintSecWarRoom;
```

### 2. 啟動 Go API

```bash
cd backend
go mod download
DATABASE_URL="sqlserver://sa:YourStrong!Passw0rd@localhost:1433?database=PrintSecWarRoom&encrypt=disable" go run .
```

後端啟動時會自動執行：

```text
backend/migrations/001_init.sql
```

### 3. 啟動 Vue

```bash
cd frontend
npm install
npm run dev
```

開啟：

```text
http://localhost:5173
```

## 讓同仁一起使用

部署到一台大家都連得到的主機，然後開放：

- `8088`：Vue 前端
- `8080`：Go API
- `1433`：MSSQL，建議只開給內部網路或後端主機

同仁只需要開：

```text
http://伺服器IP:8088
```

不同網段、不同地點使用時，建議放在公司 VPN、雲端主機或 VPS，並加 HTTPS 與登入權限。

## 資料庫設計

主要資料表：

- `dbo.projects`：專案主檔
- `dbo.project_files`：報價檔、客製檔集中歸檔 metadata
- `dbo.project_notes`：備注說明歷程
- `dbo.calendar_events`：共用行事曆

詳細 UML 與 UI 資料影響評估請見：

- `docs/database-design.md`

專案狀態：

- `need`：需求確認
- `quote`：報價中
- `poc`：POC 排程
- `dev`：客製開發
- `closing`：待結案
- `done`：已完成結案
- `lost`：未成案

結案判斷：

- `is_closed = 1` 且 `install_date IS NOT NULL`：已完成結案
- `is_closed = 1` 且 `install_date IS NULL`：未成案
- `is_closed = 0`：進行中

## API

- `GET /api/health`
- `GET /api/projects?view=active|done|lost`
- `POST /api/projects`
- `PUT /api/projects/{id}`
- `POST /api/projects/{id}/close`
- `POST /api/projects/{id}/files?kind=quote|custom`
- `GET /api/events`
- `POST /api/events`
- `PUT /api/events/{id}`
- `DELETE /api/events/{id}`

## 本機驗證狀態

已完成：

- Go 後端 `go test ./...` 編譯檢查通過
- Vue 前端 `npm run build` 通過
- npm audit 已修正到 0 vulnerabilities

未在本機執行：

- MSSQL 實體連線測試，因目前 Codex 環境沒有啟動 SQL Server

在有 Docker 的環境執行 `docker compose up --build` 即可同時啟動 MSSQL、Go API 與 Vue 前端。
