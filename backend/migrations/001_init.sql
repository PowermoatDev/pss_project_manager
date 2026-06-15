IF OBJECT_ID('dbo.projects', 'U') IS NULL
BEGIN
  CREATE TABLE dbo.projects (
    project_id bigint IDENTITY(1,1) NOT NULL CONSTRAINT pk_projects PRIMARY KEY,
    project_code nvarchar(40) NOT NULL CONSTRAINT uq_projects_code UNIQUE,
    project_name nvarchar(200) NOT NULL,
    client_name nvarchar(160) NOT NULL CONSTRAINT df_projects_client DEFAULT N'',
    spec_content nvarchar(max) NOT NULL CONSTRAINT df_projects_spec DEFAULT N'',
    quote_content nvarchar(max) NOT NULL CONSTRAINT df_projects_quote DEFAULT N'',
    quote_date date NULL,
    quote_note nvarchar(max) NOT NULL CONSTRAINT df_projects_quote_note DEFAULT N'',
    custom_need nvarchar(max) NOT NULL CONSTRAINT df_projects_custom_need DEFAULT N'',
    custom_date date NULL,
    custom_note nvarchar(max) NOT NULL CONSTRAINT df_projects_custom_note DEFAULT N'',
    custom_days decimal(10,2) NOT NULL CONSTRAINT df_projects_custom_days DEFAULT 0,
    poc_date date NULL,
    poc_result nvarchar(max) NOT NULL CONSTRAINT df_projects_poc DEFAULT N'',
    install_date date NULL,
    done_date date NULL,
    is_closed bit NOT NULL CONSTRAINT df_projects_closed DEFAULT 0,
    status varchar(20) NOT NULL CONSTRAINT df_projects_status DEFAULT 'need',
    owner nvarchar(60) NOT NULL CONSTRAINT df_projects_owner DEFAULT N'KC',
    editor nvarchar(60) NOT NULL CONSTRAINT df_projects_editor DEFAULT N'KC',
    latest_note nvarchar(max) NOT NULL CONSTRAINT df_projects_note DEFAULT N'',
    created_at datetime2(0) NOT NULL CONSTRAINT df_projects_created DEFAULT SYSUTCDATETIME(),
    updated_at datetime2(0) NOT NULL CONSTRAINT df_projects_updated DEFAULT SYSUTCDATETIME()
  );
END
GO

IF OBJECT_ID('dbo.project_files', 'U') IS NULL
BEGIN
  CREATE TABLE dbo.project_files (
    file_id bigint IDENTITY(1,1) NOT NULL CONSTRAINT pk_project_files PRIMARY KEY,
    project_id bigint NOT NULL,
    file_kind varchar(20) NOT NULL,
    file_name nvarchar(260) NOT NULL,
    file_url nvarchar(600) NOT NULL,
    uploaded_by nvarchar(60) NOT NULL CONSTRAINT df_project_files_by DEFAULT N'',
    created_at datetime2(0) NOT NULL CONSTRAINT df_project_files_created DEFAULT SYSUTCDATETIME(),
    CONSTRAINT fk_project_files_project FOREIGN KEY (project_id) REFERENCES dbo.projects(project_id)
  );
END
GO

IF OBJECT_ID('dbo.project_notes', 'U') IS NULL
BEGIN
  CREATE TABLE dbo.project_notes (
    note_id bigint IDENTITY(1,1) NOT NULL CONSTRAINT pk_project_notes PRIMARY KEY,
    project_id bigint NOT NULL,
    note_date date NOT NULL CONSTRAINT df_project_notes_date DEFAULT CONVERT(date, SYSUTCDATETIME()),
    note_text nvarchar(max) NOT NULL,
    created_by nvarchar(60) NOT NULL CONSTRAINT df_project_notes_by DEFAULT N'KC',
    created_at datetime2(0) NOT NULL CONSTRAINT df_project_notes_created DEFAULT SYSUTCDATETIME(),
    CONSTRAINT fk_project_notes_project FOREIGN KEY (project_id) REFERENCES dbo.projects(project_id)
  );
END
GO

IF OBJECT_ID('dbo.calendar_events', 'U') IS NULL
BEGIN
  CREATE TABLE dbo.calendar_events (
    event_id bigint IDENTITY(1,1) NOT NULL CONSTRAINT pk_calendar_events PRIMARY KEY,
    event_date date NOT NULL,
    event_time char(5) NOT NULL,
    event_type nvarchar(40) NOT NULL,
    title nvarchar(200) NOT NULL,
    owner nvarchar(60) NOT NULL CONSTRAINT df_calendar_owner DEFAULT N'KC',
    color varchar(20) NOT NULL CONSTRAINT df_calendar_color DEFAULT '#25d7ef',
    editor nvarchar(60) NOT NULL CONSTRAINT df_calendar_editor DEFAULT N'KC',
    created_at datetime2(0) NOT NULL CONSTRAINT df_calendar_created DEFAULT SYSUTCDATETIME(),
    updated_at datetime2(0) NOT NULL CONSTRAINT df_calendar_updated DEFAULT SYSUTCDATETIME()
  );
END
GO

