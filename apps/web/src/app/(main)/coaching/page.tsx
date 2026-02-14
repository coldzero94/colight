import { QuestionInputForm } from "@/components/coaching/question-input-form";

export default async function CoachingPage() {
  // TODO: Fetch user's analysis history from Go backend
  // For now, use empty array to show empty state
  const companies: any[] = [];

  const handleSubmit = async (data: any) => {
    "use server";
    // TODO: Call Go backend API for question analysis
    console.log("Question analysis request:", data);
  };

  return (
    <div className="container mx-auto max-w-3xl px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">AI 자소서 코칭</h1>
        <p className="mt-2 text-gray-600">
          자소서 문항을 분석하고, 적합한 경험을 추천받으세요
        </p>
      </div>

      <QuestionInputForm companies={companies} onSubmit={handleSubmit} />
    </div>
  );
}
