"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { codeApi, projectApi } from "@/lib/api";
import {
  ArrowLeft,
  Plus,
  Play,
  Save,
  Trash2,
  FileCode,
  Loader2,
  Terminal,
  FolderOpen,
} from "lucide-react";

interface CodeFile {
  id: string;
  name: string;
  language: string;
  content: string;
  version: number;
  updated_at: string;
}

interface RunResult {
  run_id: string;
  status: string;
  exit_code: number;
  output: string;
}

const SUPPORTED_LANGUAGES = [
  { value: "python", label: "Python", extension: ".py" },
  { value: "javascript", label: "JavaScript", extension: ".js" },
  { value: "r", label: "R", extension: ".R" },
];

export default function ProjectCodePage() {
  const params = useParams();
  const projectId = params.id as string;
  const [projectTitle, setProjectTitle] = useState("");
  const [files, setFiles] = useState<CodeFile[]>([]);
  const [selectedFile, setSelectedFile] = useState<CodeFile | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [isRunning, setIsRunning] = useState(false);
  const [runResult, setRunResult] = useState<RunResult | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newFileName, setNewFileName] = useState("");
  const [newFileLanguage, setNewFileLanguage] = useState("python");
  const [showTerminal, setShowTerminal] = useState(false);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [projectRes, filesRes] = await Promise.all([
          projectApi.get(projectId),
          codeApi.listFiles({ project_id: projectId, limit: 100 }),
        ]);
        setProjectTitle(projectRes.data.data.title);
        setFiles(filesRes.data.data || []);
      } catch (err) {
        console.error("Failed to fetch data:", err);
      } finally {
        setIsLoading(false);
      }
    };
    fetchData();
  }, [projectId]);

  const handleCreateFile = async () => {
    if (!newFileName.trim()) return;
    try {
      const ext = SUPPORTED_LANGUAGES.find((l) => l.value === newFileLanguage)?.extension || "";
      const res = await codeApi.createFile({
        project_id: projectId,
        name: newFileName + ext,
        language: newFileLanguage,
        content: "",
      });
      const newFile = res.data.data;
      setFiles([newFile, ...files]);
      setSelectedFile(newFile);
      setShowCreateModal(false);
      setNewFileName("");
    } catch (err) {
      console.error("Failed to create file:", err);
    }
  };

  const handleSaveFile = async () => {
    if (!selectedFile) return;
    setIsSaving(true);
    try {
      await codeApi.updateFile(selectedFile.id, {
        content: selectedFile.content,
      });
    } catch (err) {
      console.error("Failed to save file:", err);
    } finally {
      setIsSaving(false);
    }
  };

  const handleDeleteFile = async (id: string) => {
    if (!confirm("确定要删除这个文件吗？")) return;
    try {
      await codeApi.deleteFile(id);
      setFiles(files.filter((f) => f.id !== id));
      if (selectedFile?.id === id) {
        setSelectedFile(null);
      }
    } catch (err) {
      console.error("Failed to delete file:", err);
    }
  };

  const handleRunCode = async () => {
    if (!selectedFile) return;
    setIsRunning(true);
    setShowTerminal(true);
    setRunResult(null);
    try {
      const res = await codeApi.runCode({ file_id: selectedFile.id });
      setRunResult(res.data.data);
    } catch (err) {
      console.error("Failed to run code:", err);
      setRunResult({
        run_id: "",
        status: "failed",
        exit_code: 1,
        output: "执行失败，请检查代码或服务配置。",
      });
    } finally {
      setIsRunning(false);
    }
  };

  const handleContentChange = (content: string) => {
    if (selectedFile) {
      setSelectedFile({ ...selectedFile, content });
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
    <div className="h-[calc(100vh-4rem)] flex flex-col">
      <div className="flex items-center gap-4 px-6 py-3 border-b border-gray-200 bg-white">
        <Link
          href={`/projects/${projectId}`}
          className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
        >
          <ArrowLeft className="w-5 h-5 text-gray-600" />
        </Link>
        <div className="flex-1">
          <h1 className="text-xl font-bold text-gray-900">代码编辑</h1>
          <p className="text-sm text-gray-500">{projectTitle}</p>
        </div>
        <div className="flex items-center gap-2">
          {selectedFile && (
            <>
              <button
                onClick={handleSaveFile}
                disabled={isSaving}
                className="flex items-center gap-2 px-3 py-2 text-gray-600 hover:bg-gray-100 rounded-lg"
              >
                {isSaving ? (
                  <Loader2 className="w-4 h-4 animate-spin" />
                ) : (
                  <Save className="w-4 h-4" />
                )}
                保存
              </button>
              <button
                onClick={handleRunCode}
                disabled={isRunning}
                className="flex items-center gap-2 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-50"
              >
                {isRunning ? (
                  <Loader2 className="w-4 h-4 animate-spin" />
                ) : (
                  <Play className="w-4 h-4" />
                )}
                运行
              </button>
            </>
          )}
          <button
            onClick={() => setShowCreateModal(true)}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            <Plus className="w-4 h-4" />
            新建文件
          </button>
        </div>
      </div>

      <div className="flex flex-1 overflow-hidden">
        <div className="w-64 border-r border-gray-200 bg-gray-50 overflow-y-auto">
          <div className="p-4">
            <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-3">
              文件列表
            </h3>
            {files.length === 0 ? (
              <div className="text-center py-8 text-gray-500">
                <FolderOpen className="w-8 h-8 mx-auto mb-2" />
                <p className="text-sm">暂无文件</p>
              </div>
            ) : (
              <div className="space-y-1">
                {files.map((file) => (
                  <div
                    key={file.id}
                    className={`flex items-center justify-between px-3 py-2 rounded-lg cursor-pointer group ${
                      selectedFile?.id === file.id
                        ? "bg-blue-100 text-blue-700"
                        : "hover:bg-gray-200"
                    }`}
                    onClick={() => setSelectedFile(file)}
                  >
                    <div className="flex items-center gap-2 min-w-0">
                      <FileCode className="w-4 h-4 flex-shrink-0" />
                      <span className="text-sm truncate">{file.name}</span>
                    </div>
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        handleDeleteFile(file.id);
                      }}
                      className="p-1 hover:bg-red-100 rounded opacity-0 group-hover:opacity-100"
                    >
                      <Trash2 className="w-3 h-3 text-red-600" />
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

        <div className="flex-1 flex flex-col overflow-hidden">
          {selectedFile ? (
            <>
              <div className="flex items-center gap-2 px-4 py-2 bg-gray-100 border-b border-gray-200">
                <FileCode className="w-4 h-4 text-gray-600" />
                <span className="text-sm font-medium">{selectedFile.name}</span>
                <span className="text-xs text-gray-500">
                  v{selectedFile.version}
                </span>
              </div>
              <div className="flex-1 overflow-hidden">
                <textarea
                  value={selectedFile.content}
                  onChange={(e) => handleContentChange(e.target.value)}
                  className="w-full h-full p-4 font-mono text-sm bg-gray-900 text-green-400 resize-none focus:outline-none"
                  placeholder="输入代码..."
                  spellCheck={false}
                />
              </div>
            </>
          ) : (
            <div className="flex-1 flex items-center justify-center text-gray-500">
              <div className="text-center">
                <FileCode className="w-12 h-12 mx-auto mb-4" />
                <p>选择或创建一个文件开始编辑</p>
              </div>
            </div>
          )}

          {showTerminal && (
            <div className="h-48 border-t border-gray-200 bg-gray-900 text-gray-100 overflow-auto">
              <div className="flex items-center justify-between px-4 py-2 bg-gray-800">
                <div className="flex items-center gap-2">
                  <Terminal className="w-4 h-4" />
                  <span className="text-sm font-medium">终端输出</span>
                </div>
                <button
                  onClick={() => setShowTerminal(false)}
                  className="text-gray-400 hover:text-white"
                >
                  ×
                </button>
              </div>
              <pre className="p-4 text-sm font-mono whitespace-pre-wrap">
                {isRunning
                  ? "执行中..."
                  : runResult?.output || "点击运行按钮执行代码"}
              </pre>
            </div>
          )}
        </div>
      </div>

      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-full max-w-md mx-4">
            <h2 className="text-xl font-bold text-gray-900 mb-4">新建文件</h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  文件名
                </label>
                <input
                  type="text"
                  value={newFileName}
                  onChange={(e) => setNewFileName(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="例如: main"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  语言
                </label>
                <select
                  value={newFileLanguage}
                  onChange={(e) => setNewFileLanguage(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  {SUPPORTED_LANGUAGES.map((lang) => (
                    <option key={lang.value} value={lang.value}>
                      {lang.label}
                    </option>
                  ))}
                </select>
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
                onClick={handleCreateFile}
                disabled={!newFileName.trim()}
                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
              >
                创建
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
