package system

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	sysModel "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	sysReq "github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	sysResp "github.com/flipped-aurora/gin-vue-admin/server/model/system/response"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type ContactImportService struct{}

type contactImportParsed struct {
	Columns []string
	Rows    []map[string]string
	Total   int
}

func (s *ContactImportService) Upload(userID uint, fileName string, data []byte) (sysResp.ContactImportUploadResp, error) {
	if len(data) == 0 {
		return sysResp.ContactImportUploadResp{}, errors.New("file empty")
	}
	ext := strings.ToLower(strings.TrimSpace(filepath.Ext(fileName)))
	fileType := ""
	switch ext {
	case ".csv":
		fileType = "csv"
	case ".xlsx", ".xlsm", ".xls":
		fileType = "excel"
	default:
		return sysResp.ContactImportUploadResp{}, errors.New("only csv/xlsx supported")
	}

	dir := filepath.Join(global.GVA_CONFIG.Local.StorePath, "contact_import")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return sysResp.ContactImportUploadResp{}, err
	}
	key := uuid.NewString() + ext
	path := filepath.Join(dir, key)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return sysResp.ContactImportUploadResp{}, err
	}

	parsed, err := parseContactImportFile(fileType, data)
	if err != nil {
		return sysResp.ContactImportUploadResp{}, err
	}

	columnsJSON, _ := json.Marshal(parsed.Columns)
	sample := parsed.Rows
	if len(sample) > 20 {
		sample = sample[:20]
	}
	sampleJSON, _ := json.Marshal(sample)

	job := sysModel.SysContactImportJob{
		SysUserID:   userID,
		Filename:    fileName,
		FilePath:    path,
		FileType:    fileType,
		Status:      "uploaded",
		Progress:    0,
		ColumnsJSON: string(columnsJSON),
		SampleJSON:  string(sampleJSON),
		Total:       parsed.Total,
	}
	if err := global.GVA_DB.Create(&job).Error; err != nil {
		return sysResp.ContactImportUploadResp{}, err
	}

	return sysResp.ContactImportUploadResp{
		JobID:   job.ID,
		Columns: parsed.Columns,
		Sample:  sample,
	}, nil
}

