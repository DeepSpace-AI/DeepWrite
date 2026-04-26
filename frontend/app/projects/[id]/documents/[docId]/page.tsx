"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { documentApi } from "@/lib/api";
import {
  ArrowLeft,
  Save,
  CheckCircle,
  History,
  Loader2,
  Sparkles,
} from "lucide-react";

interface DocumentContent {
  abstract?: string;
  introduction?: string;
  methods?: string;
  results?: string;
  discussion?: string;
  conclusion?: string;
}

const SECTIONS = [
  { key: "abstract", label: "摘要", placeholder: "输入论文摘要..." },
  { key: "introduction", label: "引言", placeholder: "输入引言部分..." },
  { key: "methods", label: "方法", placeholder: "输入研究方法..." },
  { key: "results", label: "结果", placeholder: "输入研究结果..." },
  { key: "discussion", label: "讨论", placeholder: "输入讨论部分..." },
  { key: "conclusion", label: "结论", placeholder: "输入结论..." },
];

export default function DocumentEditorPage() {
  const params = useParams();
  const projectId = params.id as string;
  const docId = params.docId as string;

  const [title, setTitle] = useState("");
  const [content, setContent] = useState<DocumentContent>({});
  const [, setStatus] = useState("draft");
  const [wordCount, setWordCount] = useState(0);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [lastSaved, setLastSaved] = useState<Date | null>(null);
  const [activeSection, setActiveSection] = useState("abstract");
  const [showVersionModal, setShowVersionModal] = useState(false);
  const [versionSummary, setVersionSummary] = useState("");
  const [isCreatingVersion, setIsCreatingVersion] = useState(false);
  const [showAIModal, setShowAIModal] = useState(false);
  const [aiPrompt, setAiPrompt] = useState("");

  useEffect(() => {
    const fetchDocument = async () => {
      try {
        const res = await documentApi.get(docId, true);
        const doc = res.data;
        setTitle(doc.title);
        setStatus(doc.status);
        setWordCount(doc.word_count || 0);
        if (doc.content?.content) {
          setContent(doc.content.content as DocumentContent);
        }
      } catch (err) {
        console.error("Failed to fetch document:", err);
      } finally {
        setIsLoading(false);
      }
    };
    fetchDocument();
  }, [docId]);

  const handleSave = async () => {
    setIsSaving(true);
    try {
      await documentApi.updateContent(docId, {
        content: content as Record<string, unknown>,
        metadata: {
          word_count: wordCount,
          format: "markdown",
        },
      });
      setLastSaved(new Date());
    } catch (err) {
      console.error("Failed to save:", err);
    } finally {
      setIsSaving(false);
    }
  };

  const handleCreateVersion = async () => {
    if (!versionSummary.trim()) return;
    setIsCreatingVersion(true);
    try {
      await documentApi.createVersion(docId, { change_summary: versionSummary });
      setShowVersionModal(false);
      setVersionSummary("");
    } catch (err) {
      console.error("Failed to create version:", err);
    } finally {
      setIsCreatingVersion(false);
    }
  };

  const handleContentChange = (section: string, html: string) => {
    setContent((prev) => ({ ...prev, [section]: html }));
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="w-6 h-6 animate-spin text-blue-600" />
      </div>
    );
  }

  return (
    <div className="h-[calc(100vh-4rem)] flex flex-col">
      <div className="flex items-center gap-4 px-6 py-3 border-b border-gray-200 bg-white">
        <Link
          href={`/projects/${projectId}/documents`}
          className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
        >
          <ArrowLeft className="w-5 h-5 text-gray-600" />
        </Link>
        <div className="flex-1">
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            className="text-lg font-semibold text-gray-900 bg-transparent border-none focus:outline-none focus:ring-0 w-full"
            placeholder="论文标题"
          />
        </div>
        <div className="flex items-center gap-4 text-sm text-gray-500">
          <span>{wordCount} 字</span>
          {lastSaved && (
            <span className="flex items-center gap-1">
              <CheckCircle className="w-4 h-4 text-green-500" />
              已保存
            </span>
          )}
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setShowAIModal(true)}
            className="flex items-center gap-2 px-3 py-2 text-purple-600 hover:bg-purple-50 rounded-lg"
          >
            <Sparkles className="w-4 h-4" />
            AI 助手
          </button>
          <button
            onClick={() => setShowVersionModal(true)}
            className="flex items-center gap-2 px-3 py-2 text-gray-600 hover:bg-gray-100 rounded-lg"
          >
            <History className="w-4 h-4" />
            保存版本
          </button>
          <button
            onClick={handleSave}
            disabled={isSaving}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
          >
            {isSaving ? (
              <Loader2 className="w-4 h-4 animate-spin" />
            ) : (
              <Save className="w-4 h-4" />
            )}
            保存
          </button>
        </div>
      </div>

      <div className="flex flex-1 overflow-hidden">
        <div className="w-48 border-r border-gray-200 bg-gray-50 overflow-y-auto">
          <div className="p-4">
            <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-3">
              章节
            </h3>
            <nav className="space-y-1">
              {SECTIONS.map((section) => (
                <button
                  key={section.key}
                  onClick={() => setActiveSection(section.key)}
                  className={`w-full text-left px-3 py-2 rounded-lg text-sm transition-colors ${
                    activeSection === section.key
                      ? "bg-blue-100 text-blue-700 font-medium"
                      : "text-gray-600 hover:bg-gray-200"
                  }`}
                >
                  {section.label}
                  {content[section.key as keyof DocumentContent] && (
                    <span className="ml-2 text-xs text-green-500">●</span>
                  )}
                </button>
              ))}
            </nav>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto bg-white">
          <div className="max-w-4xl mx-auto py-8 px-6">
            {SECTIONS.map((section) => (
              <div
                key={section.key}
                className={activeSection === section.key ? "block" : "hidden"}
              >
                <h2 className="text-2xl font-bold text-gray-900 mb-6">
                  {section.label}
                </h2>
                <textarea
                  value={content[section.key as keyof DocumentContent] || ""}
                  onChange={(e) =>
                    handleContentChange(section.key, e.target.value)
                  }
                  className="w-full min-h-[500px] p-4 text-gray-700 leading-relaxed resize-none focus:outline-none"
                  placeholder={section.placeholder}
                />
              </div>
            ))}
          </div>
        </div>
      </div>

      {showVersionModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-full max-w-md mx-4">
            <h2 className="text-xl font-bold text-gray-900 mb-4">保存版本</h2>
            <p className="text-gray-500 mb-4">
              保存当前内容为一个新版本，方便以后回溯。
            </p>
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">
                版本说明
              </label>
              <input
                type="text"
                value={versionSummary}
                onChange={(e) => setVersionSummary(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="例如：初稿完成、方法部分修订"
              />
            </div>
            <div className="flex justify-end gap-3">
              <button
                onClick={() => setShowVersionModal(false)}
                className="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg"
              >
                取消
              </button>
              <button
                onClick={handleCreateVersion}
                disabled={isCreatingVersion || !versionSummary.trim()}
                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 flex items-center gap-2"
              >
                {isCreatingVersion && (
                  <Loader2 className="w-4 h-4 animate-spin" />
                )}
                保存版本
              </button>
            </div>
          </div>
        </div>
      )}

      {showAIModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-full max-w-lg mx-4">
            <h2 className="text-xl font-bold text-gray-900 mb-4 flex items-center gap-2">
              <Sparkles className="w-5 h-5 text-purple-600" />
              AI 写作助手
            </h2>
            <div className="space-y-4">
              <div className="p-4 bg-purple-50 rounded-lg">
                <h3 className="font-medium text-purple-900 mb-2">当前章节</h3>
                <p className="text-purple-700">
                  {SECTIONS.find((s) => s.key === activeSection)?.label}
                </p>
              </div>
              <textarea
                value={aiPrompt}
                onChange={(e) => setAiPrompt(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-purple-500 h-24 resize-none"
                placeholder="描述您想要 AI 帮助的内容，例如：帮我写一段关于深度学习在医学影像中应用的引言..."
              />
              <div className="flex gap-2">
                <button className="px-3 py-1.5 text-sm bg-purple-100 text-purple-700 rounded-lg hover:bg-purple-200">
                  续写
                </button>
                <button className="px-3 py-1.5 text-sm bg-purple-100 text-purple-700 rounded-lg hover:bg-purple-200">
                  润色
                </button>
                <button className="px-3 py-1.5 text-sm bg-purple-100 text-purple-700 rounded-lg hover:bg-purple-200">
                  摘要
                </button>
              </div>
            </div>
            <div className="flex justify-end gap-3 mt-6">
              <button
                onClick={() => setShowAIModal(false)}
                className="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg"
              >
                取消
              </button>
              <button
                onClick={() => setShowAIModal(false)}
                className="px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700"
              >
                生成内容
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
