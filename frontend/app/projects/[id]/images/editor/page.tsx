"use client";

import { useState, useRef } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import dynamic from "next/dynamic";
import { imageApi, storageApi } from "@/lib/api";
import {
  ArrowLeft,
  Upload,
  Image as ImageIcon,
  Loader2,
  Wand2,
  Grid,
  X,
} from "lucide-react";

const CanvasEditor = dynamic(() => import("@/components/canvas-editor"), {
  ssr: false,
  loading: () => (
    <div className="flex items-center justify-center h-96">
      <Loader2 className="w-8 h-8 animate-spin text-blue-600" />
    </div>
  ),
});

interface Template {
  id: string;
  name: string;
  description: string;
  thumbnail: string;
  category: string;
  width: number;
  height: number;
  objects: any[];
}

const TEMPLATES: Template[] = [
  {
    id: "flowchart",
    name: "流程图",
    description: "用于展示工作流程或算法步骤",
    thumbnail: "📊",
    category: "diagram",
    width: 800,
    height: 600,
    objects: [
      { type: "rect", left: 300, top: 50, width: 200, height: 60, fill: "#3B82F6", stroke: "#2563EB", text: "开始" },
      { type: "line", x1: 400, y1: 110, x2: 400, y2: 160, stroke: "#6B7280" },
      { type: "diamond", left: 300, top: 160, width: 200, height: 100, fill: "#F59E0B", stroke: "#D97706", text: "条件判断" },
      { type: "line", x1: 400, y1: 260, x2: 400, y2: 310, stroke: "#6B7280" },
      { type: "rect", left: 300, top: 310, width: 200, height: 60, fill: "#10B981", stroke: "#059669", text: "处理" },
      { type: "line", x1: 400, y1: 370, x2: 400, y2: 420, stroke: "#6B7280" },
      { type: "rect", left: 300, top: 420, width: 200, height: 60, fill: "#3B82F6", stroke: "#2563EB", text: "结束" },
    ],
  },
  {
    id: "architecture",
    name: "系统架构图",
    description: "用于展示系统组件和交互",
    thumbnail: "🏗️",
    category: "diagram",
    width: 1000,
    height: 700,
    objects: [
      { type: "rect", left: 50, top: 50, width: 200, height: 100, fill: "#3B82F6", stroke: "#2563EB", text: "前端" },
      { type: "rect", left: 400, top: 50, width: 200, height: 100, fill: "#8B5CF6", stroke: "#7C3AED", text: "API 网关" },
      { type: "rect", left: 750, top: 50, width: 200, height: 100, fill: "#10B981", stroke: "#059669", text: "数据库" },
      { type: "rect", left: 50, top: 250, width: 200, height: 100, fill: "#F59E0B", stroke: "#D97706", text: "用户服务" },
      { type: "rect", left: 400, top: 250, width: 200, height: 100, fill: "#F59E0B", stroke: "#D97706", text: "项目服务" },
      { type: "rect", left: 750, top: 250, width: 200, height: 100, fill: "#F59E0B", stroke: "#D97706", text: "AI 服务" },
    ],
  },
  {
    id: "bar-chart",
    name: "柱状图模板",
    description: "用于数据可视化",
    thumbnail: "📊",
    category: "chart",
    width: 600,
    height: 400,
    objects: [
      { type: "rect", left: 50, top: 300, width: 80, height: 100, fill: "#3B82F6" },
      { type: "rect", left: 150, top: 200, width: 80, height: 200, fill: "#10B981" },
      { type: "rect", left: 250, top: 250, width: 80, height: 150, fill: "#F59E0B" },
      { type: "rect", left: 350, top: 150, width: 80, height: 250, fill: "#EF4444" },
      { type: "rect", left: 450, top: 100, width: 80, height: 300, fill: "#8B5CF6" },
    ],
  },
  {
    id: "neural-network",
    name: "神经网络图",
    description: "用于展示深度学习模型结构",
    thumbnail: "🧠",
    category: "scientific",
    width: 800,
    height: 500,
    objects: [
      { type: "circle", left: 100, top: 100, radius: 30, fill: "#3B82F6", stroke: "#2563EB" },
      { type: "circle", left: 100, top: 200, radius: 30, fill: "#3B82F6", stroke: "#2563EB" },
      { type: "circle", left: 100, top: 300, radius: 30, fill: "#3B82F6", stroke: "#2563EB" },
      { type: "circle", left: 350, top: 100, radius: 30, fill: "#10B981", stroke: "#059669" },
      { type: "circle", left: 350, top: 200, radius: 30, fill: "#10B981", stroke: "#059669" },
      { type: "circle", left: 350, top: 300, radius: 30, fill: "#10B981", stroke: "#059669" },
      { type: "circle", left: 600, top: 150, radius: 30, fill: "#F59E0B", stroke: "#D97706" },
      { type: "circle", left: 600, top: 250, radius: 30, fill: "#F59E0B", stroke: "#D97706" },
    ],
  },
  {
    id: "sequence",
    name: "时序图",
    description: "用于展示对象间的交互顺序",
    thumbnail: "🔄",
    category: "diagram",
    width: 900,
    height: 600,
    objects: [
      { type: "rect", left: 50, top: 50, width: 120, height: 40, fill: "#3B82F6", text: "用户" },
      { type: "rect", left: 350, top: 50, width: 120, height: 40, fill: "#10B981", text: "服务器" },
      { type: "rect", left: 650, top: 50, width: 120, height: 40, fill: "#F59E0B", text: "数据库" },
    ],
  },
  {
    id: "pie-chart",
    name: "饼图模板",
    description: "用于展示占比关系",
    thumbnail: "🥧",
    category: "chart",
    width: 500,
    height: 500,
    objects: [
      { type: "circle", left: 150, top: 100, radius: 150, fill: "#3B82F6", startAngle: 0, endAngle: 120 },
      { type: "circle", left: 150, top: 100, radius: 150, fill: "#10B981", startAngle: 120, endAngle: 240 },
      { type: "circle", left: 150, top: 100, radius: 150, fill: "#F59E0B", startAngle: 240, endAngle: 360 },
    ],
  },
];

