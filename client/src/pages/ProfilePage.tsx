"use client";

import { useState } from "react";
import {
  User,
  Mail,
  Coins,
  Gem,
  Shield,
  Award,
  ClipboardList,
  Crown,
  Ticket,
  CircleDollarSign,
  Edit3,
  Save,
  Users,
  Search,
  Plus,
  Heart,
  MessageCircle,
  MapPin,
  Calendar,
  ChevronDown,
  ChevronRight,
  Flame,
  Star,
  Gamepad2,
  Phone,
  Smile,
  Scissors,
  FolderOpen,
  Image as ImageIcon,
  Mic,
  AtSign,
  Send,
} from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Empty,
  EmptyHeader,
  EmptyTitle,
  EmptyDescription,
  EmptyMedia,
} from "@/components/ui/empty";
import { cn } from "@/lib/utils";

interface MenuItem {
  id: string;
  label: string;
  icon: React.ReactNode;
  badge?: string;
}

const menuItems: MenuItem[] = [
  { id: "profile", label: "我的资料", icon: <User className="w-4 h-4" /> },
  { id: "friends", label: "我的好友", icon: <Users className="w-4 h-4" /> },
  { id: "messages", label: "站内信", icon: <Mail className="w-4 h-4" /> },
  { id: "coins", label: "我的金币", icon: <Coins className="w-4 h-4" /> },
  { id: "coupons", label: "我的代金券", icon: <Ticket className="w-4 h-4" /> },
  { id: "gems", label: "我的钻石", icon: <Gem className="w-4 h-4" /> },
  { id: "tasks", label: "我的任务", icon: <ClipboardList className="w-4 h-4" /> },
  { id: "vip", label: "会员体系", icon: <Crown className="w-4 h-4" />, badge: "HOT" },
];

interface Friend {
  id: string;
  nickname: string;
  status: "online" | "offline" | "ingame";
  signature: string;
  level: number;
  gender: string;
  age: number;
  birthday: string;
  location: string;
  likes: number;
  group: string;
}

const mockFriends: Friend[] = [
  { id: "1", nickname: "Blind", status: "online", signature: "专注CS2战术分析", level: 42, gender: "男", age: 24, birthday: "3月15日", location: "北京朝阳", likes: 128, group: "朋友" },
  { id: "2", nickname: "林真有", status: "online", signature: "残局大师", level: 38, gender: "男", age: 22, birthday: "7月8日", location: "上海浦东", likes: 96, group: "朋友" },
  { id: "3", nickname: "颜家茗", status: "ingame", signature: "你会不会恨你自己?", level: 55, gender: "女", age: 21, birthday: "11月2日", location: "广州天河", likes: 234, group: "朋友" },
  { id: "4", nickname: "拜桑", status: "online", signature: "A1高闪来一个", level: 29, gender: "男", age: 19, birthday: "1月20日", location: "成都武侯", likes: 67, group: "朋友" },
  { id: "5", nickname: "陈泓", status: "offline", signature: "马枪怪", level: 18, gender: "男", age: 25, birthday: "5月5日", location: "深圳南山", likes: 12, group: "朋友" },
  { id: "6", nickname: "陈奕欣", status: "online", signature: "你在ECL 刘亚楼 你记一下", level: 61, gender: "女", age: 23, birthday: "9月12日", location: "杭州西湖", likes: 521, group: "朋友" },
  { id: "7", nickname: "陈伟泽", status: "offline", signature: "npm i --legacy-peer-deps", level: 35, gender: "男", age: 26, birthday: "6月18日", location: "武汉洪山", likes: 88, group: "朋友" },
  { id: "8", nickname: "尘心", status: "online", signature: "无奈本人没文化...", level: 27, gender: "男", age: 20, birthday: "4月3日", location: "西安雁塔", likes: 45, group: "朋友" },
  { id: "9", nickname: "WindyPath", status: "online", signature: "GG WP", level: 72, gender: "男", age: 28, birthday: "12月25日", location: "南京鼓楼", likes: 1024, group: "最近组队" },
];

const friendGroups = [
  { name: "特别关心", friends: mockFriends.filter((f) => f.group === "特别关心") },
  { name: "朋友", friends: mockFriends.filter((f) => f.group === "朋友") },
  { name: "最近组队", friends: mockFriends.filter((f) => f.group === "最近组队") },
];

