package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"printsec-warroom/backend/internal/models"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListProjects(ctx context.Context, view string) ([]models.Project, error) {
	where := "1=1"
	switch view {
	case "active":
		where = "p.is_closed = 0"
	case "done":
		where = "p.is_closed = 1 AND p.install_date IS NOT NULL"
	case "lost":
		where = "p.is_closed = 1 AND p.install_date IS NULL"
	}
	query := fmt.Sprintf(`
SELECT p.project_id, p.project_code, p.project_name, p.client_name, p.spec_content,
       p.quote_content, ISNULL(CONVERT(varchar(10), p.quote_date, 23), ''), p.quote_note,
       p.custom_need, ISNULL(CONVERT(varchar(10), p.custom_date, 23), ''), p.custom_note, p.custom_days,
       ISNULL(CONVERT(varchar(10), p.poc_date, 23), ''), p.poc_result,
       ISNULL(CONVERT(varchar(10), p.install_date, 23), ''), ISNULL(CONVERT(varchar(10), p.done_date, 23), ''),
       p.is_closed, p.status, p.owner, p.editor, p.latest_note,
       ISNULL(q.file_name, ''), ISNULL(q.file_url, ''),
       ISNULL(c.file_name, ''), ISNULL(c.file_url, ''),
       CONVERT(varchar(19), p.created_at, 120), CONVERT(varchar(19), p.updated_at, 120)
FROM dbo.projects p
OUTER APPLY (SELECT TOP 1 file_name, file_url FROM dbo.project_files WHERE project_id = p.project_id AND file_kind = 'quote' ORDER BY created_at DESC) q
OUTER APPLY (SELECT TOP 1 file_name, file_url FROM dbo.project_files WHERE project_id = p.project_id AND file_kind = 'custom' ORDER BY created_at DESC) c
WHERE %s
ORDER BY p.updated_at DESC, p.project_id DESC`, where)

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var p models.Project
		if err := rows.Scan(
			&p.ID, &p.Code, &p.Name, &p.Client, &p.Spec,
			&p.QuoteContent, &p.QuoteDate, &p.QuoteNote,
			&p.CustomNeed, &p.CustomDate, &p.CustomNote, &p.CustomDays,
			&p.POCDate, &p.POCResult, &p.InstallDate, &p.DoneDate,
			&p.IsClosed, &p.Status, &p.Owner, &p.Editor, &p.LatestNote,
			&p.QuoteFileName, &p.QuoteFileURL, &p.CustomFileName, &p.CustomFileURL,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (r *Repository) CreateProject(ctx context.Context, input models.ProjectInput) (models.Project, error) {
	if input.Status == "" {
		input.Status = "need"
	}
	if input.Owner == "" {
		input.Owner = "KC"
	}
	if input.Editor == "" {
		input.Editor = input.Owner
	}
	if input.LatestNote == "" {
		input.LatestNote = "建立專案。"
	}
	var id int64
	err := r.db.QueryRowContext(ctx, `
INSERT INTO dbo.projects
  (project_code, project_name, client_name, spec_content, quote_content, quote_date, quote_note,
   custom_need, custom_date, custom_note, custom_days, poc_date, poc_result, install_date,
   done_date, is_closed, status, owner, editor, latest_note)
OUTPUT INSERTED.project_id
VALUES
  (dbo.next_project_code(), @p1, @p2, @p3, @p4, NULLIF(@p5, ''), @p6,
   @p7, NULLIF(@p8, ''), @p9, @p10, NULLIF(@p11, ''), @p12, NULLIF(@p13, ''),
   NULLIF(@p14, ''), @p15, @p16, @p17, @p18, @p19)`,
		input.Name, input.Client, input.Spec, input.QuoteContent, input.QuoteDate, input.QuoteNote,
		input.CustomNeed, input.CustomDate, input.CustomNote, input.CustomDays, input.POCDate,
		input.POCResult, input.InstallDate, input.DoneDate, input.IsClosed, input.Status,
		input.Owner, input.Editor, input.LatestNote,
	).Scan(&id)
	if err != nil {
		return models.Project{}, err
	}
	return r.GetProject(ctx, id)
}

func (r *Repository) GetProject(ctx context.Context, id int64) (models.Project, error) {
	projects, err := r.listByID(ctx, id)
	if err != nil {
		return models.Project{}, err
	}
	if len(projects) == 0 {
		return models.Project{}, sql.ErrNoRows
	}
	return projects[0], nil
}

func (r *Repository) listByID(ctx context.Context, id int64) ([]models.Project, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT p.project_id, p.project_code, p.project_name, p.client_name, p.spec_content,
       p.quote_content, ISNULL(CONVERT(varchar(10), p.quote_date, 23), ''), p.quote_note,
       p.custom_need, ISNULL(CONVERT(varchar(10), p.custom_date, 23), ''), p.custom_note, p.custom_days,
       ISNULL(CONVERT(varchar(10), p.poc_date, 23), ''), p.poc_result,
       ISNULL(CONVERT(varchar(10), p.install_date, 23), ''), ISNULL(CONVERT(varchar(10), p.done_date, 23), ''),
       p.is_closed, p.status, p.owner, p.editor, p.latest_note,
       ISNULL(q.file_name, ''), ISNULL(q.file_url, ''),
       ISNULL(c.file_name, ''), ISNULL(c.file_url, ''),
       CONVERT(varchar(19), p.created_at, 120), CONVERT(varchar(19), p.updated_at, 120)
FROM dbo.projects p
OUTER APPLY (SELECT TOP 1 file_name, file_url FROM dbo.project_files WHERE project_id = p.project_id AND file_kind = 'quote' ORDER BY created_at DESC) q
OUTER APPLY (SELECT TOP 1 file_name, file_url FROM dbo.project_files WHERE project_id = p.project_id AND file_kind = 'custom' ORDER BY created_at DESC) c
WHERE p.project_id = @p1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var projects []models.Project
	for rows.Next() {
		var p models.Project
		if err := rows.Scan(
			&p.ID, &p.Code, &p.Name, &p.Client, &p.Spec,
			&p.QuoteContent, &p.QuoteDate, &p.QuoteNote,
			&p.CustomNeed, &p.CustomDate, &p.CustomNote, &p.CustomDays,
			&p.POCDate, &p.POCResult, &p.InstallDate, &p.DoneDate,
			&p.IsClosed, &p.Status, &p.Owner, &p.Editor, &p.LatestNote,
			&p.QuoteFileName, &p.QuoteFileURL, &p.CustomFileName, &p.CustomFileURL,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (r *Repository) UpdateProject(ctx context.Context, id int64, input models.ProjectInput) (models.Project, error) {
	if input.Status == "" {
		if input.IsClosed && input.InstallDate == "" {
			input.Status = "lost"
		} else if input.IsClosed {
			input.Status = "done"
		} else {
			input.Status = "need"
		}
	}
	_, err := r.db.ExecContext(ctx, `
UPDATE dbo.projects
SET project_name = @p1, client_name = @p2, spec_content = @p3,
    quote_content = @p4, quote_date = NULLIF(@p5, ''), quote_note = @p6,
    custom_need = @p7, custom_date = NULLIF(@p8, ''), custom_note = @p9, custom_days = @p10,
    poc_date = NULLIF(@p11, ''), poc_result = @p12,
    install_date = NULLIF(@p13, ''), done_date = NULLIF(@p14, ''),
    is_closed = @p15, status = @p16, owner = @p17, editor = @p18, latest_note = @p19,
    updated_at = SYSUTCDATETIME()
WHERE project_id = @p20`,
		input.Name, input.Client, input.Spec, input.QuoteContent, input.QuoteDate, input.QuoteNote,
		input.CustomNeed, input.CustomDate, input.CustomNote, input.CustomDays, input.POCDate,
		input.POCResult, input.InstallDate, input.DoneDate, input.IsClosed, input.Status,
		input.Owner, input.Editor, input.LatestNote, id,
	)
	if err != nil {
		return models.Project{}, err
	}
	return r.GetProject(ctx, id)
}

func (r *Repository) CloseProject(ctx context.Context, id int64, installDate string, editor string) (models.Project, error) {
	status := "done"
	if strings.TrimSpace(installDate) == "" {
		status = "lost"
	}
	_, err := r.db.ExecContext(ctx, `
UPDATE dbo.projects
SET is_closed = 1,
    status = @p1,
    install_date = NULLIF(@p2, ''),
    done_date = CONVERT(date, SYSUTCDATETIME()),
    editor = @p3,
    latest_note = CASE WHEN @p1 = 'done'
      THEN N'結案完成，已歸檔至已完成結案。'
      ELSE N'結案完成但未填預計安裝日期，已歸檔至未成案。' END,
    updated_at = SYSUTCDATETIME()
WHERE project_id = @p4`, status, installDate, editor, id)
	if err != nil {
		return models.Project{}, err
	}
	return r.GetProject(ctx, id)
}

func (r *Repository) SaveFile(ctx context.Context, projectID int64, kind string, fileName string, fileURL string, uploadedBy string) (models.UploadResult, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
INSERT INTO dbo.project_files (project_id, file_kind, file_name, file_url, uploaded_by)
OUTPUT INSERTED.file_id
VALUES (@p1, @p2, @p3, @p4, @p5)`, projectID, kind, fileName, fileURL, uploadedBy).Scan(&id)
	if err != nil {
		return models.UploadResult{}, err
	}
	return models.UploadResult{ID: id, Kind: kind, FileName: fileName, FileURL: fileURL}, nil
}

func (r *Repository) ListEvents(ctx context.Context) ([]models.CalendarEvent, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT event_id, CONVERT(varchar(10), event_date, 23), event_time, event_type, title,
       owner, color, editor, CONVERT(varchar(19), created_at, 120), CONVERT(varchar(19), updated_at, 120)
FROM dbo.calendar_events
ORDER BY event_date, event_time, event_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []models.CalendarEvent
	for rows.Next() {
		var event models.CalendarEvent
		if err := rows.Scan(&event.ID, &event.EventDate, &event.EventTime, &event.Type, &event.Title, &event.Owner, &event.Color, &event.Editor, &event.CreatedAt, &event.UpdatedAt); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (r *Repository) CreateEvent(ctx context.Context, input models.CalendarEventInput) (models.CalendarEvent, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
INSERT INTO dbo.calendar_events (event_date, event_time, event_type, title, owner, color, editor)
OUTPUT INSERTED.event_id
VALUES (NULLIF(@p1, ''), @p2, @p3, @p4, @p5, @p6, @p7)`,
		input.EventDate, input.EventTime, input.Type, input.Title, input.Owner, input.Color, input.Editor,
	).Scan(&id)
	if err != nil {
		return models.CalendarEvent{}, err
	}
	return r.GetEvent(ctx, id)
}

func (r *Repository) GetEvent(ctx context.Context, id int64) (models.CalendarEvent, error) {
	var event models.CalendarEvent
	err := r.db.QueryRowContext(ctx, `
SELECT event_id, CONVERT(varchar(10), event_date, 23), event_time, event_type, title,
       owner, color, editor, CONVERT(varchar(19), created_at, 120), CONVERT(varchar(19), updated_at, 120)
FROM dbo.calendar_events
WHERE event_id = @p1`, id).Scan(&event.ID, &event.EventDate, &event.EventTime, &event.Type, &event.Title, &event.Owner, &event.Color, &event.Editor, &event.CreatedAt, &event.UpdatedAt)
	return event, err
}

func (r *Repository) UpdateEvent(ctx context.Context, id int64, input models.CalendarEventInput) (models.CalendarEvent, error) {
	_, err := r.db.ExecContext(ctx, `
UPDATE dbo.calendar_events
SET event_date = NULLIF(@p1, ''), event_time = @p2, event_type = @p3, title = @p4,
    owner = @p5, color = @p6, editor = @p7, updated_at = SYSUTCDATETIME()
WHERE event_id = @p8`, input.EventDate, input.EventTime, input.Type, input.Title, input.Owner, input.Color, input.Editor, id)
	if err != nil {
		return models.CalendarEvent{}, err
	}
	return r.GetEvent(ctx, id)
}

func (r *Repository) DeleteEvent(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM dbo.calendar_events WHERE event_id = @p1`, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
