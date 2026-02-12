# 데이터 내보내기 & 백업 전략

> 사용자 데이터 내보내기, 계정 삭제 처리, DB 백업 및 재해 복구 전략
>
> 관련 문서: `08-legal-privacy.md` (개인정보 보관/삭제 정책), `02-data-structure.md` (DB 스키마), `10-deployment-ci.md` (인프라)

---

## 1. 사용자 데이터 내보내기

### 1.1 법적 근거

> **개인정보 자기결정권**: 사용자는 자신의 데이터를 열람하고 내보낼 권리가 있음 (개인정보보호법 제35조).
> `08-legal-privacy.md` 참조.

| 항목 | 정책 |
|------|------|
| **내보내기 요청 권한** | 모든 가입 사용자 (무료 포함) |
| **처리 기한** | 요청 후 72시간 이내 (법적 기한 10일보다 엄격하게 설정) |
| **요청 빈도 제한** | 월 3회 (어뷰징 방지) |
| **파일 보관 기간** | 다운로드 링크 발급 후 7일, 이후 자동 삭제 |

### 1.2 내보내기 범위

| 데이터 | 포함 여부 | JSON | PDF | 비고 |
|--------|----------|------|-----|------|
| **프로필 정보** (닉네임, 희망 직무/산업) | O | O | O | `user_profiles` 테이블 |
| **경험 데이터** (STAR 구조, 키워드) | O | O | O | `experiences` + `experience_tags` |
| **무기 태깅 결과** | O | O | O | `experience_weapons` (주/부 무기, confidence) |
| **자소서 전체 버전** | O | O | O | `cover_letters` + `cover_letter_versions` |
| **기업 분석 결과** | O | O | O | `company_analyses` (사용자별) |
| **코칭 세션 요약** | O | O | - | `coaching_sessions` (점수, 피드백 요약) |
| **지원 현황** | O | O | O | `applications` |
| **경험 사용 이력** | O | O | - | `experience_usages` |
| **임베딩 벡터** | X | - | - | 기술 데이터, 사용자에게 무의미 |
| **AI 프롬프트/세션 원문** | X | - | - | 서비스 내부 데이터 |
| **결제 이력** | 별도 | O | - | 결제 정보는 Toss Payments 보관 |

### 1.3 내보내기 포맷

#### JSON 포맷

사용자 전체 데이터를 구조화된 JSON으로 제공한다.

```json
{
  "export_version": "1.0",
  "exported_at": "2026-03-15T14:30:00+09:00",
  "user": {
    "nickname": "취준생A",
    "target_job": "마케팅",
    "target_industry": "IT",
    "created_at": "2026-02-01T10:00:00+09:00"
  },
  "experiences": [
    {
      "title": "마케팅 동아리 SNS 홍보",
      "category": "대외활동",
      "period": { "start": "2025-03-01", "end": "2025-12-31" },
      "role": "홍보팀장",
      "star": {
        "situation": "신입 부원 모집이 필요한 상황...",
        "task": "SNS 홍보를 통해 지원자 50명 이상 확보...",
        "action": "인스타그램 릴스 콘텐츠 기획...",
        "result": "지원자 83명 확보, 전년 대비 166% 증가"
      },
      "weapons": [
        { "code": "W05", "name": "문제해결", "is_primary": true, "confidence": 0.92 },
        { "code": "W06", "name": "소통/설득", "is_primary": false, "confidence": 0.78 }
      ],
      "tags": ["SNS 마케팅", "콘텐츠 기획", "데이터 분석"],
      "created_at": "2026-02-10T15:00:00+09:00"
    }
  ],
  "cover_letters": [
    {
      "company_name": "삼성전자",
      "question": "본인의 성장 과정에서 가장 힘들었던 경험은?",
      "versions": [
        {
          "version": 1,
          "content": "초안 내용...",
          "scores": { "specificity": 72, "job_fit": 68, "company_fit": 75, "authenticity": 80 },
          "created_at": "2026-03-01T10:00:00+09:00"
        },
        {
          "version": 2,
          "content": "수정된 내용...",
          "scores": { "specificity": 85, "job_fit": 82, "company_fit": 88, "authenticity": 85 },
          "created_at": "2026-03-02T14:00:00+09:00"
        }
      ]
    }
  ],
  "analyses": [
    {
      "company_name": "삼성전자",
      "position": "마케팅",
      "overall_fit_score": 78,
      "analysis_result": { "core_values": [], "talent_profile": [], "strategy_keywords": [] },
      "created_at": "2026-03-01T09:00:00+09:00"
    }
  ],
  "applications": [
    {
      "company_name": "삼성전자",
      "position": "마케팅",
      "status": "preparing",
      "deadline": "2026-03-30",
      "created_at": "2026-03-01T09:00:00+09:00"
    }
  ]
}
```

