'use client';

import { useCallback, useEffect, useState } from 'react';
import { useAuth } from '@/components/AuthProvider';
import {
  getAdminSubmissions,
  reviewSubmission,
  type AdminSubmission,
  type PaginationMeta,
} from '@/lib/admin';
import styles from './page.module.css';

type StatusFilter = '' | 'pending' | 'approved' | 'rejected';

const FILTERS: { label: string; value: StatusFilter }[] = [
  { label: '全部', value: '' },
  { label: '待审核', value: 'pending' },
  { label: '已通过', value: 'approved' },
  { label: '已拒绝', value: 'rejected' },
];

const PAGE_SIZE = 20;

function formatTime(iso: string): string {
  const d = new Date(iso);
  return d.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  });
}

function StatusBadge({ status }: { status: string }) {
  const map: Record<string, { cls: string; text: string }> = {
    pending: { cls: styles.badgePending, text: '待审核' },
    approved: { cls: styles.badgeApproved, text: '已通过' },
    rejected: { cls: styles.badgeRejected, text: '已拒绝' },
  };
  const info = map[status] ?? { cls: styles.badgePending, text: status };
  return <span className={`${styles.badge} ${info.cls}`}>{info.text}</span>;
}

function SubmissionCard({
  submission,
  onReviewed,
}: {
  submission: AdminSubmission;
  onReviewed: () => void;
}) {
  const { token } = useAuth();
  const [expanded, setExpanded] = useState(false);
  const [note, setNote] = useState('');
  const [reviewing, setReviewing] = useState(false);

  const handleReview = async (newStatus: 'approved' | 'rejected') => {
    if (!token || reviewing) return;
    setReviewing(true);
    try {
      await reviewSubmission(token, submission.id, newStatus, note);
      onReviewed();
    } catch (err) {
      console.error('Review failed:', err);
    } finally {
      setReviewing(false);
    }
  };

  return (
    <div className={styles.card}>
      <div className={styles.cardHeader}>
        <div className={styles.cardMeta}>
          <span className={styles.cardType}>{submission.type}</span>
          {submission.submitterEmail && (
            <span className={styles.cardEmail}>{submission.submitterEmail}</span>
          )}
          <span className={styles.cardTime}>{formatTime(submission.createdAt)}</span>
        </div>
        <StatusBadge status={submission.status} />
      </div>

      <button
        className={styles.payloadToggle}
        onClick={() => setExpanded((v) => !v)}
        type="button"
      >
        {expanded ? '▾ 收起数据' : '▸ 查看数据'}
      </button>

      {expanded && (
        <pre className={styles.payload}>
          {JSON.stringify(submission.payload, null, 2)}
        </pre>
      )}

      {submission.reviewNote && (
        <div className={styles.reviewNote}>审核备注: {submission.reviewNote}</div>
      )}

      {submission.status === 'pending' && (
        <div className={styles.reviewActions}>
          <input
            className={styles.noteInput}
            type="text"
            placeholder="审核备注（可选）"
            value={note}
            onChange={(e) => setNote(e.target.value)}
          />
          <button
            className={styles.btnApprove}
            onClick={() => handleReview('approved')}
            disabled={reviewing}
            type="button"
          >
            ✓ 通过
          </button>
          <button
            className={styles.btnReject}
            onClick={() => handleReview('rejected')}
            disabled={reviewing}
            type="button"
          >
            ✗ 拒绝
          </button>
        </div>
      )}
    </div>
  );
}

export default function AdminSubmissionsPage() {
  const { token } = useAuth();
  const [filter, setFilter] = useState<StatusFilter>('');
  const [page, setPage] = useState(1);
  const [submissions, setSubmissions] = useState<AdminSubmission[]>([]);
  const [meta, setMeta] = useState<PaginationMeta | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchData = useCallback(async () => {
    if (!token) return;
    setLoading(true);
    try {
      const res = await getAdminSubmissions(
        token,
        filter || undefined,
        page,
        PAGE_SIZE,
      );
      setSubmissions(res.data);
      setMeta(res.meta);
    } catch (err) {
      console.error('Failed to load submissions:', err);
      setSubmissions([]);
      setMeta(null);
    } finally {
      setLoading(false);
    }
  }, [token, filter, page]);

  useEffect(() => {
    void fetchData();
  }, [fetchData]);

  const handleFilterChange = (value: StatusFilter) => {
    setFilter(value);
    setPage(1);
  };

  const totalPages = meta?.totalPages ?? 1;

  const renderPagination = () => {
    if (!meta || totalPages <= 1) return null;

    const pages: number[] = [];
    for (let i = 1; i <= totalPages; i++) {
      pages.push(i);
    }

    return (
      <div className={styles.pagination}>
        <button
          className={styles.pageBtn}
          onClick={() => setPage((p) => Math.max(1, p - 1))}
          disabled={page <= 1}
          type="button"
        >
          上一页
        </button>

        {pages.map((p) => (
          <button
            key={p}
            className={`${styles.pageBtn} ${p === page ? styles.pageBtnActive : ''}`}
            onClick={() => setPage(p)}
            type="button"
          >
            {p}
          </button>
        ))}

        <span className={styles.pageInfo}>
          {meta.total} 条记录
        </span>

        <button
          className={styles.pageBtn}
          onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
          disabled={page >= totalPages}
          type="button"
        >
          下一页
        </button>
      </div>
    );
  };

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <h1 className={styles.headerTitle}>&gt; SUBMISSIONS // REVIEW</h1>
      </div>

      <div className={styles.filterTabs}>
        {FILTERS.map((f) => (
          <button
            key={f.value}
            className={`${styles.filterTab} ${filter === f.value ? styles.filterTabActive : ''}`}
            onClick={() => handleFilterChange(f.value)}
            type="button"
          >
            {f.label}
          </button>
        ))}
      </div>

      {loading ? (
        <div className={styles.spinner} />
      ) : submissions.length === 0 ? (
        <div className={styles.empty}>暂无提交记录</div>
      ) : (
        <>
          <div className={styles.list}>
            {submissions.map((s) => (
              <SubmissionCard
                key={s.id}
                submission={s}
                onReviewed={fetchData}
              />
            ))}
          </div>
          {renderPagination()}
        </>
      )}
    </div>
  );
}
