# Phase 5.2: 코칭 에디터

> **⚠️ 아키텍처 변경 사항**: 이 문서의 코드 예시 중 서버 사이드 로직(Supabase 직접 쿼리, `createClient`)은 Go 백엔드로 구현합니다. 프론트엔드 코드(컴포넌트, Tiptap 에디터)는 그대로 참고하세요.
>
> - `createClient` from `@/lib/supabase/server` → Go 백엔드 API 호출 (생성된 SDK 사용)
> - 자동 저장/버전 관리 API → Go 백엔드 `internal/controller/editor_controller.go`
> - Supabase 직접 쿼리 → Ent ORM 쿼리 (`internal/service/`)

## Overview

| 항목 | 내용 |
|------|------|
| **목표** | Tiptap 리치 텍스트 에디터를 통합하여 AI 초안을 편집하고, 글자수 카운터와 자동 저장, 내보내기 기능을 제공한다 |
| **선행 조건** | Phase 5.1 (초안 코칭) 완료, `cover_letters` + `cover_letter_versions` 테이블 존재, 초안 데이터가 DB에 저장된 상태 |
| **스프린트** | Sprint 4 |
| **관련 기능** | F13 (초안 코칭), F20 (버전 관리 일부) |
| **예상 공수** | 1일 (Day 5) |
| **산출물** | Tiptap 에디터 페이지, 글자수 카운터, 자동 저장, 버전 관리, 내보내기 |

---

## Progress

| Step | 이름 | 상태 |
|------|------|------|
| 5.2.1 | Tiptap 통합 | ⬜ 대기 |
| 5.2.2 | 에디터 페이지 | ⬜ 대기 |
| 5.2.3 | 자동 저장 | ⬜ 대기 |
| 5.2.4 | 내보내기 | ⬜ 대기 |

---

## Step 5.2.1: Tiptap 통합

### 목표

Tiptap 에디터를 프로젝트에 설치하고, 자소서 편집에 필요한 최소한의 툴바(굵게, 기울임, 리스트, 제목)와 글자수 카운트 확장을 설정한다.

### 체크리스트

- [ ] Tiptap 패키지 설치 (`@tiptap/react`, `@tiptap/starter-kit`, `@tiptap/extension-character-count`)
- [ ] 기본 에디터 컴포넌트 구현
- [ ] 툴바 구현 (Bold, Italic, Bullet List, Heading 2/3)
- [ ] Character Count 확장 설정 (한국어 글자수 기준)
- [ ] STAR 태그 하이라이트 커스텀 확장 (또는 CSS 기반)
- [ ] 에디터 스타일링 (Tailwind CSS, prose 클래스)
- [ ] 에디터 콘텐츠 초기화 (초안 데이터 로드)

### 패키지 설치

```bash
npm install @tiptap/react @tiptap/starter-kit @tiptap/extension-character-count @tiptap/pm
```

### 프론트엔드 컴포넌트

| 컴포넌트 | 위치 | Props | 설명 |
|----------|------|-------|------|
| `TiptapEditor` | `src/components/coaching/tiptap-editor.tsx` | `content: string, charLimit: number, onChange: fn, editable: boolean` | Tiptap 에디터 래퍼 |
| `EditorToolbar` | `src/components/coaching/editor-toolbar.tsx` | `editor: Editor` | 포맷팅 툴바 (Bold, Italic, List, Heading) |
| `CharCounter` | `src/components/coaching/char-counter.tsx` | `current: number, limit: number` | 글자수 카운터 (색상 변화) |

### 구현 코드

