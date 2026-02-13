package crawler

// RawJobPosting represents unstructured job posting data extracted from HTML
type RawJobPosting struct {
	Source       string   // "jobkorea", "catch", etc.
	SourceURL    string   // Original URL
	CompanyName  string   // 회사명
	Position     string   // 포지션명
	Department   string   // 부서/팀
	Career       string   // 경력 조건
	Education    string   // 학력 조건
	JobType      string   // 고용형태
	Salary       string   // 급여
	Location     string   // 근무지
	Deadline     string   // 마감일
	MainTasks    string   // 담당업무
	Requirements string   // 자격요건
	Preferred    string   // 우대사항
	Skills       []string // 스킬 태그
	RawHTML      string   // HTML 원문 (AI fallback용)
}

// JobPosting represents normalized, structured job posting data after AI processing
type JobPosting struct {
	CompanyName      string   `json:"company_name"`
	Position         string   `json:"position"`
	Department       string   `json:"department,omitempty"`
	JobType          string   `json:"job_type,omitempty"`
	ExperienceLevel  string   `json:"experience_level,omitempty"`
	MainTasks        []string `json:"main_tasks,omitempty"`
	Requirements     []string `json:"requirements,omitempty"`
	Preferred        []string `json:"preferred,omitempty"`
	RequiredSkills   []string `json:"required_skills,omitempty"`
	SoftSkills       []string `json:"soft_skills,omitempty"`
	CompanyValuesHints []string `json:"company_values_hints,omitempty"`
	Deadline         string   `json:"deadline,omitempty"`
}
