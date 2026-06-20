"use client";

import { useState } from "react";
import { Link, NavLink, useLocation } from "react-router-dom";
import { Coins, Gem, Bell, Settings, Mail, User, Menu, X } from "lucide-react";
import { useAuth } from "@/context/AuthContext";

const navItems = [
  { label: "首页", href: "/home" },
  { label: "对战", href: "/match" },
  { label: "俱乐部", href: "#" },
  { label: "市场", href: "#" },
  { label: "抽卡", href: "/gacha" },
  { label: "任务", href: "#" },
  { label: "排行", href: "/ranking" },
];

export function Header() {
  const { session } = useAuth();
  const username = session?.username ?? null;
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const { pathname } = useLocation();

  const isActive = (href: string) => {
    if (href === "/home") return pathname === "/home";
    return pathname.startsWith(href);
  };

  return (
    <header className="bg-card border-b border-border">
      <div className="max-w-[1400px] mx-auto flex items-center justify-between px-4 lg:px-0 py-3">
        {/* Logo */}
        <div className="flex items-center gap-4 lg:gap-8">
          <Link to="/home" className="flex items-center gap-2">
            <div className="text-xl lg:text-2xl font-bold text-primary">CS2</div>
            <div className="text-[10px] lg:text-xs text-muted-foreground">SIMULATOR</div>
          </Link>

          {/* Navigation - Desktop */}
          <nav className="hidden lg:flex items-center gap-6">
            {navItems.map((item) => (
              <NavLink
                key={item.label}
                to={item.href}
                className={({ isActive: active }) =>
                  `text-sm font-medium transition-colors ${
                    active || isActive(item.href)
                      ? "text-foreground"
                      : "text-muted-foreground hover:text-foreground"
                  }`
                }
              >
                {item.label}
              </NavLink>
            ))}
          </nav>
        </div>

        {/* User Info */}
        <div className="flex items-center gap-3 lg:gap-6">
          {/* Currency - Simplified on mobile */}
          <div className="flex items-center gap-2 lg:gap-4">
            <div className="flex items-center gap-1">
              <Coins className="w-4 h-4 text-yellow-500" />
              <span className="text-xs lg:text-sm font-medium hidden sm:inline">23,568</span>
            </div>
            <div className="hidden sm:flex items-center gap-1">
              <Gem className="w-4 h-4 text-pink-500" />
              <span className="text-xs lg:text-sm font-medium">1,250</span>
            </div>
            <Link to="/profile/me" className="hidden md:flex items-center gap-1.5 hover:opacity-80 transition-opacity" title="个人中心">
              <div className="w-5 h-5 rounded-full bg-gradient-to-br from-blue-500 to-purple-500 flex items-center justify-center">
                <User className="w-3 h-3 text-white" />
              </div>
              <span className="text-xs lg:text-sm font-medium max-w-[80px] truncate">
                {username ?? "玩家"}
              </span>
            </Link>
          </div>

          {/* Icons - Hidden on small mobile */}
          <div className="hidden sm:flex items-center gap-1 lg:gap-3">
            <button className="p-1.5 lg:p-2 rounded hover:bg-secondary transition-colors">
              <Mail className="w-4 lg:w-5 h-4 lg:h-5 text-muted-foreground" />
            </button>
            <button className="p-1.5 lg:p-2 rounded hover:bg-secondary transition-colors">
              <Bell className="w-4 lg:w-5 h-4 lg:h-5 text-muted-foreground" />
            </button>
            <button className="p-1.5 lg:p-2 rounded hover:bg-secondary transition-colors">
              <Settings className="w-4 lg:w-5 h-4 lg:h-5 text-muted-foreground" />
            </button>
          </div>

          {/* Mobile Menu Button */}
          <button
            className="lg:hidden p-2 rounded hover:bg-secondary transition-colors"
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
          >
            {mobileMenuOpen ? (
              <X className="w-5 h-5 text-muted-foreground" />
            ) : (
              <Menu className="w-5 h-5 text-muted-foreground" />
            )}
          </button>
        </div>
      </div>

      {/* Mobile Navigation Menu */}
      {mobileMenuOpen && (
        <div className="lg:hidden border-t border-border bg-card">
          <nav className="flex flex-col p-4 gap-2">
            {navItems.map((item) => (
              <NavLink
                key={item.label}
                to={item.href}
                onClick={() => setMobileMenuOpen(false)}
                className={({ isActive: active }) =>
                  `text-left py-2 px-3 rounded text-sm font-medium transition-colors ${
                    active || isActive(item.href)
                      ? "text-foreground bg-secondary"
                      : "text-muted-foreground hover:text-foreground hover:bg-secondary/50"
                  }`
                }
              >
                {item.label}
              </NavLink>
            ))}
          </nav>
          
          {/* Mobile-only profile link + icons */}
          <div className="sm:hidden flex items-center justify-around py-3 border-t border-border">
            <Link
              to="/profile"
              onClick={() => setMobileMenuOpen(false)}
              className="flex flex-col items-center gap-1 text-muted-foreground hover:text-foreground transition-colors"
            >
              <User className="w-5 h-5" />
              <span className="text-[10px]">我的</span>
            </Link>
            <button className="p-2 rounded hover:bg-secondary transition-colors">
              <Mail className="w-5 h-5 text-muted-foreground" />
            </button>
            <button className="p-2 rounded hover:bg-secondary transition-colors">
              <Bell className="w-5 h-5 text-muted-foreground" />
            </button>
            <button className="p-2 rounded hover:bg-secondary transition-colors">
              <Settings className="w-5 h-5 text-muted-foreground" />
            </button>
          </div>
        </div>
      )}
    </header>
  );
}
