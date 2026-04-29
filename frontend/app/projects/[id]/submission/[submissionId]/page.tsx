"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { submissionApi, reviewApi } from "@/lib/api";
import {
  ArrowLeft,
  Send,
  Clock,
  Loader2,
  BookOpen,
  TrendingUp,
  CheckCircle2,
  XCircle,
  Edit3,
  Trash2,
  Plus,
  MessageSquare,
  Star,
} from "lucide-react";

interface Journal {
  id: number;
  name: string;
  impact_factor: number;
  quartile: string;
  publisher: string;
  review_time_days: number;
  acceptance_rate: number;
}

interface Review {
  id: string;
  reviewer_name: string;
  review_type: string;
  status: string;
  content: string;
  rating: number;
  recommendation: string;
  created_at: string;
}

interface History {
  id: string;
  from_status: string;
  to_status: string;
  note: string;
  created_at: string;
}

interface Submission {
  id: string;
  title: string;
  abstract?: string;
  keywords: string[];
  status: string;
  stage: string;
  journal?: Journal;
  submission_date?: string;
  notes?: string;
  recommendation_score?: number;
  reviews: Review[];
  history: History[];
  created_at: string;
  updated_at: string;
}

const STATUS_LABELS: Record<string, string> = {
  draft: "草稿",
  submitted: "已投稿",
  under_review: "审稿中",
  revision_required: "需修改",
  revised: "已修改",
  accepted: "已接收",
  rejected: "已拒绝",
  withdrawn: "已撤稿",
};

const STATUS_COLORS: Record<string, string> = {
  draft: "bg-gray-100 text-gray-700",
  submitted: "bg-blue-100 text-blue-700",
  under_review: "bg-yellow-100 text-yellow-700",
  revision_required: "bg-orange-100 text-orange-700",
  revised: "bg-purple-100 text-purple-700",
  accepted: "bg-green-100 text-green-700",
  rejected: "bg-red-100 text-red-700",
  withdrawn: "bg-gray-100 text-gray-500",
};

const REVIEW_STATUS_LABELS: Record<string, string> = {
  pending: "待审",
  completed: "已完成",
  rejected: "已拒绝",
};

const RECOMMENDATION_LABELS: Record<string, string> = {
  accept: "接收",
  minor_revision: "小修",
  major_revision: "大修",
  reject: "拒绝",
};

