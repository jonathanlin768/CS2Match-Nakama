"use client";

import { Book, Database, GraduationCap, FileText, MessageCircle, MessageSquare, HelpCircle } from "lucide-react";

const leftLinks = [
  { icon: Book, label: "新手指南" },
  { icon: Database, label: "数据百科" },
  { icon: GraduationCap, label: "战术学院" },
  { icon: FileText, label: "更新日志" },
];

const rightLinks = [
  { icon: MessageCircle, label: "社区" },
  { icon: MessageSquare, label: "反馈建议" },
  { icon: HelpCircle, label: "帮助中心" },
];

export function Footer() {
  return (
    <footer className="bg-card border-t border-border">
      <div className="max-w-[1400px] mx-auto flex flex-col sm:flex-row items-center justify-between px-4 lg:px-6 py-3 gap-3">
        {/* Left Links */}
        <div className="flex items-center gap-3 sm:gap-4 lg:gap-6 flex-wrap justify-center sm:justify-start">
          {leftLinks.map((link) => (
            <button
              key={link.label}
              className="flex items-center gap-1.5 lg:gap-2 text-xs lg:text-sm text-muted-foreground hover:text-foreground transition-colors"
            >
              <link.icon className="w-3.5 lg:w-4 h-3.5 lg:h-4" />
              <span className="hidden sm:inline">{link.label}</span>
            </button>
          ))}
        </div>

        {/* Right Links */}
        <div className="flex items-center gap-3 sm:gap-4 lg:gap-6">
          {rightLinks.map((link) => (
            <button
              key={link.label}
              className="flex items-center gap-1.5 lg:gap-2 text-xs lg:text-sm text-muted-foreground hover:text-foreground transition-colors"
            >
              <link.icon className="w-3.5 lg:w-4 h-3.5 lg:h-4" />
              <span className="hidden sm:inline">{link.label}</span>
            </button>
          ))}
        </div>
      </div>
    </footer>
  );
}
