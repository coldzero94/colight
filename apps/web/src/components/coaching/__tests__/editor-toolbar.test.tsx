import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { EditorToolbar } from "../editor-toolbar";
import type { Editor } from "@tiptap/react";

const createMockEditor = (activeStates: Record<string, boolean> = {}): Editor => {
  const chain = {
    focus: vi.fn().mockReturnThis(),
    toggleBold: vi.fn().mockReturnThis(),
    toggleItalic: vi.fn().mockReturnThis(),
    toggleBulletList: vi.fn().mockReturnThis(),
    toggleHeading: vi.fn().mockReturnThis(),
    run: vi.fn().mockReturnValue(true),
  };

  const canChain = {
    focus: vi.fn().mockReturnThis(),
    toggleBold: vi.fn().mockReturnThis(),
    toggleItalic: vi.fn().mockReturnThis(),
    toggleBulletList: vi.fn().mockReturnThis(),
    toggleHeading: vi.fn().mockReturnThis(),
    run: vi.fn().mockReturnValue(true),
  };

  const can = vi.fn().mockReturnValue({
    chain: () => canChain,
  });

  const chainFn = vi.fn().mockReturnValue(chain);

  return {
    chain: chainFn,
    can,
    isActive: vi.fn((format: string, params?: Record<string, unknown>) => {
      if (params?.level) {
        return activeStates[`${format}-${params.level}`] ?? false;
      }
      return activeStates[format] ?? false;
    }),
  } as unknown as Editor;
};

describe("EditorToolbar", () => {
  it("renders all formatting buttons", () => {
    const editor = createMockEditor();

    render(<EditorToolbar editor={editor} />);

    expect(screen.getByLabelText("Bold")).toBeInTheDocument();
    expect(screen.getByLabelText("Italic")).toBeInTheDocument();
    expect(screen.getByLabelText("Bullet List")).toBeInTheDocument();
    expect(screen.getByLabelText("Heading 2")).toBeInTheDocument();
    expect(screen.getByLabelText("Heading 3")).toBeInTheDocument();
  });

  it("calls editor.chain().toggleBold() on bold button click", async () => {
    const user = userEvent.setup();
    const editor = createMockEditor();
    const chainFn = editor.chain as ReturnType<typeof vi.fn>;

    render(<EditorToolbar editor={editor} />);

    const boldButton = screen.getByLabelText("Bold");
    await user.click(boldButton);

    expect(chainFn).toHaveBeenCalled();
  });

  it("calls editor.chain().toggleItalic() on italic button click", async () => {
    const user = userEvent.setup();
    const editor = createMockEditor();
    const chainFn = editor.chain as ReturnType<typeof vi.fn>;

    render(<EditorToolbar editor={editor} />);

    const italicButton = screen.getByLabelText("Italic");
    await user.click(italicButton);

    expect(chainFn).toHaveBeenCalled();
  });

  it("calls editor.chain().toggleBulletList() on list button click", async () => {
    const user = userEvent.setup();
    const editor = createMockEditor();

    render(<EditorToolbar editor={editor} />);

    const listButton = screen.getByLabelText("Bullet List");
    await user.click(listButton);

    expect(editor.chain).toHaveBeenCalled();
  });

  it("calls editor.chain().toggleHeading() on heading buttons click", async () => {
    const user = userEvent.setup();
    const editor = createMockEditor();

    render(<EditorToolbar editor={editor} />);

    const heading2Button = screen.getByLabelText("Heading 2");
    await user.click(heading2Button);

    expect(editor.chain).toHaveBeenCalled();

    const heading3Button = screen.getByLabelText("Heading 3");
    await user.click(heading3Button);

    expect(editor.chain).toHaveBeenCalled();
  });

  it("shows active state for bold when text is bold", () => {
    const editor = createMockEditor({ bold: true });

    render(<EditorToolbar editor={editor} />);

    const boldButton = screen.getByLabelText("Bold");
    expect(boldButton).toHaveClass("bg-white/[0.12]");
  });

  it("shows active state for italic when text is italic", () => {
    const editor = createMockEditor({ italic: true });

    render(<EditorToolbar editor={editor} />);

    const italicButton = screen.getByLabelText("Italic");
    expect(italicButton).toHaveClass("bg-white/[0.12]");
  });

  it("shows active state for bullet list when list is active", () => {
    const editor = createMockEditor({ bulletList: true });

    render(<EditorToolbar editor={editor} />);

    const listButton = screen.getByLabelText("Bullet List");
    expect(listButton).toHaveClass("bg-white/[0.12]");
  });

  it("shows active state for heading 2 when active", () => {
    const editor = createMockEditor({ "heading-2": true });

    render(<EditorToolbar editor={editor} />);

    const heading2Button = screen.getByLabelText("Heading 2");
    expect(heading2Button).toHaveClass("bg-white/[0.12]");
  });
});