```typescript
// src/components/coaching/tiptap-editor.tsx
'use client';

import { useEditor, EditorContent } from '@tiptap/react';
import StarterKit from '@tiptap/starter-kit';
import CharacterCount from '@tiptap/extension-character-count';

interface TiptapEditorProps {
  content: string;
  charLimit: number;
  onChange: (content: string) => void;
  editable?: boolean;
}

export function TiptapEditor({
  content,
  charLimit,
  onChange,
  editable = true,
}: TiptapEditorProps) {
  const editor = useEditor({
    extensions: [
      StarterKit,
      CharacterCount.configure({
        limit: charLimit,
      }),
    ],
    content,
    editable,
    onUpdate: ({ editor }) => {
      onChange(editor.getText());  // 순수 텍스트 (글자수 기준)
    },
  });

  if (!editor) return null;

  const charCount = editor.storage.characterCount.characters();
  const percentage = Math.round((charCount / charLimit) * 100);

  return (
    <div className="border rounded-lg">
      <EditorToolbar editor={editor} />
      <EditorContent
        editor={editor}
        className="prose prose-sm max-w-none p-4 min-h-[400px] focus:outline-none"
      />
      <CharCounter current={charCount} limit={charLimit} />
    </div>
  );
}
```

### 글자수 카운터 색상 규칙

| 비율 | 색상 | 의미 |
|------|------|------|
| 0~70% | `text-gray-500` | 여유 있음 |
| 70~90% | `text-blue-600` | 적절한 범위 |
| 90~100% | `text-amber-600` | 거의 다 참 |
| 100% 초과 | `text-red-600 font-bold` | 초과 (경고) |

### 검증 방법

- [ ] 에디터 영역에 텍스트 입력/수정 가능
- [ ] Bold, Italic, Bullet List, Heading 툴바 동작
- [ ] 글자수 카운터가 실시간 업데이트
- [ ] `charLimit` 초과 시 빨간색 경고 표시
- [ ] 초안 콘텐츠가 에디터에 정상 로드
- [ ] STAR 태그가 시각적으로 구분 표시
- [ ] `editable=false` 시 읽기 전용

### 산출물

- `src/components/coaching/tiptap-editor.tsx`
- `src/components/coaching/editor-toolbar.tsx`
- `src/components/coaching/char-counter.tsx`

---

## Step 5.2.2: 에디터 페이지

### 목표

코칭 세션별 에디터 페이지를 구현한다. 초안이 Tiptap 에디터에 사전 로드되고, 사이드 패널에 분석 요약과 경험 카드가 표시되어 참고하며 편집할 수 있다.

### 체크리스트

- [ ] `(main)/coaching/[id]/edit/page.tsx` 페이지 생성
- [ ] 서버 컴포넌트에서 `cover_letters` + 최신 `cover_letter_versions` 로드
- [ ] 에디터 영역 (좌측, 메인): Tiptap 에디터 + 글자수 카운터
- [ ] 사이드 패널 (우측): 분석 요약 + 경험 카드
- [ ] 헤더: 기업명 + 문항 + 저장 상태 표시
- [ ] `loading.tsx`, `error.tsx` 추가
- [ ] URL 파라미터로 cover_letter_id 식별

### 프론트엔드 컴포넌트

| 컴포넌트 | 위치 | Props | 설명 |
|----------|------|-------|------|
| `CoachingEditPage` | `src/app/(main)/coaching/[id]/edit/page.tsx` | `params: { id }` | 에디터 페이지 (서버 컴포넌트) |
| `EditorLayout` | `src/components/coaching/editor-layout.tsx` | `coverLetter, analysis, experiences` | 에디터 + 사이드 패널 레이아웃 |
| `AnalysisSidebar` | `src/components/coaching/analysis-sidebar.tsx` | `analysis: AnalysisResult, experiences: Experience[]` | 분석 요약 사이드 패널 |
| `SaveIndicator` | `src/components/coaching/save-indicator.tsx` | `status: 'saved' \| 'saving' \| 'unsaved'` | 저장 상태 인디케이터 |

### 페이지 레이아웃