export default function ImageEditorPage() {
  const params = useParams();
  const router = useRouter();
  const projectId = params.id as string;
  const [showTemplates, setShowTemplates] = useState(false);
  const [currentImage, setCurrentImage] = useState<string | null>(null);
  const [isSaving, setIsSaving] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleImageUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (event) => {
      setCurrentImage(event.target?.result as string);
    };
    reader.readAsDataURL(file);
  };

  const handleTemplateSelect = (template: Template) => {
    setCurrentImage(null);
    setShowTemplates(false);
  };

  const handleSave = async (dataUrl: string) => {
    setIsSaving(true);
    try {
      const blob = await fetch(dataUrl).then((r) => r.blob());
      const file = new File([blob], "canvas-export.png", { type: "image/png" });
      await imageApi.upload(projectId, file, "Canvas Drawing");
      alert("Image saved successfully!");
    } catch (err) {
      console.error("Failed to save image:", err);
      alert("Failed to save image");
    } finally {
      setIsSaving(false);
    }
  };

  const handleExport = (format: string, dataUrl: string) => {
    console.log(`Exported as ${format}`);
  };

  return (
    <div className="h-[calc(100vh-4rem)] flex flex-col">
      <div className="flex items-center gap-4 px-6 py-3 border-b border-gray-200 bg-white">
        <Link
          href={`/projects/${projectId}/images`}
          className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
        >
          <ArrowLeft className="w-5 h-5 text-gray-600" />
        </Link>
        <div className="flex-1">
          <h1 className="text-xl font-bold text-gray-900">图像编辑器</h1>
          <p className="text-sm text-gray-500">使用 Canvas 编辑器创建和编辑图像</p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setShowTemplates(true)}
            className="flex items-center gap-2 px-4 py-2 bg-white border border-gray-200 rounded-lg hover:bg-gray-50"
          >
            <Grid className="w-4 h-4" />
            模板
          </button>
          <button
            onClick={() => fileInputRef.current?.click()}
            className="flex items-center gap-2 px-4 py-2 bg-white border border-gray-200 rounded-lg hover:bg-gray-50"
          >
            <Upload className="w-4 h-4" />
            上传图片
          </button>
          <input
            ref={fileInputRef}
            type="file"
            accept="image/*"
            onChange={handleImageUpload}
            className="hidden"
          />
        </div>
      </div>

      <div className="flex-1 overflow-hidden">
        <CanvasEditor
          width={800}
          height={600}
          initialImage={currentImage || undefined}
          onSave={handleSave}
          onExport={handleExport}
        />
      </div>

      {showTemplates && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-full max-w-4xl mx-4 max-h-[90vh] overflow-y-auto">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl font-bold text-gray-900">图像模板</h2>
              <button
                onClick={() => setShowTemplates(false)}
                className="p-2 hover:bg-gray-100 rounded-lg"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
            <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
              {TEMPLATES.map((template) => (
                <div
                  key={template.id}
                  className="bg-gray-50 rounded-xl p-4 cursor-pointer hover:shadow-md transition-shadow border border-gray-200"
                  onClick={() => handleTemplateSelect(template)}
                >
                  <div className="aspect-video bg-white rounded-lg flex items-center justify-center mb-3 text-4xl">
                    {template.thumbnail}
                  </div>
                  <h3 className="font-semibold text-gray-900">{template.name}</h3>
                  <p className="text-sm text-gray-500 mt-1">{template.description}</p>
                  <div className="flex items-center gap-2 mt-2">
                    <span className="text-xs px-2 py-0.5 bg-blue-100 text-blue-700 rounded">
                      {template.category}
                    </span>
                    <span className="text-xs text-gray-400">
                      {template.width} × {template.height}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {isSaving && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 flex items-center gap-3">
            <Loader2 className="w-6 h-6 animate-spin text-blue-600" />
            <span>保存中...</span>
          </div>
        </div>
      )}
    </div>
  );
}
