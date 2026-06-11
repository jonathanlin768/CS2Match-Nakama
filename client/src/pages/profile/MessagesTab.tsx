import { useState, useRef, useEffect, useCallback } from "react";
import {
  Search,
  Phone,
  Smile,
  Scissors,
  FolderOpen,
  Image as ImageIcon,
  Mic,
  AtSign,
  Send,
  ArrowLeft,
  AlertCircle,
  RefreshCw,
  Loader2,
} from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { cn } from "@/lib/utils";
import { useAuth } from "@/context/AuthContext";
import { useFriends } from "@/hooks/useFriends";
import { useFriendDM } from "@/hooks/useFriendDM";
import type { ChatMessage } from "@/hooks/useFriendDM";

// ---- Helpers ----

function formatTime(isoString: string): string {
  if (!isoString) return "";
  try {
    const date = new Date(isoString);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return "刚刚";
    if (diffMins < 60) return `${diffMins}分钟前`;
    if (diffHours < 24) return `${diffHours}小时前`;
    if (diffDays < 7) return `${diffDays}天前`;
    return date.toLocaleDateString("zh-CN", {
      month: "2-digit",
      day: "2-digit",
    });
  } catch {
    return "";
  }
}

// ---- Components ----

/** 空状态占位 */
function EmptyState({
  icon: Icon,
  message,
  action,
}: {
  icon: React.ComponentType<{ className?: string }>;
  message: string;
  action?: { label: string; onClick: () => void };
}) {
  return (
    <div className="flex flex-col items-center justify-center h-full text-muted-foreground gap-3 p-8">
      <Icon className="w-12 h-12 opacity-30" />
      <p className="text-sm">{message}</p>
      {action && (
        <Button variant="outline" size="sm" onClick={action.onClick}>
          {action.label}
        </Button>
      )}
    </div>
  );
}

/** 加载状态 */
function LoadingState({ text = "加载中..." }: { text?: string }) {
  return (
    <div className="flex flex-col items-center justify-center h-full text-muted-foreground gap-3 p-8">
      <Loader2 className="w-8 h-8 animate-spin opacity-50" />
      <p className="text-sm">{text}</p>
    </div>
  );
}

/** 消息气泡 */
function MessageBubble({ msg }: { msg: ChatMessage }) {
  return (
    <div
      className={cn(
        "flex gap-3 animate-in fade-in slide-in-from-bottom-1 duration-200",
        msg.isSelf ? "flex-row-reverse" : "flex-row",
      )}
    >
      <Avatar className="w-9 h-9 shrink-0">
        <AvatarImage src="" />
        <AvatarFallback className="bg-gradient-to-br from-blue-500 to-purple-500 text-white text-xs font-bold">
          {msg.username.slice(0, 2).toUpperCase()}
        </AvatarFallback>
      </Avatar>
      <div
        className={cn(
          "max-w-[70%] flex flex-col",
          msg.isSelf ? "items-end" : "items-start",
        )}
      >
        <div className="flex items-center gap-2 mb-1">
          <span className="text-xs text-muted-foreground">{msg.username}</span>
          <span className="text-xs text-muted-foreground/60">
            {formatTime(msg.createTime)}
          </span>
        </div>
        <div
          className={cn(
            "px-3 py-2 rounded-lg text-sm break-words relative",
            msg.isSelf
              ? "bg-primary text-primary-foreground"
              : "bg-secondary text-secondary-foreground",
            msg.sendStatus === "failed" &&
              "border border-destructive/50 opacity-80",
          )}
        >
          {msg.content}
          {msg.sendStatus === "sending" && (
            <span className="ml-2 inline-block w-3 h-3 border-2 border-current border-t-transparent rounded-full animate-spin align-middle" />
          )}
          {msg.sendStatus === "failed" && (
            <AlertCircle className="inline-block w-4 h-4 text-destructive ml-2 align-middle" />
          )}
        </div>
        {msg.sendStatus === "failed" && (
          <span className="text-xs text-destructive mt-1">发送失败</span>
        )}
      </div>
    </div>
  );
}

// ---- Main Component ----