```
┌──────────────────────────────────────────────────────────────────────┐
│  ← 코칭 목록    삼성전자 - 소프트웨어 개발직              ✓ 저장됨   │
│  문항: "팀 프로젝트에서 어려움을 극복한 경험..."                      │
├────────────────────────────────────────────┬─────────────────────────┤
│                                            │  📊 분석 요약            │
│  ┌──────────────────────────────────────┐  │                         │
│  │ B  I  • ─  H2  H3                   │  │  필요 무기               │
│  ├──────────────────────────────────────┤  │  [🏆 문제해결] [협업]    │
│  │                                      │  │                         │
│  │  [상황]                              │  │  추천 구조               │
│  │  2023년 캡스톤 프로젝트에서 4명으로   │  │  상황 20% (160자)       │
│  │  구성된 팀의 팀장을 맡게 되었습니다.  │  │  과제 15% (120자)       │
│  │  우리 팀은 AI 기반 추천 시스템을      │  │  행동 40% (320자)       │
│  │  개발하는 과제를 진행하고 있었는데,   │  │  결과 25% (200자)       │
│  │                                      │  │                         │
│  │  [과제]                              │  │  핵심 키워드              │
│  │  팀원 간 기술 수준 차이가 크다는      │  │  [데이터 기반] [개선율]   │
│  │  문제가 발생했습니다. 특히 백엔드     │  │  [주도적] [소통] [성과]   │
│  │  담당 팀원이 기본적인 API 설계에도   │  │                         │
│  │  어려움을 겪고 있어, 프로젝트 전체   │  │  ──── 사용 경험 ────     │
│  │  일정이 지연될 위기에 처했습니다.     │  │  ┌─────────────────────┐│
│  │                                      │  │  │ 캡스톤 프로젝트      ││
│  │  [행동]                              │  │  │ 적합도 92%          ││
│  │  ...                                 │  │  │ [문제해결] [협업]    ││
│  │                                      │  │  └─────────────────────┘│
│  │                                      │  │                         │
│  └──────────────────────────────────────┘  │                         │
│                                            │                         │
│  542/800자 (67.8%)  ████████████░░░░░░░░   │                         │
│                                            │                         │
│  [💾 버전 저장]  [📋 복사]  [🔍 첨삭 요청]  │                         │
├────────────────────────────────────────────┴─────────────────────────┤
│  버전 이력: v1 (초안) 14:30 | v2 (수정) 15:12 | v3 (현재)            │
└──────────────────────────────────────────────────────────────────────┘
```

### 데이터 로드 (서버 컴포넌트)

```typescript
// src/app/(main)/coaching/[id]/edit/page.tsx
import { createClient } from '@/lib/supabase/server';

export default async function CoachingEditPage({
  params,
}: {
  params: { id: string };
}) {
  const supabase = await createClient();

  // cover_letter + 최신 버전 + application + analysis 로드
  const { data: coverLetter } = await supabase
    .from('cover_letters')
    .select(`
      *,
      cover_letter_versions (
        id, version_number, content, char_count, created_at
      ),
      applications!inner (
        id, company_name, position,
        company_analyses (result)
      )
    `)
    .eq('id', params.id)
    .order('version_number', {
      referencedTable: 'cover_letter_versions',
      ascending: false,
    })
    .single();

  const latestVersion = coverLetter.cover_letter_versions[0];

  return (
    <EditorLayout
      coverLetter={coverLetter}
      initialContent={latestVersion?.content || ''}
      analysis={coverLetter.applications.company_analyses[0]?.result}
    />
  );
}
```

### 반응형 디자인

| 화면 크기 | 레이아웃 |
|-----------|---------|
| Desktop (≥1024px) | 에디터 (70%) + 사이드 패널 (30%) 나란히 |
| Tablet (768~1023px) | 에디터 전체 폭 + 사이드 패널 접기/펼치기 토글 |
| Mobile (≤767px) | 에디터 전체 폭 + 사이드 패널은 바텀 시트 |

### 검증 방법

