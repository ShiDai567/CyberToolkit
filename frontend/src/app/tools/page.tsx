import { ToolsExplorer } from '@/components/ToolsExplorer';
import { Footer } from '@/components/Footer';
import { getCategories, getTools } from '@/lib/api';

export const dynamic = 'force-dynamic';

export default async function ToolsPage({
  searchParams,
}: {
  searchParams: Promise<{ category?: string }>;
}) {
  const params = await searchParams;
  const [tools, categories] = await Promise.all([
    getTools({ page: 1, pageSize: 100 }),
    getCategories(),
  ]);

  return (
    <>
      <ToolsExplorer tools={tools} categories={categories} initialCategory={params.category || ''} />
      <Footer />
    </>
  );
}
