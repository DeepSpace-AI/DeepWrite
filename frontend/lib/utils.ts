import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatDate(date: string | Date): string {
  return new Date(date).toLocaleDateString("zh-CN", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

export function truncate(str: string, length: number): string {
  if (str.length <= length) return str;
  return str.slice(0, length) + "...";
}

/**
 * Strip JATS/XML tags and decode XML entities from academic article metadata.
 * JATS (Journal Article Tag Suite) XML is commonly returned by Crossref and other academic APIs.
 */
export function stripJatsXml(text: string | undefined | null): string {
  if (!text) return "";

  const XML_ENTITIES: Record<string, string> = {
    "&lt;": "<",
    "&gt;": ">",
    "&amp;": "&",
    "&quot;": '"',
    "&apos;": "'",
    "&nbsp;": " ",
  };

  const decoded = Object.entries(XML_ENTITIES).reduce(
    (acc, [entity, char]) => acc.replaceAll(entity, char),
    text
  );

  const BLOCK_TAGS = [
    "jats:p",
    "jats:sec",
    "jats:abstract",
    "jats:body",
    "p",
    "div",
    "sec",
  ];

  const withParagraphBreaks = BLOCK_TAGS.reduce(
    (acc, tag) => acc.replace(new RegExp(`</?${tag}[^>]*>`, "gi"), "\n"),
    decoded
  );

  const noTags = withParagraphBreaks.replace(/<[^>]+>/g, "");

  return noTags
    .replace(/\n{3,}/g, "\n\n")
    .replace(/[ \t]+/g, " ")
    .trim();
}
