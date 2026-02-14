"use client";

import type { Editor } from "@tiptap/react";
import { Bold, Italic, List, Heading2, Heading3 } from "lucide-react";

interface EditorToolbarProps {
  editor: Editor;
}

export function EditorToolbar({ editor }: EditorToolbarProps) {
  return (
    <div className="flex items-center gap-1 border-b border-gray-200 bg-gray-50 p-2">
      <button
        type="button"
        onClick={() => editor.chain().focus().toggleBold().run()}
        disabled={!editor.can().chain().focus().toggleBold().run()}
        className={`rounded p-2 hover:bg-gray-200 ${
          editor.isActive("bold") ? "bg-gray-300" : ""
        }`}
        aria-label="Bold"
        title="굵게 (Ctrl+B)"
      >
        <Bold className="h-4 w-4" />
      </button>

      <button
        type="button"
        onClick={() => editor.chain().focus().toggleItalic().run()}
        disabled={!editor.can().chain().focus().toggleItalic().run()}
        className={`rounded p-2 hover:bg-gray-200 ${
          editor.isActive("italic") ? "bg-gray-300" : ""
        }`}
        aria-label="Italic"
        title="기울임 (Ctrl+I)"
      >
        <Italic className="h-4 w-4" />
      </button>

      <div className="mx-1 h-6 w-px bg-gray-300" />

      <button
        type="button"
        onClick={() => editor.chain().focus().toggleBulletList().run()}
        className={`rounded p-2 hover:bg-gray-200 ${
          editor.isActive("bulletList") ? "bg-gray-300" : ""
        }`}
        aria-label="Bullet List"
        title="목록"
      >
        <List className="h-4 w-4" />
      </button>

      <div className="mx-1 h-6 w-px bg-gray-300" />

      <button
        type="button"
        onClick={() => editor.chain().focus().toggleHeading({ level: 2 }).run()}
        className={`rounded p-2 hover:bg-gray-200 ${
          editor.isActive("heading", { level: 2 }) ? "bg-gray-300" : ""
        }`}
        aria-label="Heading 2"
        title="제목 2"
      >
        <Heading2 className="h-4 w-4" />
      </button>

      <button
        type="button"
        onClick={() => editor.chain().focus().toggleHeading({ level: 3 }).run()}
        className={`rounded p-2 hover:bg-gray-200 ${
          editor.isActive("heading", { level: 3 }) ? "bg-gray-300" : ""
        }`}
        aria-label="Heading 3"
        title="제목 3"
      >
        <Heading3 className="h-4 w-4" />
      </button>
    </div>
  );
}
