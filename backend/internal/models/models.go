package models

type Project struct {
	ID             int64   `json:"id"`
	Code           string  `json:"code"`
	Name           string  `json:"name"`
	Client         string  `json:"client"`
	Spec           string  `json:"spec"`
	QuoteContent   string  `json:"quoteContent"`
	QuoteDate      string  `json:"quoteDate"`
	QuoteNote      string  `json:"quoteNote"`
	CustomNeed     string  `json:"customNeed"`
	CustomDate     string  `json:"customDate"`
	CustomNote     string  `json:"customNote"`
	CustomDays     float64 `json:"customDays"`
	POCDate        string  `json:"pocDate"`
	POCResult      string  `json:"pocResult"`
	InstallDate    string  `json:"installDate"`
	DoneDate       string  `json:"doneDate"`
	IsClosed       bool    `json:"isClosed"`
	Status         string  `json:"status"`
	Owner          string  `json:"owner"`
	Editor         string  `json:"editor"`
	LatestNote     string  `json:"latestNote"`
	QuoteFileName  string  `json:"quoteFileName"`
	QuoteFileURL   string  `json:"quoteFileUrl"`
	CustomFileName string  `json:"customFileName"`
	CustomFileURL  string  `json:"customFileUrl"`
	CreatedAt      string  `json:"createdAt"`
	UpdatedAt      string  `json:"updatedAt"`
}

type ProjectInput struct {
	Name         string  `json:"name"`
	Client       string  `json:"client"`
	Spec         string  `json:"spec"`
	QuoteContent string  `json:"quoteContent"`
	QuoteDate    string  `json:"quoteDate"`
	QuoteNote    string  `json:"quoteNote"`
	CustomNeed   string  `json:"customNeed"`
	CustomDate   string  `json:"customDate"`
	CustomNote   string  `json:"customNote"`
	CustomDays   float64 `json:"customDays"`
	POCDate      string  `json:"pocDate"`
	POCResult    string  `json:"pocResult"`
	InstallDate  string  `json:"installDate"`
	DoneDate     string  `json:"doneDate"`
	IsClosed     bool    `json:"isClosed"`
	Status       string  `json:"status"`
	Owner        string  `json:"owner"`
	Editor       string  `json:"editor"`
	LatestNote   string  `json:"latestNote"`
}

type CalendarEvent struct {
	ID        int64  `json:"id"`
	EventDate string `json:"eventDate"`
	EventTime string `json:"eventTime"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Owner     string `json:"owner"`
	Color     string `json:"color"`
	Editor    string `json:"editor"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type CalendarEventInput struct {
	EventDate string `json:"eventDate"`
	EventTime string `json:"eventTime"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Owner     string `json:"owner"`
	Color     string `json:"color"`
	Editor    string `json:"editor"`
}

type UploadResult struct {
	ID       int64  `json:"id"`
	Kind     string `json:"kind"`
	FileName string `json:"fileName"`
	FileURL  string `json:"fileUrl"`
}
