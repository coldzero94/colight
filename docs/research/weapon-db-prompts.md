# DB 레벨 프롬프트 설계

> 작성일: 2026-02-09

---

## 1. 왜 DB 레벨인가?

프롬프트를 코드에 하드코딩하면:
- 수정할 때마다 배포 필요
- A/B 테스트 불가
- 문항/기업/직무별 맞춤 프롬프트 관리 불가
- 프롬프트 성능 추적 불가

→ **프롬프트를 DB 테이블로 관리**하여 실시간 수정, 버전 관리, 성능 추적 가능

---

## 2. 프롬프트 DB 스키마

### prompt_templates (프롬프트 템플릿)

주요 필드: category, sub_category, system_prompt, user_prompt_template, output_schema(JSONB), model, temperature, max_tokens, version, is_active, usage_count, avg_latency_ms, avg_quality_score

### weapon_categories (무기 카테고리 마스터)

주요 필드: code(UNIQUE), parent_code, name, description, keywords(TEXT[]), question_patterns(TEXT[]), display_order, icon, color

### question_patterns (자소서 공통 문항 패턴)

주요 필드: pattern_type, pattern_name, detection_keywords(TEXT[]), detection_regex, primary_weapons(TEXT[]), secondary_weapons(TEXT[]), writing_guide(JSONB), coaching_prompt_id(FK)

### experience_weapons (경험→무기 매핑)

주요 필드: experience_id(FK), weapon_code(FK), confidence(FLOAT), is_primary(BOOL), reasoning(TEXT), user_confirmed(BOOL), user_modified(BOOL)

---

## 3. 핵심 프롬프트 템플릿 4종

### 프롬프트 1: 경험 무기 자동 분류

- category: `experience_classify` / sub: `weapon_tagging`
- model: `gemini-2.0-flash` / temperature: 0.2
- 역할: 사용자 경험을 분석하여 무기 카테고리로 분류
- 출력: primary_weapon, secondary_weapons, STAR 추출, matchable_questions, strength_keywords

### 프롬프트 2: 경험 인터뷰 (대화형)

- category: `experience_classify` / sub: `interview`
- model: `gemini-2.0-flash` / temperature: 0.7
- 역할: 대화형으로 경험을 끌어내는 코치
- 전략: 가볍게 물어보기 → 기억에 남는 순간 → 어려웠던 점 → 해결법 → 결과/배운 점 → STAR 정리

### 프롬프트 3: 자소서 문항 분석 + 무기 추천

- category: `coaching_draft` / sub: `question_analysis`
- model: `claude-sonnet-4-5` / temperature: 0.3
- 분석: 문항 의도, 필요 무기, 작성 구조(글자수 배분), 핵심 키워드, 피해야 할 것, 좋은 예시 구조

### 프롬프트 4: 무기별 경험 강화 코칭

- category: `coaching_draft` / sub: `weapon_enhance`
- model: `claude-sonnet-4-5` / temperature: 0.4
- 코칭: 역량 포인트 짚기, 구체성 보강 제안, 인재상 연결, STAR 약한 부분 보강 질문, 추상적→구체적 표현 변환

---

## 4. 프롬프트 변수 시스템

DB에 저장된 프롬프트의 `{{변수}}`는 런타임에 치환:

- `{{experience_text}}` - 사용자 입력 경험
- `{{experience_star}}` - STAR 구조화된 경험
- `{{weapon_categories}}` - DB에서 로드된 무기 목록
- `{{weapon_name}}` - 특정 무기명
- `{{company_analysis}}` - 기업 분석 결과
- `{{talent_profile}}` - 기업 인재상
- `{{question_text}}` - 자소서 문항
- `{{char_limit}}` - 글자수 제한