func (s *ContactImportService) UploadFromGoogleSheet(userID uint, url string) (sysResp.ContactImportUploadResp, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return sysResp.ContactImportUploadResp{}, errors.New("url required")
	}
	client := &http.Client{Timeout: 20 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return sysResp.ContactImportUploadResp{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return sysResp.ContactImportUploadResp{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return sysResp.ContactImportUploadResp{}, errors.New("download failed")
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return sysResp.ContactImportUploadResp{}, err
	}
	return s.Upload(userID, "google_sheet.csv", b)
}

func (s *ContactImportService) SuggestMapping(userID uint, jobID uint) (sysResp.ContactImportSuggestResp, error) {
	job, err := s.getJob(userID, jobID)
	if err != nil {
		return sysResp.ContactImportSuggestResp{}, err
	}
	var cols []string
	_ = json.Unmarshal([]byte(job.ColumnsJSON), &cols)
	suggest := suggestContactMapping(cols)
	return sysResp.ContactImportSuggestResp{
		Columns: cols,
		Suggest: suggest,
	}, nil
}

func (s *ContactImportService) Validate(userID uint, req sysReq.ContactImportValidateReq) (sysResp.ContactImportValidateResp, error) {
	job, err := s.getJob(userID, req.JobID)
	if err != nil {
		return sysResp.ContactImportValidateResp{}, err
	}
	parsed, err := s.loadAllRows(job)
	if err != nil {
		return sysResp.ContactImportValidateResp{}, err
	}

	if err := global.GVA_DB.Delete(&sysModel.SysContactImportRowError{}, "sys_user_id = ? AND job_id = ?", userID, job.ID).Error; err != nil {
		return sysResp.ContactImportValidateResp{}, err
	}

	emailToIDs, err := s.existingEmailSet(userID)
	if err != nil {
		return sysResp.ContactImportValidateResp{}, err
	}

	seenInFile := map[string]int{}
	total := 0
	valid := 0
	invalid := 0
	dup := 0
	missing := 0

	for idx, row := range parsed.Rows {
		total++
		record, errs, isDup := s.mapRow(row, req.Mapping, emailToIDs, seenInFile)
		_ = record
		if isDup {
			dup++
		}
		if hasErr(errs, "missing_required") {
			missing++
		}
		if hasErr(errs, "invalid_email") {
			invalid++
		}
		if len(errs) == 0 {
			valid++
			continue
		}
		rawJSON, _ := json.Marshal(row)
		errsJSON, _ := json.Marshal(errs)
		if err := global.GVA_DB.Create(&sysModel.SysContactImportRowError{
			SysUserID:  userID,
			JobID:      job.ID,
			RowIndex:   idx + 2,
			RawJSON:    string(rawJSON),
			ErrorsJSON: string(errsJSON),
		}).Error; err != nil {
			return sysResp.ContactImportValidateResp{}, err
		}
	}

	if err := global.GVA_DB.Model(&sysModel.SysContactImportJob{}).Where("id = ? AND sys_user_id = ?", job.ID, userID).Updates(map[string]any{
		"mapping_json":           mustJSON(req.Mapping),
		"options_json":           mustJSON(req.Options),
		"total":                  total,
		"valid_count":            valid,
		"invalid_count":          invalid,
		"duplicate_count":        dup,
		"missing_required_count": missing,
		"status":                 "validated",
	}).Error; err != nil {
		return sysResp.ContactImportValidateResp{}, err
	}

	return sysResp.ContactImportValidateResp{
		Total:                total,
		ValidCount:           valid,
		InvalidCount:         invalid,
		DuplicateCount:       dup,
		MissingRequiredCount: missing,
	}, nil
}

func (s *ContactImportService) Start(userID uint, req sysReq.ContactImportStartReq) error {
	job, err := s.getJob(userID, req.JobID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(job.MappingJSON) == "" {
		_ = global.GVA_DB.Model(&sysModel.SysContactImportJob{}).Where("id = ? AND sys_user_id = ?", job.ID, userID).Update("mapping_json", mustJSON(req.Mapping)).Error
	}
	_ = global.GVA_DB.Model(&sysModel.SysContactImportJob{}).Where("id = ? AND sys_user_id = ?", job.ID, userID).Update("options_json", mustJSON(req.Options)).Error

	if err := global.GVA_DB.Model(&sysModel.SysContactImportJob{}).Where("id = ? AND sys_user_id = ?", job.ID, userID).Updates(map[string]any{
		"status":        "running",
		"progress":      0,
		"error_message": "",
	}).Error; err != nil {
		return err
	}

	go s.runJob(job.ID, userID, req.Mapping, req.Options)
	return nil
}

func (s *ContactImportService) GetJob(userID uint, id uint) (sysModel.SysContactImportJob, error) {
	return s.getJob(userID, id)
}

func (s *ContactImportService) ListJobs(userID uint, pageInfo sysReq.ContactImportJobListReq) ([]sysResp.ContactImportJobSummary, int64, error) {
	limit := pageInfo.PageSize
	offset := pageInfo.PageSize * (pageInfo.Page - 1)
	db := global.GVA_DB.Model(&sysModel.SysContactImportJob{}).Where("sys_user_id = ?", userID)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var jobs []sysModel.SysContactImportJob
	if limit != 0 {
		db = db.Limit(limit).Offset(offset)
	}
	if err := db.Order("id desc").Find(&jobs).Error; err != nil {
		return nil, 0, err
	}
	out := make([]sysResp.ContactImportJobSummary, 0, len(jobs))
	for _, j := range jobs {
		out = append(out, sysResp.ContactImportJobSummary{
			ID:                   j.ID,
			Filename:             j.Filename,
			FileType:             j.FileType,
			Status:               j.Status,
			Progress:             j.Progress,
			Total:                j.Total,
			ValidCount:           j.ValidCount,
			InvalidCount:         j.InvalidCount,
			DuplicateCount:       j.DuplicateCount,
			MissingRequiredCount: j.MissingRequiredCount,
			CreatedCount:         j.CreatedCount,
			UpdatedCount:         j.UpdatedCount,
			FailedCount:          j.FailedCount,
			ErrorMessage:         j.ErrorMessage,
			CreatedAt:            j.CreatedAt,
			FinishedAt:           j.FinishedAt,
		})
	}
	return out, total, nil
}

func (s *ContactImportService) DeleteJob(userID uint, id uint) error {
	job, err := s.getJob(userID, id)
	if err != nil {
		return err
	}
	_ = os.Remove(job.FilePath)
	if strings.TrimSpace(job.ErrorFilePath) != "" {
		_ = os.Remove(job.ErrorFilePath)
	}
	if err := global.GVA_DB.Delete(&sysModel.SysContactImportRowError{}, "sys_user_id = ? AND job_id = ?", userID, job.ID).Error; err != nil {
		return err
	}
	return global.GVA_DB.Delete(&sysModel.SysContactImportJob{}, "sys_user_id = ? AND id = ?", userID, job.ID).Error
}

func (s *ContactImportService) ExportFailedCSV(userID uint, jobID uint) ([]byte, string, string, error) {
	job, err := s.getJob(userID, jobID)
	if err != nil {
		return nil, "", "", err
	}
	var errs []sysModel.SysContactImportRowError
	if err := global.GVA_DB.Where("sys_user_id = ? AND job_id = ?", userID, job.ID).Order("row_index asc").Find(&errs).Error; err != nil {
		return nil, "", "", err
	}
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	_ = w.Write([]string{"rowIndex", "errors", "raw"})
	for _, e := range errs {
		_ = w.Write([]string{intToString(e.RowIndex), e.ErrorsJSON, e.RawJSON})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, "", "", err
	}
	name := "import_failed_" + intToString(int(job.ID)) + ".csv"
	return buf.Bytes(), "text/csv", name, nil
}

func (s *ContactImportService) TemplateXLSX() ([]byte, string, string, error) {
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	headers := []string{"公司名", "网站", "联系人", "职位", "邮箱", "电话", "国家"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}
	b, err := f.WriteToBuffer()
	if err != nil {
		return nil, "", "", err
	}
	return b.Bytes(), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", "contacts_template.xlsx", nil
}

func (s *ContactImportService) TemplateCSV() ([]byte, string, string, error) {
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	_ = w.Write([]string{"公司名", "网站", "联系人", "职位", "邮箱", "电话", "国家"})
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, "", "", err
	}
	return buf.Bytes(), "text/csv", "contacts_template.csv", nil
}

func (s *ContactImportService) getJob(userID uint, id uint) (sysModel.SysContactImportJob, error) {
	var job sysModel.SysContactImportJob
	err := global.GVA_DB.Where("id = ? AND sys_user_id = ?", id, userID).First(&job).Error
	return job, err
}

func (s *ContactImportService) loadAllRows(job sysModel.SysContactImportJob) (contactImportParsed, error) {
	b, err := os.ReadFile(job.FilePath)
	if err != nil {
		return contactImportParsed{}, err
	}
	return parseContactImportFile(job.FileType, b)
}

func parseContactImportFile(fileType string, data []byte) (contactImportParsed, error) {
	switch fileType {
	case "csv":
		return parseCSV(data)
	case "excel":
		return parseExcel(data)
	default:
		return contactImportParsed{}, errors.New("unsupported file type")
	}
}

func parseCSV(data []byte) (contactImportParsed, error) {
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	r := csv.NewReader(bytes.NewReader(data))
	r.FieldsPerRecord = -1
	all, err := r.ReadAll()
	if err != nil {
		return contactImportParsed{}, err
	}
	if len(all) == 0 {
		return contactImportParsed{}, errors.New("no rows")
	}
	cols := normalizeHeaders(all[0])
	rows := make([]map[string]string, 0, len(all)-1)
	for _, line := range all[1:] {
		row := map[string]string{}
		for i, c := range cols {
			if i < len(line) {
				row[c] = strings.TrimSpace(line[i])
			} else {
				row[c] = ""
			}
		}
		rows = append(rows, row)
	}
	return contactImportParsed{Columns: cols, Rows: rows, Total: len(rows)}, nil
}

func parseExcel(data []byte) (contactImportParsed, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return contactImportParsed{}, err
	}
	defer func() { _ = f.Close() }()
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return contactImportParsed{}, errors.New("no sheet")
	}
	all, err := f.GetRows(sheets[0])
	if err != nil {
		return contactImportParsed{}, err
	}
	if len(all) == 0 {
		return contactImportParsed{}, errors.New("no rows")
	}
	cols := normalizeHeaders(all[0])
	rows := make([]map[string]string, 0, len(all)-1)
	for _, line := range all[1:] {
		empty := true
		for _, v := range line {
			if strings.TrimSpace(v) != "" {
				empty = false
				break
			}
		}
		if empty {
			continue
		}
		row := map[string]string{}
		for i, c := range cols {
			if i < len(line) {
				row[c] = strings.TrimSpace(line[i])
			} else {
				row[c] = ""
			}
		}
		rows = append(rows, row)
	}
	return contactImportParsed{Columns: cols, Rows: rows, Total: len(rows)}, nil
}