CREATE OR ALTER FUNCTION dbo.next_project_code()
RETURNS nvarchar(40)
AS
BEGIN
  DECLARE @next bigint;
  SELECT @next = ISNULL(MAX(project_id), 0) + 1 FROM dbo.projects;
  RETURN CONCAT('PS-', YEAR(SYSUTCDATETIME()), '-', FORMAT(@next, '0000'));
END
GO

IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'ix_projects_status' AND object_id = OBJECT_ID('dbo.projects'))
  CREATE INDEX ix_projects_status ON dbo.projects(is_closed, status, install_date);
GO

IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'ix_project_files_project_kind' AND object_id = OBJECT_ID('dbo.project_files'))
  CREATE INDEX ix_project_files_project_kind ON dbo.project_files(project_id, file_kind, created_at DESC);
GO

IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'ix_calendar_events_date' AND object_id = OBJECT_ID('dbo.calendar_events'))
  CREATE INDEX ix_calendar_events_date ON dbo.calendar_events(event_date, event_time);
GO

IF NOT EXISTS (SELECT 1 FROM dbo.projects)
BEGIN
  INSERT INTO dbo.projects
    (project_code, project_name, client_name, spec_content, quote_content, quote_date, quote_note,
     custom_need, custom_date, custom_note, custom_days, poc_date, poc_result, install_date,
     done_date, is_closed, status, owner, editor, latest_note)
  VALUES
    (N'HY-2026-0178', N'宏遠集團 列印資安專案', N'宏遠集團', N'Secure Print 標準版、列印稽核、權限控管',
     N'標準版 + 追蹤稽核模組', '2026-06-03', N'報價內容已建立，待客戶確認。',
     N'客製報表、與 AD 整合', '2026-06-03', N'客製需求已彙整於集中歸檔。', 18,
     '2026-06-12', N'通過。功能符合需求，效能測試良好。', '2026-07-03',
     NULL, 0, 'need', N'KC', N'AZ', N'2026/06/12 POC 通過，待排安裝。'),
    (N'SKH-2026-0162', N'新光醫院 列印資安專案', N'新光醫院', N'Secure Print 標準版、浮水印、掃描管控',
     N'標準版', '2026-06-05', N'報價內容已建立，待客戶確認。',
     N'自訂浮水印、掃描管控', '2026-06-05', N'客製需求已彙整於集中歸檔。', 10,
     '2026-06-13', N'正在測試列印浮水印規則。', '2026-06-26',
     NULL, 0, 'quote', N'WL', N'WL', N'2026/06/13 正在測試列印浮水印規則。'),
    (N'CTBC-2026-0139', N'國泰金控 列印資安專案', N'國泰金控', N'企業版、金融稽核格式、權限控管',
     N'企業版', '2026-05-18', N'報價已完成。',
     N'金控稽核格式客製', '2026-05-18', N'已確認稽核欄位格式。', 35,
     '2026-06-10', N'已確認客製需求。', '2026-06-25',
     NULL, 0, 'closing', N'KC', N'KC', N'2026/06/10 已確認客製需求。'),
    (N'FCG-2026-0118', N'台塑集團 列印資安專案', N'台塑集團', N'標準版、稽核模組、多廠區集中管理',
     N'標準版 + 稽核模組', '2026-04-25', N'客戶調整預算，暫不導入。',
     N'多廠區集中管理', '2026-04-25', N'未進入客製排程。', 20,
     NULL, N'客戶調整時程。', NULL,
     '2026-06-06', 1, 'lost', N'YL', N'YL', N'2026/06/06 結案但未填預計安裝日期，歸檔為未成案。'),
    (N'NCCU-2026-0108', N'政大校務列印資安專案', N'政治大學', N'校園版、浮水印、帳號同步',
     N'校園版 + 浮水印', '2026-04-11', N'採購完成。',
     N'校務帳號同步', '2026-04-11', N'驗收完成。', 12,
     '2026-05-12', N'驗收通過。', '2026-06-01',
     '2026-06-08', 1, 'done', N'TW', N'TW', N'2026/06/08 結案完成，已歸檔至已完成結案。');
END
GO

IF NOT EXISTS (SELECT 1 FROM dbo.calendar_events)
BEGIN
  INSERT INTO dbo.calendar_events (event_date, event_time, event_type, title, owner, color, editor)
  VALUES
    ('2026-06-15', '09:00', N'POC', N'宏遠集團 POC 複測', N'AZ', '#25d7ef', N'AZ'),
    ('2026-06-16', '10:00', N'報價', N'新光醫院報價確認', N'WL', '#f7aa35', N'WL'),
    ('2026-06-16', '14:00', N'需求', N'長榮大學規格會議', N'HJ', '#9a73ff', N'HJ'),
    ('2026-06-17', '09:30', N'客製', N'高鐵 MFA 整合討論', N'JH', '#5a8cff', N'JH'),
    ('2026-06-18', '11:00', N'同步', N'中華電信專案同步', N'SC', '#9a73ff', N'SC'),
    ('2026-06-19', '16:00', N'結案', N'政大結案資料歸檔', N'TW', '#52ca78', N'TW');
END
GO

