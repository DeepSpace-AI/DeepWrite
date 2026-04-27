import axios, { AxiosError, type InternalAxiosRequestConfig } from "axios";
import { useAuthStore } from "@/stores/auth";

const API_BASE_URL = "";

export const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    "Content-Type": "application/json",
  },
  timeout: 10000,
});

api.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = useAuthStore.getState().accessToken;
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

api.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    if (error.response?.status === 401) {
      const isOnLoginPage = typeof window !== "undefined" && window.location.pathname === "/login";
      if (!isOnLoginPage) {
        useAuthStore.getState().logout();
        if (typeof window !== "undefined") {
          window.location.href = "/login";
        }
      }
    }
    return Promise.reject(error);
  }
);

export const authApi = {
  register: (data: { email: string; password: string; name: string }) =>
    api.post("/api/users/register", data),
  login: (data: { email: string; password: string }) =>
    api.post("/api/users/login", data),
  getMe: () => api.get("/api/users/me"),
  updateMe: (data: { name?: string; institution?: string }) =>
    api.put("/api/users/me", data),
};

export const teamApi = {
  list: () => api.get("/api/teams"),
  create: (data: { name: string; description?: string }) =>
    api.post("/api/teams", data),
  get: (id: string) => api.get(`/api/teams/${id}`),
  update: (id: string, data: Partial<{ name: string; description: string }>) =>
    api.put(`/api/teams/${id}`, data),
  delete: (id: string) => api.delete(`/api/teams/${id}`),
};

export const projectApi = {
  list: (params?: { limit?: number; offset?: number }) =>
    api.get("/api/projects", { params }),
  create: (data: { title: string; description?: string; research_field?: string }) =>
    api.post("/api/projects", data),
  get: (id: string) => api.get(`/api/projects/${id}`),
  update: (id: string, data: Partial<{ title: string; description: string; status: string }>) =>
    api.put(`/api/projects/${id}`, data),
  delete: (id: string) => api.delete(`/api/projects/${id}`),
  archive: (id: string) => api.post(`/api/projects/${id}/archive`),
  unarchive: (id: string) => api.post(`/api/projects/${id}/unarchive`),
};

export const documentApi = {
  list: (params: { project_id: string; status?: string; page?: number; limit?: number }) =>
    api.get("/api/documents", { params }),
  create: (data: { project_id: string; title: string; abstract?: string; format?: string }) =>
    api.post("/api/documents", data),
  get: (id: string, includeContent?: boolean) =>
    api.get(`/api/documents/${id}`, { params: { include_content: includeContent } }),
  update: (id: string, data: Partial<{ title: string; abstract: string; status: string }>) =>
    api.put(`/api/documents/${id}`, data),
  updateContent: (id: string, data: { content: Record<string, unknown>; metadata?: { word_count: number; format: string }; citations?: unknown[] }) =>
    api.put(`/api/documents/${id}/content`, data),
  delete: (id: string) => api.delete(`/api/documents/${id}`),
  createVersion: (id: string, data: { change_summary: string }) =>
    api.post(`/api/documents/${id}/versions`, data),
  listVersions: (id: string, params?: { page?: number; limit?: number }) =>
    api.get(`/api/documents/${id}/versions`, { params }),
};

export const referenceApi = {
  list: (params: { project_id: string; page?: number; limit?: number; search?: string; tags?: string }) =>
    api.get("/api/references", { params }),
  create: (data: { project_id: string; title: string; authors?: string[]; year?: number; journal?: string; doi?: string; abstract?: string; tags?: string[] }) =>
    api.post("/api/references", data),
  get: (id: string) => api.get(`/api/references/${id}`),
  update: (id: string, data: Partial<{ title: string; authors: string[]; year: number; journal: string; doi: string; abstract: string; tags: string[] }>) =>
    api.put(`/api/references/${id}`, data),
  delete: (id: string) => api.delete(`/api/references/${id}`),
  importDOI: (data: { project_id: string; doi: string }) =>
    api.post("/api/references/import/doi", data),
  importBibTeX: (projectId: string, file: File) => {
    const formData = new FormData();
    formData.append("file", file);
    return api.post(`/api/references/import/bibtex?project_id=${projectId}`, formData, {
      headers: { "Content-Type": "multipart/form-data" },
    });
  },
  getCitation: (id: string, style: string) =>
    api.post(`/api/references/${id}/citation`, { style }),
  searchExternal: (query: string, source: string = "crossref") =>
    api.get("/api/references/search/external", { params: { query, source } }),
};

export const aiApi = {
  chat: (data: { message: string; context?: string; model?: string }) =>
    api.post("/api/ai/chat", data),
  write: (data: { topic: string; section?: string; style?: string; word_count?: number; references?: string[] }) =>
    api.post("/api/ai/academic/write", data),
  polish: (data: { text: string; style?: string; focus?: string }) =>
    api.post("/api/ai/academic/polish", data),
  analyzeReference: (data: { title?: string; abstract?: string; text?: string }) =>
    api.post("/api/ai/references/analyze", data),
};

export const codeApi = {
  listFiles: (params: { project_id: string; page?: number; limit?: number }) =>
    api.get("/api/code/files", { params }),
  createFile: (data: { project_id: string; name: string; language: string; content?: string }) =>
    api.post("/api/code/files", data),
  getFile: (id: string) => api.get(`/api/code/files/${id}`),
  updateFile: (id: string, data: Partial<{ name: string; content: string }>) =>
    api.put(`/api/code/files/${id}`, data),
  deleteFile: (id: string) => api.delete(`/api/code/files/${id}`),
  runCode: (data: { file_id: string; timeout?: number }) =>
    api.post("/api/code/run", data),
  getRunStatus: (runId: string) => api.get(`/api/code/runs/${runId}`),
};

export const imageApi = {
  list: (params: { project_id: string; page?: number; limit?: number }) =>
    api.get("/api/images", { params }),
  upload: (projectId: string, file: File, name?: string) => {
    const formData = new FormData();
    formData.append("file", file);
    return api.post(`/api/images/upload?project_id=${projectId}${name ? `&name=${name}` : ""}`, formData, {
      headers: { "Content-Type": "multipart/form-data" },
    });
  },
  get: (id: string) => api.get(`/api/images/${id}`),
  update: (id: string, data: Partial<{ name: string; description: string; tags: string[] }>) =>
    api.put(`/api/images/${id}`, data),
  delete: (id: string) => api.delete(`/api/images/${id}`),
};

export const storageApi = {
  upload: (bucket: string, projectId: string, file: File) => {
    const formData = new FormData();
    formData.append("file", file);
    return api.post(`/api/storage/upload?bucket=${bucket}&project_id=${projectId}`, formData, {
      headers: { "Content-Type": "multipart/form-data" },
    });
  },
  download: (bucket: string, objectName: string) =>
    api.get(`/api/storage/${bucket}/${objectName}`, { responseType: "blob" }),
  delete: (bucket: string, objectName: string) =>
    api.delete(`/api/storage/${bucket}/${objectName}`),
  list: (params: { bucket: string; project_id: string; prefix?: string }) =>
    api.get("/api/storage/list", { params }),
};
