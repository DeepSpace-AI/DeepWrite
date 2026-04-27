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
  Upload,
  FileText,
  Copy,
  Check,
  Globe,
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

interface SearchResult {
  title: string;
  authors: string[];
  year?: number;
  journal?: string;
  doi?: string;
  abstract?: string;
}

export default function ProjectReferencesPage() {
  const params = useParams();
  const projectId = params.id as string;
  const [references, setReferences] = useState<Reference[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState("");
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showImportModal, setShowImportModal] = useState(false);
  const [showSearchModal, setShowSearchModal] = useState(false);
  const [isCreating, setIsCreating] = useState(false);
  const [isImporting, setIsImporting] = useState(false);
  const [isSearching, setIsSearching] = useState(false);
  const [doiInput, setDoiInput] = useState("");
  const [externalQuery, setExternalQuery] = useState("");
  const [searchResults, setSearchResults] = useState<SearchResult[]>([]);
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const [citationStyle, setCitationStyle] = useState("apa");
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
    const res = await referenceApi.list({ project_id: projectId, limit: 100 });
    return res.data.data || [];
  }, [projectId]);

  useEffect(() => {
    fetchReferences()
      .then(setReferences)
      .catch((err) => console.error("Failed to fetch references:", err))
      .finally(() => setIsLoading(false));
  }, [fetchReferences]);

  const handleCreate = async () => {
    if (!newRef.title.trim()) return;
    setIsCreating(true);
    try {
      const authors = newRef.authors
        .split(",")
        .map((a) => a.trim())
        .filter(Boolean);
      const tags = newRef.tags
        .split(",")
        .map((t) => t.trim())
        .filter(Boolean);
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
      setNewRef({
        title: "",
        authors: "",
        year: "",
        journal: "",
        doi: "",
        abstract: "",
        tags: "",
      });
      fetchReferences().then(setReferences);
    } catch (err) {
      console.error("Failed to create reference:", err);
    } finally {
      setIsCreating(false);
    }
  };

  const handleImportDOI = async () => {
    if (!doiInput.trim()) return;
    setIsImporting(true);
    try {
      await referenceApi.importDOI({ project_id: projectId, doi: doiInput });
      setDoiInput("");
      setShowImportModal(false);
      fetchReferences().then(setReferences);
    } catch (err) {
      console.error("Failed to import DOI:", err);
      alert("导入失败，请检查DOI是否正确");
    } finally {
      setIsImporting(false);
    }
  };

  const handleImportBibTeX = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    setIsImporting(true);
    try {
      await referenceApi.importBibTeX(projectId, file);
      fetchReferences().then(setReferences);
    } catch (err) {
      console.error("Failed to import BibTeX:", err);
      alert("导入失败，请检查BibTeX文件格式");
    } finally {
      setIsImporting(false);
      e.target.value = "";
    }
  };

  const handleExternalSearch = async () => {
    if (!externalQuery.trim()) return;
    setIsSearching(true);
    try {
      const res = await referenceApi.searchExternal(externalQuery);
      setSearchResults(res.data.data || []);
    } catch (err) {
      console.error("Failed to search:", err);
    } finally {
      setIsSearching(false);
    }
  };

  const handleImportFromSearch = async (result: SearchResult) => {
    try {
      await referenceApi.create({
        project_id: projectId,
        title: result.title,
        authors: result.authors,
        year: result.year,
        journal: result.journal,
        doi: result.doi,
        abstract: result.abstract,
      });
      fetchReferences().then(setReferences);
      alert("导入成功");
    } catch (err) {
      console.error("Failed to import:", err);
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

  const handleCopyCitation = async (ref: Reference) => {
    try {
      const res = await referenceApi.getCitation(ref.id, citationStyle);
      const citation = res.data.data.citation;
      await navigator.clipboard.writeText(citation);
      setCopiedId(ref.id);
      setTimeout(() => setCopiedId(null), 2000);
    } catch (err) {
      console.error("Failed to copy citation:", err);
    }
  };

  const filteredRefs = references.filter(
    (ref) =>
      ref.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
      ref.authors.some((a) =>
        a.toLowerCase().includes(searchQuery.toLowerCase())
      ) ||
      ref.tags.some((t) =>
        t.toLowerCase().includes(searchQuery.toLowerCase())
      )
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
        <div className="flex items-center gap-2">
          <select
            value={citationStyle}
            onChange={(e) => setCitationStyle(e.target.value)}
            className="px-3 py-2 border border-gray-300 rounded-lg text-sm"
          >
            <option value="apa">APA</option>
            <option value="mla">MLA</option>
            <option value="gb7714">GB/T 7714</option>
            <option value="chicago">Chicago</option>
            <option value="harvard">Harvard</option>
          </select>
          <button
            onClick={() => setShowSearchModal(true)}
            className="flex items-center gap-2 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700"
          >
            <Globe className="w-4 h-4" />
            搜索
          </button>
          <button
            onClick={() => setShowImportModal(true)}
            className="flex items-center gap-2 px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700"
          >
            <Upload className="w-4 h-4" />
            导入
          </button>
          <button
            onClick={() => setShowCreateModal(true)}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            <Plus className="w-4 h-4" />
            添加
          </button>
        </div>
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
          <div className="flex items-center justify-center gap-3">
            <button
              onClick={() => setShowImportModal(true)}
              className="px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700"
            >
              导入文献
            </button>
            <button
              onClick={() => setShowCreateModal(true)}
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
            >
              手动添加
            </button>
          </div>
        </div>
      ) : (
        <div className="bg-white rounded-xl border border-gray-200 overflow-hidden">
          <div className="divide-y divide-gray-200">
            {filteredRefs.map((ref) => (
              <div
                key={ref.id}
                className="p-6 hover:bg-gray-50 transition-colors"
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <h3 className="text-lg font-semibold text-gray-900 mb-1">
                      {ref.title}
                    </h3>
                    <p className="text-gray-600 text-sm mb-2">
                      {ref.authors.join(", ")}
                      {ref.year && ` (${ref.year})`}
                      {ref.journal && ` · ${ref.journal}`}
                    </p>
                    {ref.abstract && (
                      <p className="text-gray-500 text-sm mb-3 line-clamp-2">
                        {ref.abstract}
                      </p>
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
                    <div className="flex items-center gap-4 mt-3">
                      {ref.doi && (
                        <a
                          href={`https://doi.org/${ref.doi}`}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="inline-flex items-center gap-1 text-sm text-blue-600 hover:underline"
                        >
                          <ExternalLink className="w-3 h-3" />
                          DOI: {ref.doi}
                        </a>
                      )}
                      <button
                        onClick={() => handleCopyCitation(ref)}
                        className="inline-flex items-center gap-1 text-sm text-gray-600 hover:text-gray-900"
                      >
                        {copiedId === ref.id ? (
                          <>
                            <Check className="w-3 h-3 text-green-600" />
                            <span className="text-green-600">已复制</span>
                          </>
                        ) : (
                          <>
                            <Copy className="w-3 h-3" />
                            复制引用
                          </>
                        )}
                      </button>
                    </div>
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
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  标题 *
                </label>
                <input
                  type="text"
                  value={newRef.title}
                  onChange={(e) =>
                    setNewRef({ ...newRef, title: e.target.value })
                  }
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="文献标题"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  作者（逗号分隔）
                </label>
                <input
                  type="text"
                  value={newRef.authors}
                  onChange={(e) =>
                    setNewRef({ ...newRef, authors: e.target.value })
                  }
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="例如：张三, 李四, 王五"
                />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    年份
                  </label>
                  <input
                    type="number"
                    value={newRef.year}
                    onChange={(e) =>
                      setNewRef({ ...newRef, year: e.target.value })
                    }
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="2024"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    期刊
                  </label>
                  <input
                    type="text"
                    value={newRef.journal}
                    onChange={(e) =>
                      setNewRef({ ...newRef, journal: e.target.value })
                    }
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="期刊名称"
                  />
                </div>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  DOI
                </label>
                <input
                  type="text"
                  value={newRef.doi}
                  onChange={(e) =>
                    setNewRef({ ...newRef, doi: e.target.value })
                  }
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="10.xxxx/xxxxx"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  摘要
                </label>
                <textarea
                  value={newRef.abstract}
                  onChange={(e) =>
                    setNewRef({ ...newRef, abstract: e.target.value })
                  }
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 h-24 resize-none"
                  placeholder="文献摘要"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  标签（逗号分隔）
                </label>
                <input
                  type="text"
                  value={newRef.tags}
                  onChange={(e) =>
                    setNewRef({ ...newRef, tags: e.target.value })
                  }
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

      {showImportModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-full max-w-lg mx-4">
            <h2 className="text-xl font-bold text-gray-900 mb-4">导入文献</h2>
            <div className="space-y-6">
              <div>
                <h3 className="font-medium text-gray-900 mb-2 flex items-center gap-2">
                  <FileText className="w-4 h-4" />
                  通过 DOI 导入
                </h3>
                <div className="flex gap-2">
                  <input
                    type="text"
                    value={doiInput}
                    onChange={(e) => setDoiInput(e.target.value)}
                    className="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="输入 DOI，例如：10.1000/xyz123"
                  />
                  <button
                    onClick={handleImportDOI}
                    disabled={isImporting || !doiInput.trim()}
                    className="px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700 disabled:opacity-50 flex items-center gap-2"
                  >
                    {isImporting ? (
                      <Loader2 className="w-4 h-4 animate-spin" />
                    ) : (
                      <Upload className="w-4 h-4" />
                    )}
                    导入
                  </button>
                </div>
              </div>
              <div>
                <h3 className="font-medium text-gray-900 mb-2 flex items-center gap-2">
                  <FileText className="w-4 h-4" />
                  通过 BibTeX 文件导入
                </h3>
                <label className="block">
                  <span className="sr-only">选择 BibTeX 文件</span>
                  <input
                    type="file"
                    accept=".bib,.bibtex"
                    onChange={handleImportBibTeX}
                    className="block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-lg file:border-0 file:text-sm file:font-semibold file:bg-purple-50 file:text-purple-700 hover:file:bg-purple-100"
                  />
                </label>
              </div>
            </div>
            <div className="flex justify-end mt-6">
              <button
                onClick={() => setShowImportModal(false)}
                className="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg"
              >
                关闭
              </button>
            </div>
          </div>
        </div>
      )}

      {showSearchModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-full max-w-2xl mx-4 max-h-[90vh] overflow-y-auto">
            <h2 className="text-xl font-bold text-gray-900 mb-4">
              搜索外部文献
            </h2>
            <div className="flex gap-2 mb-4">
              <input
                type="text"
                value={externalQuery}
                onChange={(e) => setExternalQuery(e.target.value)}
                className="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="搜索关键词..."
                onKeyDown={(e) => e.key === "Enter" && handleExternalSearch()}
              />
              <button
                onClick={handleExternalSearch}
                disabled={isSearching || !externalQuery.trim()}
                className="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-50 flex items-center gap-2"
              >
                {isSearching ? (
                  <Loader2 className="w-4 h-4 animate-spin" />
                ) : (
                  <Search className="w-4 h-4" />
                )}
                搜索
              </button>
            </div>
            <div className="space-y-4">
              {searchResults.map((result, index) => (
                <div
                  key={index}
                  className="p-4 border border-gray-200 rounded-lg"
                >
                  <h3 className="font-medium text-gray-900 mb-1">
                    {result.title}
                  </h3>
                  <p className="text-sm text-gray-600 mb-2">
                    {result.authors.join(", ")}
                    {result.year && ` (${result.year})`}
                    {result.journal && ` · ${result.journal}`}
                  </p>
                  {result.abstract && (
                    <p className="text-sm text-gray-500 mb-3 line-clamp-2">
                      {result.abstract}
                    </p>
                  )}
                  <button
                    onClick={() => handleImportFromSearch(result)}
                    className="text-sm text-blue-600 hover:text-blue-700 font-medium"
                  >
                    + 添加到文献库
                  </button>
                </div>
              ))}
              {searchResults.length === 0 && !isSearching && (
                <div className="text-center py-8 text-gray-500">
                  输入关键词搜索 Crossref 或 Semantic Scholar
                </div>
              )}
            </div>
            <div className="flex justify-end mt-6">
              <button
                onClick={() => {
                  setShowSearchModal(false);
                  setSearchResults([]);
                }}
                className="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg"
              >
                关闭
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
