import { useState } from "react";
import {
  Search, Plus, Phone, Smile, Scissors, FolderOpen,
  Image as ImageIcon, Mic, AtSign, Send,
} from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { cn } from "@/lib/utils";

const conversations = [
  { id: "1", name: "系统通知", avatar: "", lastMessage: "赛季 S3 已开启，全新地图池上线！", time: "00:28", unread: 1, isGroup: false },
  { id: "2", name: "Astralis水友群", avatar: "", lastMessage: "stArlight: 比如有个挂pv的...", time: "00:28", unread: 8, isGroup: true, memberCount: 256 },
  { id: "3", name: "客服小助手", avatar: "", lastMessage: "您好，充值问题已为您处理完毕", time: "00:24", unread: 0, isGroup: false },
  { id: "4", name: "100T best!", avatar: "", lastMessage: "朱艺莉: 千万不要踢飞神", time: "00:22", unread: 99, isGroup: true, memberCount: 128 },
];

const chatMessages = [
  { id: "m1", sender: "修", avatar: "", content: "[图片] 赛季更新公告截图", time: "00:13", isSelf: false, isAdmin: true },
  { id: "m2", sender: "我是真的想去火烈鸟", avatar: "", content: "美团有事真上啊", time: "00:22", isSelf: false, isAdmin: false },
  { id: "m3", sender: "哭泣的键盘", avatar: "", content: "这是魔兽精兵的奴隶嘛?", time: "00:22", isSelf: false, isAdmin: false },
];

export default function MessagesTab() {
  const [selectedId, setSelectedId] = useState("1");
  const [messageInput, setMessageInput] = useState("");
  const selectedConversation = conversations.find((c) => c.id === selectedId) || conversations[0];

  return (
    <div className="flex flex-col lg:flex-row gap-0 -mx-6 -my-6 min-h-[520px]">
      <div className="w-full lg:w-72 flex-shrink-0 border-b lg:border-b-0 lg:border-r border-border flex flex-col bg-card/30">
        <div className="p-3 flex gap-2 border-b border-border">
          <div className="relative flex-1">
            <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
            <Input placeholder="搜索" className="pl-8 h-9 text-sm bg-secondary/50 border-0" />
          </div>
          <Button size="icon" variant="outline" className="h-9 w-9 shrink-0">
            <Plus className="w-4 h-4" />
          </Button>
        </div>
        <div className="flex-1 overflow-y-auto">
          {conversations.map((conv) => (
            <button
              key={conv.id}
              onClick={() => setSelectedId(conv.id)}
              className={cn(
                "w-full flex items-center gap-3 px-3 py-2.5 text-left transition-colors",
                selectedId === conv.id ? "bg-primary/10" : "hover:bg-secondary/50",
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
                  <span className="text-xs text-muted-foreground truncate">{conv.lastMessage}</span>
                  {conv.unread > 0 && (
                    <Badge variant="secondary" className="h-5 min-w-5 px-1.5 text-[10px] bg-primary/20 text-primary border-primary/20 shrink-0">
                      {conv.unread > 99 ? "99+" : conv.unread}
                    </Badge>
                  )}
                </div>
              </div>
            </button>
          ))}
        </div>
      </div>
      <div className="flex-1 flex flex-col min-h-[400px]">
        <div className="flex items-center justify-between px-4 py-3 border-b border-border">
          <div className="flex items-center gap-2">
            <span className="font-semibold text-sm">{selectedConversation.name}</span>
          </div>
          <Button variant="ghost" size="icon" className="h-8 w-8">
            <Phone className="w-4 h-4 text-muted-foreground" />
          </Button>
        </div>
        <div className="flex-1 overflow-y-auto p-4 space-y-4">
          {chatMessages.map((msg) => (
            <div key={msg.id} className={cn("flex gap-3", msg.isSelf ? "flex-row-reverse" : "flex-row")}>
              <Avatar className="w-9 h-9 shrink-0">
                <AvatarImage src={msg.avatar} />
                <AvatarFallback className="bg-gradient-to-br from-blue-500 to-purple-500 text-white text-xs font-bold">
                  {msg.sender.slice(0, 2).toUpperCase()}
                </AvatarFallback>
              </Avatar>
              <div className={cn("max-w-[70%] flex flex-col", msg.isSelf ? "items-end" : "items-start")}>
                <div className="flex items-center gap-2 mb-1">
                  <span className="text-xs text-muted-foreground">{msg.sender}</span>
                  <span className="text-xs text-muted-foreground/60">{msg.time}</span>
                </div>
                <div className={cn("px-3 py-2 rounded-lg text-sm break-words", msg.isSelf ? "bg-primary text-primary-foreground" : "bg-secondary text-secondary-foreground")}>
                  {msg.content}
                </div>
              </div>
            </div>
          ))}
        </div>
        <div className="border-t border-border p-3">
          <div className="flex items-center gap-1 mb-2">
            <Button variant="ghost" size="icon" className="h-8 w-8"><Smile className="w-4 h-4 text-muted-foreground" /></Button>
            <Button variant="ghost" size="icon" className="h-8 w-8"><Scissors className="w-4 h-4 text-muted-foreground" /></Button>
            <Button variant="ghost" size="icon" className="h-8 w-8"><FolderOpen className="w-4 h-4 text-muted-foreground" /></Button>
            <Button variant="ghost" size="icon" className="h-8 w-8"><ImageIcon className="w-4 h-4 text-muted-foreground" /></Button>
            <Button variant="ghost" size="icon" className="h-8 w-8"><Mic className="w-4 h-4 text-muted-foreground" /></Button>
            <Button variant="ghost" size="icon" className="h-8 w-8"><AtSign className="w-4 h-4 text-muted-foreground" /></Button>
          </div>
          <div className="flex gap-2">
            <Textarea
              placeholder="请输入消息..."
              className="min-h-[60px] resize-none"
              value={messageInput}
              onChange={(e) => setMessageInput(e.target.value)}
              onKeyDown={(e) => { if (e.key === "Enter" && !e.shiftKey) { e.preventDefault(); setMessageInput(""); } }}
            />
            <div className="flex flex-col justify-end">
              <Button size="sm" className="h-9 px-4 bg-primary hover:bg-primary/90" onClick={() => setMessageInput("")}>
                <Send className="w-4 h-4" />
              </Button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
