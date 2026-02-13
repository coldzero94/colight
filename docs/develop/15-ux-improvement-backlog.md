# UX 개선 백로그

> 작성일: 2026-02-13
> Vercel React Best Practices 감사 + UX 리뷰에서 도출된 개선 사항 목록

---

## 완료 항목 (Phase 2.1)

- [x] 폼 `<label>` + `<input>` id/htmlFor 연결 (접근성)
- [x] 유효성 에러 메시지에 `role="alert"` 추가
- [x] STAR 필드 `aria-describedby` 가이드 연결
- [x] 글자수 카운터 `aria-live="polite"` 추가
- [x] Back link에 `aria-label` 추가
- [x] 무기 필터 탭 `aria-label` 추가 + 터치 타겟 44px 확보
- [x] `maxLength` HTML 속성 보강 (title, role, content, STAR fields)
- [x] `router.push` → `<Link>` 변경 (prefetch + a11y)
- [x] ReactQueryDevtools `next/dynamic` lazy-load
- [x] `optimizePackageImports` 설정

---

## 미완료 항목

### High Priority (Phase 3~4에서 권장)

| # | 항목 | 설명 | 관련 파일 |
|---|------|------|-----------|
| 1 | 페이지네이션/무한 스크롤 | 경험 목록이 100건 이상 시 성능 저하 | `experiences/page.tsx` |
| 2 | 폼 미저장 변경 경고 | 브라우저 뒤로가기 시 작성 중 데이터 유실 방지 | `experience-form.tsx` |
| 3 | 에러 메시지 상세화 | toast 에러에 네트워크/서버/인증 구분 메시지 | 전체 mutation 핸들러 |

### Medium Priority (Phase 5~6에서 권장)

| # | 항목 | 설명 | 관련 파일 |
|---|------|------|-----------|
| 4 | 낙관적 UI 업데이트 | 생성/수정/삭제 시 즉시 UI 반영 | `use-experiences.ts` |
| 5 | 삭제 다이얼로그 모바일 바텀시트 | 모바일에서 중앙 모달 → 바텀시트 패턴 | `delete-dialog.tsx` |
| 6 | 필터 결과 빈 상태 강화 | weapon 코드 미존재 시 fallback 텍스트 | `experiences/page.tsx:130` |
| 7 | 폼 submit 로딩 스피너 | 버튼 텍스트 "저장 중..." + 스피너 아이콘 추가 | `experience-form.tsx` |
| 8 | STAR 외 필드 글자수 카운터 | title, role, content에도 카운터 표시 | `experience-form.tsx` |
| 9 | 입력 필드 에러 시 border 색상 변경 | 에러 상태에서 `border-red-300` 적용 | `experience-form.tsx` |

### Low Priority (Phase 6.1+ 이후)

| # | 항목 | 설명 | 관련 파일 |
|---|------|------|-----------|
| 10 | 페이지 전환 애니메이션 | 콘텐츠 fade-in 트랜지션 | layout 전체 |
| 11 | 브레드크럼 컴포넌트 | 중첩 라우트 계층 표시 | `[id]/page.tsx`, `[id]/edit/page.tsx` |
| 12 | 버튼 active 상태 | `:active` 피드백 추가 (모바일 터치) | 전체 버튼 |
| 13 | 가로 스크롤 힌트 | 무기 필터 탭 스크롤 gradient 힌트 | `weapon-filter-tabs.tsx` |
| 14 | 스티키 헤더 | 스크롤 시 헤더 고정 | `(main)/layout.tsx` |
| 15 | Grid 브레이크포인트 세분화 | `sm:` 구간 추가 (소형 폰 대응) | `experience-list.tsx` |