#### PDF 포맷

사람이 읽기 쉬운 포맷으로 주요 데이터를 정리한다.

```
┌───────────────────────────────────────────────┐
│  Colight 데이터 내보내기                        │
│  내보내기 일시: 2026-03-15                      │
│  사용자: 취준생A                                │
├───────────────────────────────────────────────┤
│                                               │
│  1. 프로필 정보                                │
│     희망 직무: 마케팅                           │
│     희망 산업: IT                              │
│                                               │
│  2. 경험 목록 (3건)                            │
│     [경험 1] 마케팅 동아리 SNS 홍보             │
│       S: 신입 부원 모집이 필요한 상황...         │
│       T: SNS 홍보를 통해 지원자 50명 이상...     │
│       A: 인스타그램 릴스 콘텐츠 기획...          │
│       R: 지원자 83명 확보, 전년 대비 166%       │
│       주 무기: 문제해결 (92%)                   │
│       부 무기: 소통/설득 (78%)                  │
│     ...                                       │
│                                               │
│  3. 자소서 (2건)                               │
│     [삼성전자] 성장 과정 - 2개 버전              │
│     ...                                       │
│                                               │
│  4. 기업 분석 (1건)                            │
│     삼성전자 마케팅 - 적합도 78%                 │
│     ...                                       │
└───────────────────────────────────────────────┘
```

**PDF 생성**: Go 백엔드에서 `go-wkhtmltopdf` 또는 `chromedp`로 HTML → PDF 변환. River background job으로 비동기 처리.

---

## 2. 내보내기 구현

### 2.1 API 엔드포인트

| Method | Path | 설명 |
|--------|------|------|
| `POST` | `/v1/export/request` | 내보내기 요청 (포맷: json/pdf/both) |
| `GET` | `/v1/export/status` | 내보내기 상태 조회 |
| `GET` | `/v1/export/download/:token` | 다운로드 (시간 제한 토큰) |

### 2.2 Go 서비스 패턴

```go
// apps/backend/internal/service/export.go
package service

import (
    "context"
    "log/slog"
    "time"

    "github.com/google/uuid"
    "github.com/riverqueue/river"
)

type ExportFormat string

const (
    ExportJSON ExportFormat = "json"
    ExportPDF  ExportFormat = "pdf"
    ExportBoth ExportFormat = "both"
)

type ExportService struct {
    logger *slog.Logger
    db     *ent.Client
    river  *river.Client[pgx.Tx]
}

func (s *ExportService) RequestExport(ctx context.Context, userID uuid.UUID, format ExportFormat) (*ExportRequest, error) {
    // 1. Rate limit check: max 3 requests per month
    count, err := s.db.DataExport.
        Query().
        Where(
            dataexport.UserIDEQ(userID),
            dataexport.CreatedAtGTE(time.Now().AddDate(0, -1, 0)),
        ).
        Count(ctx)
    if err != nil {
        return nil, err
    }
    if count >= 3 {
        return nil, ErrExportRateLimited
    }

    // 2. Create export request record
    export, err := s.db.DataExport.Create().
        SetUserID(userID).
        SetFormat(string(format)).
        SetStatus("pending").
        Save(ctx)
    if err != nil {
        return nil, err
    }

    // 3. Enqueue River background job
    _, err = s.river.Insert(ctx, ExportDataArgs{
        ExportID: export.ID,
        UserID:   userID,
        Format:   format,
    }, nil)
    if err != nil {
        return nil, err
    }

    return export, nil
}
```

### 2.3 River Worker: 대용량 내보내기

