"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { submissionApi, projectApi } from "@/lib/api";
import {
  ArrowLeft,
  Plus,
  Send,
  FileText,
  Clock,
  Loader2,
  Search,
  BookOpen,
  TrendingUp,
  CheckCircle2,
  XCircle,
} from "lucide-react";

interface Submission {
  id: string;
  title: string;
  status: string;
  stage: string;
  journal?: {
    name: string;
    impact_factor: number;
    quartile: string;
  };
  submission_date?: string;
  updated_at: string;
  recommendation_score?: number;
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

const STAGE_LABELS: Record<string, string> = {
  preparation: "准备中",
  submitted: "已投稿",
  under_review: "审稿中",
  revision: "修改中",
  decision: "决定中",
  published: "已发表",
};

export default function ProjectSubmissionPage() {
  const params = useParams();
  const router = useRouter();
  const projectId = params.id as string;
  const [projectTitle, setProjectTitle] = useState("");
  const [submissions, setSubmissions] = useState<Submission[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newTitle, setNewTitle] = useState("");
  const [newAbstract, setNewAbstract] = useState("");
  const [isCreating, setIsCreating] = useState(false);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [projectRes, subsRes] = await Promise.all([
          projectApi.get(projectId),
          submissionApi.listByProject(projectId, { limit: 50 }),
        ]);
        setProjectTitle(projectRes.data.data.title);
        setSubmissions(subsRes.data.data || []);
      } catch (err) {
        console.error("Failed to fetch data:", err);
      } finally {
        setIsLoading(false);
      }
    };
    fetchData();
  }, [projectId]);

  const handleCreate = async () => {
    if (!newTitle.trim()) return;
    setIsCreating(true);
    try {
      const res = await submissionApi.create({
        project_id: projectId,
        title: newTitle,
        abstract: newAbstract,
      });
      setSubmissions([res.data.data, ...submissions]);
      setShowCreateModal(false);
      setNewTitle("");
      setNewAbstract("");
    } catch (err) {
      console.error("Failed to create submission:", err);
    } finally {
      setIsCreating(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm("确定要删除这个投稿记录吗？")) return;
    try {
      await submissionApi.delete(id);
      setSubmissions(submissions.filter((s) => s.id !== id));
    } catch (err) {
      console.error("Failed to delete submission:", err);
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "accepted":
        return <CheckCircle2 className="w-4 h-4 text-green-500" />;
      case "rejected":
        return <XCircle className="w-4 h-4 text-red-500" />;
      default:
        return <Clock className="w-4 h-4 text-gray-400" />;
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
          href={`/projects/${projectId}`}
          className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
        >
          <ArrowLeft className="w-5 h-5 text-gray-600" />
        </Link>
        <div className="flex-1">
          <h1 className="text-2xl font-bold text-gray-900">投递管理</h1>
          <p className="text-gray-500">{projectTitle}</p>
        </div>
        <div className="flex items-center gap-2">
          <Link
            href={`/projects/${projectId}/submission/journals`}
            className="flex items-center gap-2 px-4 py-2 bg-white border border-gray-200 rounded-lg hover:bg-gray-50"
          >
            <BookOpen className="w-4 h-4" />
            期刊浏览
          </Link>
          <button
            onClick={() => setShowCreateModal(true)}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            <Plus className="w-4 h-4" />
            新建投稿
          </button>
        </div>
      </div>

      {submissions.length === 0 ? (
        <div className="text-center py-16 bg-white rounded-xl border border-gray-200">
          <Send className="w-12 h-12 text-gray-300 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">暂无投稿记录</h3>
          <p className="text-gray-500 mb-6">创建投稿以开始论文投递流程</p>
          <button
            onClick={() => setShowCreateModal(true)}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            新建投稿
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {submissions.map((sub) => (
            <div
              key={sub.id}
              className="bg-white rounded-xl border border-gray-200 p-5 hover:shadow-md hover:border-blue-300 transition-all cursor-pointer"
              onClick={() => router.push(`/projects/${projectId}/submission/${sub.id}`)}
            >
              <div className="flex items-start justify-between mb-3">
                <div className="flex items-center gap-2">
                  {getStatusIcon(sub.status)}
                  <span
                    className={`inline-flex px-2 py-0.5 text-xs font-medium rounded-full ${
                      STATUS_COLORS[sub.status] || "bg-gray-100 text-gray-700"
                    }`}
                  >
                    {STATUS_LABELS[sub.status] || sub.status}
                  </span>
                </div>
                <span className="text-xs text-gray-500">
                  {STAGE_LABELS[sub.stage] || sub.stage}
                </span>
              </div>

              <h3 className="font-semibold text-gray-900 mb-2 line-clamp-2">
                {sub.title}
              </h3>

              {sub.journal && (
                <div className="flex items-center gap-2 mb-3">
                  <BookOpen className="w-3.5 h-3.5 text-gray-400" />
                  <span className="text-sm text-gray-600 truncate">
                    {sub.journal.name}
                  </span>
                  <span className="text-xs text-gray-400">
                    IF: {sub.journal.impact_factor}
                  </span>
                </div>
              )}

              <div className="flex items-center justify-between text-xs text-gray-500">
                <span className="flex items-center gap-1">
                  <Clock className="w-3 h-3" />
                  {new Date(sub.updated_at).toLocaleDateString("zh-CN")}
                </span>
                {sub.recommendation_score && sub.recommendation_score > 0 && (
                  <span className="flex items-center gap-1">
                    <TrendingUp className="w-3 h-3" />
                    匹配度: {Math.round(sub.recommendation_score)}%
                  </span>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-full max-w-lg mx-4">
            <h2 className="text-xl font-bold text-gray-900 mb-4">新建投稿</h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  论文标题
                </label>
                <input
                  type="text"
                  value={newTitle}
                  onChange={(e) => setNewTitle(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="输入论文标题"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  摘要（可选）
                </label>
                <textarea
                  value={newAbstract}
                  onChange={(e) => setNewAbstract(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 h-24 resize-none"
                  placeholder="输入论文摘要以获得更好的期刊推荐..."
                />
              </div>
            </div>
            <div className="flex justify-end gap-3 mt-6">
              <button
                onClick={() => setShowCreateModal(false)}
                className="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg"
              >
                取消
              </button>
              <button
                onClick={handleCreate}
                disabled={isCreating || !newTitle.trim()}
                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 flex items-center gap-2"
              >
                {isCreating && <Loader2 className="w-4 h-4 animate-spin" />}
                创建
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
