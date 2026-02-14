import { EditorPageClient } from "@/components/coaching/editor-page-client";

export default async function CoachingEditPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  return <EditorPageClient coverLetterId={id} />;
}
