import { useState } from "react";
import { toast } from "sonner";
import type { Friend } from "@heroiclabs/nakama-js";
import { useAuth } from "@/context/AuthContext";
import { useFriends } from "@/hooks/useFriends";
import {
  Search,
  Plus,
  MapPin,
  ChevronDown,
  ChevronRight,
  Edit3,
  UserX,
  UserPlus,
  UserCheck,
  AlertCircle,
  Loader2,
} from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Input } from "@/components/ui/input";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import {
  Empty,
  EmptyHeader,
  EmptyTitle,
  EmptyDescription,
} from "@/components/ui/empty";
import { cn } from "@/lib/utils";

const FRIEND_STATE = { FRIEND: 0, INVITE_SENT: 1, INVITE_RECEIVED: 2 } as const;

function formatRelativeTime(timestamp?: string): string {
  if (!timestamp) return "-";
  const now = Date.now();
  const then = parseInt(timestamp, 10) * 1000;
  const diffMs = now - then;
  const diffDays = Math.floor(diffMs / 86400000);
  if (diffDays < 0) return "刚刚";
  if (diffDays === 0) return "今天";
  if (diffDays === 1) return "昨天";
  if (diffDays < 30) return `${diffDays}天前`;
  const diffMonths = Math.floor(diffDays / 30);
  if (diffMonths < 12) return `${diffMonths}个月前`;
  return `${Math.floor(diffMonths / 12)}年前`;
}

