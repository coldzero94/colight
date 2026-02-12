> **⚠️ 구현 언어 참고**: 이 문서의 코드 예시는 TypeScript로 작성되어 있으나, 실제 AI 파이프라인은 Go 백엔드에서 구현합니다. 로직 흐름은 참고 가능하며, Go 구현 패턴은 `docs/develop/04-ai-pipeline.md`를 참조하세요.

# AI 분석 파이프라인

> 작성일: 2026-02-09

---

## 1. 전체 파이프라인 구조

```
[채용공고 URL 입력]
    |
    v
[Step 1: 공고 파싱] ---------> 직무, 자격요건, 우대사항, 키워드 추출
    |
    v
[Step 2: 기업 정보 수집] ----> DART, 뉴스, 사전 DB에서 기업 데이터 병합
    |
    v
[Step 3: AI 기업 분석] ------> 인재상 추론, 핵심가치 요약, 기업 문화 분석
    |
    v
[Step 4: 경험 매칭] ---------> 사용자 경험 DB와 기업 요구사항 매칭
    |
    v
[Step 5: 결과 생성] ---------> 적합도 점수 + 매칭 근거 + 자소서 전략 제안
```

---

## 2. 프롬프트 설계

### Step 1: 채용공고 정보 구조화

추출 항목: company_name, position, department, job_type, experience_level, main_tasks, requirements, preferred, required_skills, soft_skills, company_values_hints, deadline

### Step 2: 기업 종합 분석

분석 항목:
1. **기업 핵심가치** (3~5개 키워드 + 각 설명)
2. **인재상 분석** (이 기업이 원하는 인재의 특성 3~5가지)
3. **최근 사업 동향** (최근 뉴스/공시 기반 요약)
4. **직무 핵심 역량** (이 포지션에서 가장 중요한 역량 순위)
5. **자소서 전략 키워드** (자소서에 반드시 녹여야 할 키워드 5~7개)
6. **피해야 할 표현** (이 기업/직무에 어울리지 않는 표현)

### Step 3: 경험 매칭

평가 기준:
- **직무 관련도** (0~100): 40% 가중치
- **인재상 부합도** (0~100): 35% 가중치
- **차별화 점수** (0~100): 25% 가중치
- **종합 적합도** (0~100): 가중 평균

---

## 3. 매칭 알고리즘

### 방법 1: LLM 기반 직접 매칭 (권장, 구현 간단)

- LLM이 직접 경험과 인재상을 비교 분석
- **장점**: 구현 간단, 의미적 매칭 정확도 높음
- **단점**: API 호출 비용, 응답 시간 2~5초

### 방법 2: 임베딩 + 코사인 유사도 (대량 처리시)

- 텍스트를 벡터로 변환 후 유사도 계산
- text-embedding-3-small 모델 사용
- **장점**: 빠른 계산, 대량 비교 가능, 비용 저렴
- **단점**: 의미적 맥락 이해 부족

### 권장: 하이브리드 방식

```
[1차: 임베딩 유사도로 빠른 필터링] → 상위 N개 경험 선별
        |
[2차: LLM으로 정밀 분석] → 선별된 경험에 대해 상세 매칭 + 점수 + 근거
```

---

## 4. 적합도 퍼센트 계산 로직

```typescript
function calculateOverallFit(scores: {
  jobRelevance: number;   // 직무 관련도
  talentFit: number;      // 인재상 부합도
  uniqueness: number;     // 차별화
}): number {
  const weights = {
    jobRelevance: 0.40,
    talentFit: 0.35,
    uniqueness: 0.25,
  };
  return Math.round(
    scores.jobRelevance * weights.jobRelevance +
    scores.talentFit * weights.talentFit +
    scores.uniqueness * weights.uniqueness
  );
}
```

---

## 5. 전체 분석 결과 스키마

```typescript
interface CompanyAnalysisResult {
  company: { name, industry, size, founded, employees, revenue, homepage };
  jobPosting: { position, mainTasks[], requirements[], preferred[], skills[] };
  analysis: {
    coreValues: { keyword, description }[];
    talentProfile: { trait, description, evidence }[];
    recentTrends: { title, summary, relevance }[];
    strategyKeywords: string[];
    avoidExpressions: string[];
  };
  matching: {
    overallFit: number;
    categoryScores: CategoryScore[];
    experienceMatches: ExperienceMatch[];
    recommendations: string[];
  };
  meta: { analyzedAt, sources[], confidence };
}
```
