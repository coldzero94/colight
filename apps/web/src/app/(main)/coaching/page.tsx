import { CoachingFlow } from "@/components/coaching/coaching-flow";

export default function CoachingPage() {
  return (
    <div className="container mx-auto max-w-3xl px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-foreground">AI 자소서 코칭</h1>
        <p className="mt-2 text-muted-foreground">
          자소서 문항을 분석하고, 적합한 경험을 추천받으세요
        </p>
      </div>

      <CoachingFlow />
    </div>
  );
}