func normalizeHeaders(in []string) []string {
	out := make([]string, 0, len(in))
	seen := map[string]int{}
	for i, h := range in {
		name := strings.TrimSpace(h)
		if name == "" {
			name = "列" + intToString(i+1)
		}
		if n, ok := seen[name]; ok {
			seen[name] = n + 1
			name = name + "_" + intToString(n+1)
		} else {
			seen[name] = 1
		}
		out = append(out, name)
	}
	return out
}

func suggestContactMapping(columns []string) map[string]string {
	out := map[string]string{}
	best := func(field string, cand string) {
		if _, ok := out[field]; ok {
			return
		}
		out[field] = cand
	}
	for _, c := range columns {
		n := normalizeHeaderKey(c)
		switch {
		case strings.Contains(n, "email") || strings.Contains(n, "邮箱") || strings.Contains(n, "e-mail") || strings.Contains(n, "mail"):
			best("email", c)
		case strings.Contains(n, "company") || strings.Contains(n, "公司") || strings.Contains(n, "企业") || strings.Contains(n, "单位"):
			best("companyName", c)
		case strings.Contains(n, "website") || strings.Contains(n, "站点") || strings.Contains(n, "网址") || strings.Contains(n, "web"):
			best("website", c)
		case strings.Contains(n, "name") || strings.Contains(n, "联系人") || strings.Contains(n, "姓名") || strings.Contains(n, "contact"):
			best("contactName", c)
		case strings.Contains(n, "title") || strings.Contains(n, "职位") || strings.Contains(n, "职务") || strings.Contains(n, "position"):
			best("title", c)
		case strings.Contains(n, "phone") || strings.Contains(n, "电话") || strings.Contains(n, "手机") || strings.Contains(n, "tel"):
			best("phone", c)
		case strings.Contains(n, "country") || strings.Contains(n, "国家"):
			best("country", c)
		}
	}
	return out
}

