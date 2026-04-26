"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { documentApi, projectApi } from "@/lib/api";
import {
  ArrowLeft,
  FileText,
  Plus,
  Clock,
  Edit3,
  Trash2,
  Loader2,
} from "lucide-react";

interface Document {
  id: string;
  title: string;
  abstract: string;
  status: string;
  word_count: number;
  updated_at: string;
}

export default function ProjectDocumentsPage() {
  const params = useParams();
  const router = useRouter();
  const projectId = params.id as string;
  const [documents, setDocuments] = useState<Document[]>([]);
  const [projectTitle, setProjectTitle] = useState("");
  const [isLoading, setIsLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newDocTitle, setNewDocTitle] = useState("");
  const [newDocAbstract, setNewDocAbstract] = useState("");
  const [isCreating, setIsCreating] = useState(false);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [projectRes, docsRes] = await Promise.all([
          projectApi.get(projectId),
          documentApi.list({ project_id: projectId, limit: 50 }),
        ]);
        setProjectTitle(projectRes.data.data.title);
        setDocuments(docsRes.data.documents || []);
      } catch (err) {
        console.error("Failed to fetch data:", err);
      } finally {
        setIsLoading(false);
      }
    };
    fetchData();
  }, [projectId]);

  const handleCreate = async () => {
    if (!newDocTitle.trim()) return;
    setIsCreating(true);
    try {
      const res = await documentApi.create({
        project_id: projectId,
        title: newDocTitle,
        abstract: newDocAbstract,
      });
      setDocuments([res.data, ...documents]);
      setShowCreateModal(false);
      setNewDocTitle("");
      setNewDocAbstract("");
    } catch (err) {
      console.error("Failed to create document:", err);
    } finally {
      setIsCreating(false);
    }
  };

  const handleDelete = async (docId: string) => {
    if (!confirm("确定要删除这篇论文吗？此操作不可恢复。")) return;
    try {
      await documentApi.delete(docId);
      setDocuments(documents.filter((d) => d.id !== docId));
    } catch (err) {
      console.error("Failed to delete document:", err);
    }
  };

  const getStatusLabel = (status: string) => {
    const labels: Record<string, string> = {
      draft: "草稿",
      reviewing: "审阅中",
      published: "已发表",
      archived: "已归档",
    };
    return labels[status] || status;
  };

  const getStatusColor = (status: string) => {
    const colors: Record<string, string> = {
      draft: "bg-yellow-100 text-yellow-700",
      reviewing: "bg-blue-100 text-blue-700",
      published: "bg-green-100 text-green-700",
      archived: "bg-gray-100 text-gray-700",
    };
    return colors[status] || "bg-gray-100 text-gray-700";
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
          <h1 className="text-2xl font-bold text-gray-900">论文管理</h1>
          <p className="text-gray-500">{projectTitle}</p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
        >
          <Plus className="w-4 h-4" />
          新建论文
        </button>
      </div>

      {documents.length === 0 ? (
        <div className="text-center py-16 bg-white rounded-xl border border-gray-200">
          <FileText className="w-12 h-12 text-gray-300 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">暂无论文</h3>
          <p className="text-gray-500 mb-6">创建第一篇论文开始写作</p>
          <button
            onClick={() => setShowCreateModal(true)}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            新建论文
          </button>
        </div>
      ) : (
        <div className="bg-white rounded-xl border border-gray-200 overflow-hidden">
          <div className="divide-y divide-gray-200">
            {documents.map((doc) => (
              <div
                key={doc.id}
                className="p-6 hover:bg-gray-50 transition-colors"
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center gap-3 mb-2">
                      <h3
                        className="text-lg font-semibold text-gray-900 cursor-pointer hover:text-blue-600"
                        onClick={() =>
                          router.push(`/projects/${projectId}/documents/${doc.id}`)
                        }
                      >
                        {doc.title}
                      </h3>
                      <span
                        className={`px-2 py-0.5 text-xs font-medium rounded-full ${getStatusColor(
                          doc.status
                        )}`}
                      >
                        {getStatusLabel(doc.status)}
                      </span>
                    </div>
                    {doc.abstract && (
                      <p className="text-gray-600 text-sm mb-3 line-clamp-2">
                        {doc.abstract}
                      </p>
                    )}
                    <div className="flex items-center gap-4 text-sm text-gray-500">
                      <span className="flex items-center gap-1">
                        <Clock className="w-4 h-4" />
                        {new Date(doc.updated_at).toLocaleDateString("zh-CN")}
                      </span>
                      <span>{doc.word_count} 字</span>
                    </div>
                  </div>
                  <div className="flex items-center gap-2 ml-4">
                    <button
                      onClick={() =>
                        router.push(`/projects/${projectId}/documents/${doc.id}`)
                      }
                      className="p-2 hover:bg-gray-200 rounded-lg"
                    >
                      <Edit3 className="w-4 h-4 text-gray-600" />
                    </button>
                    <button
                      onClick={() => handleDelete(doc.id)}
                      className="p-2 hover:bg-red-100 rounded-lg"
                    >
                      <Trash2 className="w-4 h-4 text-red-600" />
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-full max-w-lg mx-4">
            <h2 className="text-xl font-bold text-gray-900 mb-4">新建论文</h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  标题
                </label>
                <input
                  type="text"
                  value={newDocTitle}
                  onChange={(e) => setNewDocTitle(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="输入论文标题"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  摘要
                </label>
                <textarea
                  value={newDocAbstract}
                  onChange={(e) => setNewDocAbstract(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 h-24 resize-none"
                  placeholder="输入论文摘要（可选）"
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
                disabled={isCreating || !newDocTitle.trim()}
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