export default function SubmissionDetailPage() {
  const params = useParams();
  const router = useRouter();
  const projectId = params.id as string;
  const submissionId = params.submissionId as string;
  const [submission, setSubmission] = useState<Submission | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [showStatusModal, setShowStatusModal] = useState(false);
  const [newStatus, setNewStatus] = useState("");
  const [statusNote, setStatusNote] = useState("");
  const [showReviewModal, setShowReviewModal] = useState(false);
  const [reviewContent, setReviewContent] = useState("");
  const [reviewRating, setReviewRating] = useState(0);
  const [reviewRecommendation, setReviewRecommendation] = useState("");
  const [isUpdating, setIsUpdating] = useState(false);

  useEffect(() => {
    const fetchSubmission = async () => {
      try {
        const res = await submissionApi.get(submissionId);
        setSubmission(res.data.data);
      } catch (err) {
        console.error("Failed to fetch submission:", err);
      } finally {
        setIsLoading(false);
      }
    };
    fetchSubmission();
  }, [submissionId]);

  const handleStatusUpdate = async () => {
    if (!newStatus) return;
    setIsUpdating(true);
    try {
      const res = await submissionApi.updateStatus(submissionId, {
        status: newStatus,
        note: statusNote,
      });
      setSubmission(res.data.data);
      setShowStatusModal(false);
      setNewStatus("");
      setStatusNote("");
    } catch (err) {
      console.error("Failed to update status:", err);
    } finally {
      setIsUpdating(false);
    }
  };

  const handleAddReview = async () => {
    if (!reviewContent.trim()) return;
    setIsUpdating(true);
    try {
      const res = await reviewApi.create(submissionId, {
        content: reviewContent,
        rating: reviewRating,
        recommendation: reviewRecommendation,
      });
      if (submission) {
        setSubmission({
          ...submission,
          reviews: [res.data.data, ...submission.reviews],
        });
      }
      setShowReviewModal(false);
      setReviewContent("");
      setReviewRating(0);
      setReviewRecommendation("");
    } catch (err) {
      console.error("Failed to add review:", err);
    } finally {
      setIsUpdating(false);
    }
  };

  const handleDeleteReview = async (reviewId: string) => {
    if (!confirm("确定要删除这条审稿意见吗？")) return;
    try {
      await reviewApi.delete(submissionId, reviewId);
      if (submission) {
        setSubmission({
          ...submission,
          reviews: submission.reviews.filter((r) => r.id !== reviewId),
        });
      }
    } catch (err) {
      console.error("Failed to delete review:", err);
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "accepted":
        return <CheckCircle2 className="w-5 h-5 text-green-500" />;
      case "rejected":
        return <XCircle className="w-5 h-5 text-red-500" />;
      default:
        return <Clock className="w-5 h-5 text-gray-400" />;
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="w-6 h-6 animate-spin text-blue-600" />
      </div>
    );
  }

  if (!submission) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500">投稿记录不存在</p>
        <Link
          href={`/projects/${projectId}/submission`}
          className="mt-4 text-blue-600 hover:underline"
        >
          返回投递列表
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link
          href={`/projects/${projectId}/submission`}
          className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
        >
          <ArrowLeft className="w-5 h-5 text-gray-600" />
        </Link>
        <div className="flex-1">
          <div className="flex items-center gap-3">
            {getStatusIcon(submission.status)}
            <span
              className={`inline-flex px-2.5 py-1 text-sm font-medium rounded-full ${
                STATUS_COLORS[submission.status] || "bg-gray-100 text-gray-700"
              }`}
            >
              {STATUS_LABELS[submission.status] || submission.status}
            </span>
          </div>
          <h1 className="text-xl font-bold text-gray-900 mt-2">
            {submission.title}
          </h1>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setShowStatusModal(true)}
            className="flex items-center gap-2 px-4 py-2 bg-white border border-gray-200 rounded-lg hover:bg-gray-50"
          >
            <Edit3 className="w-4 h-4" />
            更新状态
          </button>
          <Link
            href={`/projects/${projectId}/submission/${submissionId}/recommend`}
            className="flex items-center gap-2 px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700"
          >
            <TrendingUp className="w-4 h-4" />
            期刊推荐
          </Link>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 space-y-6">
          {submission.abstract && (
            <div className="bg-white rounded-xl border border-gray-200 p-6">
              <h2 className="text-lg font-semibold text-gray-900 mb-3">摘要</h2>
              <p className="text-gray-700 leading-relaxed">{submission.abstract}</p>
            </div>
          )}

          {submission.keywords && submission.keywords.length > 0 && (
            <div className="bg-white rounded-xl border border-gray-200 p-6">
              <h2 className="text-lg font-semibold text-gray-900 mb-3">关键词</h2>
              <div className="flex flex-wrap gap-2">
                {submission.keywords.map((kw) => (
                  <span
                    key={kw}
                    className="px-3 py-1 bg-blue-50 text-blue-700 text-sm rounded-full"
                  >
                    {kw}
                  </span>
                ))}
              </div>
            </div>
          )}

          <div className="bg-white rounded-xl border border-gray-200 p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold text-gray-900">审稿意见</h2>
              <button
                onClick={() => setShowReviewModal(true)}
                className="flex items-center gap-2 px-3 py-1.5 bg-blue-50 text-blue-700 rounded-lg hover:bg-blue-100"
              >
                <Plus className="w-4 h-4" />
                添加意见
              </button>
            </div>
            {submission.reviews && submission.reviews.length > 0 ? (
              <div className="space-y-4">
                {submission.reviews.map((review) => (
                  <div
                    key={review.id}
                    className="p-4 bg-gray-50 rounded-lg"
                  >
                    <div className="flex items-center justify-between mb-2">
                      <div className="flex items-center gap-2">
                        <MessageSquare className="w-4 h-4 text-gray-400" />
                        <span className="font-medium text-gray-900">
                          {review.reviewer_name || "匿名审稿人"}
                        </span>
                        {review.rating > 0 && (
                          <div className="flex items-center gap-1">
                            {[...Array(5)].map((_, i) => (
                              <Star
                                key={i}
                                className={`w-3.5 h-3.5 ${
                                  i < review.rating
                                    ? "text-yellow-400 fill-yellow-400"
                                    : "text-gray-300"
                                }`}
                              />
                            ))}
                          </div>
                        )}
                      </div>
                      <div className="flex items-center gap-2">
                        {review.recommendation && (
                          <span className="text-xs px-2 py-0.5 bg-blue-100 text-blue-700 rounded">
                            {RECOMMENDATION_LABELS[review.recommendation] || review.recommendation}
                          </span>
                        )}
                        <button
                          onClick={() => handleDeleteReview(review.id)}
                          className="p-1 hover:bg-red-100 rounded"
                        >
                          <Trash2 className="w-3.5 h-3.5 text-red-500" />
                        </button>
                      </div>
                    </div>
                    <p className="text-gray-700 text-sm">{review.content}</p>
                    <p className="text-xs text-gray-500 mt-2">
                      {new Date(review.created_at).toLocaleString("zh-CN")}
                    </p>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-gray-500 text-center py-4">暂无审稿意见</p>
            )}
          </div>

          <div className="bg-white rounded-xl border border-gray-200 p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">状态历史</h2>
            {submission.history && submission.history.length > 0 ? (
              <div className="space-y-3">
                {submission.history.map((h) => (
                  <div key={h.id} className="flex items-start gap-3">
                    <div className="w-2 h-2 bg-blue-500 rounded-full mt-2" />
                    <div className="flex-1">
                      <div className="flex items-center gap-2 text-sm">
                        {h.from_status && (
                          <span className="text-gray-500">
                            {STATUS_LABELS[h.from_status] || h.from_status}
                          </span>
                        )}
                        {h.from_status && <span className="text-gray-400">→</span>}
                        <span className="font-medium">
                          {STATUS_LABELS[h.to_status] || h.to_status}
                        </span>
                      </div>
                      {h.note && (
                        <p className="text-sm text-gray-600 mt-1">{h.note}</p>
                      )}
                      <p className="text-xs text-gray-400 mt-1">
                        {new Date(h.created_at).toLocaleString("zh-CN")}
                      </p>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-gray-500 text-center py-4">暂无历史记录</p>
            )}
          </div>
        </div>

        <div className="space-y-6">
          <div className="bg-white rounded-xl border border-gray-200 p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">投稿信息</h2>
            <div className="space-y-4">
              {submission.journal && (
                <div>
                  <p className="text-sm text-gray-500">目标期刊</p>
                  <p className="font-medium">{submission.journal.name}</p>
                  <div className="flex items-center gap-2 mt-1">
                    <span className="text-xs text-gray-500">
                      IF: {submission.journal.impact_factor}
                    </span>
                    <span className="text-xs px-1.5 py-0.5 bg-blue-100 text-blue-700 rounded">
                      {submission.journal.quartile}
                    </span>
                  </div>
                </div>
              )}
              <div>
                <p className="text-sm text-gray-500">投稿阶段</p>
                <p className="font-medium">
                  {STATUS_LABELS[submission.stage] || submission.stage}
                </p>
              </div>
              {submission.submission_date && (
                <div>
                  <p className="text-sm text-gray-500">投稿日期</p>
                  <p className="font-medium">
                    {new Date(submission.submission_date).toLocaleDateString("zh-CN")}
                  </p>
                </div>
              )}
              {submission.recommendation_score && submission.recommendation_score > 0 && (
                <div>
                  <p className="text-sm text-gray-500">期刊匹配度</p>
                  <div className="flex items-center gap-2">
                    <div className="flex-1 bg-gray-200 rounded-full h-2">
                      <div
                        className="bg-green-500 h-2 rounded-full"
                        style={{ width: `${submission.recommendation_score}%` }}
                      />
                    </div>
                    <span className="text-sm font-medium">
                      {Math.round(submission.recommendation_score)}%
                    </span>
                  </div>
                </div>
              )}
              <div>
                <p className="text-sm text-gray-500">创建时间</p>
                <p className="font-medium">
                  {new Date(submission.created_at).toLocaleDateString("zh-CN")}
                </p>
              </div>
              <div>
                <p className="text-sm text-gray-500">最后更新</p>
                <p className="font-medium">
                  {new Date(submission.updated_at).toLocaleDateString("zh-CN")}
                </p>
              </div>
            </div>
          </div>

          {submission.notes && (
            <div className="bg-white rounded-xl border border-gray-200 p-6">
              <h2 className="text-lg font-semibold text-gray-900 mb-3">备注</h2>
              <p className="text-gray-700">{submission.notes}</p>
            </div>
          )}
        </div>
      </div>

      {showStatusModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-full max-w-md mx-4">
            <h2 className="text-xl font-bold text-gray-900 mb-4">更新状态</h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  新状态
                </label>
                <select
                  value={newStatus}
                  onChange={(e) => setNewStatus(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="">选择状态</option>
                  {Object.entries(STATUS_LABELS).map(([key, label]) => (
                    <option key={key} value={key}>
                      {label}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  备注
                </label>
                <textarea
                  value={statusNote}
                  onChange={(e) => setStatusNote(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 h-20 resize-none"
                  placeholder="添加状态变更说明..."
                />
              </div>
            </div>
            <div className="flex justify-end gap-3 mt-6">
              <button
                onClick={() => setShowStatusModal(false)}
                className="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg"
              >
                取消
              </button>
              <button
                onClick={handleStatusUpdate}
                disabled={isUpdating || !newStatus}
                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 flex items-center gap-2"
              >
                {isUpdating && <Loader2 className="w-4 h-4 animate-spin" />}
                更新
              </button>
            </div>
          </div>
        </div>
      )}

      {showReviewModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-full max-w-lg mx-4">
            <h2 className="text-xl font-bold text-gray-900 mb-4">添加审稿意见</h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  审稿内容
                </label>
                <textarea
                  value={reviewContent}
                  onChange={(e) => setReviewContent(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 h-32 resize-none"
                  placeholder="输入审稿意见..."
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  评分
                </label>
                <div className="flex items-center gap-1">
                  {[1, 2, 3, 4, 5].map((star) => (
                    <button
                      key={star}
                      onClick={() => setReviewRating(star)}
                      className="p-1"
                    >
                      <Star
                        className={`w-6 h-6 ${
                          star <= reviewRating
                            ? "text-yellow-400 fill-yellow-400"
                            : "text-gray-300"
                        }`}
                      />
                    </button>
                  ))}
                </div>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  建议
                </label>
                <select
                  value={reviewRecommendation}
                  onChange={(e) => setReviewRecommendation(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="">选择建议</option>
                  <option value="accept">接收</option>
                  <option value="minor_revision">小修</option>
                  <option value="major_revision">大修</option>
                  <option value="reject">拒绝</option>
                </select>
              </div>
            </div>
            <div className="flex justify-end gap-3 mt-6">
              <button
                onClick={() => setShowReviewModal(false)}
                className="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg"
              >
                取消
              </button>
              <button
                onClick={handleAddReview}
                disabled={isUpdating || !reviewContent.trim()}
                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 flex items-center gap-2"
              >
                {isUpdating && <Loader2 className="w-4 h-4 animate-spin" />}
                添加
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
