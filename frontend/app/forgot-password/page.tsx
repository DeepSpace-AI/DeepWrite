"use client";

import { useState } from "react";
import Link from "next/link";
import { Mail, ArrowLeft, CheckCircle } from "lucide-react";

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [isSubmitted, setIsSubmitted] = useState(false);
  const [error, setError] = useState("");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setIsLoading(true);

    try {
      await new Promise((resolve) => setTimeout(resolve, 1000));
      setIsSubmitted(true);
    } catch (err: unknown) {
      const error = err as { message?: string };
      setError(error.message || "发送失败，请重试");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div
      className="min-h-screen flex items-center justify-center px-4"
      style={{ background: "var(--bg-secondary)" }}
    >
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold" style={{ color: "var(--text-primary)" }}>
            DeepWrite
          </h1>
          <p className="mt-2" style={{ color: "var(--text-secondary)" }}>
            AI辅助科研写作平台
          </p>
        </div>

        <div
          className="rounded-xl shadow-sm p-8"
          style={{
            background: "var(--bg-card)",
            border: "1px solid var(--border)",
          }}
        >
          {isSubmitted ? (
            <div className="text-center">
              <CheckCircle
                className="w-12 h-12 mx-auto mb-4"
                style={{ color: "var(--success)" }}
              />
              <h2
                className="text-xl font-semibold mb-2"
                style={{ color: "var(--text-primary)" }}
              >
                重置链接已发送
              </h2>
              <p className="mb-6" style={{ color: "var(--text-secondary)" }}>
                请检查您的邮箱 {email}，点击邮件中的链接重置密码
              </p>
              <Link
                href="/login"
                className="inline-flex items-center gap-2 font-medium"
                style={{ color: "var(--accent)" }}
              >
                <ArrowLeft className="w-4 h-4" />
                返回登录
              </Link>
            </div>
          ) : (
            <>
              <h2
                className="text-xl font-semibold mb-2"
                style={{ color: "var(--text-primary)" }}
              >
                忘记密码？
              </h2>
              <p className="mb-6" style={{ color: "var(--text-secondary)" }}>
                输入您的邮箱地址，我们将发送密码重置链接
              </p>

              {error && (
                <div
                  className="mb-4 p-3 rounded-lg text-sm"
                  style={{
                    background: "var(--error)10",
                    border: "1px solid var(--error)30",
                    color: "var(--error)",
                  }}
                >
                  {error}
                </div>
              )}

              <form onSubmit={handleSubmit} className="space-y-5">
                <div>
                  <label
                    className="block text-sm font-medium mb-1.5"
                    style={{ color: "var(--text-primary)" }}
                  >
                    邮箱地址
                  </label>
                  <div className="relative">
                    <Mail
                      className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5"
                      style={{ color: "var(--text-muted)" }}
                    />
                    <input
                      type="email"
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                      placeholder="your@email.com"
                      required
                      className="w-full pl-10 pr-4 py-2.5 rounded-lg focus:outline-none focus:ring-2"
                      style={{
                        background: "var(--bg-primary)",
                        border: "1px solid var(--border)",
                        color: "var(--text-primary)",
                      }}
                    />
                  </div>
                </div>

                <button
                  type="submit"
                  disabled={isLoading}
                  className="w-full flex items-center justify-center gap-2 py-2.5 text-white rounded-lg disabled:opacity-50 disabled:cursor-not-allowed"
                  style={{ background: "var(--accent)" }}
                >
                  {isLoading ? (
                    <div
                      className="w-5 h-5 border-2 border-t-transparent rounded-full animate-spin"
                      style={{ borderColor: "white", borderTopColor: "transparent" }}
                    />
                  ) : (
                    "发送重置链接"
                  )}
                </button>
              </form>

              <div className="mt-6 text-center">
                <Link
                  href="/login"
                  className="inline-flex items-center gap-2 text-sm"
                  style={{ color: "var(--text-secondary)" }}
                >
                  <ArrowLeft className="w-4 h-4" />
                  返回登录
                </Link>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
