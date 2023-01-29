package pkg

type AliyunDriverFile struct {
	DriveID      string `json:"drive_id"`
	DomainID     string `json:"domain_id"`
	FileID       string `json:"file_id"`
	ShareID      string `json:"share_id"`
	Name         string `json:"name"`
	Type         string `json:"type:"`
	CreateAt     string `json:"create_at"`
	UpdateAt     string `json:"update_at"`
	ParentFileID string `json:"parent_file_id"`
	RevisionID   string `json:"revision_id"`
	FromShareID  string `json:"from_share_id"`
}
