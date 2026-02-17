package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/prompttemplate"
	"github.com/coby/colight/apps/backend/ent/userprofile"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run seed.go [all|admin|weapons|prompts|patterns]")
		os.Exit(1)
	}

	command := os.Args[1]

	// Load .env file (optional - ignore if not found)
	_ = godotenv.Load("../../.env")

	// Load DATABASE_URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5532/colight?sslmode=disable"
		log.Printf("DATABASE_URL not set, using default: %s", dbURL)
	}

	// Create Ent client
	client, err := ent.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Execute seed commands
	switch command {
	case "all":
		log.Println("Seeding all data...")
		if err := seedAdmin(ctx, client); err != nil {
			log.Fatalf("failed seeding admin: %v", err)
		}
		if err := seedWeaponCategories(ctx, client); err != nil {
			log.Fatalf("failed seeding weapon categories: %v", err)
		}
		if err := seedPromptTemplates(ctx, client); err != nil {
			log.Fatalf("failed seeding prompt templates: %v", err)
		}
		if err := seedQuestionPatterns(ctx, client); err != nil {
			log.Fatalf("failed seeding question patterns: %v", err)
		}
		log.Println("✅ All seed data inserted successfully")

	case "admin":
		if err := seedAdmin(ctx, client); err != nil {
			log.Fatalf("failed seeding admin: %v", err)
		}
		log.Println("✅ Admin account seeded successfully")

	case "weapons":
		if err := seedWeaponCategories(ctx, client); err != nil {
			log.Fatalf("failed seeding weapon categories: %v", err)
		}
		log.Println("✅ Weapon categories seeded successfully")

	case "prompts":
		if err := seedPromptTemplates(ctx, client); err != nil {
			log.Fatalf("failed seeding prompt templates: %v", err)
		}
		log.Println("✅ Prompt templates seeded successfully")

	case "patterns":
		if err := seedQuestionPatterns(ctx, client); err != nil {
			log.Fatalf("failed seeding question patterns: %v", err)
		}
		log.Println("✅ Question patterns seeded successfully")

	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Usage: go run seed.go [all|admin|weapons|prompts|patterns]")
		os.Exit(1)
	}
}

