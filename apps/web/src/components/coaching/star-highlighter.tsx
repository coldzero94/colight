"use client";

interface StarHighlighterProps {
  text: string;
}

const STAR_TAGS = {
  "[상황]": "text-blue-700 bg-blue-50",
  "[과제]": "text-amber-700 bg-amber-50",
  "[행동]": "text-green-700 bg-green-50",
  "[결과]": "text-purple-700 bg-purple-50",
};

export function StarHighlighter({ text }: StarHighlighterProps) {
  // Split text by STAR tags and highlight them
  const parts: Array<{ text: string; isTag: boolean; className?: string }> = [];
  let remaining = text;

  while (remaining.length > 0) {
    let found = false;

    // Check for STAR tags
    for (const [tag, className] of Object.entries(STAR_TAGS)) {
      if (remaining.startsWith(tag)) {
        parts.push({ text: tag, isTag: true, className });
        remaining = remaining.slice(tag.length);
        found = true;
        break;
      }
    }

    if (!found) {
      // Find next tag position
      let nextTagIndex = remaining.length;
      for (const tag of Object.keys(STAR_TAGS)) {
        const index = remaining.indexOf(tag);
        if (index !== -1 && index < nextTagIndex) {
          nextTagIndex = index;
        }
      }

      // Add text until next tag
      if (nextTagIndex > 0) {
        parts.push({ text: remaining.slice(0, nextTagIndex), isTag: false });
        remaining = remaining.slice(nextTagIndex);
      } else {
        break;
      }
    }
  }

  return (
    <div className="whitespace-pre-wrap text-sm">
      {parts.map((part, index) =>
        part.isTag ? (
          <span
            key={index}
            className={`inline-block rounded px-2 py-1 font-semibold ${part.className}`}
          >
            {part.text}
          </span>
        ) : (
          <span key={index}>{part.text}</span>
        )
      )}
    </div>
  );
}
