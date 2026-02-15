import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { RollbackDialog } from "../rollback-dialog";

describe("RollbackDialog", () => {
  it("renders version number in title", () => {
    render(
      <RollbackDialog
        open={true}
        onOpenChange={vi.fn()}
        versionNumber={3}
        isLoading={false}
        onConfirm={vi.fn()}
      />
    );
    expect(
      screen.getByText("v3 버전으로 복원하시겠습니까?")
    ).toBeInTheDocument();
  });

  it("shows description text", () => {
    render(
      <RollbackDialog
        open={true}
        onOpenChange={vi.fn()}
        versionNumber={1}
        isLoading={false}
        onConfirm={vi.fn()}
      />
    );
    expect(
      screen.getByText(
        "현재 내용은 새 버전으로 저장된 후, 선택한 버전의 내용이 복원됩니다."
      )
    ).toBeInTheDocument();
  });

  it("calls onConfirm when confirm button clicked", async () => {
    const user = userEvent.setup();
    const handleConfirm = vi.fn();
    render(
      <RollbackDialog
        open={true}
        onOpenChange={vi.fn()}
        versionNumber={1}
        isLoading={false}
        onConfirm={handleConfirm}
      />
    );

    await user.click(screen.getByText("복원하기"));
    expect(handleConfirm).toHaveBeenCalled();
  });

  it("calls onOpenChange(false) when cancel button clicked", async () => {
    const user = userEvent.setup();
    const handleOpenChange = vi.fn();
    render(
      <RollbackDialog
        open={true}
        onOpenChange={handleOpenChange}
        versionNumber={1}
        isLoading={false}
        onConfirm={vi.fn()}
      />
    );

    await user.click(screen.getByText("취소"));
    expect(handleOpenChange).toHaveBeenCalledWith(false);
  });

  it("shows loading state during rollback", () => {
    render(
      <RollbackDialog
        open={true}
        onOpenChange={vi.fn()}
        versionNumber={1}
        isLoading={true}
        onConfirm={vi.fn()}
      />
    );

    expect(screen.getByText("복원 중...")).toBeInTheDocument();
  });

  it("disables buttons during loading", () => {
    render(
      <RollbackDialog
        open={true}
        onOpenChange={vi.fn()}
        versionNumber={1}
        isLoading={true}
        onConfirm={vi.fn()}
      />
    );

    expect(screen.getByText("취소").closest("button")).toBeDisabled();
    expect(screen.getByText("복원 중...").closest("button")).toBeDisabled();
  });
});