```go
// apps/backend/internal/worker/export_data.go
package worker

import (
    "context"
    "encoding/json"
    "log/slog"

    "github.com/riverqueue/river"
)

type ExportDataArgs struct {
    ExportID uuid.UUID    `json:"export_id"`
    UserID   uuid.UUID    `json:"user_id"`
    Format   ExportFormat `json:"format"`
}

func (ExportDataArgs) Kind() string { return "export_user_data" }

type ExportDataWorker struct {
    river.WorkerDefaults[ExportDataArgs]
    logger  *slog.Logger
    db      *ent.Client
    storage FileStorage // S3 or local temp storage
}

func (w *ExportDataWorker) Work(ctx context.Context, job *river.Job[ExportDataArgs]) error {
    w.logger.Info("starting data export",
        slog.String("export_id", job.Args.ExportID.String()),
        slog.String("user_id", job.Args.UserID.String()),
    )

    // 1. Collect all user data
    data, err := w.collectUserData(ctx, job.Args.UserID)
    if err != nil {
        return w.markFailed(ctx, job.Args.ExportID, err)
    }

    // 2. Generate files based on format
    var files []ExportFile

    if job.Args.Format == ExportJSON || job.Args.Format == ExportBoth {
        jsonBytes, err := json.MarshalIndent(data, "", "  ")
        if err != nil {
            return w.markFailed(ctx, job.Args.ExportID, err)
        }
        files = append(files, ExportFile{
            Name:    "colight-export.json",
            Content: jsonBytes,
        })
    }

    if job.Args.Format == ExportPDF || job.Args.Format == ExportBoth {
        pdfBytes, err := w.generatePDF(data)
        if err != nil {
            w.logger.Warn("PDF generation failed, JSON-only fallback",
                slog.String("error", err.Error()),
            )
            // PDF generation failure is non-fatal
        } else {
            files = append(files, ExportFile{
                Name:    "colight-export.pdf",
                Content: pdfBytes,
            })
        }
    }

    // 3. Upload to temporary storage
    downloadToken := uuid.New().String()
    for _, f := range files {
        if err := w.storage.Upload(ctx, downloadToken, f.Name, f.Content); err != nil {
            return w.markFailed(ctx, job.Args.ExportID, err)
        }
    }

    // 4. Update export record with download token
    _, err = w.db.DataExport.UpdateOneID(job.Args.ExportID).
        SetStatus("completed").
        SetDownloadToken(downloadToken).
        SetExpiresAt(time.Now().Add(7 * 24 * time.Hour)). // 7 days
        Save(ctx)
    if err != nil {
        return err
    }

    // 5. Send notification to user
    // NotificationService.Send(EXPORT_COMPLETE, ...)

    w.logger.Info("data export completed",
        slog.String("export_id", job.Args.ExportID.String()),
        slog.Int("file_count", len(files)),
    )
    return nil
}

func (w *ExportDataWorker) collectUserData(ctx context.Context, userID uuid.UUID) (*UserExportData, error) {
    // Query all user-owned data using Ent eager loading
    profile, err := w.db.UserProfile.
        Query().
        Where(userprofile.UserIDEQ(userID)).
        WithExperiences(func(q *ent.ExperienceQuery) {
            q.WithTags()
            q.WithWeapons()
        }).
        WithApplications().
        WithCompanyAnalyses().
        WithCoverLetters(func(q *ent.CoverLetterQuery) {
            q.WithVersions()
        }).
        WithCoachingSessions().
        Only(ctx)
    if err != nil {
        return nil, err
    }

    return mapToExportData(profile), nil
}
```

### 2.4 data_exports 테이블

```go
// apps/backend/ent/schema/dataexport.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/dialect/entsql"
    "entgo.io/ent/schema/edge"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
    "github.com/google/uuid"
)

type DataExport struct {
    ent.Schema
}

func (DataExport) Mixin() []ent.Mixin {
    return []ent.Mixin{
        BaseMixin{},
    }
}

func (DataExport) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("user_id", uuid.UUID{}).
            Comment("References auth.users(id)"),
        field.String("format").
            NotEmpty().
            MaxLen(10).
            Comment("Export format: json, pdf, both"),
        field.Enum("status").
            Values("pending", "processing", "completed", "failed").
            Default("pending").
            Comment("Export status"),
        field.String("download_token").
            Optional().
            MaxLen(100).
            Comment("One-time download token"),
        field.Time("expires_at").
            Optional().
            Nillable().
            Comment("Download link expiration"),
        field.Text("error_message").
            Optional().
            Comment("Error message if failed"),
    }
}

func (DataExport) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("user", UserProfile.Type).
            Ref("data_exports").
            Unique().
            Required().
            Field("user_id").
            Annotations(entsql.OnDelete(entsql.Cascade)),
    }
}

func (DataExport) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("user_id", "created_at"),
        index.Fields("download_token"),
        index.Fields("expires_at"),
    }
}
```