func normalizeHeaderKey(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "_", "")
	return s
}

func (s *ContactImportService) existingEmailSet(userID uint) (map[string]uint, error) {
	type row struct {
		ID    uint
		Email string
	}
	var rows []row
	if err := global.GVA_DB.Model(&sysModel.SysContact{}).Select("id,email").Where("sys_user_id = ?", userID).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := map[string]uint{}
	for _, r := range rows {
		out[normalizeEmail(r.Email)] = r.ID
	}
	return out, nil
}

func (s *ContactImportService) mapRow(row map[string]string, mapping map[string]string, existing map[string]uint, seenInFile map[string]int) (sysModel.SysContact, []string, bool) {
	get := func(field string) string {
		col := strings.TrimSpace(mapping[field])
		if col == "" {
			return ""
		}
		return strings.TrimSpace(row[col])
	}

	company := get("companyName")
	email := normalizeEmail(get("email"))
	contact := sysModel.SysContact{
		CompanyName: company,
		Website:     get("website"),
		ContactName: get("contactName"),
		Title:       get("title"),
		Email:       email,
		Phone:       get("phone"),
		Country:     get("country"),
		Status:      "uncontacted",
	}

	errs := make([]string, 0)
	if strings.TrimSpace(company) == "" || strings.TrimSpace(email) == "" {
		errs = append(errs, "missing_required")
	}
	if email != "" && !isEmailValidFormat(email) {
		errs = append(errs, "invalid_email")
	}
	isDup := false
	if email != "" {
		if _, ok := existing[email]; ok {
			isDup = true
		}
		if n, ok := seenInFile[email]; ok {
			seenInFile[email] = n + 1
			errs = append(errs, "duplicate")
			isDup = true
		} else {
			seenInFile[email] = 1
		}
	}
	sort.Strings(errs)
	return contact, errs, isDup
}