export default function FriendsTab() {
  const { session } = useAuth();
  const {
    status: friendsStatus,
    friends,
    error: friendsError,
    visibleGroups,
    searchQuery,
    setSearchQuery,
    addFriend,
    removeFriend,
    retry: retryFriends,
  } = useFriends(session);

  const [selectedFriendId, setSelectedFriendId] = useState<string | null>(null);
  const [expandedGroups, setExpandedGroups] = useState<Record<string, boolean>>({
    我的好友: true,
    已发送请求: true,
    收到的请求: true,
  });
  const [showAddDialog, setShowAddDialog] = useState(false);
  const [addFriendName, setAddFriendName] = useState("");
  const [addFriendError, setAddFriendError] = useState("");
  const [addFriendLoading, setAddFriendLoading] = useState(false);

  const selectedFriend: Friend | null =
    selectedFriendId !== null
      ? friends.find((f) => f.user?.id === selectedFriendId) ?? null
      : null;

  const state = selectedFriend?.state;

  return (
    <div className="flex flex-col lg:flex-row gap-0 -mx-6 -my-6 min-h-[520px]">
      {/* Left: Friend List */}
      <div className="w-full lg:w-64 flex-shrink-0 border-b lg:border-b-0 lg:border-r border-border flex flex-col bg-card/30">
        {/* Search + Add + Refresh */}
        <div className="p-3 flex gap-2 border-b border-border">
          <div className="relative flex-1">
            <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
            <Input
              placeholder="搜索好友..."
              className="pl-8 h-9 text-sm"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>
          <Button
            size="icon"
            variant="outline"
            className="h-9 w-9 shrink-0"
            onClick={() => setShowAddDialog(true)}
            title="添加好友"
          >
            <Plus className="w-4 h-4" />
          </Button>
          <Button
            size="icon"
            variant="ghost"
            className="h-9 w-9 shrink-0"
            onClick={retryFriends}
            title="刷新好友列表"
          >
            <Loader2
              className={cn(
                "w-3.5 h-3.5",
                friendsStatus === "loading" && "animate-spin",
              )}
            />
          </Button>
        </div>

        {/* Friend Groups */}
        <div className="flex-1 overflow-y-auto py-1">
          {friendsStatus === "loading" && (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="w-5 h-5 animate-spin text-muted-foreground" />
            </div>
          )}

          {friendsStatus === "error" && (
            <div className="flex flex-col items-center gap-3 py-12 px-4">
              <AlertCircle className="w-8 h-8 text-destructive/70" />
              <p className="text-sm text-muted-foreground text-center">
                {friendsError}
              </p>
              <Button variant="outline" size="sm" onClick={retryFriends}>
                重试
              </Button>
            </div>
          )}

          {friendsStatus === "success" && friends.length === 0 && (
            <div className="flex items-center justify-center py-12">
              <Empty>
                <EmptyHeader>
                  <EmptyTitle>暂无好友</EmptyTitle>
                  <EmptyDescription>点击 + 添加好友</EmptyDescription>
                </EmptyHeader>
              </Empty>
            </div>
          )}

          {friendsStatus === "success" &&
            visibleGroups.map((group) => {
              const isExpanded = expandedGroups[group.name] ?? true;
              return (
                <div key={group.state}>
                  <button
                    onClick={() =>
                      setExpandedGroups((prev) => ({
                        ...prev,
                        [group.name]: !isExpanded,
                      }))
                    }
                    className="w-full flex items-center gap-1 px-3 py-2 text-xs text-muted-foreground hover:text-foreground transition-colors"
                  >
                    {isExpanded ? (
                      <ChevronDown className="w-3.5 h-3.5" />
                    ) : (
                      <ChevronRight className="w-3.5 h-3.5" />
                    )}
                    <span className="font-medium">{group.name}</span>
                    {group.state === FRIEND_STATE.INVITE_RECEIVED && (
                      <Badge
                        variant="destructive"
                        className="text-[10px] px-1.5 py-0 h-4 ml-1"
                      >
                        {group.friends.length}
                      </Badge>
                    )}
                    <span className="text-muted-foreground/60 ml-auto">
                      {group.friends.length}
                    </span>
                  </button>

                  {isExpanded &&
                    group.friends.map((friend) => (
                      <button
                        key={friend.user?.id}
                        onClick={() =>
                          setSelectedFriendId(friend.user?.id ?? null)
                        }
                        className={cn(
                          "w-full flex items-center gap-3 px-3 py-2.5 text-left transition-colors",
                          selectedFriendId === friend.user?.id
                            ? "bg-primary/10"
                            : "hover:bg-secondary/50",
                        )}
                      >
                        <div className="relative shrink-0">
                          <Avatar className="w-9 h-9">
                            <AvatarImage src={friend.user?.avatar_url} />
                            <AvatarFallback className="bg-gradient-to-br from-blue-500 to-purple-500 text-white text-xs font-bold">
                              {(friend.user?.username ?? "??")
                                .slice(0, 2)
                                .toUpperCase()}
                            </AvatarFallback>
                          </Avatar>
                          <span
                            className={cn(
                              "absolute -bottom-0.5 -right-0.5 w-2.5 h-2.5 rounded-full border-2 border-background",
                              friend.user?.online
                                ? "bg-green-500"
                                : "bg-gray-500",
                            )}
                          />
                        </div>
                        <div className="min-w-0 flex-1">
                          <div className="text-sm font-medium truncate">
                            {friend.user?.username ?? "未知用户"}
                          </div>
                          <div className="text-xs text-muted-foreground truncate">
                            {friend.user?.display_name || "-"}
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
            {/* Header */}
            <div className="flex items-start gap-5">
              <Avatar className="w-20 h-20 border-2 border-primary/20">
                <AvatarImage src={selectedFriend.user?.avatar_url} />
                <AvatarFallback className="bg-gradient-to-br from-blue-500 to-purple-500 text-white text-2xl font-bold">
                  {(selectedFriend.user?.username ?? "??")
                    .slice(0, 2)
                    .toUpperCase()}
                </AvatarFallback>
              </Avatar>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-3 flex-wrap">
                  <h3 className="text-xl font-bold">
                    {selectedFriend.user?.username ?? "未知用户"}
                  </h3>
                  <div className="flex items-center gap-1.5 text-sm">
                    <span
                      className={cn(
                        "w-2.5 h-2.5 rounded-full",
                        selectedFriend.user?.online
                          ? "bg-green-500"
                          : "bg-gray-500",
                      )}
                    />
                    <span className="text-muted-foreground">
                      {selectedFriend.user?.online ? "在线" : "离线"}
                    </span>
                  </div>
                </div>
                <p className="text-sm text-muted-foreground mt-1">
                  ID: {selectedFriend.user?.id ?? "-"}
                </p>
                <div className="flex items-center gap-2 mt-2">
                  {state === FRIEND_STATE.INVITE_RECEIVED && (
                    <Badge variant="destructive" className="text-xs">
                      待处理
                    </Badge>
                  )}
                  {state === FRIEND_STATE.INVITE_SENT && (
                    <Badge variant="secondary" className="text-xs">
                      等待回应
                    </Badge>
                  )}
                </div>
              </div>
            </div>

            <Separator />

            {/* Info */}
            <div className="flex flex-wrap items-center gap-x-6 gap-y-2 text-sm">
              <div className="flex items-center gap-1.5">
                <MapPin className="w-3.5 h-3.5 text-muted-foreground" />
                <span>
                  {selectedFriend.user?.location
                    ? `现居 ${selectedFriend.user.location}`
                    : "所在地: -"}
                </span>
              </div>
              <div className="flex items-center gap-1.5">
                <Edit3 className="w-3.5 h-3.5 text-muted-foreground" />
                <span>
                  最近活跃:{" "}
                  {formatRelativeTime(selectedFriend.user?.update_time)}
                </span>
              </div>
            </div>

            {/* Signature */}
            <div className="space-y-2">
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <Edit3 className="w-3.5 h-3.5" />
                <span>签名</span>
              </div>
              <p className="text-sm">
                {selectedFriend.user?.display_name || "-"}
              </p>
            </div>

            {/* Actions */}
            <Separator />
            <div className="flex items-center gap-3 pt-2">
              {state === FRIEND_STATE.FRIEND && (
                <Button
                  variant="destructive"
                  size="sm"
                  className="gap-1.5"
                  onClick={() => {
                    const name =
                      selectedFriend.user?.username ?? "该好友";
                    if (window.confirm(`确定要删除好友 ${name} 吗？`)) {
                      removeFriend(selectedFriend.user!.id!).then(() => {
                        toast.success(`已删除好友 ${name}`);
                        setSelectedFriendId(null);
                      });
                    }
                  }}
                >
                  <UserX className="w-4 h-4" />
                  删除好友
                </Button>
              )}

              {state === FRIEND_STATE.INVITE_SENT && (
                <Button
                  variant="outline"
                  size="sm"
                  className="gap-1.5"
                  onClick={() => {
                    const name =
                      selectedFriend.user?.username ?? "该用户";
                    if (
                      window.confirm(
                        `确定要取消发送给 ${name} 的好友请求吗？`,
                      )
                    ) {
                      removeFriend(selectedFriend.user!.id!).then(() => {
                        toast.success("已取消好友请求");
                        setSelectedFriendId(null);
                      });
                    }
                  }}
                >
                  <UserX className="w-4 h-4" />
                  取消请求
                </Button>
              )}

              {state === FRIEND_STATE.INVITE_RECEIVED && (
                <>
                  <Button
                    size="sm"
                    className="gap-1.5 bg-primary hover:bg-primary/90"
                    onClick={() => {
                      const name =
                        selectedFriend.user?.username ?? "该用户";
                      addFriend(name).then(() => {
                        toast.success(`已接受 ${name} 的好友请求`);
                      });
                    }}
                  >
                    <UserCheck className="w-4 h-4" />
                    接受
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    className="gap-1.5"
                    onClick={() => {
                      const name =
                        selectedFriend.user?.username ?? "该用户";
                      if (
                        window.confirm(
                          `确定要拒绝 ${name} 的好友请求吗？`,
                        )
                      ) {
                        removeFriend(selectedFriend.user!.id!).then(() => {
                          toast.success(`已拒绝 ${name} 的好友请求`);
                          setSelectedFriendId(null);
                        });
                      }
                    }}
                  >
                    <UserX className="w-4 h-4" />
                    拒绝
                  </Button>
                </>
              )}
            </div>
          </div>
        ) : (
          <Empty className="min-h-[400px]">
            <EmptyHeader>
              <EmptyTitle>请选择好友</EmptyTitle>
              <EmptyDescription>
                在左侧列表中选择一位好友查看详情
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        )}
      </div>

      {/* Add Friend Dialog */}
      <Dialog
        open={showAddDialog}
        onOpenChange={(open) => {
          setShowAddDialog(open);
          if (!open) {
            setAddFriendName("");
            setAddFriendError("");
          }
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>添加好友</DialogTitle>
            <DialogDescription>
              输入对方的用户名，发送好友请求
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-3 py-3">
            <Input
              placeholder="请输入用户名"
              value={addFriendName}
              onChange={(e) => {
                setAddFriendName(e.target.value);
                setAddFriendError("");
              }}
            />
            {addFriendError && (
              <p className="text-sm text-destructive">{addFriendError}</p>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowAddDialog(false)}>
              取消
            </Button>
            <Button
              disabled={addFriendLoading}
              onClick={async () => {
                const name = addFriendName.trim();
                if (!name) {
                  setAddFriendError("请输入用户名");
                  return;
                }
                if (
                  name.toLowerCase() === session?.username?.toLowerCase()
                ) {
                  setAddFriendError("无法添加自己为好友");
                  return;
                }
                const existing = friends.find(
                  (f) =>
                    f.user?.username?.toLowerCase() === name.toLowerCase(),
                );
                if (existing) {
                  if (existing.state === FRIEND_STATE.FRIEND) {
                    setAddFriendError("该用户已是您的好友");
                    return;
                  }
                  if (existing.state === FRIEND_STATE.INVITE_SENT) {
                    setAddFriendError("已向该用户发送过好友请求");
                    return;
                  }
                }
                setAddFriendLoading(true);
                setAddFriendError("");
                try {
                  await addFriend(name);
                  toast.success("好友请求已发送");
                  setShowAddDialog(false);
                } catch (err) {
                  const msg =
                    err instanceof Error ? err.message : String(err);
                  if (msg.includes("not found") || msg.includes("NOT_FOUND")) {
                    setAddFriendError("用户不存在，请检查用户名");
                  } else if (
                    msg.includes("already") ||
                    msg.includes("ALREADY_EXISTS")
                  ) {
                    setAddFriendError("该用户已是您的好友");
                  } else {
                    setAddFriendError(msg || "添加失败，请稍后重试");
                  }
                } finally {
                  setAddFriendLoading(false);
                }
              }}
            >
              {addFriendLoading ? (
                <>
                  <Loader2 className="w-4 h-4 mr-1 animate-spin" />
                  发送中...
                </>
              ) : (
                <>
                  <UserPlus className="w-4 h-4 mr-1" />
                  添加
                </>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