// seedWeaponCategories inserts 7 parent categories + 28 subcategories (35 total)
func seedWeaponCategories(ctx context.Context, client *ent.Client) error {
	log.Println("Seeding weapon categories...")

	// Check if already seeded
	count, err := client.WeaponCategory.Query().Count(ctx)
	if err != nil {
		return fmt.Errorf("failed to count weapon categories: %w", err)
	}
	if count > 0 {
		log.Printf("⚠️  Weapon categories already exist (%d rows), skipping", count)
		return nil
	}

	// Parent categories (7)
	parents := []struct {
		Code             string
		Name             string
		Description      string
		Keywords         []string
		QuestionPatterns []string
		DisplayOrder     int
		Icon             string
		Color            string
	}{
		{
			Code:        "W01",
			Name:        "위기극복",
			Description: "실패, 역경, 위기 상황을 극복한 경험",
			Keywords: []string{
				"실패", "좌절", "극복", "위기", "역경", "난관", "어려움", "시련",
				"재도전", "회복", "돌파", "변수", "긴급", "위험",
			},
			QuestionPatterns: []string{
				"실패를 극복한 경험",
				"어려움을 이겨낸 사례",
				"위기 상황에서의 대응",
			},
			DisplayOrder: 1,
			Icon:         "🛡️",
			Color:        "#EF4444",
		},
		{
			Code:        "W02",
			Name:        "리더십",
			Description: "팀을 이끌고 방향을 제시한 경험",
			Keywords: []string{
				"리더", "팀장", "주도", "이끌", "방향", "비전", "동기부여",
				"의사결정", "책임", "조직", "통솔", "관리", "목표설정",
			},
			QuestionPatterns: []string{
				"리더십을 발휘한 경험",
				"팀을 이끈 사례",
				"주도적으로 해결한 경험",
			},
			DisplayOrder: 2,
			Icon:         "👑",
			Color:        "#F59E0B",
		},
		{
			Code:        "W03",
			Name:        "팀워크/협업",
			Description: "타인과 협력하여 시너지를 낸 경험",
			Keywords: []string{
				"협업", "팀워크", "갈등", "조율", "소통", "화합", "협력",
				"역할분담", "시너지", "다양성", "존중", "공동목표",
			},
			QuestionPatterns: []string{
				"협업 경험",
				"팀워크를 발휘한 사례",
				"갈등을 해결한 경험",
			},
			DisplayOrder: 3,
			Icon:         "🤝",
			Color:        "#10B981",
		},
		{
			Code:        "W04",
			Name:        "도전정신",
			Description: "새로운 것에 도전하고 목표를 달성한 경험",
			Keywords: []string{
				"도전", "목표", "달성", "시도", "새로운", "혁신", "개척",
				"최초", "창업", "창작", "한계", "돌파", "성취", "기록",
			},
			QuestionPatterns: []string{
				"도전한 경험",
				"목표를 달성한 사례",
				"새로운 시도를 한 경험",
			},
			DisplayOrder: 4,
			Icon:         "🚀",
			Color:        "#8B5CF6",
		},
		{
			Code:        "W05",
			Name:        "문제해결",
			Description: "복잡한 문제를 분석하고 해결한 경험",
			Keywords: []string{
				"문제", "해결", "분석", "원인", "개선", "효율", "최적화",
				"창의", "혁신", "프로세스", "알고리즘", "디버깅", "진단",
			},
			QuestionPatterns: []string{
				"문제를 해결한 경험",
				"창의적으로 접근한 사례",
				"개선한 경험",
			},
			DisplayOrder: 5,
			Icon:         "🔧",
			Color:        "#3B82F6",
		},
		{
			Code:        "W06",
			Name:        "소통/설득",
			Description: "효과적으로 소통하고 설득한 경험",
			Keywords: []string{
				"소통", "설득", "협상", "발표", "프레젠테이션", "경청",
				"공감", "조율", "이해관계자", "커뮤니케이션", "전달",
			},
			QuestionPatterns: []string{
				"소통한 경험",
				"설득한 사례",
				"발표 경험",
			},
			DisplayOrder: 6,
			Icon:         "💬",
			Color:        "#EC4899",
		},
		{
			Code:        "W07",
			Name:        "성장/학습",
			Description: "지속적으로 성장하고 배운 경험",
			Keywords: []string{
				"성장", "학습", "배움", "자격증", "전문", "공부", "연구",
				"가치관", "멘토링", "피드백", "자기계발", "역량", "심화",
			},
			QuestionPatterns: []string{
				"성장한 경험",
				"배운 점",
				"자기계발 사례",
			},
			DisplayOrder: 7,
			Icon:         "📈",
			Color:        "#6366F1",
		},
	}

	// Insert parent categories
	bulk := make([]*ent.WeaponCategoryCreate, len(parents))
	for i, p := range parents {
		bulk[i] = client.WeaponCategory.Create().
			SetCode(p.Code).
			SetName(p.Name).
			SetDescription(p.Description).
			SetKeywords(p.Keywords).
			SetQuestionPatterns(p.QuestionPatterns).
			SetDisplayOrder(p.DisplayOrder).
			SetIcon(p.Icon).
			SetColor(p.Color).
			SetIsActive(true)
	}

	if _, err := client.WeaponCategory.CreateBulk(bulk...).Save(ctx); err != nil {
		return fmt.Errorf("failed to insert parent categories: %w", err)
	}
	log.Println("✓ Inserted 7 parent categories")

	// Subcategories (28 total: 4 per parent)
	subcategories := []struct {
		Code         string
		ParentCode   string
		Name         string
		Description  string
		DisplayOrder int
	}{
		// W01 subcategories
		{"W01-A", "W01", "프로젝트 위기", "프로젝트나 업무 중 발생한 위기 상황", 1},
		{"W01-B", "W01", "개인적 역경", "개인적 어려움이나 불리한 상황", 2},
		{"W01-C", "W01", "실패 후 재도전", "실패를 딛고 다시 시도한 경험", 3},
		{"W01-D", "W01", "예상치 못한 변수 대응", "갑작스러운 변수에 유연하게 대응", 4},

		// W02 subcategories
		{"W02-A", "W02", "공식적 리더", "팀장, 회장 등 공식 리더 역할", 1},
		{"W02-B", "W02", "비공식적 리더", "직책 없이 자연스럽게 리더 역할", 2},
		{"W02-C", "W02", "의사결정", "중요한 결정을 내리고 책임진 경험", 3},
		{"W02-D", "W02", "동기부여", "팀원들에게 동기를 부여하고 격려", 4},

		// W03 subcategories
		{"W03-A", "W03", "갈등 해결", "팀 내 갈등을 중재하고 해결", 1},
		{"W03-B", "W03", "역할 분담/조율", "역할을 나누고 일정 조율", 2},
		{"W03-C", "W03", "다양성 존중", "다양한 배경과 의견을 존중", 3},
		{"W03-D", "W03", "시너지 창출", "협력을 통해 1+1>2 효과 달성", 4},

		// W04 subcategories
		{"W04-A", "W04", "새로운 영역 도전", "처음 해보는 분야에 도전", 1},
		{"W04-B", "W04", "높은 목표 설정 & 달성", "어려운 목표를 세우고 달성", 2},
		{"W04-C", "W04", "창업/창작", "새로운 사업이나 작품 창조", 3},
		{"W04-D", "W04", "자기 한계 돌파", "자신의 한계를 뛰어넘은 경험", 4},

		// W05 subcategories
		{"W05-A", "W05", "분석적 접근", "데이터와 논리로 문제 분석", 1},
		{"W05-B", "W05", "창의적 접근", "기존과 다른 방식으로 해결", 2},
		{"W05-C", "W05", "프로세스 개선", "비효율을 발견하고 개선", 3},
		{"W05-D", "W05", "기술적 문제해결", "기술/코드 문제를 해결", 4},

		// W06 subcategories
		{"W06-A", "W06", "이해관계자 설득", "다양한 이해관계자를 설득", 1},
		{"W06-B", "W06", "프레젠테이션/발표", "효과적으로 발표하고 전달", 2},
		{"W06-C", "W06", "경청과 공감", "상대방의 입장을 이해하고 공감", 3},
		{"W06-D", "W06", "협상/조율", "서로 다른 의견을 조율", 4},

		// W07 subcategories
		{"W07-A", "W07", "자기주도 학습", "스스로 목표를 정하고 학습", 1},
		{"W07-B", "W07", "멘토링/피드백 수용", "조언을 받아들이고 성장", 2},
		{"W07-C", "W07", "가치관 형성", "경험을 통해 가치관이 형성됨", 3},
		{"W07-D", "W07", "전문성 심화", "특정 분야의 전문성을 깊게 쌓음", 4},
	}

	// Insert subcategories
	subBulk := make([]*ent.WeaponCategoryCreate, len(subcategories))
	for i, s := range subcategories {
		subBulk[i] = client.WeaponCategory.Create().
			SetCode(s.Code).
			SetParentCode(s.ParentCode).
			SetName(s.Name).
			SetDescription(s.Description).
			SetDisplayOrder(s.DisplayOrder).
			SetKeywords([]string{}).
			SetQuestionPatterns([]string{}).
			SetIsActive(true)
	}

	if _, err := client.WeaponCategory.CreateBulk(subBulk...).Save(ctx); err != nil {
		return fmt.Errorf("failed to insert subcategories: %w", err)
	}
	log.Println("✓ Inserted 28 subcategories")

	return nil
}