---

## 3. 계정 삭제 시 데이터 처리

> `08-legal-privacy.md` 보관 및 삭제 정책 준수

### 3.1 삭제 흐름

```
[사용자: 계정 삭제 요청]
    │
    ▼
[확인 모달]
    "정말 계정을 삭제하시겠습니까?"
    "삭제 후 30일간 복구 가능하며, 이후 완전히 삭제됩니다."
    "저장된 경험, 자소서, 분석 결과가 모두 삭제됩니다."
    [취소] [삭제 진행]
    │
    ▼
[Soft Delete 실행]
    ├── user_profiles.deleted_at = NOW()
    ├── Supabase Auth: user disabled (비활성화)
    └── 확인 이메일 발송: "계정 삭제가 요청되었습니다"
    │
    ▼
[30일 유예 기간]
    ├── 로그인 시도 → "계정이 삭제 예정입니다. 복구하시겠습니까?" 안내
    ├── 복구 요청 → deleted_at = NULL, Supabase Auth 재활성화
    └── 데이터 접근 차단 (API 401 반환)
    │
    ▼
[D+30: Hard Delete (River cron job)]
    ├── 모든 사용자 데이터 CASCADE 삭제
    │     ├── experiences → experience_tags, experience_weapons, experience_usages
    │     ├── applications → cover_letters → cover_letter_versions
    │     ├── company_analyses
    │     ├── coaching_sessions
    │     ├── notifications, notification_settings
    │     └── data_exports
    │
    ├── 임베딩 벡터 삭제 (experiences.embedding)
    │
    ├── Supabase Auth: user 완전 삭제
    │
    └── 삭제 완료 로그 (감사 추적용, 비식별)
```

### 3.2 삭제 예외 데이터

| 데이터 | 보관 기간 | 이유 |
|--------|----------|------|
| **결제 이력** | 5년 | 전자상거래법 (세금계산서 보관 의무) |
| **접속 로그** | 3개월 | 통신비밀보호법 |
| **기업 분석 캐시** | 삭제 불필요 | 사용자 비식별 공유 캐시 (`company_analysis_cache`) |

### 3.3 River Cron Job: Hard Delete

```go
// apps/backend/internal/worker/hard_delete.go
package worker

type HardDeleteArgs struct{}

func (HardDeleteArgs) Kind() string { return "hard_delete_expired_accounts" }

type HardDeleteWorker struct {
    river.WorkerDefaults[HardDeleteArgs]
    logger *slog.Logger
    db     *ent.Client
}

func (w *HardDeleteWorker) Work(ctx context.Context, job *river.Job[HardDeleteArgs]) error {
    // Find users with deleted_at older than 30 days
    threshold := time.Now().AddDate(0, 0, -30)

    users, err := w.db.UserProfile.
        Query().
        Where(
            userprofile.DeletedAtNotNil(),
            userprofile.DeletedAtLTE(threshold),
        ).
        All(ctx)
    if err != nil {
        return err
    }

    for _, user := range users {
        w.logger.Info("hard deleting user account",
            slog.String("user_id", user.UserID.String()),
            slog.Time("deleted_at", *user.DeletedAt),
        )

        // CASCADE delete handles all related data
        err := w.db.UserProfile.DeleteOneID(user.ID).Exec(ctx)
        if err != nil {
            w.logger.Error("hard delete failed",
                slog.String("user_id", user.UserID.String()),
                slog.String("error", err.Error()),
            )
            continue // Don't fail entire batch
        }

        // Delete from Supabase Auth (admin API)
        // supabaseAdmin.DeleteUser(user.UserID)
    }

    w.logger.Info("hard delete batch completed",
        slog.Int("deleted_count", len(users)),
    )
    return nil
}
```

```go
// River scheduler (cmd/api/main.go)
// Run daily at 03:00 KST (18:00 UTC previous day)
river.NewPeriodicJob(
    river.PeriodicInterval(24*time.Hour),
    func() (river.JobArgs, *river.InsertOpts) {
        return HardDeleteArgs{}, nil
    },
    &river.PeriodicJobOpts{RunOnStart: false},
),
```

---

## 4. DB 백업 전략

### 4.1 Supabase 자동 백업

| 항목 | Free Tier | Pro Tier |
|------|----------|---------|
| **백업 방식** | 일간 백업 | Point-in-Time Recovery (PITR) |
| **보관 기간** | 7일 | 7일 (최대 30일) |
| **복구 단위** | 일간 스냅샷 | 초 단위 (WAL 기반) |
| **자동 여부** | 자동 | 자동 |

