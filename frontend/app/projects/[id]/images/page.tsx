"use client";

import { useEffect, useState, useRef } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { imageApi, aiApi, projectApi } from "@/lib/api";
import {
  ArrowLeft,
  Upload,
  Trash2,
  Image as ImageIcon,
  Loader2,
  Download,
  X,
  Tag,
  Sparkles,
  Wand2,
  Copy,
  Check,
} from "lucide-react";

interface ImageData {
  id: string;
  name: string;
  description: string;
  url: string;
  width: number;
  height: number;
  size: number;
  format: string;
  tags: string[];
  created_at: string;
}

export default function ProjectImagesPage() {
  const params = useParams();
  const projectId = params.id as string;
  const [projectTitle, setProjectTitle] = useState("");
  const [images, setImages] = useState<ImageData[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isUploading, setIsUploading] = useState(false);
  const [selectedImage, setSelectedImage] = useState<ImageData | null>(null);
  const [showUploadModal, setShowUploadModal] = useState(false);
  const [uploadName, setUploadName] = useState("");
  const fileInputRef = useRef<HTMLInputElement>(null);

  const [showAIModal, setShowAIModal] = useState(false);
  const [aiPrompt, setAiPrompt] = useState("");
  const [aiLoading, setAiLoading] = useState(false);
  const [aiResult, setAiResult] = useState<string | null>(null);
  const [aiStyle, setAiStyle] = useState("realistic");
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [projectRes, imagesRes] = await Promise.all([
          projectApi.get(projectId),
          imageApi.list({ project_id: projectId, limit: 50 }),
        ]);
        setProjectTitle(projectRes.data.data.title);
        setImages(imagesRes.data.data || []);
      } catch (err) {
        console.error("Failed to fetch data:", err);
      } finally {
        setIsLoading(false);
      }
    };
    fetchData();
  }, [projectId]);

  const handleUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setIsUploading(true);
    try {
      const res = await imageApi.upload(projectId, file, uploadName || file.name);
      setImages([res.data.data, ...images]);
      setShowUploadModal(false);
      setUploadName("");
    } catch (err) {
      console.error("Failed to upload image:", err);
      alert("上传失败，请检查文件格式");
    } finally {
      setIsUploading(false);
      if (fileInputRef.current) {
        fileInputRef.current.value = "";
      }
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm("确定要删除这张图片吗？")) return;
    try {
      await imageApi.delete(id);
      setImages(images.filter((img) => img.id !== id));
      if (selectedImage?.id === id) {
        setSelectedImage(null);
      }
    } catch (err) {
      console.error("Failed to delete image:", err);
    }
  };

  const formatFileSize = (bytes: number) => {
    if (bytes < 1024) return bytes + " B";
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
    return (bytes / (1024 * 1024)).toFixed(1) + " MB";
  };

  const handleAIGenerate = async () => {
    if (!aiPrompt.trim()) return;
    setAiLoading(true);
    setAiResult(null);
    try {
      const stylePrompts: Record<string, string> = {
        realistic: "photorealistic, high quality, detailed",
        illustration: "illustration, vector art, clean design",
        scientific: "scientific diagram, technical illustration, academic style",
        abstract: "abstract art, creative, modern",
      };

      const fullPrompt = `${aiPrompt}, ${stylePrompts[aiStyle] || stylePrompts.realistic}`;

      const res = await aiApi.chat({
        message: `请为以下图像生成一个详细的描述，用于AI图像生成。这个描述应该足够详细，可以用来生成高质量的学术或科研相关图像。\n\n用户需求：${fullPrompt}\n\n请直接返回描述，不要添加其他解释。`,
        context: "图像描述生成",
      });

      setAiResult(res.data.data.response);
    } catch (err) {
      console.error("AI generation failed:", err);
      setAiResult("AI 生成失败，请检查网络连接或 API 配置。");
    } finally {
      setAiLoading(false);
    }
  };

  const handleCopyDescription = () => {
    if (aiResult) {
      navigator.clipboard.writeText(aiResult);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
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
          <h1 className="text-2xl font-bold text-gray-900">图像管理</h1>
          <p className="text-gray-500">{projectTitle}</p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setShowAIModal(true)}
            className="flex items-center gap-2 px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700"
          >
            <Sparkles className="w-4 h-4" />
            AI 生成
          </button>
          <button
            onClick={() => setShowUploadModal(true)}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            <Upload className="w-4 h-4" />
            上传图片
          </button>
        </div>
      </div>

      {images.length === 0 ? (
        <div className="text-center py-16 bg-white rounded-xl border border-gray-200">
          <ImageIcon className="w-12 h-12 text-gray-300 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">暂无图片</h3>
          <p className="text-gray-500 mb-6">上传图片或使用AI生成以开始管理</p>
          <div className="flex items-center justify-center gap-3">
            <button
              onClick={() => setShowAIModal(true)}
              className="px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700 flex items-center gap-2"
            >
              <Sparkles className="w-4 h-4" />
              AI 生成
            </button>
            <button
              onClick={() => setShowUploadModal(true)}
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
            >
              上传图片
            </button>
          </div>
        </div>
      ) : (
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
          {images.map((image) => (
            <div
              key={image.id}
              className="bg-white rounded-xl border border-gray-200 overflow-hidden cursor-pointer hover:shadow-lg transition-shadow"
              onClick={() => setSelectedImage(image)}
            >
              <div className="aspect-square bg-gray-100 relative">
                <img
                  src={image.url}
                  alt={image.name}
                  className="w-full h-full object-cover"
                />
              </div>
              <div className="p-3">
                <p className="text-sm font-medium text-gray-900 truncate">
                  {image.name}
                </p>
                <p className="text-xs text-gray-500 mt-1">
                  {image.width} × {image.height} · {formatFileSize(image.size)}
                </p>
                {image.tags.length > 0 && (
                  <div className="flex items-center gap-1 mt-2 flex-wrap">
                    {image.tags.slice(0, 2).map((tag) => (
                      <span
                        key={tag}
                        className="inline-flex items-center gap-1 px-1.5 py-0.5 bg-blue-50 text-blue-700 text-xs rounded"
                      >
                        <Tag className="w-2.5 h-2.5" />
                        {tag}
                      </span>
                    ))}
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {selectedImage && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-full max-w-2xl mx-4 max-h-[90vh] overflow-y-auto">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl font-bold text-gray-900">
                {selectedImage.name}
              </h2>
              <button
                onClick={() => setSelectedImage(null)}
                className="p-2 hover:bg-gray-100 rounded-lg"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
            <div className="mb-4">
              <img
                src={selectedImage.url}
                alt={selectedImage.name}
                className="w-full rounded-lg"
              />
            </div>
            <div className="grid grid-cols-2 gap-4 mb-4">
              <div>
                <p className="text-sm text-gray-500">尺寸</p>
                <p className="font-medium">
                  {selectedImage.width} × {selectedImage.height}
                </p>
              </div>
              <div>
                <p className="text-sm text-gray-500">大小</p>
                <p className="font-medium">
                  {formatFileSize(selectedImage.size)}
                </p>
              </div>
              <div>
                <p className="text-sm text-gray-500">格式</p>
                <p className="font-medium">{selectedImage.format}</p>
              </div>
              <div>
                <p className="text-sm text-gray-500">上传时间</p>
                <p className="font-medium">
                  {new Date(selectedImage.created_at).toLocaleDateString("zh-CN")}
                </p>
              </div>
            </div>
            {selectedImage.tags.length > 0 && (
              <div className="mb-4">
                <p className="text-sm text-gray-500 mb-2">标签</p>
                <div className="flex items-center gap-2 flex-wrap">
                  {selectedImage.tags.map((tag) => (
                    <span
                      key={tag}
                      className="inline-flex items-center gap-1 px-2 py-1 bg-blue-50 text-blue-700 text-sm rounded-full"
                    >
                      <Tag className="w-3 h-3" />
                      {tag}
                    </span>
                  ))}
                </div>
              </div>
            )}
            <div className="flex justify-end gap-3">
              <a
                href={selectedImage.url}
                download
                className="flex items-center gap-2 px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg"
              >
                <Download className="w-4 h-4" />
                下载
              </a>
              <button
                onClick={() => handleDelete(selectedImage.id)}
                className="flex items-center gap-2 px-4 py-2 text-red-600 hover:bg-red-50 rounded-lg"
              >
                <Trash2 className="w-4 h-4" />
                删除
              </button>
            </div>
          </div>
        </div>
      )}

      {showUploadModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-full max-w-md mx-4">
            <h2 className="text-xl font-bold text-gray-900 mb-4">上传图片</h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  图片名称（可选）
                </label>
                <input
                  type="text"
                  value={uploadName}
                  onChange={(e) => setUploadName(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="默认使用文件名"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  选择图片
                </label>
                <input
                  ref={fileInputRef}
                  type="file"
                  accept="image/*"
                  onChange={handleUpload}
                  disabled={isUploading}
                  className="block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-lg file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100 disabled:opacity-50"
                />
              </div>
              {isUploading && (
                <div className="flex items-center gap-2 text-blue-600">
                  <Loader2 className="w-4 h-4 animate-spin" />
                  <span className="text-sm">上传中...</span>
                </div>
              )}
            </div>
            <div className="flex justify-end mt-6">
              <button
                onClick={() => {
                  setShowUploadModal(false);
                  setUploadName("");
                }}
                className="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg"
              >
                关闭
              </button>
            </div>
          </div>
        </div>
      )}

      {showAIModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-full max-w-2xl mx-4 max-h-[90vh] overflow-y-auto">
            <h2 className="text-xl font-bold text-gray-900 mb-4 flex items-center gap-2">
              <Sparkles className="w-5 h-5 text-purple-600" />
              AI 图像描述生成
            </h2>
            <div className="space-y-4">
              <div className="p-4 bg-purple-50 rounded-lg">
                <h3 className="font-medium text-purple-900 mb-2">使用说明</h3>
                <p className="text-sm text-purple-700">
                  AI 将根据您的描述生成详细的图像描述，您可以将此描述用于 DALL-E、Midjourney
                  等 AI 图像生成工具。
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  图像风格
                </label>
                <div className="flex flex-wrap gap-2">
                  {[
                    { value: "realistic", label: "写实风格" },
                    { value: "illustration", label: "插画风格" },
                    { value: "scientific", label: "科学图表" },
                    { value: "abstract", label: "抽象艺术" },
                  ].map((style) => (
                    <button
                      key={style.value}
                      onClick={() => setAiStyle(style.value)}
                      className={`px-4 py-2 rounded-lg text-sm font-medium ${
                        aiStyle === style.value
                          ? "bg-purple-600 text-white"
                          : "bg-purple-100 text-purple-700 hover:bg-purple-200"
                      }`}
                    >
                      {style.label}
                    </button>
                  ))}
                </div>
              </div>

              <textarea
                value={aiPrompt}
                onChange={(e) => setAiPrompt(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-purple-500 h-24 resize-none"
                placeholder="描述您想要生成的图像，例如：一个展示深度学习神经网络架构的科学图表..."
              />

              <button
                onClick={handleAIGenerate}
                disabled={aiLoading || !aiPrompt.trim()}
                className="w-full px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700 disabled:opacity-50 flex items-center justify-center gap-2"
              >
                {aiLoading ? (
                  <>
                    <Loader2 className="w-4 h-4 animate-spin" />
                    生成中...
                  </>
                ) : (
                  <>
                    <Wand2 className="w-4 h-4" />
                    生成图像描述
                  </>
                )}
              </button>

              {aiResult && (
                <div className="p-4 bg-gray-50 rounded-lg">
                  <div className="flex items-center justify-between mb-2">
                    <h3 className="font-medium text-gray-900">生成的描述</h3>
                    <button
                      onClick={handleCopyDescription}
                      className="flex items-center gap-1 text-sm text-gray-500 hover:text-gray-700"
                    >
                      {copied ? (
                        <>
                          <Check className="w-4 h-4 text-green-500" />
                          已复制
                        </>
                      ) : (
                        <>
                          <Copy className="w-4 h-4" />
                          复制
                        </>
                      )}
                    </button>
                  </div>
                  <p className="text-gray-700 leading-relaxed">{aiResult}</p>
                  <div className="mt-4 p-3 bg-blue-50 rounded-lg">
                    <p className="text-sm text-blue-700">
                      <strong>提示：</strong>复制上述描述，粘贴到 DALL-E、Midjourney 或 Stable
                      Diffusion 等 AI 图像生成工具中即可生成图像。
                    </p>
                  </div>
                </div>
              )}
            </div>
            <div className="flex justify-end gap-3 mt-6">
              <button
                onClick={() => {
                  setShowAIModal(false);
                  setAiResult(null);
                  setAiPrompt("");
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
