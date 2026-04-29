"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { journalApi } from "@/lib/api";
import {
  ArrowLeft,
  Search,
  BookOpen,
  Loader2,
  ExternalLink,
  TrendingUp,
  Clock,
  Percent,
} from "lucide-react";

interface Journal {
  id: number;
  name: string;
  publisher: string;
  issn: string;
  category: string;
  subcategory: string;
  impact_factor: number;
  quartile: string;
  open_access: boolean;
  website_url: string;
  review_time_days: number;
  acceptance_rate: number;
  keywords: string[];
}

export default function JournalsPage() {
  const params = useParams();
  const projectId = params.id as string;
  const [journals, setJournals] = useState<Journal[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedCategory, setSelectedCategory] = useState("");
  const [categories, setCategories] = useState<string[]>([]);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);

  useEffect(() => {
    const fetchCategories = async () => {
      try {
        const res = await journalApi.getCategories();
        setCategories(res.data.data || []);
      } catch (err) {
        console.error("Failed to fetch categories:", err);
      }
    };
    fetchCategories();
  }, []);

  useEffect(() => {
    const fetchJournals = async () => {
      setIsLoading(true);
      try {
        const res = await journalApi.list({
          query: searchQuery,
          category: selectedCategory,
          page,
          limit: 20,
        });
        setJournals(res.data.data || []);
        setTotal(res.data.meta?.total || 0);
      } catch (err) {
        console.error("Failed to fetch journals:", err);
      } finally {
        setIsLoading(false);
      }
    };
    fetchJournals();
  }, [searchQuery, selectedCategory, page]);

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link
          href={`/projects/${projectId}/submission`}
          className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
        >
          <ArrowLeft className="w-5 h-5 text-gray-600" />
        </Link>
        <div>
          <h1 className="text-2xl font-bold text-gray-900">期刊浏览</h1>
          <p className="text-gray-500">浏览和搜索学术期刊</p>
        </div>
      </div>

      <div className="flex items-center gap-4">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => {
              setSearchQuery(e.target.value);
              setPage(1);
            }}
            placeholder="搜索期刊名称、出版商或关键词..."
            className="w-full pl-10 pr-4 py-2 bg-white border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
        <select
          value={selectedCategory}
          onChange={(e) => {
            setSelectedCategory(e.target.value);
            setPage(1);
          }}
          className="px-4 py-2 bg-white border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          <option value="">所有学科</option>
          {categories.map((cat) => (
            <option key={cat} value={cat}>
              {cat}
            </option>
          ))}
        </select>
      </div>

      {isLoading ? (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="w-6 h-6 animate-spin text-blue-600" />
        </div>
      ) : journals.length === 0 ? (
        <div className="text-center py-12 bg-white rounded-xl border border-gray-200">
          <BookOpen className="w-12 h-12 text-gray-300 mx-auto mb-4" />
          <p className="text-gray-500">未找到匹配的期刊</p>
        </div>
      ) : (
        <>
          <p className="text-sm text-gray-500">共 {total} 个期刊</p>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {journals.map((journal) => (
              <div
                key={journal.id}
                className="bg-white rounded-xl border border-gray-200 p-5 hover:shadow-md transition-shadow"
              >
                <div className="flex items-start justify-between mb-3">
                  <div className="flex-1">
                    <h3 className="font-semibold text-gray-900 mb-1">
                      {journal.name}
                    </h3>
                    <p className="text-sm text-gray-500">{journal.publisher}</p>
                  </div>
                  <div className="flex items-center gap-2">
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
                </div>

                <div className="flex items-center gap-4 mb-3">
                  <div className="flex items-center gap-1">
                    <TrendingUp className="w-3.5 h-3.5 text-gray-400" />
                    <span className="text-sm">
                      <span className="font-medium">{journal.impact_factor}</span>
                      <span className="text-gray-500"> IF</span>
                    </span>
                  </div>
                  <div className="flex items-center gap-1">
                    <Clock className="w-3.5 h-3.5 text-gray-400" />
                    <span className="text-sm text-gray-600">
                      {journal.review_time_days}天审稿
                    </span>
                  </div>
                  <div className="flex items-center gap-1">
                    <Percent className="w-3.5 h-3.5 text-gray-400" />
                    <span className="text-sm text-gray-600">
                      {journal.acceptance_rate}%录用率
                    </span>
                  </div>
                </div>

                <div className="flex items-center gap-2 mb-3">
                  <span className="px-2 py-0.5 text-xs bg-gray-100 text-gray-700 rounded">
                    {journal.category}
                  </span>
                  {journal.subcategory && (
                    <span className="px-2 py-0.5 text-xs bg-gray-100 text-gray-700 rounded">
                      {journal.subcategory}
                    </span>
                  )}
                </div>

                {journal.keywords && journal.keywords.length > 0 && (
                  <div className="flex flex-wrap gap-1 mb-3">
                    {journal.keywords.slice(0, 5).map((kw) => (
                      <span
                        key={kw}
                        className="px-2 py-0.5 text-xs bg-blue-50 text-blue-600 rounded"
                      >
                        {kw}
                      </span>
                    ))}
                  </div>
                )}

                {journal.website_url && (
                  <a
                    href={journal.website_url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center gap-1 text-sm text-blue-600 hover:text-blue-800"
                  >
                    <ExternalLink className="w-3.5 h-3.5" />
                    访问期刊网站
                  </a>
                )}
              </div>
            ))}
          </div>

          {total > 20 && (
            <div className="flex items-center justify-center gap-2">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                className="px-4 py-2 bg-white border border-gray-200 rounded-lg hover:bg-gray-50 disabled:opacity-50"
              >
                上一页
              </button>
              <span className="text-sm text-gray-600">
                第 {page} 页 / 共 {Math.ceil(total / 20)} 页
              </span>
              <button
                onClick={() => setPage((p) => p + 1)}
                disabled={page >= Math.ceil(total / 20)}
                className="px-4 py-2 bg-white border border-gray-200 rounded-lg hover:bg-gray-50 disabled:opacity-50"
              >
                下一页
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
}