export default function MessagesTab() {
  const { session } = useAuth();
  const { friends } = useFriends(session);
  const currentUserId = session?.user_id ?? "";

  const {
    status,
    conversations,
    selectedId,
    messages,
    sendMessage,
    selectConversation,
    loadMoreMessages,
    retry,
  } = useFriendDM(session, friends, currentUserId);

  const [messageInput, setMessageInput] = useState("");
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [showMobileChat, setShowMobileChat] = useState(false);

  const messagesContainerRef = useRef<HTMLDivElement>(null);
  const isAtBottomRef = useRef(true);
  const needsInitialScrollRef = useRef(false); // 选中新会话时设为 true，强制滚到底
  const [hasNewMessage, setHasNewMessage] = useState(false);

  const selectedConversation = conversations.find(
    (c) => c.friendUserId === selectedId,
  );

  // 选中会话变化时（切页签 / F5 自动选中 / 手动点击），标记需要滚到底
  useEffect(() => {
    if (selectedId) {
      needsInitialScrollRef.current = true;
    }
  }, [selectedId]);

  // ---- 滚动到最底部 ----
  const scrollToBottom = useCallback((smooth = false) => {
    const container = messagesContainerRef.current;
    if (!container) return;
    container.scrollTo({
      top: container.scrollHeight,
      behavior: smooth ? "smooth" : "instant",
    });
    setHasNewMessage(false);
  }, []);

  // ---- 监听用户手动滚动，记录是否在底部 ----
  useEffect(() => {
    const container = messagesContainerRef.current;
    if (!container) return;

    const checkAtBottom = () => {
      const { scrollTop, scrollHeight, clientHeight } = container;
      isAtBottomRef.current = scrollHeight - scrollTop - clientHeight < 60;
    };

    container.addEventListener("scroll", checkAtBottom, { passive: true });
    return () => container.removeEventListener("scroll", checkAtBottom);
  }, []);

  // ---- 消息列表变化时自动滚动 ----
  useEffect(() => {
    if (messages.length === 0) return;

    // 首次加载（切会话 / F5 / 切页签）→ 强制滚到底
    if (needsInitialScrollRef.current) {
      needsInitialScrollRef.current = false;
      // requestAnimationFrame 确保 DOM 已渲染新消息
      requestAnimationFrame(() => scrollToBottom(false));
      return;
    }

    // 新消息到达 & 用户在底部 → 平滑滚到底
    if (isAtBottomRef.current) {
      requestAnimationFrame(() => scrollToBottom(true));
    } else {
      // 用户在翻历史 → 不滚动，显示"新消息"提示
      const lastMsg = messages[messages.length - 1];
      if (lastMsg && !lastMsg.isSelf) {
        setHasNewMessage(true);
      }
    }
  }, [messages, scrollToBottom]);

  // ---- 向上滚动加载更早消息 ----

  const handleScrollToTop = useCallback(async () => {
    const container = messagesContainerRef.current;
    if (!container || container.scrollTop > 50 || isLoadingMore) return;

    setIsLoadingMore(true);
    const prevScrollHeight = container.scrollHeight;

    try {
      await loadMoreMessages();
    } finally {
      // 保持滚动位置稳定
      requestAnimationFrame(() => {
        const newScrollHeight = container.scrollHeight;
        container.scrollTop = newScrollHeight - prevScrollHeight;
        setIsLoadingMore(false);
      });
    }
  }, [loadMoreMessages, isLoadingMore]);

  // ---- 发送消息 ----

  const handleSend = useCallback(async () => {
    const trimmed = messageInput.trim();
    if (!trimmed || !selectedId) return;

    setMessageInput("");
    await sendMessage(selectedId, trimmed);
    scrollToBottom(true);
  }, [messageInput, selectedId, sendMessage, scrollToBottom]);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === "Enter" && !e.shiftKey) {
        e.preventDefault();
        handleSend();
      }
    },
    [handleSend],
  );

  // ---- 选中会话时切换到对话视图（移动端） ----

  const handleSelectConversation = useCallback(
    (friendUserId: string) => {
      needsInitialScrollRef.current = true;
      selectConversation(friendUserId);
      setShowMobileChat(true);
    },
    [selectConversation],
  );

  // ---- 渲染 ----

  if (!session) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[520px] text-muted-foreground gap-3">
        <p className="text-sm">请先登录</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col lg:flex-row gap-0 -mx-6 -my-6 h-[calc(100vh-30rem)] min-h-[420px]">
      {/* ---- 左侧：会话列表 ---- */}
      <div
        className={cn(
          "w-full lg:w-72 flex-shrink-0 border-b lg:border-b-0 lg:border-r border-border flex flex-col bg-card/30",
          showMobileChat && "hidden lg:flex", // 移动端：进入聊天后隐藏
        )}
      >
        {/* 搜索栏 */}
        <div className="p-3 flex gap-2 border-b border-border">
          <div className="relative flex-1">
            <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
            <Input
              placeholder="搜索"
              className="pl-8 h-9 text-sm bg-secondary/50 border-0"
            />
          </div>
          <Button size="icon" variant="outline" className="h-9 w-9 shrink-0 hidden">
            {/* Plus 按钮预留，本阶段隐藏 */}
          </Button>
        </div>

        {/* 会话列表内容 */}
        <div className="flex-1 overflow-y-auto">
          {status === "loading" && <LoadingState text="加载会话中..." />}

          {status === "error" && (
            <EmptyState
              icon={AlertCircle}
              message="加载失败，请稍后重试"
              action={{ label: "重试", onClick: retry }}
            />
          )}

          {status === "success" && conversations.length === 0 && (
            <EmptyState icon={Search} message="暂无好友会话，请先添加好友" />
          )}

          {status === "success" &&
            conversations.map((conv) => (
              <button
                key={conv.friendUserId}
                onClick={() => handleSelectConversation(conv.friendUserId)}
                className={cn(
                  "w-full flex items-center gap-3 px-3 py-2.5 text-left transition-colors",
                  selectedId === conv.friendUserId
                    ? "bg-primary/10"
                    : "hover:bg-secondary/50",
                )}
              >
                {/* 头像 + 在线状态 */}
                <div className="relative shrink-0">
                  <Avatar className="w-10 h-10">
                    <AvatarImage src={conv.friendAvatarUrl} />
                    <AvatarFallback className="bg-gradient-to-br from-blue-500 to-purple-500 text-white text-xs font-bold">
                      {conv.friendUsername.slice(0, 2).toUpperCase()}
                    </AvatarFallback>
                  </Avatar>
                  {/* 在线状态圆点 */}
                  <span
                    className={cn(
                      "absolute -bottom-0.5 -right-0.5 w-3 h-3 rounded-full border-2 border-background",
                      conv.isFriendOnline ? "bg-green-500" : "bg-gray-500",
                    )}
                  />
                </div>

                {/* 内容 */}
                <div className="min-w-0 flex-1">
                  <div className="flex items-center justify-between gap-2">
                    <span className="text-sm font-medium truncate">
                      {conv.friendUsername}
                    </span>
                    <span className="text-xs text-muted-foreground shrink-0">
                      {formatTime(conv.lastMessageTime)}
                    </span>
                  </div>
                  <div className="flex items-center justify-between gap-2 mt-0.5">
                    <span className="text-xs text-muted-foreground truncate">
                      {conv.lastMessage || "暂无消息"}
                    </span>
                    {conv.unreadCount > 0 && (
                      <Badge
                        variant="secondary"
                        className="h-5 min-w-5 px-1.5 text-[10px] bg-primary/20 text-primary border-primary/20 shrink-0"
                      >
                        {conv.unreadCount > 99 ? "99+" : conv.unreadCount}
                      </Badge>
                    )}
                  </div>
                </div>
              </button>
            ))}
        </div>
      </div>

      {/* ---- 右侧：聊天区域 ---- */}
      <div
        className={cn(
          "flex-1 flex flex-col min-h-[400px]",
          !showMobileChat && "hidden lg:flex", // 移动端：默认隐藏
        )}
      >
        {/* 聊天头部 */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-border">
          <div className="flex items-center gap-2">
            {/* 移动端返回按钮 */}
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8 lg:hidden"
              onClick={() => setShowMobileChat(false)}
            >
              <ArrowLeft className="w-4 h-4" />
            </Button>

            {selectedConversation ? (
              <>
                <div className="relative">
                  <Avatar className="w-8 h-8">
                    <AvatarImage src={selectedConversation.friendAvatarUrl} />
                    <AvatarFallback className="bg-gradient-to-br from-blue-500 to-purple-500 text-white text-xs font-bold">
                      {selectedConversation.friendUsername.slice(0, 2).toUpperCase()}
                    </AvatarFallback>
                  </Avatar>
                </div>
                <div className="flex flex-col">
                  <span className="font-semibold text-sm">
                    {selectedConversation.friendUsername}
                  </span>
                  <span className="text-xs text-muted-foreground">
                    {selectedConversation.isFriendOnline ? (
                      <span className="text-green-500">● 在线</span>
                    ) : (
                      <span className="text-gray-500">● 离线</span>
                    )}
                  </span>
                </div>
              </>
            ) : (
              <span className="font-semibold text-sm text-muted-foreground">
                选择好友开始聊天
              </span>
            )}
          </div>
          <Button variant="ghost" size="icon" className="h-8 w-8">
            <Phone className="w-4 h-4 text-muted-foreground" />
          </Button>
        </div>

        {/* 消息列表 */}
        <div
          ref={messagesContainerRef}
          className="flex-1 overflow-y-auto p-4 space-y-4"
          onScroll={(e) => {
            const target = e.currentTarget;
            if (target.scrollTop <= 50 && !isLoadingMore) {
              handleScrollToTop();
            }
          }}
        >
          {/* 加载更早消息 */}
          {isLoadingMore && (
            <div className="flex justify-center py-2">
              <Loader2 className="w-5 h-5 animate-spin text-muted-foreground" />
            </div>
          )}

          {/* 无消息提示 */}
          {!selectedConversation && (
            <EmptyState icon={Search} message="选择好友开始聊天" />
          )}

          {selectedConversation && messages.length === 0 && !isLoadingMore && (
            <EmptyState icon={Send} message="开始新对话" />
          )}

          {/* 消息气泡 */}
          {messages.map((msg) => (
            <MessageBubble key={msg.messageId} msg={msg} />
          ))}

          {/* 新消息浮动按钮 */}
          {hasNewMessage && (
            <div className="sticky bottom-2 flex justify-center">
              <Button
                variant="secondary"
                size="sm"
                className="shadow-lg animate-in slide-in-from-bottom-2"
                onClick={scrollToBottom}
              >
                新消息 ↓
              </Button>
            </div>
          )}

        </div>

        {/* 输入工具栏 */}
        <div className="border-t border-border p-3">
          <div className="flex items-center gap-1 mb-2">
            <Button variant="ghost" size="icon" className="h-8 w-8" disabled>
              <Smile className="w-4 h-4 text-muted-foreground/40" />
            </Button>
            <Button variant="ghost" size="icon" className="h-8 w-8" disabled>
              <Scissors className="w-4 h-4 text-muted-foreground/40" />
            </Button>
            <Button variant="ghost" size="icon" className="h-8 w-8" disabled>
              <FolderOpen className="w-4 h-4 text-muted-foreground/40" />
            </Button>
            <Button variant="ghost" size="icon" className="h-8 w-8" disabled>
              <ImageIcon className="w-4 h-4 text-muted-foreground/40" />
            </Button>
            <Button variant="ghost" size="icon" className="h-8 w-8" disabled>
              <Mic className="w-4 h-4 text-muted-foreground/40" />
            </Button>
            <Button variant="ghost" size="icon" className="h-8 w-8" disabled>
              <AtSign className="w-4 h-4 text-muted-foreground/40" />
            </Button>
          </div>
          <div className="flex gap-2">
            <Textarea
              placeholder={
                selectedConversation ? "请输入消息..." : "请选择好友开始聊天"
              }
              className="min-h-[60px] resize-none"
              value={messageInput}
              onChange={(e) => setMessageInput(e.target.value)}
              onKeyDown={handleKeyDown}
              disabled={!selectedConversation}
            />
            <div className="flex flex-col justify-end">
              <Button
                size="sm"
                className="h-9 px-4 bg-primary hover:bg-primary/90"
                onClick={handleSend}
                disabled={!selectedConversation || !messageInput.trim()}
              >
                <Send className="w-4 h-4" />
              </Button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
