package crawler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWantedExtractor_NextData_Legacy(t *testing.T) {
	html := `<html><head></head><body>
		<script id="__NEXT_DATA__" type="application/json">
		{
			"props": {
				"pageProps": {
					"job": {
						"detail": {
							"position": "백엔드 개발자",
							"company": {
								"name": "원티드랩"
							},
							"intro": "서비스 API 설계 및 개발\nMSA 아키텍처 설계\n데이터 파이프라인 구축",
							"requirements": "Go 또는 Java 3년 이상 경력\nPostgreSQL 경험\nRESTful API 설계 경험",
							"preferred": "Kubernetes 운영 경험\ngRPC 사용 경험\nCI/CD 파이프라인 구축 경험",
							"skill_tags": [
								{"title": "Go"},
								{"title": "PostgreSQL"},
								{"title": "Docker"}
							]
						}
					}
				}
			}
		}
		</script>
		<div id="__next"></div>
	</body></html>`

	extractor := NewWantedExtractor()
	result, err := extractor.ExtractFromNextData(html)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "원티드랩", result.CompanyName)
	assert.Equal(t, "백엔드 개발자", result.Position)
	assert.Contains(t, result.MainTasks, "서비스 API 설계")
	assert.Contains(t, result.MainTasks, "MSA 아키텍처")
	assert.Contains(t, result.Requirements, "Go 또는 Java")
	assert.Contains(t, result.Preferred, "Kubernetes 운영")
	assert.Contains(t, result.Skills, "Go")
	assert.Contains(t, result.Skills, "PostgreSQL")
	assert.Equal(t, "wanted", result.Source)
}

func TestWantedExtractor_NextData_InitialData(t *testing.T) {
	html := `<html><head></head><body>
		<script id="__NEXT_DATA__" type="application/json">
		{
			"props": {
				"pageProps": {
					"initialData": {
						"position": "프론트엔드 엔지니어",
						"company": {
							"company_name": "카카오"
						},
						"main_tasks": "React/Next.js 기반 서비스 개발\n디자인 시스템 구축 및 운영\n성능 최적화",
						"intro": "카카오는 기술로 세상을 바꿉니다",
						"requirements": "React 3년 이상 경력\nTypeScript 숙련",
						"preferred_points": "Next.js 경험\n디자인 시스템 구축 경험",
						"benefits": "자율출퇴근\n점심식사 제공",
						"category_tag": {
							"child_tags": [
								{"text": "React"},
								{"text": "TypeScript"},
								{"text": "Next.js"}
							]
						},
						"address": {
							"full_location": "경기 성남시 분당구 판교역로"
						}
					}
				}
			}
		}
		</script>
	</body></html>`

	extractor := NewWantedExtractor()
	result, err := extractor.ExtractFromNextData(html)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "카카오", result.CompanyName)
	assert.Equal(t, "프론트엔드 엔지니어", result.Position)
	assert.Contains(t, result.MainTasks, "React/Next.js 기반 서비스 개발")
	assert.Contains(t, result.MainTasks, "성능 최적화")
	assert.Contains(t, result.Requirements, "React 3년 이상")
	assert.Contains(t, result.Preferred, "Next.js 경험")
	assert.Contains(t, result.Preferred, "디자인 시스템")
	assert.Equal(t, "경기 성남시 분당구 판교역로", result.Location)
	assert.Contains(t, result.Skills, "React")
	assert.Contains(t, result.Skills, "TypeScript")
	assert.Contains(t, result.Skills, "Next.js")
	assert.Equal(t, "wanted", result.Source)
}

func TestWantedExtractor_FullContent(t *testing.T) {
	// Verify that full text is preserved, not truncated
	// JSON uses escaped \n for newlines within string values
	html := `<html><head></head><body>
		<script id="__NEXT_DATA__" type="application/json">
		{
			"props": {
				"pageProps": {
					"job": {
						"detail": {
							"position": "시니어 개발자",
							"company": {"name": "테스트"},
							"intro": "1. 대규모 트래픽 서비스 백엔드 API 설계 및 개발\n2. MSA 기반 시스템 아키텍처 설계 및 마이그레이션\n3. 데이터 파이프라인 구축 및 최적화\n4. 코드 리뷰 및 기술 멘토링\n5. 서비스 모니터링 및 장애 대응",
							"requirements": "긴 자격요건 텍스트",
							"preferred": "긴 우대사항 텍스트",
							"skill_tags": []
						}
					}
				}
			}
		}
		</script>
	</body></html>`

	extractor := NewWantedExtractor()
	result, err := extractor.ExtractFromNextData(html)

	require.NoError(t, err)
	require.NotNil(t, result)

	// FULL content should be preserved
	assert.Contains(t, result.MainTasks, "대규모 트래픽")
	assert.Contains(t, result.MainTasks, "서비스 모니터링 및 장애 대응")
	assert.Contains(t, result.MainTasks, "코드 리뷰 및 기술 멘토링")
}

func TestWantedExtractor_NoNextData(t *testing.T) {
	html := `<html><head></head><body><div>No Next.js data here</div></body></html>`

	extractor := NewWantedExtractor()
	result, err := extractor.ExtractFromNextData(html)

	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestWantedExtractor_MalformedJSON(t *testing.T) {
	html := `<html><head></head><body>
		<script id="__NEXT_DATA__" type="application/json">
		{invalid json content!!!
		</script>
	</body></html>`

	extractor := NewWantedExtractor()
	result, err := extractor.ExtractFromNextData(html)

	// Should not crash — graceful nil return
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestWantedExtractor_MissingJobField(t *testing.T) {
	html := `<html><head></head><body>
		<script id="__NEXT_DATA__" type="application/json">
		{
			"props": {
				"pageProps": {
					"someOtherData": {}
				}
			}
		}
		</script>
	</body></html>`

	extractor := NewWantedExtractor()
	result, err := extractor.ExtractFromNextData(html)

	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestWantedExtractor_InitialDataPreferredOverLegacy(t *testing.T) {
	// When both initialData and job.detail exist, initialData should be preferred
	html := `<html><head></head><body>
		<script id="__NEXT_DATA__" type="application/json">
		{
			"props": {
				"pageProps": {
					"initialData": {
						"position": "NEW FORMAT",
						"company": {"company_name": "NewCo"}
					},
					"job": {
						"detail": {
							"position": "OLD FORMAT",
							"company": {"name": "OldCo"}
						}
					}
				}
			}
		}
		</script>
	</body></html>`

	extractor := NewWantedExtractor()
	result, err := extractor.ExtractFromNextData(html)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "NewCo", result.CompanyName)
	assert.Equal(t, "NEW FORMAT", result.Position)
}