> Supabase Free Tier에서는 **일간 자동 백업만** 제공된다. MVP 단계에서는 이것으로 충분하나, 유료 전환 시 PITR 활성화를 권장한다.

### 4.2 추가 백업: pg_dump 스케줄

Supabase 자동 백업 외에 별도의 `pg_dump` 백업을 River cron job으로 실행한다.

```go
// apps/backend/internal/worker/db_backup.go
package worker

type DBBackupArgs struct{}

func (DBBackupArgs) Kind() string { return "db_backup" }

type DBBackupWorker struct {
    river.WorkerDefaults[DBBackupArgs]
    logger     *slog.Logger
    dbURL      string
    storage    FileStorage // S3-compatible or local
}

func (w *DBBackupWorker) Work(ctx context.Context, job *river.Job[DBBackupArgs]) error {
    timestamp := time.Now().Format("20060102-150405")
    filename := fmt.Sprintf("colight-backup-%s.sql.gz", timestamp)

    w.logger.Info("starting database backup", slog.String("filename", filename))

    // Execute pg_dump and compress
    // Note: pg_dump must be available in the Docker image
    cmd := exec.CommandContext(ctx,
        "pg_dump",
        "--format=custom",
        "--compress=9",
        "--no-owner",
        "--no-acl",
        w.dbURL,
    )

    output, err := cmd.Output()
    if err != nil {
        w.logger.Error("pg_dump failed", slog.String("error", err.Error()))
        return err
    }

    // Upload to storage
    if err := w.storage.Upload(ctx, "backups/"+filename, filename, output); err != nil {
        return err
    }

    // Cleanup old backups (retention policy)
    if err := w.cleanupOldBackups(ctx); err != nil {
        w.logger.Warn("backup cleanup failed", slog.String("error", err.Error()))
    }

    w.logger.Info("database backup completed",
        slog.String("filename", filename),
        slog.Int("size_bytes", len(output)),
    )
    return nil
}
```

### 4.3 백업 보관 정책

```
┌─────────────────────────────────────────────┐
│  백업 보관 정책 (Grandfather-Father-Son)       │
├─────────────────────────────────────────────┤
│                                             │
│  일간 백업 (Daily):   7개 보관 (최근 7일)      │
│  주간 백업 (Weekly):  4개 보관 (최근 4주)      │
│  월간 백업 (Monthly): 3개 보관 (최근 3개월)     │
│                                             │
│  실행 스케줄:                                 │
│  ├── 매일 04:00 KST: 일간 백업              │
│  ├── 매주 일요일 04:00 KST: 주간 백업 태그    │
│  └── 매월 1일 04:00 KST: 월간 백업 태그      │
│                                             │
│  총 저장 용량 예상: ~100MB (초기)              │
│  (사용자 1,000명 기준, 압축 적용)              │
└─────────────────────────────────────────────┘
```

### 4.4 백업 정리 로직

```go
func (w *DBBackupWorker) cleanupOldBackups(ctx context.Context) error {
    files, err := w.storage.List(ctx, "backups/")
    if err != nil {
        return err
    }

    // Sort by timestamp (filename contains timestamp)
    sort.Slice(files, func(i, j int) bool {
        return files[i].Name > files[j].Name // newest first
    })

    // Retention: keep 7 daily + 4 weekly + 3 monthly
    // Daily: keep latest 7
    // Weekly (Sunday): keep latest 4 beyond daily retention
    // Monthly (1st): keep latest 3 beyond weekly retention

    for _, f := range filesToDelete(files) {
        if err := w.storage.Delete(ctx, f.Key); err != nil {
            w.logger.Warn("failed to delete old backup",
                slog.String("file", f.Name),
                slog.String("error", err.Error()),
            )
        }
    }

    return nil
}
```

---

## 5. 벡터 데이터 (pgvector) 백업 고려사항

### 5.1 임베딩 벡터 특성

| 항목 | 값 |
|------|---|
| **차원** | 1536 (text-embedding-3-small) |
| **저장 크기** | ~6KB/벡터 (float32 x 1536) |
| **사용자당 예상** | 3~30 경험 = 18KB~180KB |
| **1,000 사용자 기준** | ~18MB~180MB |

### 5.2 백업 시 포함