export default function ProfilePage() {
  const [activeMenu, setActiveMenu] = useState("profile");
  const [isEditing, setIsEditing] = useState(false);
  const [selectedFriendId, setSelectedFriendId] = useState<string | null>("1");
  const [expandedGroups, setExpandedGroups] = useState<Record<string, boolean>>({
    特别关心: true,
    朋友: true,
    最近组队: true,
  });

  const toggleGroup = (name: string) => {
    setExpandedGroups((prev) => ({ ...prev, [name]: !prev[name] }));
  };

  // Chat mock data
  const [selectedConversationId, setSelectedConversationId] = useState("1");
  const [messageInput, setMessageInput] = useState("");

  const conversations = [
    { id: "1", name: "系统通知", avatar: "", lastMessage: "赛季 S3 已开启，全新地图池上线！", time: "00:28", unread: 1, isGroup: false },
    { id: "2", name: "Astralis水友群", avatar: "", lastMessage: "stArlight: 比如有个挂pv的...", time: "00:28", unread: 8, isGroup: true, memberCount: 256 },
    { id: "3", name: "客服小助手", avatar: "", lastMessage: "您好，充值问题已为您处理完毕", time: "00:24", unread: 0, isGroup: false },
    { id: "4", name: "100T best!", avatar: "", lastMessage: "朱艺莉: 千万不要踢飞神", time: "00:22", unread: 99, isGroup: true, memberCount: 128 },
    { id: "5", name: "幸存者聚落", avatar: "", lastMessage: "哭泣的键盘: 这是魔兽精兵的...", time: "00:22", unread: 0, isGroup: true, memberCount: 128 },
    { id: "6", name: "《三国：百将牌》...", avatar: "", lastMessage: "群管 [21-24] 囡囡: 为了...", time: "00:05", unread: 0, isGroup: true, memberCount: 64 },
    { id: "7", name: "CiGA Game Jam ...", avatar: "", lastMessage: "海蓝石: [动画表情]", time: "00:02", unread: 0, isGroup: true, memberCount: 32 },
    { id: "8", name: "游戏开发学习创...", avatar: "", lastMessage: "佐仓蜜柑暑假出demo: [...", time: "昨天23:40", unread: 99, isGroup: true, memberCount: 512 },
    { id: "9", name: "QQ邮箱提醒", avatar: "", lastMessage: "Eamonn Brennan: The big tr...", time: "昨天22:09", unread: 0, isGroup: false },
    { id: "10", name: "BW漫展搭子交...", avatar: "", lastMessage: "小袖: ✨职业毛娘接单✨...", time: "昨天21:34", unread: 10, isGroup: true, memberCount: 256 },
  ];

  const chatMessages = [
    { id: "m1", sender: "修", avatar: "", content: "[图片] 赛季更新公告截图", time: "00:13", isSelf: false, isAdmin: true },
    { id: "m2", sender: "我是真的想去火烈鸟", avatar: "", content: "美团有事真上啊", time: "00:22", isSelf: false, isAdmin: false },
    { id: "m3", sender: "哭泣的键盘", avatar: "", content: "这是魔兽精兵的奴隶嘛?", time: "00:22", isSelf: false, isAdmin: false },
  ];

  const selectedConversation = conversations.find((c) => c.id === selectedConversationId) || conversations[0];

  // Mock user data
  const [userData, setUserData] = useState({
    username: "Coach_Jonathan",
    email: "jonathan@cs2sim.com",
    bio: "CS2 战术教练 | 追求极致的战术模拟体验",
    level: 52,
    vipLevel: 3,
    userId: "2585674862",
    coins: 23568,
    gems: 1250,
    securityScore: 85,
  });

  const selectedFriend = mockFriends.find((f) => f.id === selectedFriendId) || mockFriends[0];

  const handleSaveProfile = () => {
    setIsEditing(false);
  };

  const statusConfig = {
    online: { color: "bg-green-500", label: "在线" },
    offline: { color: "bg-gray-500", label: "离线" },
    ingame: { color: "bg-primary", label: "游戏中" },
  };

  const renderContent = () => {
    switch (activeMenu) {
      case "profile":
        return (
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold">基本资料</h3>
              <Button
                variant="outline"
                size="sm"
                onClick={() =>
                  isEditing ? handleSaveProfile() : setIsEditing(true)
                }
                className="gap-1.5"
              >
                {isEditing ? (
                  <>
                    <Save className="w-4 h-4" />
                    保存
                  </>
                ) : (
                  <>
                    <Edit3 className="w-4 h-4" />
                    编辑
                  </>
                )}
              </Button>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="space-y-2">
                <Label>昵称</Label>
                <Input
                  value={userData.username}
                  disabled={!isEditing}
                  onChange={(e) =>
                    setUserData({ ...userData, username: e.target.value })
                  }
                />
              </div>
              <div className="space-y-2">
                <Label>邮箱</Label>
                <Input
                  value={userData.email}
                  disabled={!isEditing}
                  onChange={(e) =>
                    setUserData({ ...userData, email: e.target.value })
                  }
                />
              </div>
              <div className="space-y-2 md:col-span-2">
                <Label>个人简介</Label>
                <Input
                  value={userData.bio}
                  disabled={!isEditing}
                  onChange={(e) =>
                    setUserData({ ...userData, bio: e.target.value })
                  }
                />
              </div>
              <div className="space-y-2">
                <Label>用户 ID</Label>
                <Input value={userData.userId} disabled />
              </div>
              <div className="space-y-2">
                <Label>当前等级</Label>
                <div className="flex items-center gap-2 h-9 px-3 rounded-md border border-input bg-transparent text-sm">
                  <Award className="w-4 h-4 text-primary" />
                  Lv.{userData.level}
                </div>
              </div>
            </div>

            <Separator />

            <div className="space-y-3">
              <h3 className="text-lg font-semibold">账号安全</h3>
              <div className="flex items-center gap-4">
                <div className="flex-1 max-w-xs">
                  <div className="flex items-center justify-between mb-1.5">
                    <span className="text-sm text-muted-foreground">
                      安全评分
                    </span>
                    <span className="text-sm font-medium">
                      {userData.securityScore}分
                    </span>
                  </div>
                  <div className="h-2 w-full rounded-full bg-muted overflow-hidden">
                    <div
                      className={cn(
                        "h-full rounded-full transition-all",
                        userData.securityScore >= 80
                          ? "bg-green-500"
                          : userData.securityScore >= 50
                            ? "bg-yellow-500"
                            : "bg-red-500"
                      )}
                      style={{ width: `${userData.securityScore}%` }}
                    />
                  </div>
                </div>
                <div className="flex items-center gap-1.5 text-sm text-green-500">
                  <Shield className="w-4 h-4" />
                  <span>安全</span>
                </div>
                <Button variant="link" size="sm" className="text-primary">
                  修改密码
                </Button>
              </div>
            </div>

            <Separator />

            <div className="space-y-3">
              <h3 className="text-lg font-semibold">我的特权</h3>
              <div className="flex items-center gap-3 flex-wrap">
                {[1, 2, 3, 4, 5].map((i) => (
                  <div
                    key={i}
                    className="w-10 h-10 rounded-lg bg-secondary flex items-center justify-center"
                  >
                    <Award className="w-5 h-5 text-primary/80" />
                  </div>
                ))}
                <Button variant="ghost" size="sm" className="text-primary">
                  更多 »
                </Button>
              </div>
            </div>
          </div>
        );

      case "friends":
        return (
          <div className="flex flex-col lg:flex-row gap-0 -mx-6 -my-6 min-h-[520px]">
            {/* Left: Friend List */}
            <div className="w-full lg:w-64 flex-shrink-0 border-b lg:border-b-0 lg:border-r border-border flex flex-col bg-card/30">
              {/* Search + Add */}
              <div className="p-3 flex gap-2 border-b border-border">
                <div className="relative flex-1">
                  <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                  <Input placeholder="搜索好友..." className="pl-8 h-9 text-sm" />
                </div>
                <Button size="icon" variant="outline" className="h-9 w-9 shrink-0">
                  <Plus className="w-4 h-4" />
                </Button>
              </div>

              {/* Friend Groups */}
              <div className="flex-1 overflow-y-auto py-1">
                {friendGroups.map((group) => {
                  const onlineCount = group.friends.filter(
                    (f) => f.status === "online" || f.status === "ingame"
                  ).length;
                  const isExpanded = expandedGroups[group.name];

                  return (
                    <div key={group.name}>
                      <button
                        onClick={() => toggleGroup(group.name)}
                        className="w-full flex items-center gap-1 px-3 py-2 text-xs text-muted-foreground hover:text-foreground transition-colors"
                      >
                        {isExpanded ? (
                          <ChevronDown className="w-3.5 h-3.5" />
                        ) : (
                          <ChevronRight className="w-3.5 h-3.5" />
                        )}
                        <span className="font-medium">{group.name}</span>
                        <span className="text-muted-foreground/60 ml-1">
                          {onlineCount}/{group.friends.length}
                        </span>
                      </button>

                      {isExpanded &&
                        group.friends.map((friend) => (
                          <button
                            key={friend.id}
                            onClick={() => setSelectedFriendId(friend.id)}
                            className={cn(
                              "w-full flex items-center gap-3 px-3 py-2.5 text-left transition-colors",
                              selectedFriendId === friend.id
                                ? "bg-primary/10"
                                : "hover:bg-secondary/50"
                            )}
                          >
                            <div className="relative shrink-0">
                              <Avatar className="w-9 h-9">
                                <AvatarImage src="" />
                                <AvatarFallback className="bg-gradient-to-br from-blue-500 to-purple-500 text-white text-xs font-bold">
                                  {friend.nickname.slice(0, 2).toUpperCase()}
                                </AvatarFallback>
                              </Avatar>
                              <span
                                className={cn(
                                  "absolute -bottom-0.5 -right-0.5 w-2.5 h-2.5 rounded-full border-2 border-background",
                                  statusConfig[friend.status].color
                                )}
                              />
                            </div>
                            <div className="min-w-0 flex-1">
                              <div className="text-sm font-medium truncate">
                                {friend.nickname}
                              </div>
                              <div className="text-xs text-muted-foreground truncate">
                                {friend.signature}
                              </div>
                            </div>
                          </button>
                        ))}
                    </div>
                  );
                })}
              </div>
            </div>

            {/* Right: Friend Detail */}
            <div className="flex-1 p-5 lg:p-8 overflow-y-auto">
              {selectedFriend ? (
                <div className="space-y-6">
                  {/* Header: Avatar + Name + Status */}
                  <div className="flex items-start gap-5">
                    <Avatar className="w-20 h-20 border-2 border-primary/20">
                      <AvatarImage src="" />
                      <AvatarFallback className="bg-gradient-to-br from-blue-500 to-purple-500 text-white text-2xl font-bold">
                        {selectedFriend.nickname.slice(0, 2).toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-3 flex-wrap">
                        <h3 className="text-xl font-bold">{selectedFriend.nickname}</h3>
                        <div className="flex items-center gap-1.5 text-sm">
                          <span
                            className={cn(
                              "w-2.5 h-2.5 rounded-full",
                              statusConfig[selectedFriend.status].color
                            )}
                          />
                          <span className="text-muted-foreground">
                            {statusConfig[selectedFriend.status].label}
                          </span>
                        </div>
                      </div>
                      <p className="text-sm text-muted-foreground mt-1">
                        ID: {selectedFriend.id.padStart(10, "0")}
                      </p>
                      <div className="flex items-center gap-2 mt-2">
                        <Badge variant="secondary" className="text-xs">
                          Lv.{selectedFriend.level}
                        </Badge>
                        <Badge variant="outline" className="text-xs">
                          {selectedFriend.group}
                        </Badge>
                      </div>
                    </div>
                    <div className="flex items-center gap-1 text-sm text-muted-foreground">
                      <Heart className="w-4 h-4 text-red-500 fill-red-500" />
                      <span>{selectedFriend.likes}</span>
                    </div>
                  </div>

                  <Separator />

                  {/* Info Row */}
                  <div className="flex flex-wrap items-center gap-x-6 gap-y-2 text-sm">
                    <div className="flex items-center gap-1.5">
                      <span className="text-muted-foreground">
                        {selectedFriend.gender === "男" ? "♂" : "♀"}
                      </span>
                      <span>{selectedFriend.gender}</span>
                    </div>
                    <div className="flex items-center gap-1.5">
                      <Flame className="w-3.5 h-3.5 text-muted-foreground" />
                      <span>{selectedFriend.age}岁</span>
                    </div>
                    <div className="flex items-center gap-1.5">
                      <Calendar className="w-3.5 h-3.5 text-muted-foreground" />
                      <span>{selectedFriend.birthday}</span>
                    </div>
                    <div className="flex items-center gap-1.5">
                      <MapPin className="w-3.5 h-3.5 text-muted-foreground" />
                      <span>现居 {selectedFriend.location}</span>
                    </div>
                  </div>

                  {/* Medals / Level icons */}
                  <div className="flex items-center gap-2 flex-wrap">
                    {[Crown, Flame, Star, Award, Gamepad2, Shield, Gem, Heart].map(
                      (Icon, i) => (
                        <div
                          key={i}
                          className={cn(
                            "w-9 h-9 rounded-lg flex items-center justify-center",
                            i < 3
                              ? "bg-yellow-500/15 text-yellow-500"
                              : i < 5
                                ? "bg-primary/15 text-primary"
                                : "bg-muted text-muted-foreground"
                          )}
                        >
                          <Icon className="w-4 h-4" />
                        </div>
                      )
                    )}
                  </div>

                  {/* Signature */}
                  <div className="space-y-2">
                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                      <Edit3 className="w-3.5 h-3.5" />
                      <span>签名</span>
                    </div>
                    <p className="text-sm">{selectedFriend.signature}</p>
                  </div>

                  {/* Photo Wall (Mock) */}
                  <div className="space-y-3">
                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                      <Gamepad2 className="w-3.5 h-3.5" />
                      <span>精选战绩</span>
                    </div>
                    <div className="flex gap-3">
                      {[1, 2, 3].map((i) => (
                        <div
                          key={i}
                          className="w-24 h-24 lg:w-28 lg:h-28 rounded-lg bg-secondary/80 flex items-center justify-center"
                        >
                          <span className="text-xs text-muted-foreground">
                            截图 {i}
                          </span>
                        </div>
                      ))}
                    </div>
                  </div>

                  {/* Actions */}
                  <div className="flex items-center gap-3 pt-2">
                    <Button variant="outline" className="gap-1.5">
                      <Edit3 className="w-4 h-4" />
                      编辑资料
                    </Button>
                    <Button className="gap-1.5 bg-primary hover:bg-primary/90">
                      <MessageCircle className="w-4 h-4" />
                      发消息
                    </Button>
                  </div>
                </div>
              ) : (
                <Empty className="min-h-[400px]">
                  <EmptyHeader>
                    <EmptyTitle>请选择好友</EmptyTitle>
                    <EmptyDescription>在左侧列表中选择一位好友查看详情</EmptyDescription>
                  </EmptyHeader>
                </Empty>
              )}
            </div>
          </div>
        );

      case "messages":
        return (
          <div className="flex flex-col lg:flex-row gap-0 -mx-6 -my-6 min-h-[520px]">
            {/* Left: Conversation List */}
            <div className="w-full lg:w-72 flex-shrink-0 border-b lg:border-b-0 lg:border-r border-border flex flex-col bg-card/30">
              {/* Search */}
              <div className="p-3 flex gap-2 border-b border-border">
                <div className="relative flex-1">
                  <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                  <Input placeholder="搜索" className="pl-8 h-9 text-sm bg-secondary/50 border-0" />
                </div>
                <Button size="icon" variant="outline" className="h-9 w-9 shrink-0">
                  <Plus className="w-4 h-4" />
                </Button>
              </div>

              {/* Conversation List */}
              <div className="flex-1 overflow-y-auto">
                {conversations.map((conv) => (
                  <button
                    key={conv.id}
                    onClick={() => setSelectedConversationId(conv.id)}
                    className={cn(
                      "w-full flex items-center gap-3 px-3 py-2.5 text-left transition-colors",
                      selectedConversationId === conv.id
                        ? "bg-primary/10"
                        : "hover:bg-secondary/50"
                    )}
                  >
                    <Avatar className="w-10 h-10 shrink-0">
                      <AvatarImage src={conv.avatar} />
                      <AvatarFallback className="bg-gradient-to-br from-blue-500 to-purple-500 text-white text-xs font-bold">
                        {conv.name.slice(0, 2).toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center justify-between gap-2">
                        <span className="text-sm font-medium truncate">{conv.name}</span>
                        <span className="text-xs text-muted-foreground shrink-0">{conv.time}</span>
                      </div>
                      <div className="flex items-center justify-between gap-2 mt-0.5">
                        <span className="text-xs text-muted-foreground truncate">
                          {conv.lastMessage}
                        </span>
                        {conv.unread > 0 && (
                          <Badge
                            variant="secondary"
                            className="h-5 min-w-5 px-1.5 text-[10px] bg-primary/20 text-primary border-primary/20 shrink-0"
                          >
                            {conv.unread > 99 ? "99+" : conv.unread}
                          </Badge>
                        )}
                      </div>
                    </div>
                  </button>
                ))}
              </div>
            </div>

            {/* Right: Chat Window */}
            <div className="flex-1 flex flex-col min-h-[400px]">
              {/* Header */}
              <div className="flex items-center justify-between px-4 py-3 border-b border-border">
                <div className="flex items-center gap-2">
                  <span className="font-semibold text-sm">{selectedConversation.name}</span>
                  {selectedConversation.isGroup && (
                    <span className="text-xs text-muted-foreground">
                      ({selectedConversation.memberCount})
                    </span>
                  )}
                </div>
                <Button variant="ghost" size="icon" className="h-8 w-8">
                  <Phone className="w-4 h-4 text-muted-foreground" />
                </Button>
              </div>

              {/* Messages Area */}
              <div className="flex-1 overflow-y-auto p-4 space-y-4">
                {chatMessages.map((msg) => (
                  <div
                    key={msg.id}
                    className={cn(
                      "flex gap-3",
                      msg.isSelf ? "flex-row-reverse" : "flex-row"
                    )}
                  >
                    <Avatar className="w-9 h-9 shrink-0">
                      <AvatarImage src={msg.avatar} />
                      <AvatarFallback className="bg-gradient-to-br from-blue-500 to-purple-500 text-white text-xs font-bold">
                        {msg.sender.slice(0, 2).toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                    <div
                      className={cn(
                        "max-w-[70%] flex flex-col",
                        msg.isSelf ? "items-end" : "items-start"
                      )}
                    >
                      <div className="flex items-center gap-2 mb-1">
                        <span className="text-xs text-muted-foreground">{msg.sender}</span>
                        {msg.isAdmin && (
                          <Badge variant="secondary" className="text-[10px] h-4 px-1 bg-primary/20 text-primary border-primary/20">
                            管理员
                          </Badge>
                        )}
                        <span className="text-xs text-muted-foreground/60">{msg.time}</span>
                      </div>
                      <div
                        className={cn(
                          "px-3 py-2 rounded-lg text-sm break-words",
                          msg.isSelf
                            ? "bg-primary text-primary-foreground"
                            : "bg-secondary text-secondary-foreground"
                        )}
                      >
                        {msg.content}
                      </div>
                    </div>
                  </div>
                ))}
              </div>

              {/* Input Area */}
              <div className="border-t border-border p-3">
                {/* Toolbar */}
                <div className="flex items-center gap-1 mb-2">
                  <Button variant="ghost" size="icon" className="h-8 w-8">
                    <Smile className="w-4 h-4 text-muted-foreground" />
                  </Button>
                  <Button variant="ghost" size="icon" className="h-8 w-8">
                    <Scissors className="w-4 h-4 text-muted-foreground" />
                  </Button>
                  <Button variant="ghost" size="icon" className="h-8 w-8">
                    <FolderOpen className="w-4 h-4 text-muted-foreground" />
                  </Button>
                  <Button variant="ghost" size="icon" className="h-8 w-8">
                    <ImageIcon className="w-4 h-4 text-muted-foreground" />
                  </Button>
                  <Button variant="ghost" size="icon" className="h-8 w-8">
                    <Mic className="w-4 h-4 text-muted-foreground" />
                  </Button>
                  <Button variant="ghost" size="icon" className="h-8 w-8">
                    <AtSign className="w-4 h-4 text-muted-foreground" />
                  </Button>
                </div>
                <div className="flex gap-2">
                  <Textarea
                    placeholder="请输入消息..."
                    className="min-h-[60px] resize-none"
                    value={messageInput}
                    onChange={(e) => setMessageInput(e.target.value)}
                    onKeyDown={(e) => {
                      if (e.key === "Enter" && !e.shiftKey) {
                        e.preventDefault();
                        setMessageInput("");
                      }
                    }}
                  />
                  <div className="flex flex-col justify-end">
                    <Button
                      size="sm"
                      className="h-9 px-4 bg-primary hover:bg-primary/90"
                      onClick={() => setMessageInput("")}
                    >
                      <Send className="w-4 h-4" />
                    </Button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        );

      case "coins":
        return (
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold">我的金币</h3>
              <Button size="sm" className="bg-primary hover:bg-primary/90">
                <CircleDollarSign className="w-4 h-4 mr-1" />
                充值金币
              </Button>
            </div>
            <Card className="bg-gradient-to-r from-yellow-500/10 to-transparent border-yellow-500/20">
              <CardContent className="py-6">
                <div className="flex items-center gap-3">
                  <Coins className="w-8 h-8 text-yellow-500" />
                  <div>
                    <p className="text-sm text-muted-foreground">当前余额</p>
                    <p className="text-2xl font-bold">{userData.coins.toLocaleString()}</p>
                  </div>
                </div>
              </CardContent>
            </Card>
            <Empty className="min-h-[240px] border-border border">
              <EmptyHeader>
                <EmptyDescription>暂无交易记录</EmptyDescription>
              </EmptyHeader>
            </Empty>
          </div>
        );

      case "coupons":
        return (
          <div className="space-y-6">
            <h3 className="text-lg font-semibold">我的代金券</h3>
            <Empty className="min-h-[320px] border-border border">
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <Ticket className="w-6 h-6" />
                </EmptyMedia>
                <EmptyTitle>暂无代金券</EmptyTitle>
                <EmptyDescription>参与活动可获得代金券</EmptyDescription>
              </EmptyHeader>
            </Empty>
          </div>
        );

      case "gems":
        return (
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold">我的钻石</h3>
              <Button size="sm" className="bg-primary hover:bg-primary/90">
                <Gem className="w-4 h-4 mr-1" />
                领取钻石
              </Button>
            </div>
            <Card className="bg-gradient-to-r from-pink-500/10 to-transparent border-pink-500/20">
              <CardContent className="py-6">
                <div className="flex items-center gap-3">
                  <Gem className="w-8 h-8 text-pink-500" />
                  <div>
                    <p className="text-sm text-muted-foreground">当前余额</p>
                    <p className="text-2xl font-bold">{userData.gems.toLocaleString()}</p>
                  </div>
                </div>
              </CardContent>
            </Card>
            <Empty className="min-h-[240px] border-border border">
              <EmptyHeader>
                <EmptyDescription>暂无交易记录</EmptyDescription>
              </EmptyHeader>
            </Empty>
          </div>
        );

      case "tasks":
        return (
          <div className="space-y-6">
            <h3 className="text-lg font-semibold">我的任务</h3>
            <Empty className="min-h-[320px] border-border border">
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <ClipboardList className="w-6 h-6" />
                </EmptyMedia>
                <EmptyTitle>暂无进行中的任务</EmptyTitle>
                <EmptyDescription>前往首页查看今日任务</EmptyDescription>
              </EmptyHeader>
            </Empty>
          </div>
        );

      case "vip":
        return (
          <div className="space-y-6">
            <h3 className="text-lg font-semibold">会员体系</h3>
            <Card className="bg-gradient-to-r from-primary/10 to-transparent border-primary/20">
              <CardContent className="py-8">
                <div className="flex items-center gap-4">
                  <div className="w-16 h-16 rounded-full bg-primary/20 flex items-center justify-center">
                    <Crown className="w-8 h-8 text-primary" />
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="text-xl font-bold">VIP {userData.vipLevel}</span>
                      <Badge variant="secondary">当前等级</Badge>
                    </div>
                    <p className="text-sm text-muted-foreground mt-1">
                      再消费 5,000 钻石可升级至 VIP 4
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        );

      default:
        return null;
    }
  };

  const activeMenuItem = menuItems.find((item) => item.id === activeMenu);

  return (
    <div className="flex-1 p-4 lg:p-6">
      <div className="max-w-[1400px] mx-auto flex flex-col gap-4 lg:gap-6">
        {/* Top User Info Card */}
        <Card className="overflow-hidden">
          <CardContent className="p-4 lg:p-6">
            <div className="flex flex-col lg:flex-row lg:items-center gap-4 lg:gap-8">
              {/* Avatar & Basic Info */}
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
                    <Badge
                      variant="secondary"
                      className="bg-primary/20 text-primary border-primary/30"
                    >
                      VIP{userData.vipLevel}
                    </Badge>
                  </div>
                  <div className="flex items-center gap-2 text-sm text-muted-foreground mt-1">
                    <span>ID: {userData.userId}</span>
                  </div>
                </div>
              </div>

              <Separator orientation="vertical" className="hidden lg:block h-12" />

              {/* Coins */}
              <div className="flex items-center gap-4 flex-1">
                <div className="flex-1">
                  <p className="text-sm text-muted-foreground mb-1">我的金币</p>
                  <div className="flex items-center gap-3">
                    <span className="text-xl font-bold">
                      {userData.coins.toLocaleString()}
                    </span>
                    <Button
                      size="sm"
                      className="bg-primary hover:bg-primary/90 text-primary-foreground"
                    >
                      充值金币
                    </Button>
                  </div>
                </div>

                <Separator orientation="vertical" className="hidden sm:block h-10" />

                <div className="flex-1">
                  <p className="text-sm text-muted-foreground mb-1">钻石余额</p>
                  <div className="flex items-center gap-3">
                    <span className="text-xl font-bold">
                      {userData.gems.toLocaleString()}
                    </span>
                    <Button
                      size="sm"
                      className="bg-primary hover:bg-primary/90 text-primary-foreground"
                    >
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

            {/* Security & Privileges */}
            <div className="flex flex-col sm:flex-row sm:items-center gap-3 sm:gap-6">
              <div className="flex items-center gap-2">
                <span className="text-sm text-muted-foreground">账号安全:</span>
                <div className="w-24 h-2 rounded-full bg-muted overflow-hidden">
                  <div
                    className="h-full rounded-full bg-green-500"
                    style={{ width: `${userData.securityScore}%` }}
                  />
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
                    <div
                      key={i}
                      className="w-7 h-7 rounded bg-secondary flex items-center justify-center"
                    >
                      <Award className="w-3.5 h-3.5 text-primary/80" />
                    </div>
                  ))}
                  <Button variant="ghost" size="sm" className="text-primary h-auto px-1.5">
                    更多 »
                  </Button>
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
                  <button
                    key={item.id}
                    onClick={() => setActiveMenu(item.id)}
                    className={cn(
                      "w-full flex items-center gap-3 px-4 py-3 text-sm font-medium transition-colors text-left relative",
                      activeMenu === item.id
                        ? "bg-primary/10 text-primary border-l-2 border-l-primary"
                        : "text-muted-foreground hover:text-foreground hover:bg-secondary/50 border-l-2 border-l-transparent"
                    )}
                  >
                    {item.icon}
                    <span>{item.label}</span>
                    {item.badge && (
                      <Badge
                        variant="destructive"
                        className="ml-auto text-[10px] px-1.5 py-0 h-4"
                      >
                        {item.badge}
                      </Badge>
                    )}
                  </button>
                ))}
              </div>
            </Card>
          </div>

          {/* Right Content Panel */}
          <div className="flex-1 min-w-0">
            <Card className="h-full overflow-hidden">
              <CardHeader className="pb-4">
                <CardTitle className="text-base">
                  {activeMenuItem?.label}
                </CardTitle>
              </CardHeader>
              <CardContent>{renderContent()}</CardContent>
            </Card>
          </div>
        </div>
      </div>
    </div>
  );
}
