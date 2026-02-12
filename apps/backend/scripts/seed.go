package main

import (
	"context"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/prompttemplate"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run seed.go [all|weapons|prompts|patterns]")
		os.Exit(1)
	}

	command := os.Args[1]

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
		fmt.Println("Usage: go run seed.go [all|weapons|prompts|patterns]")
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

// seedPromptTemplates inserts 4 prompt templates
func seedPromptTemplates(ctx context.Context, client *ent.Client) error {
	log.Println("Seeding prompt templates...")

	// Check if already seeded
	count, err := client.PromptTemplate.Query().Count(ctx)
	if err != nil {
		return fmt.Errorf("failed to count prompt templates: %w", err)
	}
	if count > 0 {
		log.Printf("⚠️  Prompt templates already exist (%d rows), skipping", count)
		return nil
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
			SystemPrompt: `당신은 취업 준비생의 경험을 분석하여 7대 역량 무기로 분류하는 AI입니다.

7대 무기 카테고리:
- W01 위기극복: 실패, 역경, 위기 극복
- W02 리더십: 팀을 이끌고 방향 제시
- W03 팀워크/협업: 협력하여 시너지 창출
- W04 도전정신: 새로운 도전, 목표 달성
- W05 문제해결: 복잡한 문제 분석과 해결
- W06 소통/설득: 효과적인 커뮤니케이션
- W07 성장/학습: 지속적 성장과 학습

사용자의 경험을 읽고 가장 적합한 무기 카테고리를 선택하세요.
주 무기 1개와 부 무기 1-2개를 선정하고, 그 이유를 간단히 설명하세요.`,
			UserPromptTemplate: `경험 제목: {{title}}

경험 내용:
{{content}}

결과:
{{result}}

위 경험에 가장 적합한 무기 카테고리를 분류하고, JSON 형식으로 응답하세요.`,
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
			Model:       "gpt-4.1-mini",
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
			Model:       "gpt-4.1-mini",
			Temperature: 0.5,
			MaxTokens:   500,
			Version:     1,
			IsActive:    true,
		},
		{
			Category:    "coaching_draft",
			SubCategory: "question_analysis",
			Name:        "자소서 문항 분석",
			SystemPrompt: `당신은 자기소개서 문항을 분석하는 AI 코치입니다.

분석 항목:
1. 문항 유형 (지원동기, 장단점, 위기극복, 리더십, 팀워크, 목표달성, 성장과정)
2. 요구하는 핵심 역량 (7대 무기 중 primary 1개, secondary 1-2개)
3. 평가 기준 (기업이 이 문항으로 보려는 것)
4. 작성 가이드 (구조, 비중, 주의사항)
5. 예시 개요 (좋은 답변의 골자)

기업 정보가 있다면 인재상과 직무 요구사항을 반영하세요.`,
			UserPromptTemplate: `자소서 문항: {{question_text}}

글자 제한: {{char_limit}}자

기업명: {{company_name}}
직무: {{position}}
기업 분석 정보:
{{company_analysis}}

위 문항을 분석하고, JSON 형식으로 응답하세요.`,
			OutputSchema: map[string]interface{}{
				"question_type":     "string",
				"primary_weapon":    "string (W01~W07)",
				"secondary_weapons": []string{},
				"evaluation_criteria": []map[string]interface{}{
					{"criterion": "string", "weight": "number"},
				},
				"writing_guide": map[string]interface{}{
					"structure": "string",
					"ratios":    map[string]interface{}{},
					"tips":      []string{},
				},
				"example_outline": "string",
			},
			Model:       "claude-sonnet-4-5",
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
			Model:       "claude-sonnet-4-5",
			Temperature: 0.4,
			MaxTokens:   2500,
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
	log.Println("✓ Inserted 4 prompt templates")

	return nil
}

// seedQuestionPatterns inserts 7 question patterns
func seedQuestionPatterns(ctx context.Context, client *ent.Client) error {
	log.Println("Seeding question patterns...")

	// Check if already seeded
	count, err := client.QuestionPattern.Query().Count(ctx)
	if err != nil {
		return fmt.Errorf("failed to count question patterns: %w", err)
	}
	if count > 0 {
		log.Printf("⚠️  Question patterns already exist (%d rows), skipping", count)
		return nil
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
