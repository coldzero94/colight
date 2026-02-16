"use client";

import { useState } from "react";
import { Mic, ArrowRight, CheckCircle2, Loader2 } from "lucide-react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { InterviewChat } from "@/components/interview/interview-chat";
import { InterviewProgress } from "@/components/interview/interview-progress";
import { InterviewTimer } from "@/components/interview/interview-timer";
import { STARPreview } from "@/components/interview/star-preview";
import { useInterviewChat } from "@/hooks/use-interview-chat";
import {
  extractSTAR,
  saveInterviewExperience,
  type ExtractSTARResult,
} from "@/lib/api/interview";

type PageStep =
  | "welcome"
  | "interviewing"
  | "complete"
  | "extracting"
  | "preview"
  | "saved";

export default function InterviewPage() {
  const [step, setStep] = useState<PageStep>("welcome");
  const chat = useInterviewChat();
  const [starData, setStarData] = useState<ExtractSTARResult | null>(null);
  const [isSaving, setIsSaving] = useState(false);
  const [savedId, setSavedId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const handleStart = async () => {
    setStep("interviewing");
    await chat.startInterview();
  };

  const handleSend = async (content: string) => {
    const completed = await chat.sendMessage(content);
    if (completed) {
      setStep("complete");
    }
  };

  const handleExtract = async () => {
    setStep("extracting");
    setError(null);
    try {
      const result = await extractSTAR(chat.messages);
      setStarData(result);
      setStep("preview");
    } catch (e) {
      setError(
        e instanceof Error ? e.message : "경험 추출에 실패했습니다"
      );
      setStep("complete");
    }
  };

  const handleSave = async (data: ExtractSTARResult) => {
    setIsSaving(true);
    setError(null);
    try {
      const result = await saveInterviewExperience(data);
      setSavedId(result.experience_id);
      setStep("saved");
    } catch (e) {
      setError(
        e instanceof Error ? e.message : "경험 저장에 실패했습니다"
      );
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      {/* Welcome */}
      {step === "welcome" && (
        <div className="rounded-lg border border-border bg-card p-8 text-center space-y-6">
          <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-white/[0.06]">
            <Mic className="h-8 w-8 text-muted-foreground" />
          </div>
          <div className="space-y-2">
            <h2 className="text-xl font-semibold text-foreground">
              경험 인터뷰
            </h2>
            <p className="text-sm text-muted-foreground">
              AI가 5단계 질문을 통해 당신의 경험을 깊이 있게 탐색합니다.
              <br />
              답변을 바탕으로 STAR 구조의 경험 카드를 자동으로 생성합니다.
            </p>
          </div>
          <div className="flex flex-wrap justify-center gap-2 text-xs text-muted-foreground/60">
            <span className="rounded-full bg-white/[0.04] px-3 py-1">가볍게</span>
            <ArrowRight className="h-4 w-4 self-center" />
            <span className="rounded-full bg-white/[0.04] px-3 py-1">
              기억에 남는 순간
            </span>
            <ArrowRight className="h-4 w-4 self-center" />
            <span className="rounded-full bg-white/[0.04] px-3 py-1">
              어려웠던 점
            </span>
            <ArrowRight className="h-4 w-4 self-center" />
            <span className="rounded-full bg-white/[0.04] px-3 py-1">해결법</span>
            <ArrowRight className="h-4 w-4 self-center" />
            <span className="rounded-full bg-white/[0.04] px-3 py-1">
              결과/배운 점
            </span>
          </div>
          <Button onClick={handleStart} className="gap-2">
            <Mic className="h-4 w-4" />
            인터뷰 시작하기
          </Button>
        </div>
      )}

      {/* Interviewing */}
      {step === "interviewing" && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <InterviewProgress
              currentStage={chat.stage}
              isComplete={chat.isComplete}
            />
            <InterviewTimer seconds={chat.elapsedSeconds} />
          </div>
          <InterviewChat
            messages={chat.messages}
            isLoading={chat.isLoading}
            onSend={handleSend}
            disabled={chat.isComplete}
          />
          {chat.error && (
            <div className="rounded-lg border border-red-500/20 bg-red-500/10 p-3 text-sm text-red-400">
              {chat.error}
            </div>
          )}
        </div>
      )}

      {/* Complete */}
      {step === "complete" && (
        <div className="rounded-lg border border-border bg-card p-8 text-center space-y-6">
          <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-primary/10">
            <CheckCircle2 className="h-8 w-8 text-primary" />
          </div>
          <div className="space-y-2">
            <h2 className="text-xl font-semibold text-foreground">
              인터뷰 완료!
            </h2>
            <p className="text-sm text-muted-foreground">
              {Math.floor(chat.elapsedSeconds / 60)}분{" "}
              {chat.elapsedSeconds % 60}초 동안 인터뷰를 진행했습니다.
              <br />
              답변을 바탕으로 STAR 구조의 경험 카드를 생성할 수 있습니다.
            </p>
          </div>
          {error && (
            <div className="rounded-lg border border-red-500/20 bg-red-500/10 p-3 text-sm text-red-400">
              {error}
            </div>
          )}
          <div className="flex justify-center gap-3">
            <Button
              variant="outline"
              onClick={() => {
                chat.reset();
                setStep("welcome");
              }}
            >
              다시 시작하기
            </Button>
            <Button onClick={handleExtract} className="gap-2">
              경험 카드 생성
              <ArrowRight className="h-4 w-4" />
            </Button>
          </div>
        </div>
      )}

      {/* Extracting */}
      {step === "extracting" && (
        <div className="rounded-lg border border-border bg-card p-8 text-center space-y-4">
          <Loader2 className="mx-auto h-8 w-8 animate-spin text-muted-foreground/60" />
          <p className="text-sm text-muted-foreground">
            STAR 구조로 경험을 추출하고 있습니다...
          </p>
        </div>
      )}

      {/* Preview */}
      {step === "preview" && starData && (
        <STARPreview
          data={starData}
          isSaving={isSaving}
          onSave={handleSave}
          onReExtract={handleExtract}
        />
      )}

      {/* Saved */}
      {step === "saved" && (
        <div className="rounded-lg border border-border bg-card p-8 text-center space-y-6">
          <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-primary/10">
            <CheckCircle2 className="h-8 w-8 text-primary" />
          </div>
          <div className="space-y-2">
            <h2 className="text-xl font-semibold text-foreground">
              경험 카드가 저장되었습니다!
            </h2>
            <p className="text-sm text-muted-foreground">
              경험 관리 페이지에서 확인할 수 있습니다.
            </p>
          </div>
          <div className="flex justify-center gap-3">
            <Button
              variant="outline"
              onClick={() => {
                chat.reset();
                setStarData(null);
                setSavedId(null);
                setStep("welcome");
              }}
            >
              새 인터뷰 시작
            </Button>
            <Link href={savedId ? `/experiences/${savedId}` : "/experiences"}>
              <Button className="gap-2">
                경험 카드 보기
                <ArrowRight className="h-4 w-4" />
              </Button>
            </Link>
          </div>
        </div>
      )}
    </div>
  );
}
