"use client";

import { useState } from "react";
import { Mic, ArrowRight, CheckCircle2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { InterviewChat } from "@/components/interview/interview-chat";
import { InterviewProgress } from "@/components/interview/interview-progress";
import { InterviewTimer } from "@/components/interview/interview-timer";
import { useInterviewChat } from "@/hooks/use-interview-chat";

type PageStep = "welcome" | "interviewing" | "complete";

export default function InterviewPage() {
  const [step, setStep] = useState<PageStep>("welcome");
  const chat = useInterviewChat();

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

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      {/* Welcome */}
      {step === "welcome" && (
        <div className="rounded-lg border border-gray-200 bg-white p-8 text-center space-y-6">
          <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-gray-100">
            <Mic className="h-8 w-8 text-gray-600" />
          </div>
          <div className="space-y-2">
            <h2 className="text-xl font-semibold text-gray-900">
              경험 인터뷰
            </h2>
            <p className="text-sm text-gray-500">
              AI가 5단계 질문을 통해 당신의 경험을 깊이 있게 탐색합니다.
              <br />
              답변을 바탕으로 STAR 구조의 경험 카드를 자동으로 생성합니다.
            </p>
          </div>
          <div className="flex flex-wrap justify-center gap-2 text-xs text-gray-400">
            <span className="rounded-full bg-gray-50 px-3 py-1">가볍게</span>
            <ArrowRight className="h-4 w-4 self-center" />
            <span className="rounded-full bg-gray-50 px-3 py-1">
              기억에 남는 순간
            </span>
            <ArrowRight className="h-4 w-4 self-center" />
            <span className="rounded-full bg-gray-50 px-3 py-1">
              어려웠던 점
            </span>
            <ArrowRight className="h-4 w-4 self-center" />
            <span className="rounded-full bg-gray-50 px-3 py-1">해결법</span>
            <ArrowRight className="h-4 w-4 self-center" />
            <span className="rounded-full bg-gray-50 px-3 py-1">
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
            <div className="rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-600">
              {chat.error}
            </div>
          )}
        </div>
      )}

      {/* Complete */}
      {step === "complete" && (
        <div className="rounded-lg border border-gray-200 bg-white p-8 text-center space-y-6">
          <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-green-50">
            <CheckCircle2 className="h-8 w-8 text-green-600" />
          </div>
          <div className="space-y-2">
            <h2 className="text-xl font-semibold text-gray-900">
              인터뷰 완료!
            </h2>
            <p className="text-sm text-gray-500">
              {Math.floor(chat.elapsedSeconds / 60)}분{" "}
              {chat.elapsedSeconds % 60}초 동안 인터뷰를 진행했습니다.
              <br />
              답변을 바탕으로 STAR 구조의 경험 카드를 생성할 수 있습니다.
            </p>
          </div>
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
            <Button className="gap-2">
              경험 카드 생성
              <ArrowRight className="h-4 w-4" />
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
