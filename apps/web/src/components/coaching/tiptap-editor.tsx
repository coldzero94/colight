"use client";

import { useEditor, EditorContent } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import CharacterCount from "@tiptap/extension-character-count";
import { StarTagHighlight } from "@/lib/tiptap/star-tag-highlight";
import { EditorToolbar } from "./editor-toolbar";
import { CharCounter } from "./char-counter";

interface TiptapEditorProps {
  content: string;
  charLimit: number;
  onChange: (content: string) => void;
  editable?: boolean;
}

export function TiptapEditor({
  content,
  charLimit,
  onChange,
  editable = true,
}: TiptapEditorProps) {
  const editor = useEditor({
    extensions: [
      StarterKit,
      CharacterCount.configure({
        limit: charLimit,
      }),
      StarTagHighlight,
    ],
    content,
    editable,
    onUpdate: ({ editor }) => {
      onChange(editor.getText());
    },
    editorProps: {
      attributes: {
        class: "prose prose-sm max-w-none p-4 min-h-[400px] focus:outline-none tiptap-editor",
        role: "textbox",
      },
    },
  });

  if (!editor) return null;

  const charCount = editor.storage.characterCount.characters();

  return (
    <div className="rounded-lg border border-gray-200">
      {editable && <EditorToolbar editor={editor} />}
      <EditorContent editor={editor} />
      <CharCounter current={charCount} limit={charLimit} />
    </div>
  );
}