`pg_dump`는 pgvector `VECTOR` 타입을 포함하여 백업한다. 별도 처리 불필요.

```sql
-- pg_dump 복원 시 pgvector extension 먼저 활성화 필요
CREATE EXTENSION IF NOT EXISTS vector;

-- 이후 pg_restore 정상 동작
pg_restore --dbname=colight_restore backup.dump
```

### 5.3 벡터 재생성 전략

임베딩 벡터는 **원본 텍스트(경험 내용)에서 재생성 가능**하다. 재해 복구 시 벡터가 손실되더라도 복원 가능:

```
[벡터 손실 발생]
    │
    ├── 1순위: pg_dump 백업에서 복원 (벡터 포함)
    │
    └── 2순위: 원본 텍스트에서 재생성 (River batch job)
          │
          ├── experiences 테이블에서 embedding IS NULL 인 레코드 조회
          ├── 각 경험의 content → text-embedding-3-small → embedding 업데이트
          └── 비용: ~0.5원/건, 1,000건 = ~500원
```

---

## 6. 재해 복구 (Disaster Recovery)

### 6.1 RTO / RPO 목표

| 지표 | 목표 | 설명 |
|------|------|------|
| **RPO** (Recovery Point Objective) | 24시간 | 최대 24시간 분의 데이터 유실 허용 (일간 백업 기준) |
| **RTO** (Recovery Time Objective) | 4시간 | 장애 발생 후 4시간 이내 서비스 복구 |

> MVP 단계(무료 인프라)에서의 현실적 목표. 유료 Supabase Pro 전환 시 RPO를 수 분 단위로 줄일 수 있음 (PITR).

### 6.2 재해 복구 시나리오

| 시나리오 | 복구 방법 | 예상 RTO |
|----------|----------|---------|
| **Supabase DB 장애** | Supabase 자동 복구 대기 또는 백업 복원 | 1~4시간 |
| **Koyeb 백엔드 장애** | Koyeb 자동 재시작 또는 수동 재배포 | 5~30분 |
| **Vercel 프론트엔드 장애** | Vercel 자동 복구 (Edge Network redundancy) | 5분 미만 |
| **DB 데이터 오염/삭제** | pg_dump 백업에서 복원 | 2~4시간 |
| **임베딩 벡터 손실** | 원본 텍스트에서 재생성 (River batch) | 30분~2시간 |

### 6.3 복구 절차 (DB 데이터 오염 시)

```
[장애 감지]
    │
    ├── 1. 서비스 점검 모드 전환 (503 응답)
    │
    ├── 2. 최신 백업 확인
    │     ├── Supabase PITR (Pro) → 장애 직전 시점으로 복원
    │     └── pg_dump 백업 → 최신 일간 백업 복원
    │
    ├── 3. 새 Supabase 프로젝트에 백업 복원 (테스트)
    │
    ├── 4. 데이터 정합성 검증
    │     ├── 사용자 수, 경험 수, 자소서 수 비교
    │     └── 임베딩 벡터 존재 여부 확인
    │
    ├── 5. DNS / 환경변수 전환 (새 DB 가리키도록)
    │
    ├── 6. 임베딩 재생성 (필요 시, River batch job)
    │
    └── 7. 서비스 점검 모드 해제 + 사용자 공지
```

---

## 7. Phase별 구현 계획

| Phase | 구현 내용 |
|-------|----------|
| **Phase 6.1** | 계정 삭제 흐름 (soft delete + 30일 유예 + hard delete cron) |
| **Phase 6.2** | 데이터 내보내기 API (JSON 포맷) + 다운로드 UI |
| **Phase 9** | pg_dump 자동 백업 River cron job |
| **Phase 10** | PDF 내보내기, 백업 보관 정책 자동화, 재해 복구 매뉴얼 문서화 |
| **Supabase Pro 전환 시** | PITR 활성화, RPO 단축 |

---

## 8. 환경변수

| 변수명 | 설명 | 필요 Phase |
|--------|------|-----------|
| `BACKUP_STORAGE_URL` | 백업 파일 저장 위치 (S3-compatible) | Phase 9 |
| `BACKUP_STORAGE_KEY` | 스토리지 액세스 키 | Phase 9 |
| `BACKUP_STORAGE_SECRET` | 스토리지 시크릿 키 | Phase 9 |
| `EXPORT_TEMP_DIR` | 임시 내보내기 파일 디렉토리 | Phase 6.2 |
