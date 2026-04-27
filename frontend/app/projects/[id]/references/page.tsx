"use client";

import { useCallback, useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { referenceApi } from "@/lib/api";
import {
  ArrowLeft,
  BookOpen,
  Plus,
  Search,
  Loader2,
  Trash2,
  ExternalLink,
  Tag,
} from "lucide-react";

interface Reference {
  id: string;
  title: string;
  authors: string[];
  year?: number;
  journal?: string;
  doi?: string;
  abstract?: string;
  keywords: string[];
  tags: string[];
  created_at: string;
}

export default function ProjectReferencesPage() {
  const params = useParams();
  const projectId = params.id as string;
  const [references, setReferences] = useState<Reference[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState("");
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [isCreating, setIsCreating] = useState(false);
  const [newRef, setNewRef] = useState({
    title: "",
    authors: "",
    year: "",
    journal: "",
    doi: "",
    abstract: "",
    tags: "",
  });

  const fetchReferences = useCallback(async () => {
    try {
      const res = await referenceApi.list({ project_id: projectId, limit: 100 });
      setReferences(res.data.data || []);
    } catch (err) {
      console.error("Failed to fetch references:", err);
    } finally {
      setIsLoading(false);
    }
  }, [projectId]);

  useEffect(() => {
    fetchReferences();
  }, [fetchReferences]);

  const handleCreate = async () => {
    if (!newRef.title.trim()) return;
    setIsCreating(true);
    try {
      const authors = newRef.authors.split(",").map((a) => a.trim()).filter(Boolean);
      const tags = newRef.tags.split(",").map((t) => t.trim()).filter(Boolean);
      await referenceApi.create({
        project_id: projectId,
        title: newRef.title,
        authors,
        year: newRef.year ? parseInt(newRef.year) : undefined,
        journal: newRef.journal || undefined,
        doi: newRef.doi || undefined,
        abstract: newRef.abstract || undefined,
        tags,
      });
      setShowCreateModal(false);
      setNewRef({ title: "", authors: "", year: "", journal: "", doi: "", abstract: "", tags: "" });
      fetchReferences();
    } catch (err) {
      console.error("Failed to create reference:", err);
    } finally {
      setIsCreating(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm("确定要删除这篇文献吗？")) return;
    try {
      await referenceApi.delete(id);
      setReferences(references.filter((r) => r.id !== id));
    } catch (err) {
      console.error("Failed to delete reference:", err);
    }
  };

  const filteredRefs = references.filter((ref) =>
    ref.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
    ref.authors.some((a) => a.toLowerCase().includes(searchQuery.toLowerCase())) ||
    ref.tags.some((t) => t.toLowerCase().includes(searchQuery.toLowerCase()))
  );

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
          <h1 className="text-2xl font-bold text-gray-900">文献管理</h1>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
        >
          <Plus className="w-4 h-4" />
          添加文献
        </button>
      </div>

      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
        <input
          type="text"
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="w-full pl-10 pr-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="搜索文献标题、作者或标签..."
        />
      </div>

      {filteredRefs.length === 0 ? (
        <div className="text-center py-16 bg-white rounded-xl border border-gray-200">
          <BookOpen className="w-12 h-12 text-gray-300 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">暂无文献</h3>
          <p className="text-gray-500 mb-6">添加文献以开始管理引用</p>
          <button
            onClick={() => setShowCreateModal(true)}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            添加文献
          </button>
        </div>
      ) : (
        <div className="bg-white rounded-xl border border-gray-200 overflow-hidden">
          <div className="divide-y divide-gray-200">
            {filteredRefs.map((ref) => (
              <div key={ref.id} className="p-6 hover:bg-gray-50 transition-colors">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <h3 className="text-lg font-semibold text-gray-900 mb-1">{ref.title}</h3>
                    <p className="text-gray-600 text-sm mb-2">
                      {ref.authors.join(", ")}
                      {ref.year && ` (${ref.year})`}
                      {ref.journal && ` · ${ref.journal}`}
                    </p>
                    {ref.abstract && (
                      <p className="text-gray-500 text-sm mb-3 line-clamp-2">{ref.abstract}</p>
                    )}
                    <div className="flex items-center gap-2 flex-wrap">
                      {ref.tags.map((tag) => (
                        <span
                          key={tag}
                          className="inline-flex items-center gap-1 px-2 py-0.5 bg-blue-50 text-blue-700 text-xs rounded-full"
                        >
                          <Tag className="w-3 h-3" />
                          {tag}
                        </span>
                      ))}
                    </div>
                    {ref.doi && (
                      <a
                        href={`https://doi.org/${ref.doi}`}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="inline-flex items-center gap-1 text-sm text-blue-600 hover:underline mt-2"
                      >
                        <ExternalLink className="w-3 h-3" />
                        DOI: {ref.doi}
                      </a>
                    )}
                  </div>
                  <button
                    onClick={() => handleDelete(ref.id)}
                    className="p-2 hover:bg-red-100 rounded-lg ml-4"
                  >
                    <Trash2 className="w-4 h-4 text-red-600" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-full max-w-lg mx-4 max-h-[90vh] overflow-y-auto">
            <h2 className="text-xl font-bold text-gray-900 mb-4">添加文献</h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">标题 *</label>
                <input
                  type="text"
                  value={newRef.title}
                  onChange={(e) => setNewRef({ ...newRef, title: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="文献标题"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">作者（逗号分隔）</label>
                <input
                  type="text"
                  value={newRef.authors}
                  onChange={(e) => setNewRef({ ...newRef, authors: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="例如：张三, 李四, 王五"
                />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">年份</label>
                  <input
                    type="number"
                    value={newRef.year}
                    onChange={(e) => setNewRef({ ...newRef, year: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="2024"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">期刊</label>
                  <input
                    type="text"
                    value={newRef.journal}
                    onChange={(e) => setNewRef({ ...newRef, journal: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="期刊名称"
                  />
                </div>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">DOI</label>
                <input
                  type="text"
                  value={newRef.doi}
                  onChange={(e) => setNewRef({ ...newRef, doi: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="10.xxxx/xxxxx"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">摘要</label>
                <textarea
                  value={newRef.abstract}
                  onChange={(e) => setNewRef({ ...newRef, abstract: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 h-24 resize-none"
                  placeholder="文献摘要"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">标签（逗号分隔）</label>
                <input
                  type="text"
                  value={newRef.tags}
                  onChange={(e) => setNewRef({ ...newRef, tags: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="例如：深度学习, 医学影像, CNN"
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
                disabled={isCreating || !newRef.title.trim()}
                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 flex items-center gap-2"
              >
                {isCreating && <Loader2 className="w-4 h-4 animate-spin" />}
                添加
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
