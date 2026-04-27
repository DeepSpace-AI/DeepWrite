"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { projectApi } from "@/lib/api";
import type { Project } from "@/types";
import {
  ArrowLeft,
  FolderOpen,
  Edit,
  Archive,
  Users,
  Clock,
  FileText,
  BookOpen,
  Code,
  Image,
  MoreVertical,
} from "lucide-react";

export default function ProjectDetailPage() {
  const params = useParams();
  const projectId = params.id as string;
  const [project, setProject] = useState<Project | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const fetchProject = async () => {
      try {
        const response = await projectApi.get(projectId);
        setProject(response.data.data);
      } catch (err) {
        console.error("Failed to fetch project:", err);
      } finally {
        setIsLoading(false);
      }
    };
    fetchProject();
  }, [projectId]);

  if (isLoading) {
    return <div className="text-center py-12">加载中...</div>;
  }

  if (!project) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500">项目不存在或已被删除</p>
        <Link href="/projects" className="mt-4 text-blue-600 hover:underline">
          返回项目列表
        </Link>
      </div>
    );
  }

  const tabs = [
    { label: "概览", href: `#overview`, icon: FolderOpen },
    { label: "文献", href: `#references`, icon: BookOpen },
    { label: "论文", href: `#documents`, icon: FileText },
    { label: "成员", href: `#members`, icon: Users },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link
          href="/projects"
          className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
        >
          <ArrowLeft className="w-5 h-5 text-gray-600" />
        </Link>
        <div className="flex-1">
          <h1 className="text-2xl font-bold text-gray-900">{project.title}</h1>
          <p className="mt-1 text-gray-600">{project.description || "暂无描述"}</p>
        </div>
        <div className="flex items-center gap-2">
          <button className="flex items-center gap-2 px-4 py-2 bg-white border border-gray-200 rounded-lg hover:bg-gray-50">
            <Edit className="w-4 h-4" />
            编辑
          </button>
          <button className="flex items-center gap-2 px-4 py-2 bg-white border border-gray-200 rounded-lg hover:bg-gray-50 text-red-600">
            <Archive className="w-4 h-4" />
            归档
          </button>
          <button className="p-2 hover:bg-gray-100 rounded-lg">
            <MoreVertical className="w-5 h-5 text-gray-600" />
          </button>
        </div>
      </div>

      <div className="flex items-center gap-6 text-sm text-gray-500">
        <span className="flex items-center gap-1">
          <Clock className="w-4 h-4" />
          创建于 {new Date(project.created_at).toLocaleDateString("zh-CN")}
        </span>
        <span className="flex items-center gap-1">
          <Users className="w-4 h-4" />
          1 成员
        </span>
        <span
          className={`inline-flex px-2 py-0.5 text-xs font-medium rounded-full ${
            project.status === "active"
              ? "bg-green-100 text-green-700"
              : "bg-gray-100 text-gray-700"
          }`}
        >
          {project.status === "active" ? "进行中" : project.status}
        </span>
      </div>

      <div className="border-b border-gray-200">
        <nav className="flex gap-1">
          {tabs.map((tab) => (
            <button
              key={tab.label}
              className="flex items-center gap-2 px-4 py-3 text-sm font-medium text-gray-600 hover:text-gray-900 border-b-2 border-transparent hover:border-gray-300 transition-colors"
            >
              <tab.icon className="w-4 h-4" />
              {tab.label}
            </button>
          ))}
        </nav>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 space-y-6">
          <div className="bg-white rounded-xl border border-gray-200 p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">项目概览</h2>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              {[
                { label: "文献", value: "0", icon: BookOpen },
                { label: "论文", value: "0", icon: FileText },
                { label: "代码文件", value: "0", icon: FolderOpen },
                { label: "图像", value: "0", icon: FolderOpen },
              ].map((stat) => (
                <div key={stat.label} className="text-center p-4 bg-gray-50 rounded-lg">
                  <stat.icon className="w-6 h-6 text-gray-400 mx-auto mb-2" />
                  <p className="text-2xl font-bold text-gray-900">{stat.value}</p>
                  <p className="text-sm text-gray-500">{stat.label}</p>
                </div>
              ))}
            </div>
          </div>

          <div className="bg-white rounded-xl border border-gray-200 p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">最近活动</h2>
            <div className="text-center py-8 text-gray-500">
              暂无活动记录
            </div>
          </div>
        </div>

        <div className="space-y-6">
          <div className="bg-white rounded-xl border border-gray-200 p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">快速操作</h2>
            <div className="space-y-2">
              {[
                { label: "管理文献", icon: BookOpen, href: `/projects/${projectId}/references` },
                { label: "论文写作", icon: FileText, href: `/projects/${projectId}/writing` },
                { label: "代码编辑", icon: Code, href: `/projects/${projectId}/code` },
                { label: "图像管理", icon: Image, href: `/projects/${projectId}/images` },
              ].map((action) => (
                <Link
                  key={action.label}
                  href={action.href}
                  className="w-full flex items-center gap-3 p-3 text-left hover:bg-gray-50 rounded-lg transition-colors"
                >
                  <action.icon className="w-5 h-5 text-gray-400" />
                  <span className="font-medium text-gray-700">{action.label}</span>
                </Link>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
