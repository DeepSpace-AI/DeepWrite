"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { projectApi } from "@/lib/api";
import type { Project } from "@/types";
import {
  FolderOpen,
  Plus,
  Search,
  Filter,
  MoreVertical,
  Users,
  Clock,
  Trash2,
} from "lucide-react";

export default function ProjectsPage() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState("");
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newProject, setNewProject] = useState({ title: "", description: "" });
  const [openMenuId, setOpenMenuId] = useState<string | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Project | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  useEffect(() => {
    fetchProjects();
  }, []);

  const fetchProjects = async () => {
    try {
      const response = await projectApi.list();
      setProjects(response.data.data || []);
    } catch (err) {
      console.error("Failed to fetch projects:", err);
    } finally {
      setIsLoading(false);
    }
  };

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await projectApi.create(newProject);
      setShowCreateModal(false);
      setNewProject({ title: "", description: "" });
      fetchProjects();
    } catch (err) {
      console.error("Failed to create project:", err);
    }
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    setIsDeleting(true);
    try {
      await projectApi.delete(deleteTarget.id);
      setDeleteTarget(null);
      fetchProjects();
    } catch (err) {
      console.error("Failed to delete project:", err);
      alert("删除项目失败，请重试");
    } finally {
      setIsDeleting(false);
    }
  };

  useEffect(() => {
    const handleClickOutside = () => setOpenMenuId(null);
    if (openMenuId) {
      document.addEventListener("click", handleClickOutside);
      return () => document.removeEventListener("click", handleClickOutside);
    }
  }, [openMenuId]);

  const filteredProjects = projects.filter(
    (p) =>
      p.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
      p.description?.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold" style={{ color: "var(--text-primary)" }}>
            项目管理
          </h1>
          <p className="mt-1" style={{ color: "var(--text-secondary)" }}>
            管理您的科研项目和团队协作
          </p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="flex items-center gap-2 px-4 py-2 text-white rounded-lg"
          style={{ background: "var(--accent)" }}
        >
          <Plus className="w-4 h-4" />
          新建项目
        </button>
      </div>

      <div className="flex items-center gap-4">
        <div className="relative flex-1 max-w-md">
          <Search
            className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4"
            style={{ color: "var(--text-muted)" }}
          />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="搜索项目..."
            className="w-full pl-10 pr-4 py-2 rounded-lg focus:outline-none focus:ring-2"
            style={{
              background: "var(--bg-card)",
              border: "1px solid var(--border)",
              color: "var(--text-primary)",
            }}
          />
        </div>
        <button
          className="flex items-center gap-2 px-4 py-2 rounded-lg"
          style={{
            background: "var(--bg-card)",
            border: "1px solid var(--border)",
            color: "var(--text-secondary)",
          }}
        >
          <Filter className="w-4 h-4" />
          筛选
        </button>
      </div>

      {isLoading ? (
        <div className="text-center py-12" style={{ color: "var(--text-muted)" }}>
          加载中...
        </div>
      ) : filteredProjects.length > 0 ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {filteredProjects.map((project) => (
            <Link
              key={project.id}
              href={`/projects/${project.id}`}
              className="group rounded-xl p-5 transition-all"
              style={{
                background: "var(--bg-card)",
                border: "1px solid var(--border)",
              }}
            >
              <div className="flex items-start justify-between">
                <div
                  className="w-10 h-10 rounded-lg flex items-center justify-center"
                  style={{
                    background: "var(--accent-light)",
                    color: "var(--accent)",
                  }}
                >
                  <FolderOpen className="w-5 h-5" />
                </div>
                <div className="relative">
                  <button
                    onClick={(e) => {
                      e.preventDefault();
                      e.stopPropagation();
                      setOpenMenuId(openMenuId === project.id ? null : project.id);
                    }}
                    className="opacity-0 group-hover:opacity-100 p-1 rounded transition-opacity"
                    style={{ color: "var(--text-muted)" }}
                  >
                    <MoreVertical className="w-4 h-4" />
                  </button>
                  {openMenuId === project.id && (
                    <div
                      className="absolute right-0 top-full mt-1 w-32 rounded-lg shadow-lg z-10 py-1"
                      style={{
                        background: "var(--bg-card)",
                        border: "1px solid var(--border)",
                      }}
                      onClick={(e) => e.stopPropagation()}
                    >
                      <button
                        onClick={(e) => {
                          e.preventDefault();
                          e.stopPropagation();
                          setDeleteTarget(project);
                          setOpenMenuId(null);
                        }}
                        className="w-full flex items-center gap-2 px-3 py-2 text-sm text-left transition-colors hover:bg-red-50 dark:hover:bg-red-950"
                        style={{ color: "var(--error, #ef4444)" }}
                      >
                        <Trash2 className="w-4 h-4" />
                        删除项目
                      </button>
                    </div>
                  )}
                </div>
              </div>

              <h3
                className="mt-3 font-semibold"
                style={{ color: "var(--text-primary)" }}
              >
                {project.title}
              </h3>
              <p
                className="mt-1 text-sm line-clamp-2"
                style={{ color: "var(--text-secondary)" }}
              >
                {project.description || "暂无描述"}
              </p>

              <div
                className="mt-4 flex items-center gap-4 text-sm"
                style={{ color: "var(--text-muted)" }}
              >
                <span className="flex items-center gap-1">
                  <Clock className="w-3.5 h-3.5" />
                  {new Date(project.created_at).toLocaleDateString("zh-CN")}
                </span>
                <span className="flex items-center gap-1">
                  <Users className="w-3.5 h-3.5" />
                  1 成员
                </span>
              </div>

              <div className="mt-3">
                <span
                  className="inline-flex px-2 py-1 text-xs font-medium rounded-full"
                  style={{
                    background: project.status === "active" ? "var(--success)20" : "var(--bg-secondary)",
                    color: project.status === "active" ? "var(--success)" : "var(--text-secondary)",
                  }}
                >
                  {project.status === "active" ? "进行中" : project.status}
                </span>
              </div>
            </Link>
          ))}
        </div>
      ) : (
        <div
          className="text-center py-12 rounded-xl"
          style={{
            background: "var(--bg-card)",
            border: "1px solid var(--border)",
          }}
        >
          <FolderOpen
            className="w-12 h-12 mx-auto mb-4"
            style={{ color: "var(--text-muted)" }}
          />
          <p style={{ color: "var(--text-secondary)" }}>
            {searchQuery ? "未找到匹配的项目" : "还没有项目，创建一个吧！"}
          </p>
          {!searchQuery && (
            <button
              onClick={() => setShowCreateModal(true)}
              className="mt-4 inline-flex items-center gap-2 px-4 py-2 text-white rounded-lg"
              style={{ background: "var(--accent)" }}
            >
              <Plus className="w-4 h-4" />
              创建项目
            </button>
          )}
        </div>
      )}

      {deleteTarget && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div
            className="rounded-xl shadow-xl w-full max-w-sm mx-4"
            style={{ background: "var(--bg-card)" }}
          >
            <div className="p-6 space-y-4">
              <h3
                className="text-lg font-semibold"
                style={{ color: "var(--text-primary)" }}
              >
                确认删除
              </h3>
              <p style={{ color: "var(--text-secondary)" }}>
                确定要删除项目「{deleteTarget.title}」吗？此操作不可恢复。
              </p>
              <div className="flex justify-end gap-3">
                <button
                  onClick={() => setDeleteTarget(null)}
                  className="px-4 py-2 rounded-lg"
                  style={{ color: "var(--text-secondary)" }}
                  disabled={isDeleting}
                >
                  取消
                </button>
                <button
                  onClick={handleDelete}
                  className="px-4 py-2 text-white rounded-lg"
                  style={{ background: "var(--error, #ef4444)" }}
                  disabled={isDeleting}
                >
                  {isDeleting ? "删除中..." : "删除"}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {showCreateModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div
            className="rounded-xl shadow-xl w-full max-w-md mx-4"
            style={{ background: "var(--bg-card)" }}
          >
            <div
              className="flex items-center justify-between px-6 py-4"
              style={{ borderBottom: "1px solid var(--border)" }}
            >
              <h2
                className="text-lg font-semibold"
                style={{ color: "var(--text-primary)" }}
              >
                新建项目
              </h2>
              <button
                onClick={() => setShowCreateModal(false)}
                style={{ color: "var(--text-muted)" }}
              >
                ×
              </button>
            </div>
            <form onSubmit={handleCreate} className="p-6 space-y-4">
              <div>
                <label
                  className="block text-sm font-medium mb-1"
                  style={{ color: "var(--text-primary)" }}
                >
                  项目名称
                </label>
                <input
                  type="text"
                  value={newProject.title}
                  onChange={(e) =>
                    setNewProject({ ...newProject, title: e.target.value })
                  }
                  placeholder="输入项目名称"
                  required
                  className="w-full px-4 py-2 rounded-lg focus:outline-none focus:ring-2"
                  style={{
                    background: "var(--bg-primary)",
                    border: "1px solid var(--border)",
                    color: "var(--text-primary)",
                  }}
                />
              </div>
              <div>
                <label
                  className="block text-sm font-medium mb-1"
                  style={{ color: "var(--text-primary)" }}
                >
                  项目描述
                </label>
                <textarea
                  value={newProject.description}
                  onChange={(e) =>
                    setNewProject({ ...newProject, description: e.target.value })
                  }
                  placeholder="简要描述项目内容..."
                  rows={3}
                  className="w-full px-4 py-2 rounded-lg focus:outline-none focus:ring-2 resize-none"
                  style={{
                    background: "var(--bg-primary)",
                    border: "1px solid var(--border)",
                    color: "var(--text-primary)",
                  }}
                />
              </div>
              <div className="flex justify-end gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  className="px-4 py-2 rounded-lg"
                  style={{ color: "var(--text-secondary)" }}
                >
                  取消
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 text-white rounded-lg"
                  style={{ background: "var(--accent)" }}
                >
                  创建
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