- [ ] URL `/coaching/[id]/edit`으로 접근 시 해당 자소서의 최신 버전 로드
- [ ] 에디터에서 텍스트 편집 가능
- [ ] 사이드 패널에 분석 요약 + 경험 카드 정상 표시
- [ ] 헤더에 기업명, 문항, 저장 상태 표시
- [ ] 존재하지 않는 ID 접근 시 404 또는 에러 페이지
- [ ] 다른 사용자의 자소서 접근 시 403 (RLS)
- [ ] 반응형: 768px 이하에서 사이드 패널 토글/바텀 시트

### 산출물

- `src/app/(main)/coaching/[id]/edit/page.tsx`
- `src/app/(main)/coaching/[id]/edit/loading.tsx`
- `src/app/(main)/coaching/[id]/edit/error.tsx`
- `src/components/coaching/editor-layout.tsx`
- `src/components/coaching/analysis-sidebar.tsx`
- `src/components/coaching/save-indicator.tsx`

---

## Step 5.2.3: 자동 저장

### 목표

에디터 내용 변경 시 2초 debounce로 자동 저장하고, 명시적 "버전 저장" 버튼 클릭 시 `cover_letter_versions`에 새 버전을 생성한다. 저장 상태(저장됨/저장 중/미저장)를 UI에 표시한다.

### 체크리스트

- [ ] Debounce 2초 자동 저장 로직 구현
- [ ] `cover_letters.updated_at` + 최신 버전의 `content` 업데이트
- [ ] "버전 저장" 버튼: `cover_letter_versions`에 새 레코드 생성 (version_number 자동 증가)
- [ ] 저장 상태 인디케이터 (`saved` / `saving` / `unsaved`)
- [ ] 버전 이력 목록 표시 (하단 또는 사이드 패널)
- [ ] 이전 버전으로 되돌리기 (해당 버전 content를 에디터에 로드)
- [ ] 페이지 이탈 시 미저장 경고 (beforeunload)

### 구현 코드

```typescript
// src/hooks/use-auto-save.ts
'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import { useDebouncedCallback } from 'use-debounce';

type SaveStatus = 'saved' | 'saving' | 'unsaved';

export function useAutoSave(coverLetterId: string) {
  const [status, setStatus] = useState<SaveStatus>('saved');
  const lastSavedContent = useRef<string>('');

  // 자동 저장 (2초 debounce)
  const debouncedSave = useDebouncedCallback(
    async (content: string) => {
      if (content === lastSavedContent.current) return;

      setStatus('saving');
      try {
        await fetch(`/api/coaching/cover-letters/${coverLetterId}`, {
          method: 'PATCH',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ content }),
        });
        lastSavedContent.current = content;
        setStatus('saved');
      } catch (error) {
        setStatus('unsaved');
      }
    },
    2000
  );

  // 명시적 버전 저장
  const saveVersion = useCallback(async (content: string) => {
    setStatus('saving');
    try {
      await fetch(`/api/coaching/cover-letters/${coverLetterId}/versions`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ content }),
      });
      lastSavedContent.current = content;
      setStatus('saved');
    } catch (error) {
      setStatus('unsaved');
    }
  }, [coverLetterId]);

  // 페이지 이탈 경고
  useEffect(() => {
    const handler = (e: BeforeUnloadEvent) => {
      if (status === 'unsaved') {
        e.preventDefault();
      }
    };
    window.addEventListener('beforeunload', handler);
    return () => window.removeEventListener('beforeunload', handler);
  }, [status]);

  return { status, debouncedSave, saveVersion };
}
```

### API 엔드포인트

| Method | Path | Request | Response |
|--------|------|---------|----------|
| `PATCH` | `/api/coaching/cover-letters/[id]` | `{ content: string }` | `{ updated_at }` |
| `POST` | `/api/coaching/cover-letters/[id]/versions` | `{ content: string }` | `{ version: CoverLetterVersion }` |
| `GET` | `/api/coaching/cover-letters/[id]/versions` | - | `{ versions: CoverLetterVersion[] }` |

