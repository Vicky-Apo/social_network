import PostPage from "./PostPage";

export default async function Page({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const postId = Number(id);
  return <PostPage postId={postId} />;
}

