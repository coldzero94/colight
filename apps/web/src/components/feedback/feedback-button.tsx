"use client";

import { useState } from "react";
import { MessageSquarePlus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { FeedbackModal } from "./feedback-modal";

export function FeedbackButton() {
  const [open, setOpen] = useState(false);

  return (
    <>
      <Button
        onClick={() => setOpen(true)}
        size="icon"
        className="fixed bottom-6 right-6 z-50 h-12 w-12 rounded-full border border-primary/35 bg-gradient-to-r from-primary to-cyan-400 shadow-lg shadow-primary/30 hover:shadow-xl hover:shadow-primary/35"
        aria-label="피드백 보내기"
      >
        <MessageSquarePlus className="h-5 w-5" />
      </Button>
      <FeedbackModal open={open} onOpenChange={setOpen} />
    </>
  );
}