### 검증 방법

- [ ] 텍스트 변경 후 2초 뒤 자동 저장 확인 (저장 상태 → "저장 중" → "저장됨")
- [ ] "버전 저장" 클릭 시 `cover_letter_versions`에 새 레코드 생성 확인
- [ ] 버전 이력 목록에 버전 번호 + 저장 시간 표시
- [ ] 이전 버전 클릭 시 해당 내용이 에디터에 로드
- [ ] 미저장 상태에서 페이지 이탈 시 확인 다이얼로그
- [ ] 네트워크 에러 시 "저장 실패" 상태 + 재시도 안내

### 산출물

- `src/hooks/use-auto-save.ts`
- `src/app/api/coaching/cover-letters/[id]/route.ts` (PATCH)
- `src/app/api/coaching/cover-letters/[id]/versions/route.ts` (GET, POST)

---

## Step 5.2.4: 내보내기

### 목표

에디터의 자소서 내용을 클립보드에 순수 텍스트로 복사하는 기능을 제공한다. 자소서 입력 폼에 붙여넣기 위해 서식을 제거한 plain text 형태로 복사한다.

### 체크리스트

- [ ] "클립보드에 복사" 버튼 구현
- [ ] 에디터 HTML → plain text 변환 (서식 제거)
- [ ] STAR 태그 (`[상황]`, `[과제]`, `[행동]`, `[결과]`) 제거 옵션 (토글)
- [ ] 복사 성공 시 토스트 메시지 ("클립보드에 복사되었습니다")
- [ ] 글자수 함께 표시 ("542자 복사됨")

### 구현 코드

```typescript
// src/components/coaching/copy-button.tsx
'use client';

import { toast } from 'sonner';

interface CopyButtonProps {
  content: string;
  removeStarTags?: boolean;
}

export function CopyButton({ content, removeStarTags = false }: CopyButtonProps) {
  const handleCopy = async () => {
    let text = content;

    if (removeStarTags) {
      text = text
        .replace(/\[상황\]/g, '')
        .replace(/\[과제\]/g, '')
        .replace(/\[행동\]/g, '')
        .replace(/\[결과\]/g, '')
        .replace(/\n{3,}/g, '\n\n')
        .trim();
    }

    await navigator.clipboard.writeText(text);
    toast.success(`클립보드에 복사되었습니다 (${text.length}자)`);
  };

  return (
    <Button variant="outline" onClick={handleCopy}>
      📋 복사
    </Button>
  );
}
```

### 검증 방법

- [ ] "복사" 버튼 클릭 시 클립보드에 텍스트 복사 확인
- [ ] 복사된 텍스트에 HTML 태그 미포함 확인
- [ ] STAR 태그 제거 옵션 동작 확인
- [ ] 복사 성공 시 토스트 메시지 표시 + 글자수 표시
- [ ] HTTPS 환경에서만 동작 (HTTP에서는 fallback 처리)

### 산출물

- `src/components/coaching/copy-button.tsx`

---

## Phase 완료 체크리스트

- [ ] Tiptap 에디터에서 초안 편집 가능
- [ ] 툴바 (Bold, Italic, List, Heading) 정상 동작
- [ ] 글자수 카운터가 실시간 업데이트 + 색상 변화
- [ ] 자동 저장 (2초 debounce) 동작 확인
- [ ] "버전 저장" 시 `cover_letter_versions`에 새 버전 생성
- [ ] 이전 버전 복원 가능
- [ ] 클립보드 복사 (plain text) 동작
- [ ] 사이드 패널에 분석 요약 + 경험 카드 표시
- [ ] 반응형 (데스크탑/태블릿/모바일) 확인
- [ ] 페이지 이탈 시 미저장 경고

---

## 다음 Phase

**[Phase 6: 첨삭 코칭](./phase-6-review-coaching.md)** — AI 첨삭 (구체성/직무적합/기업맞춤/진정성 4점 평가)
