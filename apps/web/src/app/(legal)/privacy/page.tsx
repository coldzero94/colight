import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "개인정보처리방침 - Colight",
};

export default function PrivacyPage() {
  return (
    <article className="prose prose-gray max-w-none">
      <h1>개인정보처리방침</h1>
      <p className="text-sm text-gray-500">시행일: 2026년 2월 15일</p>

      <p>
        Colight(이하 &quot;서비스&quot;)는 이용자의 개인정보를 중요시하며,
        「개인정보 보호법」 등 관련 법령을 준수합니다. 본 개인정보처리방침은
        서비스가 수집하는 개인정보의 항목, 수집 및 이용 목적, 보유 기간 등을
        안내합니다.
      </p>

      <h2>1. 개인정보의 처리 목적</h2>
      <p>서비스는 다음 목적을 위해 개인정보를 처리합니다.</p>
      <ul>
        <li>회원 가입 및 관리: 본인 확인, 서비스 제공, 고지·안내</li>
        <li>서비스 제공: AI 기반 자기소개서 코칭, 기업 분석, 경험 매칭</li>
        <li>서비스 개선: 이용 통계 분석, 서비스 품질 향상</li>
      </ul>

      <h2>2. 수집하는 개인정보 항목</h2>
      <table>
        <thead>
          <tr>
            <th>구분</th>
            <th>항목</th>
            <th>수집 방법</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>필수</td>
            <td>이메일, 비밀번호(암호화 저장)</td>
            <td>회원가입</td>
          </tr>
          <tr>
            <td>선택</td>
            <td>닉네임</td>
            <td>회원가입</td>
          </tr>
          <tr>
            <td>소셜 로그인</td>
            <td>네이버 프로필(이메일, 닉네임)</td>
            <td>네이버 OAuth</td>
          </tr>
          <tr>
            <td>자동 수집</td>
            <td>서비스 이용 기록, 접속 로그</td>
            <td>서비스 이용 시</td>
          </tr>
        </tbody>
      </table>

      <h2>3. 개인정보의 보유 및 이용 기간</h2>
      <p>
        회원 탈퇴 시까지 보유하며, 탈퇴 즉시 지체 없이 파기합니다. 단, 관련
        법령에 의해 보존이 필요한 경우 해당 기간 동안 보관합니다.
      </p>
      <ul>
        <li>전자상거래 등 소비자 보호 법률: 계약·결제 기록 5년</li>
        <li>통신비밀보호법: 접속 로그 3개월</li>
      </ul>

      <h2>4. 개인정보의 제3자 제공</h2>
      <p>
        서비스는 이용자의 동의 없이 개인정보를 제3자에게 제공하지 않습니다.
        다만, AI 기능 제공을 위해 다음과 같이 데이터를 처리합니다.
      </p>
      <ul>
        <li>
          <strong>Anthropic (Claude API)</strong>: 자기소개서 분석·코칭을 위해
          이용자가 입력한 텍스트를 AI 모델에 전달. 개인 식별 정보는 전달하지
          않음.
        </li>
        <li>
          <strong>Google (Gemini API)</strong>: 채용공고 파싱, 무기 태깅 등
          경량 AI 처리. 개인 식별 정보는 전달하지 않음.
        </li>
        <li>
          <strong>OpenAI (Embeddings API)</strong>: 텍스트 임베딩 생성.
          개인 식별 정보는 전달하지 않음.
        </li>
      </ul>

      <h2>5. 개인정보 처리 위탁</h2>
      <table>
        <thead>
          <tr>
            <th>수탁자</th>
            <th>위탁 업무</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>Supabase Inc.</td>
            <td>데이터베이스 호스팅 및 관리</td>
          </tr>
          <tr>
            <td>Vercel Inc.</td>
            <td>웹 애플리케이션 호스팅</td>
          </tr>
          <tr>
            <td>Koyeb SAS</td>
            <td>백엔드 서버 호스팅</td>
          </tr>
        </tbody>
      </table>

      <h2>6. 이용자의 권리와 행사 방법</h2>
      <p>이용자는 언제든지 다음 권리를 행사할 수 있습니다.</p>
      <ul>
        <li>개인정보 열람 요구</li>
        <li>오류 등의 정정 요구</li>
        <li>삭제 요구</li>
        <li>처리 정지 요구</li>
      </ul>
      <p>
        권리 행사는 서비스 내 설정 페이지 또는 아래 연락처를 통해 가능합니다.
      </p>

      <h2>7. 개인정보의 파기</h2>
      <p>
        보유 기간 경과 또는 처리 목적 달성 시, 전자적 파일은 복구 불가능한
        방법으로 삭제하고, 인쇄물은 분쇄 또는 소각하여 파기합니다.
      </p>

      <h2>8. 개인정보보호 책임자</h2>
      <ul>
        <li>담당자: Colight 운영팀</li>
        <li>이메일: privacy@colight.app</li>
      </ul>

      <h2>9. 고지의 의무</h2>
      <p>
        본 개인정보처리방침은 2026년 2월 15일부터 적용되며, 변경 시 서비스
        공지사항을 통해 안내합니다.
      </p>
    </article>
  );
}