// seedPromptTemplates inserts 5 prompt templates
func seedPromptTemplates(ctx context.Context, client *ent.Client) error {
	log.Println("Seeding prompt templates...")

	// Delete question_patterns first (FK references prompt_templates)
	deletedQP, err := client.QuestionPattern.Delete().Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete question patterns (FK cleanup): %w", err)
	}
	if deletedQP > 0 {
		log.Printf("🗑️  Deleted %d question patterns (FK cleanup)", deletedQP)
	}

	// Delete existing prompt templates and re-seed
	deleted, err := client.PromptTemplate.Delete().Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete existing prompt templates: %w", err)
	}
	if deleted > 0 {
		log.Printf("🗑️  Deleted %d existing prompt templates", deleted)
	}

	templates := []struct {
		Category            string
		SubCategory         string
		Name                string
		SystemPrompt        string
		UserPromptTemplate  string
		OutputSchema        map[string]interface{}
		Model               string
		Temperature         float64
		MaxTokens           int
		Version             int
		IsActive            bool
	}{
		{
			Category:    "experience_classify",
			SubCategory: "weapon_tagging",
			Name:        "경험 무기 자동 분류",
			SystemPrompt: `당신은 취업 준비생의 경험을 역량 무기로 분류하는 분류기입니다.
설명이나 해설 없이, 오직 JSON만 출력하세요.

응답 형식 (이 형식을 정확히 따르세요):
{
  "primary_weapon": {
    "code": "W01",
    "confidence": 0.85,
    "reasoning": "한 문장 이유"
  },
  "secondary_weapons": [
    {"code": "W03", "confidence": 0.6, "reasoning": "한 문장 이유"}
  ]
}`,
			UserPromptTemplate: `아래 무기 카테고리 목록:
{{weapon_categories}}

아래 경험을 분석하여 주 무기 1개, 부 무기 0~2개를 선정하세요.

{{experience_text}}

JSON만 출력하세요.`,
			OutputSchema: map[string]interface{}{
				"primary_weapon": map[string]interface{}{
					"code":       "string (W01~W07)",
					"confidence": "number (0~1)",
					"reasoning":  "string",
				},
				"secondary_weapons": []map[string]interface{}{
					{
						"code":       "string (W01~W07)",
						"confidence": "number (0~1)",
					},
				},
			},
			Model:       "groq/compound",
			Temperature: 0.3,
			MaxTokens:   1000,
			Version:     1,
			IsActive:    true,
		},
		{
			Category:    "experience_classify",
			SubCategory: "interview",
			Name:        "경험 AI 인터뷰 (대화형)",
			SystemPrompt: `당신은 취업 준비생의 경험을 구조화하는 인터뷰어입니다.
사용자가 입력한 경험을 STAR 기법으로 구조화하도록 질문을 던지세요.

STAR 구조:
- Situation: 어떤 상황이었나요?
- Task: 당신의 역할과 목표는?
- Action: 어떤 행동을 취했나요?
- Result: 그 결과는 어땠나요?

자연스럽고 친근한 말투로 한 번에 한 가지씩 물어보세요.
사용자 답변이 부족하면 추가 질문으로 구체화를 유도하세요.`,
			UserPromptTemplate: `사용자 입력: {{user_input}}

대화 히스토리:
{{conversation_history}}

다음 질문을 생성하세요.`,
			OutputSchema: map[string]interface{}{
				"next_question": "string",
				"phase":         "string (situation | task | action | result | complete)",
				"completeness":  "number (0~1)",
			},
			Model:       "groq/compound",
			Temperature: 0.5,
			MaxTokens:   500,
			Version:     1,
			IsActive:    true,
		},
		{
			Category:    "coaching",
			SubCategory: "question_analysis",
			Name:        "자소서 문항 분석",
			SystemPrompt: `당신은 자기소개서 문항을 분석하는 AI 코치입니다.

분석 항목:
1. 문항의 표면적 질문과 숨겨진 의도 3가지
2. 요구하는 핵심 역량 (7대 무기 중 primary 1개, secondary 1-2개)
3. 작성 구조 (글자 배분, 섹션별 가이드)
4. 핵심 키워드 및 피해야 할 표현
5. 좋은 구조 예시

기업 정보가 있다면 인재상과 직무 요구사항을 반영하세요.
지원자의 경험이 제공되면, 해당 경험을 어떻게 활용하면 좋을지 구체적으로 안내하세요.

반드시 아래 JSON 형식으로만 응답하세요:
{
  "surface_question": "표면적 질문 요약",
  "real_intents": [{"intent": "숨겨진 의도", "why": "이유"}],
  "required_weapons": {
    "primary": {"weapon_id": "W01", "weapon_name": "무기명", "reason": "이유"},
    "secondary": [{"weapon_id": "W02", "weapon_name": "무기명", "reason": "이유"}]
  },
  "writing_structure": {
    "total_chars": 800,
    "sections": [{"name": "섹션명", "char_ratio": 0.3, "char_count": 240, "guide": "작성 가이드"}]
  },
  "key_keywords": ["키워드1", "키워드2"],
  "avoid_list": ["피해야 할 표현"],
  "good_structure_example": "좋은 구조 예시 텍스트"
}`,
			UserPromptTemplate: `기업명: {{company_name}}
직무: {{position}}
인재상 키워드: {{talent_keywords}}
핵심가치: {{values_keywords}}

무기 카테고리:
{{weapon_categories}}

자소서 문항: {{question_text}}
글자 제한: {{char_limit}}자

지원자의 경험:
{{experiences_context}}

위 문항을 분석하고, JSON 형식으로 응답하세요.`,
			OutputSchema: map[string]interface{}{
				"surface_question": "string",
				"real_intents":     []map[string]interface{}{{"intent": "string", "why": "string"}},
				"required_weapons": map[string]interface{}{
					"primary":   map[string]interface{}{"weapon_id": "string", "weapon_name": "string", "reason": "string"},
					"secondary": []map[string]interface{}{{"weapon_id": "string", "weapon_name": "string", "reason": "string"}},
				},
				"writing_structure": map[string]interface{}{
					"total_chars": "number",
					"sections":    []map[string]interface{}{{"name": "string", "char_ratio": "number", "char_count": "number", "guide": "string"}},
				},
				"key_keywords":          []string{},
				"avoid_list":            []string{},
				"good_structure_example": "string",
			},
			Model:       "gemini-2.0-flash",
			Temperature: 0.3,
			MaxTokens:   2000,
			Version:     2,
			IsActive:    true,
		},
		{
			Category:    "coaching",
			SubCategory: "review",
			Name:        "자소서 첨삭 평가",
			SystemPrompt: `당신은 한국 대기업 자기소개서 첨삭 전문가입니다.

자소서를 4가지 차원으로 평가하세요:
1. specificity (구체성): 수치, 이름, 기간 등 구체적 사실이 있는가
2. job_fit (직무적합성): 지원 직무에 필요한 역량을 보여주는가
3. company_fit (기업적합성): 기업 인재상/핵심가치와 부합하는가
4. authenticity (진정성): 진솔하고 자연스러운 경험인가

각 차원 0-100점, 종합 점수(overall), 차원별 상세 피드백(잘한 점 + 개선점), 문장 단위 구체적 수정 제안을 JSON으로 응답하세요.

응답 형식:
{
  "scores": {"specificity": N, "job_fit": N, "company_fit": N, "authenticity": N},
  "overall": N,
  "per_dimension_feedback": [
    {"dimension": "specificity", "score": N, "good": ["..."], "improve": ["..."]}
  ],
  "specific_suggestions": [
    {"original": "원문 문장", "suggested": "수정 제안", "reason": "수정 이유"}
  ]
}

반드시 위 JSON 형식으로만 응답하세요. 다른 텍스트를 포함하지 마세요.`,
			UserPromptTemplate: `자소서 내용:
{{content}}

문항: {{question_text}}
글자수 제한: {{char_limit}}자

{{company_context}}

위 자소서를 4가지 차원으로 평가하고, 구체적인 수정 제안을 포함한 JSON으로 응답하세요.`,
			OutputSchema: map[string]interface{}{
				"scores": map[string]interface{}{
					"specificity":  "number (0-100)",
					"job_fit":      "number (0-100)",
					"company_fit":  "number (0-100)",
					"authenticity": "number (0-100)",
				},
				"overall":                "number (0-100)",
				"per_dimension_feedback":  []map[string]interface{}{},
				"specific_suggestions":    []map[string]interface{}{},
			},
			Model:       "groq/compound",
			Temperature: 0.3,
			MaxTokens:   3000,
			Version:     1,
			IsActive:    true,
		},
		{
			Category:    "coaching",
			SubCategory: "draft",
			Name:        "자소서 초안 생성",
			SystemPrompt: `당신은 한국 대기업 자기소개서 작성을 돕는 AI 코칭 전문가입니다.

작성 원칙:
1. STAR 구조를 자연스럽게 녹여서 서술
2. 경험의 구체적 수치, 이름, 기간을 활용
3. 기업 인재상 키워드를 자연스럽게 반영
4. 글자 제한을 정확히 준수
5. 진정성 있고 자연스러운 문체

응답은 자소서 초안 텍스트만 출력하세요. JSON이 아닌 순수 텍스트입니다.`,
			UserPromptTemplate: `기업명: {{company_name}}
직무: {{position}}
기업 인재상 키워드: {{talent_keywords}}
기업 핵심가치: {{values_keywords}}

자소서 문항: {{question_text}}
글자 제한: {{char_limit}}자

선택된 경험:
{{experiences}}

위 경험을 바탕으로, 문항에 맞는 자소서 초안을 {{char_limit}}자 이내로 작성하세요.`,
			OutputSchema:  map[string]interface{}{},
			Model:         "claude-sonnet",
			Temperature:   0.7,
			MaxTokens:     3000,
			Version:       1,
			IsActive:      true,
		},
		{
			Category:    "coaching",
			SubCategory: "advice",
			Name:        "자소서 초안 개선 조언",
			SystemPrompt: `당신은 자기소개서 코칭 전문가입니다.
생성된 초안을 분석하여 개선 포인트를 제시합니다.

카테고리:
- metric: 수치/데이터 보강이 필요한 부분
- structure: 구조 개선이 필요한 부분
- detail: 구체성 강화가 필요한 부분
- keyword: 키워드 활용이 필요한 부분

반드시 아래 JSON 배열 형식으로만 응답하세요:
[
  {"category": "metric", "content": "조언 내용", "priority": 1},
  {"category": "structure", "content": "조언 내용", "priority": 2}
]

priority: 1=높음, 2=중간, 3=낮음
3~5개의 조언을 제시하세요.`,
			UserPromptTemplate: `자소서 문항: {{question_text}}

생성된 초안:
{{draft}}

활용된 경험 요약:
{{experiences_summary}}

위 초안의 개선 포인트를 JSON 배열로 응답하세요.`,
			OutputSchema:  map[string]interface{}{},
			Model:         "gemini",
			Temperature:   0.3,
			MaxTokens:     1000,
			Version:       1,
			IsActive:      true,
		},
		{
			Category:    "coaching",
			SubCategory: "char_coaching",
			Name:        "자소서 글자수 코칭",
			SystemPrompt: `당신은 자기소개서 글자수 조절 전문가입니다.
글자수가 초과되면 줄이는 제안을, 부족하면 늘리는 제안을 합니다.

응답 형식 (JSON):
{
  "suggestions": [
    {
      "type": "trim|expand",
      "original": "원문 문장",
      "suggested": "수정 제안",
      "char_diff": -15,
      "reason": "수정 이유"
    }
  ],
  "summary": "전체 코칭 요약"
}

규칙:
- 핵심 내용은 유지하되 불필요한 표현 제거/추가
- 구체적 문장 단위로 제안
- char_diff는 예상 글자수 변화량`,
			UserPromptTemplate: `자소서 내용:
{{content}}

현재 글자수: {{current_count}}자
글자수 제한: {{char_limit}}자
차이: {{diff}}자 (양수=초과, 음수=여유)
상태: {{status}}

글자수 조절을 위한 구체적 수정 제안을 JSON으로 응답하세요.`,
			OutputSchema: map[string]interface{}{
				"suggestions": []map[string]interface{}{},
				"summary":     "string",
			},
			Model:       "groq/compound",
			Temperature: 0.3,
			MaxTokens:   2000,
			Version:     1,
			IsActive:    true,
		},
		{
			Category:    "coaching_draft",
			SubCategory: "weapon_enhance",
			Name:        "무기별 경험 강화 코칭",
			SystemPrompt: `당신은 경험을 특정 역량 무기로 강화하는 코칭 AI입니다.

코칭 원칙:
1. 경험의 핵심은 유지하되, 특정 무기 키워드를 자연스럽게 강조
2. STAR 구조를 명확하게 (Situation → Task → Action → Result)
3. 정량적 성과를 추가하거나 강조
4. 추상적 표현을 구체적 행동으로 전환
5. 기업 인재상 키워드를 자연스럽게 반영

강화 전후 비교를 보여주고, 변경 이유를 설명하세요.`,
			UserPromptTemplate: `경험 내용:
{{experience_content}}

목표 무기: {{target_weapon_code}} - {{target_weapon_name}}

기업 키워드: {{company_keywords}}

글자 제한: {{char_limit}}자

위 경험을 목표 무기로 강화한 코칭 결과를 JSON 형식으로 응답하세요.`,
			OutputSchema: map[string]interface{}{
				"enhanced_content": "string",
				"changes": []map[string]interface{}{
					{
						"before":  "string",
						"after":   "string",
						"reason":  "string",
						"type":    "string (keyword | structure | quantify | specificity)",
					},
				},
				"char_count":   "number",
				"weapon_score": "number (0~100)",
				"tips":         []string{},
			},
			Model:       "groq/compound",
			Temperature: 0.4,
			MaxTokens:   2500,
			Version:     1,
			IsActive:    true,
		},
		// === Company Analysis ===
		{
			Category:    "company_analysis",
			SubCategory: "analyze",
			Name:        "기업 AI 분석",
			SystemPrompt: `당신은 한국 기업을 분석하는 AI 전문가입니다.
기업의 핵심가치, 인재상, 최근 트렌드를 분석해주세요.

응답 형식 (JSON):
{
  "core_values": [{"keyword": "핵심가치", "description": "설명"}],
  "talent_traits": [{"trait": "인재상", "description": "설명"}],
  "recent_trends": [{"title": "트렌드", "summary": "요약"}],
  "strategy_keywords": ["키워드1", "키워드2"],
  "avoid_expressions": ["피해야 할 표현"]
}

규칙:
- 기업 정보와 뉴스를 종합하여 분석
- 구체적이고 실용적인 키워드 도출
- 한국어로 작성`,
			UserPromptTemplate: `기업명: {{company_name}}

기업 정보:
{{company_context}}

최근 뉴스:
{{news_text}}

위 정보를 기반으로 기업을 분석하고 JSON으로 응답하세요.`,
			OutputSchema: map[string]interface{}{
				"core_values":      []map[string]interface{}{},
				"talent_traits":    []map[string]interface{}{},
				"recent_trends":    []map[string]interface{}{},
				"strategy_keywords": []string{},
				"avoid_expressions": []string{},
			},
			Model:       "groq/compound",
			Temperature: 0.3,
			MaxTokens:   2000,
			Version:     1,
			IsActive:    true,
		},
		// === Crawling: Extract from Markdown ===
		{
			Category:    "crawling",
			SubCategory: "extract_markdown",
			Name:        "마크다운 채용공고 추출",
			SystemPrompt: "You are a Korean job posting data extractor. Extract structured information from the provided content. Always respond in valid JSON.",
			UserPromptTemplate: `다음은 채용공고 페이지에서 추출된 마크다운 콘텐츠입니다.

URL: {{source_url}}

--- 콘텐츠 ---
{{content}}
--- 끝 ---

다음 필드를 포함한 JSON을 반환하세요:
- company_name, position, department, job_type, experience_level
- main_tasks[], requirements[], preferred[]
- required_skills[], soft_skills[], company_values_hints[]
- deadline`,
			OutputSchema: map[string]interface{}{},
			Model:       "gemini-2.5-flash",
			Temperature: 0.1,
			MaxTokens:   2000,
			Version:     1,
			IsActive:    true,
		},
		// === Crawling: Extract from HTML ===
		{
			Category:    "crawling",
			SubCategory: "extract_html",
			Name:        "HTML 채용공고 추출",
			SystemPrompt: "You are a Korean job posting data extractor. Extract structured information from raw HTML. Always respond in valid JSON.",
			UserPromptTemplate: `다음은 채용공고 페이지의 HTML입니다.

URL: {{source_url}}

--- HTML ---
{{content}}
--- 끝 ---

다음 필드를 포함한 JSON을 반환하세요:
- company_name, position, department, job_type, experience_level
- main_tasks[], requirements[], preferred[]
- required_skills[], soft_skills[], company_values_hints[]
- deadline`,
			OutputSchema: map[string]interface{}{},
			Model:       "gemini-2.5-flash",
			Temperature: 0.1,
			MaxTokens:   2000,
			Version:     1,
			IsActive:    true,
		},
		// === Crawling: Normalize ===
		{
			Category:    "crawling",
			SubCategory: "normalize",
			Name:        "채용공고 정규화",
			SystemPrompt: "You are a job posting data extractor. Extract structured information from raw text. Always respond in valid JSON.",
			UserPromptTemplate: `Company: {{company_name}}
Position: {{position}}
Department: {{department}}
Career: {{career}}
Location: {{location}}
Main Tasks: {{main_tasks}}
Requirements: {{requirements}}
Preferred: {{preferred}}
Skills: {{skills}}

Return JSON with: company_name, position, department, job_type, experience_level, main_tasks[], requirements[], preferred[], required_skills[], soft_skills[], company_values_hints[], deadline`,
			OutputSchema: map[string]interface{}{},
			Model:       "gemini-2.5-flash",
			Temperature: 0.2,
			MaxTokens:   2000,
			Version:     1,
			IsActive:    true,
		},
		// === Interview: Generate Question ===
		{
			Category:    "interview",
			SubCategory: "generate_question",
			Name:        "인터뷰 질문 생성",
			SystemPrompt: `당신은 취업 준비생의 경험을 발굴하는 친절한 AI 인터뷰어입니다.
한국어로 대화하며, 자연스럽고 편안한 톤으로 질문합니다.
한 번에 하나의 질문만 합니다. 질문은 간결하게 2-3문장 이내로 합니다.

현재 인터뷰 단계: {{stage_name}}
단계 지시사항: {{stage_instruction}}

반드시 아래 JSON 형식으로만 응답하세요:
{"question": "질문 내용"}`,
			UserPromptTemplate: `대화 기록:
{{conversation_history}}

위 대화를 바탕으로 다음 질문을 생성하세요.`,
			OutputSchema: map[string]interface{}{
				"question": "string",
			},
			Model:       "groq/compound",
			Temperature: 0.7,
			MaxTokens:   300,
			Version:     1,
			IsActive:    true,
		},
		// === Interview: Extract STAR ===
		{
			Category:    "interview",
			SubCategory: "extract_star",
			Name:        "인터뷰 STAR 추출",
			SystemPrompt: `당신은 인터뷰 대화에서 경험을 STAR 구조로 추출하는 전문가입니다.
아래 대화를 분석하여 핵심 경험을 STAR 구조로 정리하세요.

규칙:
- 모든 필드를 한국어로 작성
- title: 경험을 한 줄로 요약 (20자 이내)
- category: project, work, activity, competition, education, volunteer, other
- content: 경험의 전체적인 설명 (2-3문장)
- result: 최종 결과 요약 (1-2문장)
- star_situation/star_task/star_action/star_result 필드 포함
- keywords: 핵심 키워드 3-5개 배열

반드시 JSON 형식으로만 응답하세요.`,
			UserPromptTemplate: `인터뷰 대화:
{{conversation_history}}

위 대화에서 STAR 구조를 추출하세요.`,
			OutputSchema: map[string]interface{}{
				"title": "string", "category": "string",
				"star_situation": "string", "star_task": "string",
				"star_action": "string", "star_result": "string",
				"keywords": []string{},
			},
			Model:       "groq/compound",
			Temperature: 0.3,
			MaxTokens:   800,
			Version:     1,
			IsActive:    true,
		},
		// === Matching ===
		{
			Category:    "matching",
			SubCategory: "match_experience",
			Name:        "경험-기업 매칭 분석",
			SystemPrompt: "You are an experience-company matching analyst. Provide objective fit scores. Always respond in valid JSON.",
			UserPromptTemplate: `Match this experience against the company requirements and rate fit scores.

Experience:
{{experience_text}}

Company Requirements:
{{company_context}}

Return JSON with scores (0-100):
- overall_fit: weighted average (job_relevance*0.4 + talent_fit*0.35 + uniqueness*0.25)
- job_relevance: how relevant is this experience to the job
- talent_fit: how well does this match the talent profile
- uniqueness: differentiation factor
- reasoning: brief explanation
- suggested_angle: how to position this experience`,
			OutputSchema: map[string]interface{}{
				"overall_fit": "number", "job_relevance": "number",
				"talent_fit": "number", "uniqueness": "number",
				"reasoning": "string", "suggested_angle": "string",
			},
			Model:       "groq/compound",
			Temperature: 0.2,
			MaxTokens:   1000,
			Version:     1,
			IsActive:    true,
		},
		// === STAR Generation ===
		{
			Category:    "star_generation",
			SubCategory: "generate",
			Name:        "STAR 자동 생성",
			SystemPrompt: `당신은 취업 준비생의 자유 형식 경험 텍스트를 STAR 기법으로 구조화하는 AI입니다.
오직 JSON만 출력하세요. 설명이나 해설 없이 JSON만 반환하세요.

응답 형식:
{
  "star_situation": "상황 설명 (배경, 맥락, 시기, 조직)",
  "star_task": "과제/목표 설명 (해결해야 할 문제, 기대 성과)",
  "star_action": "구체적 행동 (전략, 실행한 것, 수치 포함)",
  "star_result": "결과 (정량적 성과, 질적 변화, 배운 점)"
}

규칙:
- 원문의 핵심 내용을 보존하되 STAR 구조로 재배치
- 각 필드는 2~4문장으로 구성
- 원문에 없는 내용을 지어내지 말 것
- 한국어로 작성`,
			UserPromptTemplate: `제목: {{title}}

내용:
{{content}}

JSON만 출력하세요.`,
			OutputSchema: map[string]interface{}{
				"star_situation": "string", "star_task": "string",
				"star_action": "string", "star_result": "string",
			},
			Model:       "groq/compound",
			Temperature: 0.3,
			MaxTokens:   1500,
			Version:     1,
			IsActive:    true,
		},
	}

	bulk := make([]*ent.PromptTemplateCreate, len(templates))
	for i, t := range templates {
		bulk[i] = client.PromptTemplate.Create().
			SetCategory(t.Category).
			SetSubCategory(t.SubCategory).
			SetName(t.Name).
			SetSystemPrompt(t.SystemPrompt).
			SetUserPromptTemplate(t.UserPromptTemplate).
			SetOutputSchema(t.OutputSchema).
			SetModel(t.Model).
			SetTemperature(t.Temperature).
			SetMaxTokens(t.MaxTokens).
			SetVersion(t.Version).
			SetIsActive(t.IsActive).
			SetUsageCount(0).
			SetAvgLatencyMs(0).
			SetAvgQualityScore(0)
	}

	if _, err := client.PromptTemplate.CreateBulk(bulk...).Save(ctx); err != nil {
		return fmt.Errorf("failed to insert prompt templates: %w", err)
	}
	log.Printf("✓ Inserted %d prompt templates", len(templates))

	return nil
}

