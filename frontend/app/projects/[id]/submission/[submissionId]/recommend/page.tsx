"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { submissionApi } from "@/lib/api";
import {
  ArrowLeft,
  BookOpen,
  Loader2,
  TrendingUp,
  Clock,
  Percent,
  ExternalLink,
  CheckCircle,
} from "lucide-react";

interface Journal {
  id: number;
  name: string;
  publisher: string;
  impact_factor: number;
  quartile: string;
  open_access: boolean;
  website_url: string;
  review_time_days: number;
  acceptance_rate: number;
  keywords: string[];
  category: string;
}

export default function RecommendPage() {
  const params = useParams();
  const router = useRouter();
  const projectId = params.id as string;
  const submissionId = params.submissionId as string;
  const [recommendations, setRecommendations] = useState<Journal[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState<number | null>(null);
  const [submittedJournal, setSubmittedJournal] = useState<number | null>(null);

  useEffect(() => {
    const fetchRecommendations = async () => {
      try {
        const res = await submissionApi.recommendForSubmission(submissionId);
        setRecommendations(res.data.data || []);
      } catch (err) {
        console.error("Failed to fetch recommendations:", err);
      } finally {
        setIsLoading(false);
      }
    };
    fetchRecommendations();
  }, [submissionId]);

  const handleSubmitToJournal = async (journalId: number) => {
    if (!confirm("确定要投稿到这个期刊吗？")) return;
    setIsSubmitting(journalId);
    try {
      await submissionApi.submitToJournal(submissionId, journalId);
      setSubmittedJournal(journalId);
      setTimeout(() => {
        router.push(`/projects/${projectId}/submission/${submissionId}`);
      }, 1500);
    } catch (err) {
      console.error("Failed to submit:", err);
    } finally {
      setIsSubmitting(null);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="w-6 h-6 animate-spin text-blue-600" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link
          href={`/projects/${projectId}/submission/${submissionId}`}
          className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
        >
          <ArrowLeft className="w-5 h-5 text-gray-600" />
        </Link>
        <div>
          <h1 className="text-2xl font-bold text-gray-900">期刊推荐</h1>
          <p className="text-gray-500">基于您的论文内容智能推荐的期刊</p>
        </div>
      </div>

      {submittedJournal && (
        <div className="bg-green-50 border border-green-200 rounded-xl p-4 flex items-center gap-3">
          <CheckCircle className="w-5 h-5 text-green-500" />
          <p className="text-green-700">投稿成功！正在跳转到投稿详情...</p>
        </div>
      )}

      {recommendations.length === 0 ? (
        <div className="text-center py-12 bg-white rounded-xl border border-gray-200">
          <BookOpen className="w-12 h-12 text-gray-300 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">暂无推荐</h3>
          <p className="text-gray-500">
            请先完善论文的摘要和关键词以获得更准确的推荐
          </p>
        </div>
      ) : (
        <div className="space-y-4">
          <p className="text-sm text-gray-500">
            根据您的论文内容，我们为您推荐以下期刊：
          </p>
          {recommendations.map((journal, index) => (
            <div
              key={journal.id}
              className="bg-white rounded-xl border border-gray-200 p-6 hover:shadow-md transition-shadow"
            >
              <div className="flex items-start justify-between">
                <div className="flex items-start gap-4 flex-1">
                  <div className="w-10 h-10 bg-blue-100 rounded-lg flex items-center justify-center flex-shrink-0">
                    <span className="text-lg font-bold text-blue-600">
                      {index + 1}
                    </span>
                  </div>
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      <h3 className="text-lg font-semibold text-gray-900">
                        {journal.name}
                      </h3>
                      <span
                        className={`px-2 py-0.5 text-xs font-medium rounded ${
                          journal.quartile === "Q1"
                            ? "bg-green-100 text-green-700"
                            : journal.quartile === "Q2"
                            ? "bg-blue-100 text-blue-700"
                            : "bg-gray-100 text-gray-700"
                        }`}
                      >
                        {journal.quartile}
                      </span>
                      {journal.open_access && (
                        <span className="px-2 py-0.5 text-xs font-medium bg-purple-100 text-purple-700 rounded">
                          OA
                        </span>
                      )}
                    </div>
                    <p className="text-sm text-gray-500 mb-3">
                      {journal.publisher} · {journal.category}
                    </p>

                    <div className="flex items-center gap-6 mb-3">
                      <div className="flex items-center gap-1">
                        <TrendingUp className="w-4 h-4 text-gray-400" />
                        <span className="text-sm">
                          影响因子: <span className="font-medium">{journal.impact_factor}</span>
                        </span>
                      </div>
                      <div className="flex items-center gap-1">
                        <Clock className="w-4 h-4 text-gray-400" />
                        <span className="text-sm">
                          审稿周期: <span className="font-medium">{journal.review_time_days}天</span>
                        </span>
                      </div>
                      <div className="flex items-center gap-1">
                        <Percent className="w-4 h-4 text-gray-400" />
                        <span className="text-sm">
                          录用率: <span className="font-medium">{journal.acceptance_rate}%</span>
                        </span>
                      </div>
                    </div>

                    {journal.keywords && journal.keywords.length > 0 && (
                      <div className="flex flex-wrap gap-1">
                        {journal.keywords.slice(0, 6).map((kw) => (
                          <span
                            key={kw}
                            className="px-2 py-0.5 text-xs bg-gray-100 text-gray-600 rounded"
                          >
                            {kw}
                          </span>
                        ))}
                      </div>
                    )}
                  </div>
                </div>

                <div className="flex items-center gap-2 ml-4">
                  {journal.website_url && (
                    <a
                      href={journal.website_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="p-2 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-lg"
                      title="访问期刊网站"
                    >
                      <ExternalLink className="w-5 h-5" />
                    </a>
                  )}
                  <button
                    onClick={() => handleSubmitToJournal(journal.id)}
                    disabled={isSubmitting !== null || submittedJournal !== null}
                    className={`px-4 py-2 rounded-lg flex items-center gap-2 ${
                      submittedJournal === journal.id
                        ? "bg-green-600 text-white"
                        : "bg-blue-600 text-white hover:bg-blue-700"
                    } disabled:opacity-50`}
                  >
                    {isSubmitting === journal.id ? (
                      <Loader2 className="w-4 h-4 animate-spin" />
                    ) : submittedJournal === journal.id ? (
                      <CheckCircle className="w-4 h-4" />
                    ) : null}
                    {submittedJournal === journal.id ? "已投稿" : "投稿到此期刊"}
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
