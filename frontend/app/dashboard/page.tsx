"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useAuthStore } from "@/stores/auth";
import { projectApi } from "@/lib/api";
import type { Project } from "@/types";
import {
  FolderOpen,
  FileText,
  BookOpen,
  TrendingUp,
  Clock,
  Plus,
  ArrowRight,
} from "lucide-react";

export default function DashboardPage() {
  const { user } = useAuthStore();
  const [projects, setProjects] = useState<Project[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const fetchProjects = async () => {
      try {
        const response = await projectApi.list({ limit: 5 });
        setProjects(response.data.data || []);
      } catch (err) {
        console.error("Failed to fetch projects:", err);
      } finally {
        setIsLoading(false);
      }
    };
    fetchProjects();
  }, []);

  const stats = [
    { label: "进行中的项目", value: projects.filter((p) => p.status === "active").length, icon: FolderOpen, color: "bg-blue-100 text-blue-600" },
    { label: "文献数量", value: "0", icon: BookOpen, color: "bg-green-100 text-green-600" },
    { label: "论文草稿", value: "0", icon: FileText, color: "bg-purple-100 text-purple-600" },
    { label: "本月AI调用", value: "0", icon: TrendingUp, color: "bg-orange-100 text-orange-600" },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">
          欢迎回来，{user?.name || "用户"}！
        </h1>
        <p className="mt-1 text-gray-600">这是您今天的科研工作台概览</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {stats.map((stat) => (
          <div
            key={stat.label}
            className="bg-white rounded-xl border border-gray-200 p-5"
          >
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-500">{stat.label}</p>
                <p className="mt-1 text-2xl font-bold text-gray-900">{stat.value}</p>
              </div>
              <div className={`p-3 rounded-lg ${stat.color}`}>
                <stat.icon className="w-6 h-6" />
              </div>
            </div>
          </div>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 bg-white rounded-xl border border-gray-200">
          <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
            <h2 className="text-lg font-semibold text-gray-900">最近项目</h2>
            <Link
              href="/projects"
              className="text-sm text-blue-600 hover:text-blue-700 font-medium flex items-center gap-1"
            >
              查看全部
              <ArrowRight className="w-4 h-4" />
            </Link>
          </div>

          <div className="divide-y divide-gray-100">
            {isLoading ? (
              <div className="px-6 py-8 text-center text-gray-500">加载中...</div>
            ) : projects.length > 0 ? (
              projects.map((project) => (
                <Link
                  key={project.id}
                  href={`/projects/${project.id}`}
                  className="flex items-center justify-between px-6 py-4 hover:bg-gray-50 transition-colors"
                >
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 bg-blue-100 rounded-lg flex items-center justify-center">
                      <FolderOpen className="w-5 h-5 text-blue-600" />
                    </div>
                    <div>
                      <p className="font-medium text-gray-900">{project.title}</p>
                      <p className="text-sm text-gray-500">
                        {project.description || "暂无描述"}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-2 text-sm text-gray-500">
                    <Clock className="w-4 h-4" />
                    {new Date(project.created_at).toLocaleDateString("zh-CN")}
                  </div>
                </Link>
              ))
            ) : (
              <div className="px-6 py-8 text-center">
                <p className="text-gray-500">还没有项目</p>
                <Link
                  href="/projects"
                  className="mt-2 inline-flex items-center gap-1 text-blue-600 hover:text-blue-700 font-medium"
                >
                  <Plus className="w-4 h-4" />
                  创建第一个项目
                </Link>
              </div>
            )}
          </div>
        </div>

        <div className="bg-white rounded-xl border border-gray-200">
          <div className="px-6 py-4 border-b border-gray-200">
            <h2 className="text-lg font-semibold text-gray-900">快速操作</h2>
          </div>
          <div className="p-4 space-y-2">
            {[
              { label: "新建项目", href: "/projects", icon: Plus, color: "bg-blue-100 text-blue-600" },
              { label: "导入文献", href: "/references", icon: BookOpen, color: "bg-green-100 text-green-600" },
              { label: "开始写作", href: "/writing", icon: FileText, color: "bg-purple-100 text-purple-600" },
            ].map((action) => (
              <Link
                key={action.label}
                href={action.href}
                className="flex items-center gap-3 p-3 rounded-lg hover:bg-gray-50 transition-colors"
              >
                <div className={`p-2 rounded-lg ${action.color}`}>
                  <action.icon className="w-4 h-4" />
                </div>
                <span className="font-medium text-gray-700">{action.label}</span>
              </Link>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
