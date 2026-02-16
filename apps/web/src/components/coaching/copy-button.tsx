"use client";

import { Copy } from "lucide-react";

interface CopyButtonProps {
  content: string;
  removeStarTags?: boolean;
}

export function CopyButton({ content, removeStarTags = false }: CopyButtonProps) {
  const handleCopy = async () => {
    let text = content;

    // Strip HTML tags
    const div = document.createElement("div");
    div.innerHTML = text;
    text = div.textContent || div.innerText || "";

    // Remove STAR tags if option enabled
    if (removeStarTags) {
      text = text
        .replace(/\[상황\]/g, "")
        .replace(/\[과제\]/g, "")
        .replace(/\[행동\]/g, "")
        .replace(/\[결과\]/g, "")
        .replace(/\n{3,}/g, "\n\n")
        .trim();
    }

    try {
      await navigator.clipboard.writeText(text);
      // Toast notification would go here
      // toast.success(`클립보드에 복사되었습니다 (${text.length}자)`);
    } catch (error) {
      console.error("Failed to copy:", error);
    }
  };

  return (
    <button
      onClick={handleCopy}
      className="flex items-center gap-2 rounded-lg border border-border bg-card px-4 py-2 text-sm font-medium text-foreground/80 hover:bg-white/[0.04]"
    >
      <Copy className="h-4 w-4" />
      📋 복사
    </button>
  );
}