func (s *ContactImportService) runJob(jobID uint, userID uint, mapping map[string]string, opts sysReq.ContactImportOptions) {
	contactSvc := &ContactService{}
	listSvc := &ContactListService{}

	job, err := s.getJob(userID, jobID)
	if err != nil {
		return
	}
	_ = listSvc.EnsureDefaults(userID)

	parsed, err := s.loadAllRows(job)
	if err != nil {
		_ = s.failJob(userID, jobID, err.Error())
		return
	}
	existingEmailToID, err := s.existingEmailSet(userID)
	if err != nil {
		_ = s.failJob(userID, jobID, err.Error())
		return
	}

	onInvalid := strings.ToLower(strings.TrimSpace(opts.OnInvalid))
	if onInvalid == "" {
		onInvalid = "skip"
	}
	onDuplicate := strings.ToLower(strings.TrimSpace(opts.OnDuplicate))
	if onDuplicate == "" {
		onDuplicate = "update"
	}
	listID := uint(0)
	if opts.ListID != nil {
		listID = *opts.ListID
	}

	failedRows := make([][]string, 0)
	seenInFile := map[string]int{}

	total := len(parsed.Rows)
	created := 0
	updated := 0
	failed := 0

	for i, row := range parsed.Rows {
		contact, errs, _ := s.mapRow(row, mapping, existingEmailToID, seenInFile)
		if len(errs) > 0 {
			if onInvalid == "stop" {
				_ = s.failJob(userID, jobID, "validation failed")
				return
			}
			failed++
			failedRows = append(failedRows, []string{intToString(i + 2), strings.Join(errs, ","), mustJSON(row)})
			continue
		}

		err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
			if _, ok := existingEmailToID[contact.Email]; ok && onDuplicate == "skip" {
				return nil
			}
			saved, wasUpdate, err := contactSvc.UpsertByEmail(tx, userID, contact)
			if err != nil {
				return err
			}
			if wasUpdate {
				updated++
			} else {
				created++
			}
			existingEmailToID[contact.Email] = saved.ID
			if listID != 0 {
				if err := listSvc.AddContactToList(tx, userID, listID, saved.ID); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			failed++
			failedRows = append(failedRows, []string{intToString(i + 2), err.Error(), mustJSON(row)})
			continue
		}

		if i%50 == 0 || i == total-1 {
			progress := int(float64(i+1) / float64(max(total, 1)) * 100)
			_ = global.GVA_DB.Model(&sysModel.SysContactImportJob{}).Where("id = ? AND sys_user_id = ?", jobID, userID).Updates(map[string]any{
				"progress":      progress,
				"created_count": created,
				"updated_count": updated,
				"failed_count":  failed,
			}).Error
		}
	}

	errorFile := ""
	if len(failedRows) > 0 {
		dir := filepath.Join(global.GVA_CONFIG.Local.StorePath, "contact_import")
		_ = os.MkdirAll(dir, 0o755)
		errorFile = filepath.Join(dir, "import_failed_"+intToString(int(jobID))+"_"+uuid.NewString()+".csv")
		if err := writeFailedCSV(errorFile, failedRows); err != nil {
			errorFile = ""
		}
	}
	now := time.Now()
	_ = global.GVA_DB.Model(&sysModel.SysContactImportJob{}).Where("id = ? AND sys_user_id = ?", jobID, userID).Updates(map[string]any{
		"status":          "finished",
		"progress":        100,
		"total":           total,
		"created_count":   created,
		"updated_count":   updated,
		"failed_count":    failed,
		"error_file_path": errorFile,
		"finished_at":     &now,
	}).Error
}

func (s *ContactImportService) failJob(userID uint, jobID uint, msg string) error {
	now := time.Now()
	return global.GVA_DB.Model(&sysModel.SysContactImportJob{}).Where("id = ? AND sys_user_id = ?", jobID, userID).Updates(map[string]any{
		"status":        "failed",
		"error_message": msg,
		"finished_at":   &now,
	}).Error
}

func writeFailedCSV(path string, rows [][]string) error {
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	_ = w.Write([]string{"rowIndex", "errors", "raw"})
	for _, r := range rows {
		_ = w.Write(r)
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o600)
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func intToString(v int) string {
	return strconv.Itoa(v)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func hasErr(errs []string, code string) bool {
	for _, e := range errs {
		if e == code {
			return true
		}
	}
	return false
}