// seedQuestionPatterns inserts 7 question patterns
func seedQuestionPatterns(ctx context.Context, client *ent.Client) error {
	log.Println("Seeding question patterns...")

	// Delete existing question patterns and re-seed
	deletedQP, err := client.QuestionPattern.Delete().Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete existing question patterns: %w", err)
	}
	if deletedQP > 0 {
		log.Printf("🗑️  Deleted %d existing question patterns", deletedQP)
	}

	// Get coaching prompt template ID for linking (sub_category = "weapon_enhance")
	coachingPrompt, err := client.PromptTemplate.Query().
		Where(
			prompttemplate.CategoryEQ("coaching_draft"),
			prompttemplate.SubCategoryEQ("weapon_enhance"),
		).
		First(ctx)
	if err != nil {
		log.Printf("⚠️  Warning: coaching prompt template not found, patterns will not be linked: %v", err)
		coachingPrompt = nil
	}

	patterns := []struct {
		PatternType        string
		PatternName        string
		DetectionKeywords  []string
		DetectionRegex     string
		PrimaryWeapons     []string
		SecondaryWeapons   []string
		WritingGuide       map[string]interface{}
		DisplayOrder       int
	}{
		{
			PatternType: "growth",
			PatternName: "성장과정/자기소개",
			DetectionKeywords: []string{
				"성장과정", "자기소개", "소개", "성장배경", "인생",
				"가치관", "형성", "배경", "어떤 사람",
			},
			DetectionRegex:   `(성장과정|자기소개|어떤 사람|본인을 소개)`,
			PrimaryWeapons:   []string{"W07"},
			SecondaryWeapons: []string{"W01", "W04"},
			WritingGuide: map[string]interface{}{
				"structure": "도입(배경) → 전환점 경험 → 현재 가치관 → 지원 동기 연결",
				"ratios": map[string]float64{
					"background": 0.2,
					"experience": 0.4,
					"values":     0.3,
					"motivation": 0.1,
				},
				"tips": []string{
					"특정 경험 1-2개를 중심으로 서술",
					"추상적 표현보다 구체적 에피소드",
					"기업 인재상과 연결되는 가치관 강조",
				},
			},
			DisplayOrder: 1,
		},
		{
			PatternType: "crisis",
			PatternName: "위기극복/실패경험",
			DetectionKeywords: []string{
				"위기", "극복", "실패", "어려움", "난관", "역경",
				"좌절", "시련", "대처", "해결",
			},
			DetectionRegex:   `(위기|극복|실패|어려움|난관|역경)`,
			PrimaryWeapons:   []string{"W01"},
			SecondaryWeapons: []string{"W05", "W04"},
			WritingGuide: map[string]interface{}{
				"structure": "상황 설정 → 위기 발생 → 극복 과정 → 배운 점",
				"ratios": map[string]float64{
					"situation": 0.2,
					"crisis":    0.2,
					"action":    0.4,
					"learning":  0.2,
				},
				"tips": []string{
					"위기의 심각성을 구체적으로",
					"극복 과정의 시행착오와 의사결정 포함",
					"배운 점을 직무에 연결",
				},
			},
			DisplayOrder: 2,
		},
		{
			PatternType: "leadership",
			PatternName: "리더십/주도적 경험",
			DetectionKeywords: []string{
				"리더십", "주도", "이끈", "팀장", "대표", "회장",
				"방향", "결정", "책임", "동기부여",
			},
			DetectionRegex:   `(리더십|주도|이끈|팀장|대표)`,
			PrimaryWeapons:   []string{"W02"},
			SecondaryWeapons: []string{"W03", "W06"},
			WritingGuide: map[string]interface{}{
				"structure": "리더 역할 배경 → 주도적 행동 → 팀 변화 → 성과",
				"ratios": map[string]float64{
					"background": 0.15,
					"action":     0.45,
					"impact":     0.25,
					"result":     0.15,
				},
				"tips": []string{
					"공식/비공식 리더십 구분",
					"팀원 동기부여 방법 구체화",
					"정량적 성과 포함",
				},
			},
			DisplayOrder: 3,
		},
		{
			PatternType: "teamwork",
			PatternName: "팀워크/협업/갈등해결",
			DetectionKeywords: []string{
				"팀워크", "협업", "협력", "갈등", "조율", "소통",
				"화합", "시너지", "협동", "공동",
			},
			DetectionRegex:   `(팀워크|협업|협력|갈등|조율)`,
			PrimaryWeapons:   []string{"W03"},
			SecondaryWeapons: []string{"W06", "W02"},
			WritingGuide: map[string]interface{}{
				"structure": "협업 배경 → 갈등/어려움 → 조율 과정 → 시너지 결과",
				"ratios": map[string]float64{
					"background": 0.2,
					"conflict":   0.2,
					"resolution": 0.35,
					"synergy":    0.25,
				},
				"tips": []string{
					"자신의 역할과 기여를 명확히",
					"갈등 해결 과정의 소통 방법",
					"협업으로 인한 시너지 효과 강조",
				},
			},
			DisplayOrder: 4,
		},
		{
			PatternType: "challenge",
			PatternName: "도전/목표달성",
			DetectionKeywords: []string{
				"도전", "목표", "달성", "성취", "시도", "새로운",
				"최초", "혁신", "창업", "한계", "돌파",
			},
			DetectionRegex:   `(도전|목표|달성|성취|시도)`,
			PrimaryWeapons:   []string{"W04"},
			SecondaryWeapons: []string{"W05", "W07"},
			WritingGuide: map[string]interface{}{
				"structure": "목표 설정 배경 → 도전 과정 → 난관 극복 → 결과",
				"ratios": map[string]float64{
					"motivation": 0.2,
					"challenge":  0.4,
					"obstacles":  0.2,
					"result":     0.2,
				},
				"tips": []string{
					"목표의 도전성을 강조 (왜 어려웠는지)",
					"시행착오와 개선 과정 포함",
					"정량적 성과와 의미 있는 변화",
				},
			},
			DisplayOrder: 5,
		},
		{
			PatternType: "problem_solving",
			PatternName: "문제해결/창의성",
			DetectionKeywords: []string{
				"문제", "해결", "분석", "원인", "개선", "효율",
				"창의", "혁신", "최적화", "프로세스",
			},
			DetectionRegex:   `(문제|해결|분석|원인|개선)`,
			PrimaryWeapons:   []string{"W05"},
			SecondaryWeapons: []string{"W04", "W07"},
			WritingGuide: map[string]interface{}{
				"structure": "문제 인식 → 원인 분석 → 해결 방안 → 효과",
				"ratios": map[string]float64{
					"problem":  0.2,
					"analysis": 0.25,
					"solution": 0.35,
					"impact":   0.2,
				},
				"tips": []string{
					"문제의 복잡성과 중요성 설명",
					"분석적/창의적 접근 방법 구체화",
					"해결 후 정량적 개선 효과",
				},
			},
			DisplayOrder: 6,
		},
		{
			PatternType: "motivation",
			PatternName: "지원동기/입사 후 포부",
			DetectionKeywords: []string{
				"지원동기", "지원하게 된 계기", "입사 후", "포부",
				"기여", "비전", "목표", "하고 싶은 일",
			},
			DetectionRegex:   `(지원동기|지원하게|입사|포부|기여|하고 싶은)`,
			PrimaryWeapons:   []string{"W07", "W04"},
			SecondaryWeapons: []string{"W05"},
			WritingGuide: map[string]interface{}{
				"structure": "관심 계기 → 경험 연결 → 기업/직무 적합성 → 포부",
				"ratios": map[string]float64{
					"trigger":  0.2,
					"relevant": 0.3,
					"fit":      0.3,
					"vision":   0.2,
				},
				"tips": []string{
					"구체적 경험으로 동기 입증",
					"기업 인재상/직무 요구사항 연구 결과 반영",
					"입사 후 기여 방안 구체화",
				},
			},
			DisplayOrder: 7,
		},
	}

	bulk := make([]*ent.QuestionPatternCreate, len(patterns))
	for i, p := range patterns {
		create := client.QuestionPattern.Create().
			SetPatternType(p.PatternType).
			SetPatternName(p.PatternName).
			SetDetectionKeywords(p.DetectionKeywords).
			SetDetectionRegex(p.DetectionRegex).
			SetPrimaryWeapons(p.PrimaryWeapons).
			SetSecondaryWeapons(p.SecondaryWeapons).
			SetWritingGuide(p.WritingGuide).
			SetDisplayOrder(p.DisplayOrder).
			SetIsActive(true)

		// Link to coaching prompt if found
		if coachingPrompt != nil {
			create = create.SetCoachingPrompt(coachingPrompt)
		}

		bulk[i] = create
	}

	if _, err := client.QuestionPattern.CreateBulk(bulk...).Save(ctx); err != nil {
		return fmt.Errorf("failed to insert question patterns: %w", err)
	}
	log.Println("✓ Inserted 7 question patterns")

	return nil
}

// seedAdmin inserts the default admin account
func seedAdmin(ctx context.Context, client *ent.Client) error {
	log.Println("Seeding admin account...")

	// Check if admin already exists
	exists, err := client.UserProfile.Query().
		Where(userprofile.RoleEQ(userprofile.RoleAdmin)).
		Exist(ctx)
	if err != nil {
		return fmt.Errorf("failed to check admin: %w", err)
	}
	if exists {
		log.Println("⚠️  Admin account already exists, skipping")
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), 12)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	_, err = client.UserProfile.Create().
		SetEmail("admin@colight.kr").
		SetPasswordHash(string(hash)).
		SetNickname("관리자").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleAdmin).
		SetEmailVerified(true).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create admin: %w", err)
	}

	log.Println("✓ Inserted admin account (admin@colight.kr)")
	return nil
}
