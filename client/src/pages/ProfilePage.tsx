"use client";

import { useState } from "react";
import { Outlet, NavLink, useLocation } from "react-router-dom";
import { useAuth } from "@/context/AuthContext";
import {
  User,
  Mail,
  Coins,
  Gem,
  Award,
  ClipboardList,
  Crown,
  Ticket,
  Users,
} from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { cn } from "@/lib/utils";

interface MenuItem {
  id: string;
  path: string;
  label: string;
  icon: React.ReactNode;
  badge?: string;
}

const menuItems: MenuItem[] = [
  { id: "profile", path: "/profile/me", label: "我的资料", icon: <User className="w-4 h-4" /> },
  { id: "friends", path: "/profile/friends", label: "我的好友", icon: <Users className="w-4 h-4" /> },
  { id: "messages", path: "/profile/messages", label: "站内信", icon: <Mail className="w-4 h-4" /> },
  { id: "coins", path: "/profile/coins", label: "我的金币", icon: <Coins className="w-4 h-4" /> },
  { id: "coupons", path: "/profile/coupons", label: "我的代金券", icon: <Ticket className="w-4 h-4" /> },
  { id: "gems", path: "/profile/gems", label: "我的钻石", icon: <Gem className="w-4 h-4" /> },
  { id: "tasks", path: "/profile/tasks", label: "我的任务", icon: <ClipboardList className="w-4 h-4" /> },
  { id: "vip", path: "/profile/vip", label: "会员体系", icon: <Crown className="w-4 h-4" />, badge: "HOT" },
];

export default function ProfilePage() {
  const { session } = useAuth();
  const location = useLocation();
  const [isEditing, setIsEditing] = useState(false);

  const [userData, setUserData] = useState({
    username: session?.username ?? "玩家",
    email: "-",
    bio: "这个人很懒，什么都没写",
    level: 1,
    vipLevel: 0,
    userId: session?.user_id ?? "-",
    coins: 0,
    gems: 0,
    securityScore: 0,
  });

  const outletContext = { userData, setUserData, isEditing, setIsEditing };

  return (
    <div className="flex-1 p-4 lg:p-6">
      <div className="max-w-[1400px] mx-auto flex flex-col gap-4 lg:gap-6">
        {/* Top User Info Card */}
        <Card className="overflow-hidden">
          <CardContent className="p-4 lg:p-6">
            <div className="flex flex-col lg:flex-row lg:items-center gap-4 lg:gap-8">
              <div className="flex items-center gap-4">
                <Avatar className="w-16 h-16 border-2 border-primary/30">
                  <AvatarImage src="" />
                  <AvatarFallback className="bg-gradient-to-br from-blue-500 to-purple-500 text-white text-lg font-bold">
                    {userData.username.slice(0, 2).toUpperCase()}
                  </AvatarFallback>
                </Avatar>
                <div>
                  <div className="flex items-center gap-2">
                    <span className="text-lg font-bold">{userData.username}</span>
                    <Badge variant="secondary" className="bg-primary/20 text-primary border-primary/30">
                      VIP{userData.vipLevel}
                    </Badge>
                  </div>
                  <div className="flex items-center gap-2 text-sm text-muted-foreground mt-1">
                    <span>ID: {userData.userId}</span>
                  </div>
                </div>
              </div>

              <Separator orientation="vertical" className="hidden lg:block h-12" />

              <div className="flex items-center gap-4 flex-1">
                <div className="flex-1">
                  <p className="text-sm text-muted-foreground mb-1">我的金币</p>
                  <div className="flex items-center gap-3">
                    <span className="text-xl font-bold">{userData.coins.toLocaleString()}</span>
                    <Button size="sm" className="bg-primary hover:bg-primary/90 text-primary-foreground">
                      充值金币
                    </Button>
                  </div>
                </div>
                <Separator orientation="vertical" className="hidden sm:block h-10" />
                <div className="flex-1">
                  <p className="text-sm text-muted-foreground mb-1">钻石余额</p>
                  <div className="flex items-center gap-3">
                    <span className="text-xl font-bold">{userData.gems.toLocaleString()}</span>
                    <Button size="sm" className="bg-primary hover:bg-primary/90 text-primary-foreground">
                      领取钻石
                    </Button>
                    <Button variant="link" size="sm" className="text-primary px-0">
                      兑换礼品
                    </Button>
                  </div>
                </div>
              </div>
            </div>

            <Separator className="my-4" />

            <div className="flex flex-col sm:flex-row sm:items-center gap-3 sm:gap-6">
              <div className="flex items-center gap-2">
                <span className="text-sm text-muted-foreground">账号安全:</span>
                <div className="w-24 h-2 rounded-full bg-muted overflow-hidden">
                  <div className="h-full rounded-full bg-green-500" style={{ width: `${userData.securityScore}%` }} />
                </div>
                <span className="text-sm text-green-500">安全</span>
                <Button variant="link" size="sm" className="text-primary px-0 h-auto">
                  修改密码
                </Button>
              </div>
              <Separator orientation="vertical" className="hidden sm:block h-4" />
              <div className="flex items-center gap-2">
                <span className="text-sm text-muted-foreground">我的特权:</span>
                <div className="flex items-center gap-1.5">
                  {[1, 2, 3, 4, 5].map((i) => (
                    <div key={i} className="w-7 h-7 rounded bg-secondary flex items-center justify-center">
                      <Award className="w-3.5 h-3.5 text-primary/80" />
                    </div>
                  ))}
                  <Button variant="ghost" size="sm" className="text-primary h-auto px-1.5">更多 »</Button>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Main Content */}
        <div className="flex flex-col lg:flex-row gap-4 lg:gap-6">
          {/* Left Sidebar Menu */}
          <div className="w-full lg:w-52 flex-shrink-0">
            <Card className="overflow-hidden">
              <div className="py-2">
                {menuItems.map((item) => (
                  <NavLink
                    key={item.id}
                    to={item.path}
                    end={item.path.endsWith("/me")}
                    className={({ isActive }) =>
                      cn(
                        "w-full flex items-center gap-3 px-4 py-3 text-sm font-medium transition-colors text-left relative",
                        isActive
                          ? "bg-primary/10 text-primary border-l-2 border-l-primary"
                          : "text-muted-foreground hover:text-foreground hover:bg-secondary/50 border-l-2 border-l-transparent",
                      )
                    }
                  >
                    {item.icon}
                    <span>{item.label}</span>
                    {item.badge && (
                      <Badge variant="destructive" className="ml-auto text-[10px] px-1.5 py-0 h-4">
                        {item.badge}
                      </Badge>
                    )}
                  </NavLink>
                ))}
              </div>
            </Card>
          </div>

          {/* Right Content Panel */}
          <div className="flex-1 min-w-0">
            <Card className="h-full overflow-hidden">
              <CardHeader className="pb-4">
                <CardTitle className="text-base">
                  {menuItems.find((item) =>
                    item.path === location.pathname ||
                    (item.path.endsWith("/me") && location.pathname === "/profile"),
                  )?.label ?? "个人中心"}
                </CardTitle>
              </CardHeader>
              <CardContent>
                <Outlet context={outletContext} />
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </div>
  );
}
